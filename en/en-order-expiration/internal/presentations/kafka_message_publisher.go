package presentations

type (
	// KafkaMessageBase is a base message struct of a Kafka message.
	KafkaMessageBase struct {
		Source    KafkaMessageBaseSource `json:"source"`
		Check     *KafkaMessageChecker   `json:"check,omitempty"`
		Signature string                 `json:"signature,omitempty"`
		CreatedAt string                 `json:"created_at"`
		Payload   interface{}            `json:"payload"`
	}

	// KafkaMessageBaseSource is a source struct of the Kafka base message.
	KafkaMessageBaseSource struct {
		AppName    string `json:"app_name"`
		AppEnv     string `json:"app_env"`
		AppCommand string `json:"app_command,omitempty"`
	}

	// KafkaMessageOriginService contract
	KafkaMessageOriginService struct {
		ServiceName string `json:"service_name"`
		TargetTopic string `json:"target_topic"`
	}

	// KafkaMessageChecker contract
	KafkaMessageChecker struct {
		InitiateTime  string                    `json:"initiate_time"`
		ServiceOrigin KafkaMessageOriginService `json:"service_origin,omitempty"`
		Count         uint                      `json:"count"`
		NextSecond    uint                      `json:"next_second"`
		MaxSecond     uint                      `json:"max_second"`
	}

	// MessageOriginService contract
	MessageOriginService struct {
		ServiceName string `json:"service_name"`
		TargetTopic string `json:"target_topic"`
	}
)

type (
	KafkaMessageCreateOrderPayload struct {
		ID string `json:"id"`
	}
)
