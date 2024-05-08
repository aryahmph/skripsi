package consumer

import (
	"context"
	"ecst-ticket/pkg/kafka"
	"encoding/json"
	"github.com/spf13/cast"
	"time"

	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/entity"
	"ecst-ticket/internal/helper"
	"ecst-ticket/internal/presentations"
	"ecst-ticket/internal/repositories"
	ucaseContract "ecst-ticket/internal/ucase/contract"

	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/tracer"
	"ecst-ticket/pkg/util"
)

type removeTicketOrder struct {
	cfg        *appctx.Config
	kp         kafka.Producer
	ticketRepo repositories.TicketRepository
}

func NewRemoveTicketOrder(
	cfg *appctx.Config,
	kp kafka.Producer,
	ticketRepo repositories.TicketRepository,
) ucaseContract.MessageProcessor {
	return &removeTicketOrder{
		cfg:        cfg,
		kp:         kp,
		ticketRepo: ticketRepo,
	}
}

func (ucase *removeTicketOrder) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_get_ticket_from_db_status"

		lvState4       = consts.LogEventStateUpdateDB
		lfState4Status = "state_4_update_order_to_db_status"

		lfState5       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState5Status = "state_5_publish_message_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameRemoveTicketOrder),
		}

		err     error
		msgData presentations.MessageOrderExpire
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
	| STEP 3: Get ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	ticket, err := ucase.ticketRepo.FindOneTicket(ctx, repositories.FindOneTicketCriteria{OrderID: msgData.Payload.ID})
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicket, err),
			lf...,
		)
		return err
	}

	if ticket == nil {
		lf = append(lf, logger.Any(lfState3Status, consts.LogStatusFailed))
		logger.WarnWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameTicket),
			lf...,
		)
		return nil
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Update ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	err = ucase.ticketRepo.RemoveTicketOrderID(ctx, ticket.ID)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState4Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedUpdateDB, entity.TableNameTicket, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Publish message
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lfState5))

	message := presentations.KafkaMessageBase{
		Source: presentations.KafkaMessageBaseSource{
			AppName: ucase.cfg.App.AppName,
			AppEnv:  ucase.cfg.App.Env,
		},
		Check: &presentations.KafkaMessageChecker{
			InitiateTime: time.Now().Format(consts.LayoutDateTimeFormat),
			ServiceOrigin: presentations.KafkaMessageOriginService{
				ServiceName: ucase.cfg.App.AppName,
				TargetTopic: ucase.cfg.KafkaTopics.TopicUpdateTicket,
			},
			Count:      0,
			NextSecond: cast.ToUint(ucase.cfg.KafkaNextSecond.NextUpdateTicket),
			MaxSecond:  cast.ToUint(ucase.cfg.KafkaEETSecond.EETUpdateTicket),
		},
		CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
		Payload: &presentations.KafkaMessageCreateTicketPayload{
			ID:       ticket.ID,
			Code:     ticket.Code,
			Category: ticket.Category,
			Price:    ticket.Price,
			Version:  ticket.Version + 1,
		},
	}

	km := kafka.MessageContext{
		Value:   util.DumpToString(message),
		Topic:   ucase.cfg.KafkaTopics.TopicUpdateTicket,
		Verbose: true,
	}

	err = ucase.kp.Publish(ctx, &km)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameRemoveTicketOrder), lf...)
	return nil
}
