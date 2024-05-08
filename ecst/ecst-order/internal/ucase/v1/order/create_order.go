package order

import (
	"ecst-order/pkg/dq"
	"net/url"
	"time"

	"ecst-order/internal/appctx"
	"ecst-order/internal/consts"
	"ecst-order/internal/entity"
	"ecst-order/internal/presentations"
	"ecst-order/internal/repositories"
	ucaseContract "ecst-order/internal/ucase/contract"

	"ecst-order/pkg/generator"
	"ecst-order/pkg/kafka"
	"ecst-order/pkg/logger"
	"ecst-order/pkg/tracer"
	"ecst-order/pkg/util"

	"github.com/spf13/cast"
	"github.com/thedevsaddam/govalidator"
)

type createOrder struct {
	kp         kafka.Producer
	dqp        dq.Producer
	orderRepo  repositories.OrderRepository
	ticketRepo repositories.TicketRepository
}

func NewCreateOrder(
	kp kafka.Producer,
	dqp dq.Producer,
	orderRepo repositories.OrderRepository,
	ticketRepo repositories.TicketRepository,
) ucaseContract.UseCase {
	return &createOrder{
		kp:         kp,
		dqp:        dqp,
		orderRepo:  orderRepo,
		ticketRepo: ticketRepo,
	}
}

func (ucase *createOrder) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateRequest
		lfState2Status = "state_2_validate_request_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_fetch_ticket_status"

		lvState4       = consts.LogEventStateInsertDB
		lfState4Status = "state_4_insert_order_status"

		lvState5       = consts.LogEventStateDelayQueueAddJob
		lfState5Status = "state_5_add_dq_job_status"

		lvState6       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState6Status = "state_6_kafka_publish_message_status"

		userID    = data.Request.Header.Get(consts.HeaderXUserId)
		userEmail = data.Request.Header.Get(consts.HeaderXUserEmail)

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		req presentations.CreateOrderRequest

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCreateOrder),
			logger.Any(consts.LogFieldUserID, userID),
			logger.Any(consts.LogFieldUserEmail, userEmail),
		}
	)

	defer tracer.SpanFinish(ctx)

	/*-----------------------------------------
	| STEP 1: Decode Request
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState1))

	err := data.Cast(&req)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseValidationFailure)

		lf = append(lf,
			logger.Any(lfState1Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedDecodePayload, err), lf...)
		return
	}

	lf = append(lf,
		logger.Any(lfState1Status, consts.LogStatusSuccess),
		logger.EventInputHttp(req),
	)

	/*-----------------------------------------
	| STEP 2: Validate Request
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState2))

	ev := ucase.validateRequest(req)
	if len(ev) > 0 {
		response.SetName(consts.ResponseValidationFailure).SetError(ev)

		lf = append(lf,
			logger.Any(lfState2Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedToValidateRequestParam, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 3: Get ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	ticket, err := ucase.ticketRepo.FindOneTicket(ctx, repositories.FindOneTicketCriteria{ID: req.TicketID})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicket, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 4: Get ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	checkOrder, err := ucase.orderRepo.FindOneOrder(ctx, repositories.FindOneOrderCriteria{
		TicketID: ticket.ID,
		Statuses: []string{consts.OrderStatusCreated, consts.OrderStatusCompleted},
	})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState5Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameOrder, err), lf...)
		return
	}

	if checkOrder != nil {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState5Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageTicketAlreadyReserved, ticket.ID), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 4: Insert order
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	order := entity.Order{
		ID:       generator.GenerateString(),
		TicketID: ticket.ID,
		UserID:   userID,
		Status:   consts.OrderStatusCreated,
		Amount:   ticket.Price,
	}

	isTicketReserved, err := ucase.orderRepo.InsertOrder(ctx, order)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNameOrder, err), lf...)
		return
	}

	if isTicketReserved {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageTicketAlreadyReserved, ticket.ID), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Put order to delay queue
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	expiryTime := time.Now().Add(time.Second * time.Duration(data.Config.Job.OrderExpire.DurationSecond))

	err = ucase.dqp.Add(ctx, &dq.JobContext{
		QueueName: data.Config.Job.OrderExpire.QueueName,
		Value:     util.DumpToString(presentations.OrderExpireJobData{ID: order.ID}),
		ExpiredAt: expiryTime.UnixMilli(),
	})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState5Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedAddDelayQueue, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 6: Publish kafka message
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState6))

	message := presentations.KafkaMessageBase{
		Source: presentations.KafkaMessageBaseSource{
			AppName: data.Config.App.AppName,
			AppEnv:  data.Config.App.Env,
		},
		Check: &presentations.KafkaMessageChecker{
			InitiateTime: time.Now().Format(consts.LayoutDateTimeFormat),
			ServiceOrigin: presentations.KafkaMessageOriginService{
				ServiceName: data.Config.App.AppName,
				TargetTopic: data.Config.KafkaTopics.TopicCreateOrder,
			},
			Count:      0,
			NextSecond: cast.ToUint(data.Config.KafkaNextSecond.NextCreateOrder),
			MaxSecond:  cast.ToUint(data.Config.KafkaEETSecond.EETCreateOrder),
		},
		CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
		Payload: &presentations.KafkaMessageCreateOrderPayload{
			ID:       order.ID,
			TicketID: order.TicketID,
			UserID:   order.UserID,
			Status:   order.Status,
			Amount:   order.Amount,
			Version:  1,
		},
	}

	km := kafka.MessageContext{
		Value:   util.DumpToString(message),
		Topic:   data.Config.KafkaTopics.TopicCreateOrder,
		Verbose: true,
	}

	lf = append(lf, logger.EventInputKafka(message, util.DumpToString(message)))

	err = ucase.kp.Publish(ctx, &km)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState6Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	response.SetName(consts.ResponseSuccess)
	response.SetData(presentations.CreateOrderResponse{ID: order.ID})

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreateOrder), lf...)

	return
}

func (ucase *createOrder) validateRequest(req presentations.CreateOrderRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &req,
		Rules: govalidator.MapData{
			"ticket_id": consts.RuleULID,
		},
		TagIdentifier: "json",
	})

	errVal := v.ValidateStruct()
	if len(errVal) > 0 {
		return errVal
	}

	return errVal
}
