package consts

const (
	LogEventStateValidateRequest = "ValidateRequest"
	LogEventStateDecodeRequest   = "DecodeRequest"

	LogEventStateKafkaPublishMessageToTopic = "KafkaPublishMessageToTopic"
	LogEventStateDelayQueueAddJob           = "DelayQueueAddJob"
	LogEventStateDelayQueueRemoveJob        = "DelayQueueRemoveJob"
	LogEventStateValidateExecutionTime      = "ValidateExecutionTime"

	LogEventStateFetchDB  = "FetchDB"
	LogEventStateInsertDB = "InsertDB"
	LogEventStateUpdateDB = "UpdateDB"

	LogEventStateGetCache    = "GetCache"
	LogEventStatePutCache    = "PutCache"
	LogEventStateRemoveCache = "RemoveCache"

	LogEventStateGetTicket           = "GetTicket"
	LogEventStateCheckReservedTicket = "CheckReservedTicket"
	LogEventStateGetPayment          = "GetPayment"
)
