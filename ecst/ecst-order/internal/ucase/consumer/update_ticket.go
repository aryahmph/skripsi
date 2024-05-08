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

type updateTicket struct {
	cfg        *appctx.Config
	ticketRepo repositories.TicketRepository
}

func NewUpdateTicket(cfg *appctx.Config, ticketRepo repositories.TicketRepository) contract.MessageProcessor {
	return &updateTicket{cfg: cfg, ticketRepo: ticketRepo}
}

func (ucase *updateTicket) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_get_ticket_status"

		lvState4       = consts.LogEventStateUpdateDB
		lfState4Status = "state_4_update_ticket_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameUpdateTicket),
		}

		err     error
		msgData presentations.MessageCreateTicket
		now     = time.Now()
	)

	ctx = tracer.SpanStart(ctx, `consumer:update-ticket`)
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

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETUpdateTicket)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
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
	| STEP 3: Get ticket
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	findTicket, err := ucase.ticketRepo.FindOneTicket(ctx, repositories.FindOneTicketCriteria{
		ID:      msgData.Payload.ID,
		Version: msgData.Payload.Version - 1,
	})

	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogEventStateFetchDB, entity.TableNameOrder, err), lf...)
		return err
	}

	if findTicket == nil {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder), lf...)
		return nil
	}

	/*-------------------------------
	| STEP 4: Update ticket
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	ticket := entity.Ticket{
		ID:       msgData.Payload.ID,
		Code:     msgData.Payload.Code,
		Category: msgData.Payload.Category,
		Price:    msgData.Payload.Price,
		Version:  msgData.Payload.Version,
	}

	err = ucase.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameOrder, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameUpdateTicket), lf...)
	return nil
}
