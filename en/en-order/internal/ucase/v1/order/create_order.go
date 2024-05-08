package order

import (
	"en-order/pkg/dq"
	"net/url"
	"time"

	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/internal/dto"
	"en-order/internal/entity"
	"en-order/internal/presentations"
	"en-order/internal/providers/ticket"
	"en-order/internal/repositories"
	ucaseContract "en-order/internal/ucase/contract"

	"en-order/pkg/generator"
	"en-order/pkg/kafka"
	"en-order/pkg/logger"
	"en-order/pkg/tracer"
	"en-order/pkg/util"

	"github.com/spf13/cast"
	"github.com/thedevsaddam/govalidator"
)

type createOrder struct {
	kp             kafka.Producer
	dqp            dq.Producer
	orderRepo      repositories.OrderRepository
	ticketProvider ticket.TicketProvider
}

func NewCreateOrder(
	kp kafka.Producer,
	dqp dq.Producer,
	orderRepo repositories.OrderRepository,
	ticketProvider ticket.TicketProvider,
) ucaseContract.UseCase {
	return &createOrder{
		kp:             kp,
		dqp:            dqp,
		orderRepo:      orderRepo,
		ticketProvider: ticketProvider,
	}
}

func (ucase *createOrder) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateRequest
		lfState2Status = "state_2_validate_request_status"

		lvState3         = consts.LogEventStateGetTicket
		lfState3Status   = "state_3_get_ticket_from_ticket_service_status"
		lfState3Request  = "state_3_get_ticket_from_ticket_service_request"
		lfState3Response = "state_3_get_ticket_from_ticket_service_response"

		lvState4       = consts.LogEventStateCheckReservedTicket
		lfState4Status = "state_4_check_reserved_ticket_status"

		lvState5       = consts.LogEventStateInsertDB
		lfState5Status = "state_5_insert_order_status"

		lvState6       = consts.LogEventStateDelayQueueAddJob
		lfState6Status = "state_6_add_dq_job_status"

		lvState7       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState7Status = "state_7_kafka_publish_message_status"

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
	| STEP 3: Get ticket service
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	ticketReq := ticket.GetTicketRequest{
		ID:         req.TicketID,
		ClientName: data.Config.App.AppName,
	}
	reqAPI, respAPI, err := ucase.ticketProvider.GetTicket(ctx, ticketReq)

	lf = append(lf,
		logger.Any(lfState3Request, dto.TransformHeaderForLogging(reqAPI)),
		logger.Any(lfState3Response, util.DumpToString(respAPI)),
	)

	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallFailed, consts.TicketDependencyName, err), lf...)
		return
	}

	if respAPI.Name != consts.ResponseSuccess {
		response.SetName(respAPI.Name)
		response.SetMessage(respAPI.Message)
		response.SetError(respAPI.Errors)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallGotUnsuccessfulResponse, "ticket", respAPI.Name), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 4: Check reserved ticket
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	ticketResp := respAPI.Data
	if ticketResp.OrderID != "" {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageTicketAlreadyReserved, ticketResp.ID), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Check order
	* ----------------------------------------*/
	checkOrder, err := ucase.orderRepo.FindOneOrder(ctx, repositories.FindOneOrderCriteria{
		TicketID: ticketResp.ID,
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
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageTicketAlreadyReserved, ticketResp.ID), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 5: Insert order
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	order := entity.Order{
		ID:       generator.GenerateString(),
		TicketID: ticketResp.ID,
		UserID:   userID,
		Status:   consts.OrderStatusCreated,
		Amount:   ticketResp.Price,
	}

	isTicketReserved, err := ucase.orderRepo.InsertOrder(ctx, order)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState5Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNameOrder, err), lf...)
		return
	}

	if isTicketReserved {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState5Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageTicketAlreadyReserved, ticketResp.ID), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 6: Put order to delay queue
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState6))

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
			logger.Any(lfState6Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedAddDelayQueue, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 7: Publish kafka message
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState7))

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
		Payload:   &presentations.KafkaMessageCreateOrderPayload{ID: order.ID},
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
			logger.Any(lfState7Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState7Status, consts.LogStatusSuccess))

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
