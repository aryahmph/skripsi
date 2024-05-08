package consts

const (
	LogEventStateValidateRequest = "ValidateRequest"
	LogEventStateDecodeRequest   = "DecodeRequest"

	LogEventStateKafkaPublishMessageToTopic = "KafkaPublishMessageToTopic"
	LogEventStateDelayQueueAddJob           = "DelayQueueAddJob"
	LogEventStateValidateExecutionTime      = "ValidateExecutionTime"

	LogEventStateFetchDB  = "FetchDB"
	LogEventStateInsertDB = "InsertDB"
	LogEventStateUpdateDB = "UpdateDB"

	LogEventStateGetCache = "GetCache"
	LogEventStatePutCache = "PutCache"

	LogEventStateGetOrder             = "GetOrder"
	LogEventStateCreatePaymentGateway = "CreatePaymentGateway"
	LogEventStateCheckPayment         = "CheckPayment"
	LogEventStateCheckReservedTicket  = "CheckReservedTicket"
)
