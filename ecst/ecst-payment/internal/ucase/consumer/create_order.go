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

type createOrder struct {
	cfg       *appctx.Config
	orderRepo repositories.OrderRepository
}

func NewCreateOrder(cfg *appctx.Config, orderRepo repositories.OrderRepository) contract.MessageProcessor {
	return &createOrder{cfg: cfg, orderRepo: orderRepo}
}

func (ucase *createOrder) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateInsertDB
		lfState3Status = "state_3_insert_order_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCreateOrder),
		}

		err     error
		msgData presentations.MessageCreateOrder
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:create-order`)
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

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETCreateOrder)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
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
	| STEP 2: Insert order
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	order := entity.Order{
		ID:      msgData.Payload.ID,
		Status:  msgData.Payload.Status,
		Amount:  msgData.Payload.Amount,
		Version: msgData.Payload.Version,
	}

	err = ucase.orderRepo.InsertOrder(ctx, order)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNameOrder, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreateOrder), lf...)
	return nil
}
