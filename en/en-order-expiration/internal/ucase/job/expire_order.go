package job

import (
	"context"
	"encoding/json"
	"github.com/spf13/cast"
	"time"

	"en-order-expiration/internal/appctx"
	"en-order-expiration/internal/consts"
	"en-order-expiration/internal/presentations"
	ucaseContract "en-order-expiration/internal/ucase/contract"

	"en-order-expiration/pkg/kafka"
	"en-order-expiration/pkg/logger"
	"en-order-expiration/pkg/tracer"
	"en-order-expiration/pkg/util"
)

type expireOrder struct {
	cfg *appctx.Config
	kp  kafka.Producer
}

func NewExpireOrder(cfg *appctx.Config, kp kafka.Producer) ucaseContract.JobProcessor {
	return &expireOrder{
		cfg: cfg,
		kp:  kp,
	}
}

func (ucase *expireOrder) Serve(ctx context.Context, data *appctx.WatcherData) error {
	var (
		lvState1        = consts.LogEventStateDecodeRequest
		lfState1Status  = "state_1_decode_request_status"
		lfState1Payload = "state_1_decode_request_payload"

		lvState2       = consts.LogEventStateKafkaPublishMessageToTopic
		lfState2Status = "state_2_kafka_publish_message_to_topic_status"

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameRunJobExpireOrder),
		}

		payload presentations.OrderExpirePayload
	)

	ctx = tracer.SpanStart(ctx, `job:run-expire-order`)
	defer tracer.SpanFinish(ctx)

	/*-------------------------------
	| STEP 1: Decode Request
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState1))

	err := json.Unmarshal([]byte(data.Value), &payload)
	if err != nil {
		tracer.SpanError(ctx, err)

		lf = append(lf, logger.Any(lfState1Status, consts.LogStatusFailed))
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedDecodePayload, err), lf...)
		return err
	}

	lf = append(lf,
		logger.Any(lfState1Status, consts.LogStatusSuccess),
		logger.Any(lfState1Payload, util.DumpToString(payload)),
	)

	/*-------------------------------
	| STEP 2: Publish Kafka Message
	* -------------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState2))

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
		Payload:   &payload,
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

	lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	data.Commit()
	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameRunJobExpireOrder), lf...)
	return nil
}
