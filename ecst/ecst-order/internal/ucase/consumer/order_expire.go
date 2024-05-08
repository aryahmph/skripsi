package consumer

import (
	"context"
	"ecst-order/pkg/kafka"
	"encoding/json"
	"github.com/spf13/cast"
	"time"

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
)

type orderExpire struct {
	cfg       *appctx.Config
	kp        kafka.Producer
	orderRepo repositories.OrderRepository
}

func NewOrderExpire(cfg *appctx.Config, kp kafka.Producer, orderRepo repositories.OrderRepository) contract.MessageProcessor {
	return &orderExpire{cfg: cfg, kp: kp, orderRepo: orderRepo}
}

func (ucase *orderExpire) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_fetch_order_from_db_status"
		lfState3Data   = "state_3_fetch_order_from_db_data"

		lvState4       = consts.LogEventStateUpdateDB
		lfState4Status = "state_4_update_order_status_status"

		lvState5       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState5Status = "state_5_kafka_publish_message_to_topic_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameExpireOrder),
		}

		err     error
		msgData presentations.MessageOrderExpirationComplete
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

	valid, err := helper.IsValidExecutionTime(time.Duration(ucase.cfg.KafkaEETSecond.EETOrderExpirationComplete)*time.Second, msgData.CreatedAt, now.Format(consts.LayoutDateTimeFormat))
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
		ID:     msgData.Payload.ID,
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

	lf = append(lf,
		logger.Any(lfState3Status, consts.LogStatusSuccess),
		logger.Any(lfState3Data, util.DumpToString(order)),
	)

	/*-------------------------------
	| STEP 4: Update order status
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	isExist, err := ucase.orderRepo.UpdateOrder(ctx, entity.Order{
		ID:      order.ID,
		Status:  consts.OrderStatusExpired,
		Version: order.Version + 1,
	})
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

	/*-------------------------------
	| STEP 5: Publish to kafka
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState5))

	message := presentations.KafkaMessageBase{
		Source: presentations.KafkaMessageBaseSource{
			AppName: ucase.cfg.App.AppName,
			AppEnv:  ucase.cfg.App.Env,
		},
		Check: &presentations.KafkaMessageChecker{
			InitiateTime: time.Now().Format(consts.LayoutDateTimeFormat),
			ServiceOrigin: presentations.KafkaMessageOriginService{
				ServiceName: ucase.cfg.App.AppName,
				TargetTopic: ucase.cfg.KafkaTopics.TopicOrderExpire,
			},
			Count:      0,
			NextSecond: cast.ToUint(ucase.cfg.KafkaNextSecond.NextOrderExpire),
			MaxSecond:  cast.ToUint(ucase.cfg.KafkaEETSecond.EETOrderExpire),
		},
		CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
		Payload: &presentations.KafkaMessageOrderExpirePayload{
			ID:       order.ID,
			TicketID: order.TicketID,
			Version:  order.Version + 1,
		},
	}

	lf = append(lf, logger.EventInputKafka(message, util.DumpToString(message)))

	km := kafka.MessageContext{
		Value:   util.DumpToString(message),
		Topic:   ucase.cfg.KafkaTopics.TopicOrderExpire,
		Verbose: true,
	}

	err = ucase.kp.Publish(ctx, &km)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState2Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameExpireOrder), lf...)
	return nil
}
