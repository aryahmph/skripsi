package consts

const (
	LogEventStateValidateRequest       = "ValidateRequest"
	LogEventStateDecodeRequest         = "DecodeRequest"
	LogEventStateValidateExecutionTime = "ValidateExecutionTime"
	LogEventStateCallProvider          = "CallProvider"

	LogEventStateKafkaPublishMessageToTopic = "KafkaPublishMessageToTopic"

	LogEventStateFetchDB  = "FetchDB"
	LogEventStateInsertDB = "InsertDB"
	LogEventStateUpdateDB = "UpdateDB"

	LogEventStateGetCache    = "GetCache"
	LogEventStatePutCache    = "PutCache"
	LogEventStateUpdateCache = "UpdateCache"
	LogEventStateRemoveCache = "RemoveCache"
)
