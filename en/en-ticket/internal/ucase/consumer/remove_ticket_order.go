package consumer

import (
	"context"
	"en-ticket/internal/dto"
	"en-ticket/internal/providers/order"
	"en-ticket/pkg/cache"
	"encoding/json"
	"time"

	"en-ticket/internal/appctx"
	"en-ticket/internal/consts"
	"en-ticket/internal/entity"
	"en-ticket/internal/helper"
	"en-ticket/internal/presentations"
	"en-ticket/internal/repositories"
	ucaseContract "en-ticket/internal/ucase/contract"

	"en-ticket/pkg/logger"
	"en-ticket/pkg/tracer"
	"en-ticket/pkg/util"
)

type removeTicketOrder struct {
	cfg           *appctx.Config
	cacher        cache.Cacher
	ticketRepo    repositories.TicketRepository
	orderProvider order.OrderProvider
}

func NewRemoveTicketOrder(
	cfg *appctx.Config,
	cacher cache.Cacher,
	ticketRepo repositories.TicketRepository,
	orderProvider order.OrderProvider,
) ucaseContract.MessageProcessor {
	return &removeTicketOrder{
		cfg:           cfg,
		cacher:        cacher,
		ticketRepo:    ticketRepo,
		orderProvider: orderProvider,
	}
}

func (ucase *removeTicketOrder) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3         = consts.LogEventStateCallProvider
		lfState3Status   = "state_3_get_order_from_order_service_status"
		lfState3Request  = "state_3_get_order_from_order_service_request"
		lfState3Response = "state_3_get_order_from_order_service_response"

		lvState4       = consts.LogEventStateFetchDB
		lfState4Status = "state_4_get_ticket_from_db_status"

		lvState5       = consts.LogEventStateUpdateDB
		lfState5Status = "state_5_update_order_to_db_status"

		//lvState6       = consts.LogEventStateRemoveCache
		//lfState6Status = "state_6_remove_ticket_data_cache_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameRemoveTicketOrder),
		}

		err     error
		msgData presentations.MessageCreateOrder
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:update-ticket-order`)
	defer tracer.SpanFinish(ctx)

	/*-------------------------------
	| STEP 1: Decode Request
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState1))

	err = json.Unmarshal(data.Body, &msgData)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState1Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageInvalidErr, err), lf...)
		return err
	}

	_ = util.ToJSONByteCompact(msgData.Payload)
	lf = append(lf,
		logger.Any(lfState1Status, consts.LogStatusSuccess),
		logger.EventOutputKafka(msgData.Payload, data.Partition, data.Offset),
	)

	/*-------------------------------
	| STEP 2: Validate execution time
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState2))

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETOrderExpire)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState2Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedToCompareEET, err.Error()), lf...)
		return err
	}

	if !valid {
		lf = append(lf, logger.Any(lfState2Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageExpiredExecutionTime), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 3: Get order
	* ----------------------------------------*/

	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	orderReq := order.GetOrderRequest{
		ID:         msgData.Payload.ID,
		Status:     consts.OrderStatusExpired,
		ClientName: ucase.cfg.App.AppName,
	}
	reqAPI, respAPI, err := ucase.orderProvider.GetOrder(ctx, orderReq)

	lf = append(lf,
		logger.Any(lfState3Request, dto.TransformHeaderForLogging(reqAPI)),
		logger.Any(lfState3Response, util.DumpToString(respAPI)),
	)

	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallFailed, consts.OrderDependencyName, err), lf...)
		return err
	}

	if respAPI.Name != consts.ResponseSuccess {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallGotUnsuccessfulResponse, "order", respAPI.Name), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 4: Get ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	ticket, err := ucase.ticketRepo.FindOneTicket(ctx, repositories.FindOneTicketCriteria{OrderID: msgData.Payload.ID})
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicket, err),
			lf...,
		)
		return err
	}

	if ticket == nil {
		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameTicket),
			lf...,
		)
		return nil
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Update ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	err = ucase.ticketRepo.RemoveTicketOrderID(ctx, ticket.ID)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameTicket, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	///*-----------------------------------------
	//| STEP 6: Remove cache
	//* ----------------------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState6))
	//
	//err = ucase.cacher.Delete(ctx, helper.TicketCacheKey(ticket.ID))
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//
	//	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusFailed))
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedSetCache, err),
	//		lf...,
	//	)
	//	return err
	//}
	//
	//lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameRemoveTicketOrder), lf...)
	return nil
}
