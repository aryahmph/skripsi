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
	"ecst-order/pkg/logger"
	"ecst-order/pkg/tracer"
	"ecst-order/pkg/util"
	"encoding/json"
	"time"
)

type createTicket struct {
	cfg        *appctx.Config
	ticketRepo repositories.TicketRepository
}

func NewCreateTicket(cfg *appctx.Config, ticketRepo repositories.TicketRepository) contract.MessageProcessor {
	return &createTicket{cfg: cfg, ticketRepo: ticketRepo}
}

func (ucase *createTicket) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateInsertDB
		lfState3Status = "state_3_insert_ticket_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCreateTicket),
		}

		err     error
		msgData presentations.MessageCreateTicket
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

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETCreateTicket)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
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
	| STEP 3: Insert ticket
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	ticket := entity.Ticket{
		ID:       msgData.Payload.ID,
		Code:     msgData.Payload.Code,
		Category: msgData.Payload.Category,
		Price:    msgData.Payload.Price,
		Version:  msgData.Payload.Version,
	}

	err = ucase.ticketRepo.InsertTicket(ctx, ticket)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNameTicket, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreateTicket), lf...)
	return nil
}
