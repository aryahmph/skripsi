package consumer

import (
	"context"
	"en-order/pkg/cache"
	"en-order/pkg/kafka"
	"encoding/json"
	"github.com/spf13/cast"
	"time"

	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/internal/entity"
	"en-order/internal/helper"
	"en-order/internal/presentations"
	"en-order/internal/repositories"
	"en-order/internal/ucase/contract"

	"en-order/pkg/logger"
	"en-order/pkg/tracer"
	"en-order/pkg/util"
)

type orderExpire struct {
	cfg       *appctx.Config
	orderRepo repositories.OrderRepository
	cacher    cache.Cacher
	kp        kafka.Producer
}

func NewOrderExpire(cfg *appctx.Config, cacher cache.Cacher, orderRepo repositories.OrderRepository, kp kafka.Producer) contract.MessageProcessor {
	return &orderExpire{cfg: cfg, cacher: cacher, orderRepo: orderRepo, kp: kp}
}

func (ucase *orderExpire) Serve(ctx context.Context, data *appctx.ConsumerData) error {
	defer data.Commit()

	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = consts.LogEventStateValidateExecutionTime
		lfState2Status = "state_2_validate_execution_time_status"

		//lvState3       = consts.LogEventStateFetchDB
		//lfState3Status = "state_3_fetch_order_from_db_status"
		//lfState3Data   = "state_3_fetch_order_from_db_data"

		lvState4       = consts.LogEventStateUpdateDB
		lfState4Status = "state_4_update_order_status_status"

		//lvState5       = consts.LogEventStateRemoveCache
		//lfState5Status = "state_5_remove_order_cache_status"

		lvState6       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState6Status = "state_6_kafka_publish_message_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameExpireOrder),
		}

		err     error
		msgData presentations.MessageExpireOrder
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
	| STEP 4: Update order status
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState4))

	isExist, err := ucase.orderRepo.UpdateOrderStatus(ctx, msgData.Payload.ID, consts.OrderStatusExpired, "")
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

	///*-------------------------------
	//| STEP 5: Remove cache
	//* -------------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState5))
	//
	//err = ucase.cacher.Delete(ctx, helper.OrderCacheKey(order.ID))
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//
	//	lf = append(lf, logger.Any(lfState5Status, consts.LogStatusFailed))
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedRemoveCache, err),
	//		lf...,
	//	)
	//	return err
	//}
	//
	//lf = append(lf, logger.Any(lfState5Status, consts.LogStatusSuccess))

	/*-----------------------------------------
	| STEP 6: Publish kafka message
	* ----------------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState6))

	message := presentations.KafkaMessageBase{
		Source: presentations.KafkaMessageBaseSource{
			AppName: ucase.cfg.App.AppName,
			AppEnv:  ucase.cfg.App.Env,
		},
		Check: &presentations.KafkaMessageChecker{
			InitiateTime: time.Now().Format(consts.LayoutDateTimeFormat),
			ServiceOrigin: presentations.KafkaMessageOriginService{
				ServiceName: ucase.cfg.App.AppName,
				TargetTopic: ucase.cfg.KafkaTopics.TopicOrderExpire2,
			},
			Count:      0,
			NextSecond: cast.ToUint(ucase.cfg.KafkaNextSecond.NextCreateOrder),
			MaxSecond:  cast.ToUint(ucase.cfg.KafkaEETSecond.EETCreateOrder),
		},
		CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
		Payload:   &presentations.KafkaMessageExpireOrderPayload{ID: msgData.Payload.ID},
	}

	km := kafka.MessageContext{
		Value:   util.DumpToString(message),
		Topic:   ucase.cfg.KafkaTopics.TopicOrderExpire2,
		Verbose: true,
	}

	lf = append(lf, logger.EventInputKafka(message, util.DumpToString(message)))

	err = ucase.kp.Publish(ctx, &km)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState6Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err), lf...)
		return err
	}

	lf = append(lf, logger.Any(lfState6Status, consts.LogStatusSuccess))

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameExpireOrder), lf...)
	return nil
}
