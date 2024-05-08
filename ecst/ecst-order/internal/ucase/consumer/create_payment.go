package consumer

import (
	"context"
	"ecst-order/internal/appctx"
	"ecst-order/internal/consts"
	"ecst-order/internal/entity"
	"ecst-order/internal/helper"
	"ecst-order/internal/presentations"
	"ecst-order/internal/repositories"
	"ecst-order/internal/ucase/contract"
	"ecst-order/pkg/cache"
	"ecst-order/pkg/logger"
	"ecst-order/pkg/tracer"
	"ecst-order/pkg/util"
	"encoding/json"
	"time"
)

type createPayment struct {
	cfg       *appctx.Config
	cacher    cache.Cacher
	orderRepo repositories.OrderRepository
}

func NewCreatePayment(cfg *appctx.Config, orderRepo repositories.OrderRepository, cacher cache.Cacher) contract.MessageProcessor {
	return &createPayment{cfg: cfg, orderRepo: orderRepo, cacher: cacher}
}

func (ucase *createPayment) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_get_order_status"

		lvState4       = consts.LogEventStateInsertDB
		lfState4Status = "state_4_update_order_status"

		lvState5       = consts.LogEventStateDelayQueueRemoveJob
		lfState5Status = "state_dq_remove_job_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCreatePayment),
		}

		err     error
		msgData presentations.MessageCreatePayment
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:create-ticket`)
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

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETCreatePayment)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
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

	/*-------------------------------
	| STEP 3: Get order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	order, err := ucase.orderRepo.FindOneOrder(ctx, repositories.FindOneOrderCriteria{
		ID:     msgData.Payload.OrderID,
		Status: consts.OrderStatusCreated,
	})
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameOrder, err), lf...)
		return err
	}

	if order == nil {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-------------------------------
	| STEP 4: Update order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	updateOrder := entity.Order{
		ID:        order.ID,
		PaymentID: msgData.Payload.ID,
		Status:    consts.OrderStatusCompleted,
		Version:   order.Version + 1,
	}

	isExist, err := ucase.orderRepo.UpdateOrder(ctx, updateOrder)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameOrder, err), lf...)
		return err
	}

	if !isExist {
		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder), lf...)
		return nil
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Remove job
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	removeKey := util.DumpToString(presentations.OrderExpireJobData{ID: order.ID})
	err = ucase.cacher.ZRem(ctx, ucase.cfg.Job.OrderExpire.QueueName, removeKey)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedRemoveCache, err),
			lf...,
		)
		return err
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreatePayment), lf...)
	return nil
}
