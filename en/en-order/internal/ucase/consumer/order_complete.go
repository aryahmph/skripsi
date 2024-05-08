package consumer

import (
	"context"
	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/internal/dto"
	"en-order/internal/entity"
	"en-order/internal/helper"
	"en-order/internal/presentations"
	"en-order/internal/providers/payment"
	"en-order/internal/repositories"
	"en-order/internal/ucase/contract"
	"en-order/pkg/cache"
	"en-order/pkg/logger"
	"en-order/pkg/tracer"
	"en-order/pkg/util"
	"encoding/json"
	"time"
)

type orderComplete struct {
	cfg             *appctx.Config
	cacher          cache.Cacher
	orderRepo       repositories.OrderRepository
	paymentProvider payment.PaymentProvider
}

func NewOrderComplete(
	cfg *appctx.Config,
	cacher cache.Cacher,
	orderRepo repositories.OrderRepository,
	paymentProvider payment.PaymentProvider,
) contract.MessageProcessor {
	return &orderComplete{
		cfg:             cfg,
		cacher:          cacher,
		orderRepo:       orderRepo,
		paymentProvider: paymentProvider,
	}
}

func (ucase *orderComplete) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3         = consts.LogEventStateGetPayment
		lfState3Status   = "state_3_get_payment_from_payment_service_status"
		lfState3Request  = "state_3_get_payment_from_payment_service_request"
		lfState3Response = "state_3_get_payment_from_payment_service_response"

		//lvState4       = consts.LogEventStateFetchDB
		//lfState4Status = "state_4_fetch_order_from_db_status"
		//lfState4Data   = "state_4_fetch_order_from_db_data"

		lvState5       = consts.LogEventStateUpdateDB
		lfState5Status = "state_5_update_order_status"

		lvState6       = consts.LogEventStateDelayQueueRemoveJob
		lfState6Status = "state_dq_remove_job_status"

		//lvState7       = consts.LogEventStateRemoveCache
		//lfState7Status = "state_7_remove_cache_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCompleteOrder),
		}

		err     error
		msgData presentations.MessageOrderComplete
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:order-complete`)
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
	| STEP 3: Get payment
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	paymentReq := payment.GetPaymentRequest{
		ID:         msgData.Payload.ID,
		ClientName: ucase.cfg.App.AppName,
	}
	reqAPI, respAPI, err := ucase.paymentProvider.GetPayment(ctx, paymentReq)

	lf = append(lf,
		logger.Any(lfState3Request, dto.TransformHeaderForLogging(reqAPI)),
		logger.Any(lfState3Response, util.DumpToString(respAPI)),
	)

	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallFailed, consts.TicketDependencyName, err), lf...)
		return err
	}

	if respAPI.Name != consts.ResponseSuccess {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallGotUnsuccessfulResponse, "payment", respAPI.Name), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-------------------------------
	| STEP 5: Update order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	isExist, err := ucase.orderRepo.UpdateOrderStatus(ctx, respAPI.Data.OrderID, consts.OrderStatusCompleted, respAPI.Data.ID)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameOrder, err), lf...)
		return err
	}

	if !isExist {
		lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 6: Remove job
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState6))

	removeKey := util.DumpToString(presentations.OrderExpireJobData{ID: respAPI.Data.OrderID})
	err = ucase.cacher.ZRem(ctx, ucase.cfg.Job.OrderExpire.QueueName, removeKey)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState6Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedRemoveCache, err),
			lf...,
		)
		return err
	}

	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	///*-------------------------------
	//| STEP 7: Remove cache
	//* -------------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState7))
	//
	//err = ucase.cacher.Delete(ctx, helper.OrderCacheKey(order.ID))
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//
	//	lf = append(lf, logger.Any(lfState7Status, consts.LogStatusFailed))
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedRemoveCache, err),
	//		lf...,
	//	)
	//	return err
	//}
	//
	//lf = append(lf, logger.Any(lfState7Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCompleteOrder), lf...)
	return nil
}
