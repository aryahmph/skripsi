package consts

const (
	// LogMessageFailedDecodePayload constant
	LogMessageFailedDecodePayload = "failed to decode payload, err: %v"

	LogMessageFailedFetchDB  = "failed to fetch data %s from db, err: %v"
	LogMessageDBNotFound     = "data %s not found"
	LogMessageFailedInsertDB = "failed to insert data %s to db, err: %v"
	LogMessageFailedUpdateDB = "failed to update data %s to db, err: %v"

	LogMessageFailedSetCache = "failed to set cache, err: %v"
	LogMessageFailedGetCache = "failed to get cache, err: %v"

	LogMessageAPICallFailed                  = `failed to make API call to %s, err: %v`
	LogMessageAPICallGotUnsuccessfulResponse = `API call to %s got unsuccessful response, %v`

	// LogMessageInvalidErr const
	LogMessageInvalidErr = "invalid, err: %v"

	// LogMessageInvalidWarn const
	LogMessageInvalidWarn = "invalid, warn: %v"

	// LogMessageFailedPublishMessage constant
	LogMessageFailedPublishMessage = "failed to publish message, err: %s"

	// LogMessageFailedToValidateRequestBody constant
	LogMessageFailedToValidateRequestBody = `failed to validate request body, err: %v`

	LogMessageEmailDomainIsBlacklist = "email domain %s is blacklist"

	// LogMessageSuccess constant
	LogMessageSuccess = "%v success"

	// LogMessagesFailedAuthenticateToken constant
	LogMessagesFailedAuthenticateToken = "failed to authenticate token, warn: %s"

	// LogMessageFoundError constant
	LogMessageFoundError = "found error. status %d"

	// LogMessageRegexError constant
	LogMessageRegexError = "regex:%s"

	// LogMessageInvalidGrantType constant
	LogMessageInvalidGrantType = "Invalid grant type"

	// LogMessageAggregateLoginSuccessForwarded	constant
	LogMessageAggregateLoginSuccessForwarded = "success forwarded to login type : %s"

	LogMessageAggregateOTAVerificationSuccessForwarded = "success forwarded to ota verification type : %s"

	// LogMessageUnsupportedOTAType constant
	LogMessageUnsupportedOTAType = "Unsupported OTA type"

	// LogMessageMaximumVerifyOTP constant
	LogMessageMaximumVerifyOTP = "maximum verify OTP"

	// LogMessageMaximumVerifyPIN constant
	LogMessageMaximumVerifyPIN = "maximum verify PIN"

	// LogMessageInvalidOTPCode constant
	LogMessageInvalidOTPCode = "OTP code is invalid"

	// LogMessageTokenAndOTANotMatch constant
	LogMessageTokenAndOTANotMatch = "%s between token and OTA is not match"

	// LogMessageUserRegisteredByEmailTrick const
	LogMessageUserRegisteredByEmailTrick = "user has been registered with email trick"

	// LogMessageFailedSendNotification const
	LogMessageFailedSendNotification = "failed sent notification, msg: %v"

	// LogMessagePINAlreadySet const
	LogMessagePINAlreadySet = "PIN already set"

	// LogMessagePINNotSet const
	LogMessagePINNotSet = "PIN not set yet"

	// LogMessageFeatureDisabled const
	LogMessageFeatureDisabled = "Feature is disabled"

	// LogMessageAggregateVerificationChannelSuccessForwarded const
	LogMessageAggregateVerificationChannelSuccessForwarded = "success forwarded to ota verification channel with type : %s"

	// LogMessageInvalidValidate2FAVerification const
	LogMessageInvalidValidate2FAVerification = "this request doesn't allow using %s channel since 2FA is enabled"

	// LogMessageVerification2FANotComplete const
	LogMessageVerification2FANotComplete = "verification is not complete since 2FA is enabled"

	// LogMessageFailedToCompareEET constant
	LogMessageFailedToCompareEET = "failed to compare eet time, err: %v"

	// LogMessageExpiredExecutionTime constant
	LogMessageExpiredExecutionTime = "execution time exceeded."

	// LogMessageUnverifiedEmailOrPhoneNumber const
	LogMessageUnverifiedEmailOrPhoneNumber = "unverified email or phone number"

	// LogMessageFailedAPICallHasura const
	LogMessageFailedAPICallHasura = `failed API call to hasura, err: %v`

	// LogMessageAPICallHasuraUnsuccessfulResponse const
	LogMessageAPICallHasuraUnsuccessfulResponse = `API call to hasura get unsuccesful response: %v`

	// LogMessageHasuraAccountIDExist consts
	LogMessageHasuraAccountIDExist = "Account ID exists"
)
