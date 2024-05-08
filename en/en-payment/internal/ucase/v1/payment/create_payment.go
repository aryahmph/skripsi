package payment

import (
	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/internal/dto"
	"en-payment/internal/entity"
	"en-payment/internal/presentations"
	"en-payment/internal/providers/order"
	"en-payment/internal/repositories"
	"en-payment/internal/ucase/contract"
	"en-payment/pkg/generator"
	"en-payment/pkg/kafka"
	"en-payment/pkg/logger"
	"en-payment/pkg/tracer"
	"en-payment/pkg/util"
	"github.com/spf13/cast"
	"github.com/thedevsaddam/govalidator"
	"net/url"
	"time"
)

type createPayment struct {
	kp            kafka.Producer
	paymentRepo   repositories.PaymentRepository
	orderProvider order.OrderProvider
}

func NewCreatePayment(
	kp kafka.Producer,
	paymentRepo repositories.PaymentRepository,
	orderProvider order.OrderProvider,
) contract.UseCase {
	return &createPayment{
		kp:            kp,
		paymentRepo:   paymentRepo,
		orderProvider: orderProvider,
	}
}

func (ucase *createPayment) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateRequest
		lfState2Status = "state_2_validate_request_status"

		lvState3       = consts.LogEventStateCheckPayment
		lfState3Status = "state_3_check_payment_status"

		lvState4         = consts.LogEventStateGetOrder
		lfState4Status   = "state_4_get_order_from_order_service_status"
		lfState4Request  = "state_4_get_order_from_order_service_request"
		lfState4Response = "state_4_get_order_from_order_service_response"

		lvState5       = consts.LogEventStateCreatePaymentGateway
		lfState5Status = "state_5_create_payment_gateway_status"

		lvState6       = consts.LogEventStateInsertDB
		lfState6Status = "state_6_insert_payment_status"

		lvState7       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState7Status = "state_7_kafka_publish_message_status"

		userID    = data.Request.Header.Get(consts.HeaderXUserId)
		userEmail = data.Request.Header.Get(consts.HeaderXUserEmail)

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		req presentations.CreatePaymentRequest

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameCreatePayment),
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
	| STEP 3: Check payment by order
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	payment, err := ucase.paymentRepo.FindOnePayment(ctx, repositories.FindOnePaymentCriteria{OrderID: req.OrderID})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNamePayment, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	if payment != nil {
		response.SetName(consts.ResponseSuccess)
		response.SetData(presentations.CreatePaymentResponse{ID: payment.ID})

		logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreatePayment), lf...)
		return
	}

	/*-----------------------------------------
	| STEP 3: Get order
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	orderReq := order.GetOrderRequest{
		ID:         req.OrderID,
		UserID:     userID,
		ClientName: data.Config.App.AppName,
	}
	reqAPI, respAPI, err := ucase.orderProvider.GetOrder(ctx, orderReq)

	lf = append(lf,
		logger.Any(lfState4Request, dto.TransformHeaderForLogging(reqAPI)),
		logger.Any(lfState4Response, util.DumpToString(respAPI)),
	)

	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallFailed, consts.OrderDependencyName, err), lf...)
		return
	}

	if respAPI.Name != consts.ResponseSuccess {
		response.SetName(respAPI.Name)
		response.SetMessage(respAPI.Message)
		response.SetError(respAPI.Errors)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageAPICallGotUnsuccessfulResponse, "order", respAPI.Name), lf...)
		return
	}

	switch respAPI.Data.Status {
	case consts.OrderStatusExpired:
		response.SetName(consts.ResponseOrderExpired)

		lf = append(lf,
			logger.Any(lfState4Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx, consts.LogMessageOrderAlreadyExpired, lf...)
		return
	case consts.OrderStatusCompleted:
		response.SetName(consts.ResponseSuccess)
		response.SetData(presentations.CreatePaymentResponse{ID: respAPI.Data.PaymentID})

		logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreatePayment), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	/*-------------------------------------------
	| STEP 4: Create payment with payment gateway
	* -------------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	duration := 50 * time.Millisecond
	time.Sleep(duration)

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*--------------------------
	| STEP 5: Insert payment
	* --------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState6))

	paymentEnt := entity.Payment{
		ID:      generator.GenerateString(),
		OrderID: req.OrderID,
	}

	err = ucase.paymentRepo.InsertPayment(ctx, paymentEnt)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState6Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNamePayment, err), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	/*--------------------------
	| STEP 6: Publish message
	* --------------------------*/
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
				TargetTopic: data.Config.KafkaTopics.TopicCreatePayment,
			},
			Count:      0,
			NextSecond: cast.ToUint(data.Config.KafkaNextSecond.NextCreatePayment),
			MaxSecond:  cast.ToUint(data.Config.KafkaEETSecond.EETCreatePayment),
		},
		CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
		Payload:   &presentations.KafkaMessageCreateOrderPayload{ID: paymentEnt.ID},
	}

	km := kafka.MessageContext{
		Value:   util.DumpToString(message),
		Topic:   data.Config.KafkaTopics.TopicCreatePayment,
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
	response.SetData(presentations.CreatePaymentResponse{ID: paymentEnt.ID})

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreatePayment), lf...)
	return
}

func (ucase *createPayment) validateRequest(req presentations.CreatePaymentRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &req,
		Rules: govalidator.MapData{
			"order_id":    consts.RuleULID,
			"card_number": []string{"required"},
			"exp_month":   []string{"required"},
			"exp_year":    []string{"required"},
			"cvv":         []string{"required"},
		},
		TagIdentifier: "json",
	})

	errVal := v.ValidateStruct()
	if len(errVal) > 0 {
		return errVal
	}

	return errVal
}
