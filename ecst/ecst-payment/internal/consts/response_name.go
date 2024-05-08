// Package consts
package consts

const (

	// ResponseAuthenticationFailure response var
	ResponseAuthenticationFailure = "AUTHENTICATION_FAILURE"

	// ResponseSignatureFailure response var
	ResponseSignatureFailure = "SIGNATURE_FAILURE"

	// MiddlewarePassed response var for middleware http request passed
	MiddlewarePassed = "MIDDLEWARE_PASSED"

	// ResponseValidationFailure response var for general validation error or not pass
	ResponseValidationFailure = "VALIDATION_FAILURE"

	// ResponseDataNotFound response var for general not found data in our system or third party services
	ResponseDataNotFound = "DATA_NOT_FOUND"

	// ResponseSuccess response var for general success
	ResponseSuccess = "SUCCESS"

	// ResponseInternalFailure response var for internal server error or like something went wrong in system
	ResponseInternalFailure = "INTERNAL_FAILURE"

	// ResponseUnprocessableEntity response var for general wen we cannot continue process and can not retry
	ResponseUnprocessableEntity = "UNPROCESSABLE_ENTITY"

	// ResponseForbidden response var for forbidden access
	ResponseForbidden = `FORBIDDEN`

	// ResponseRequestTimeout response var
	ResponseRequestTimeout = `REQUEST_TIMEOUT`

	ResponseOrderExpired = "ORDER_EXPIRED"
)
