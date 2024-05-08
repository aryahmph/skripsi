package consumer

import (
	"context"
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/consts"
	"ecst-payment/internal/entity"
	"ecst-payment/internal/helper"
	"ecst-payment/internal/presentations"
	"ecst-payment/internal/repositories"
	"ecst-payment/internal/ucase/contract"
	"ecst-payment/pkg/logger"
	"ecst-payment/pkg/tracer"
	"ecst-payment/pkg/util"
	"encoding/json"
	"time"
)

type orderExpire struct {
	cfg       *appctx.Config
	orderRepo repositories.OrderRepository
}

func NewOrderExpire(cfg *appctx.Config, orderRepo repositories.OrderRepository) contract.MessageProcessor {
	return &orderExpire{cfg: cfg, orderRepo: orderRepo}
}

func (ucase *orderExpire) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_get_order_status"

		lvState4       = consts.LogEventStateUpdateDB
		lfState4Status = "state_4_update_order_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameOrderExpire),
		}

		err     error
		msgData presentations.MessageOrderExpire
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:order-expire`)
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

	/*-------------------------------
	| STEP 3: Get order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	findOrder, err := ucase.orderRepo.FindOneOrder(ctx, repositories.FindOneOrderCriteria{
		ID:      msgData.Payload.ID,
		Version: msgData.Payload.Version - 1,
	})

	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogEventStateFetchDB, entity.TableNameOrder, err), lf...)
		return err
	}

	if findOrder == nil {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder), lf...)
		return nil
	}

	/*-------------------------------
	| STEP 4: Update order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	order := entity.Order{
		ID:      msgData.Payload.ID,
		Status:  consts.OrderStatusExpired,
		Version: msgData.Payload.Version,
	}

	err = ucase.orderRepo.UpdateOrder(ctx, order)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameOrder, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameOrderExpire), lf...)
	return nil
}
