package payment

import (
	"context"
	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/internal/dto"
	"en-payment/internal/entity"
	"en-payment/internal/helper"
	"en-payment/internal/presentations"
	"en-payment/internal/repositories"
	"en-payment/internal/ucase/contract"
	"encoding/json"
	"time"

	"en-payment/pkg/cache"
	"en-payment/pkg/logger"
	"en-payment/pkg/tracer"
	"en-payment/pkg/util"

	"github.com/gorilla/mux"
)

type internalGetPayment struct {
	cacher      cache.Cacher
	paymentRepo repositories.PaymentRepository
}

func NewInternalGetPayment(cacher cache.Cacher, paymentRepo repositories.PaymentRepository) contract.UseCase {
	return &internalGetPayment{cacher: cacher, paymentRepo: paymentRepo}
}

func (ucase *internalGetPayment) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		//lvState2       = consts.LogEventStateGetCache
		//lfState2Status = "state_2_get_payment_from_cache_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_fetch_payment_from_db_status"
		//
		//lvState4       = consts.LogEventStatePutCache
		//lfState4Status = "state_4_put_payment_data_to_cache_status"

		paymentID  = mux.Vars(data.Request)["id"]
		clientName = data.Request.Header.Get(consts.HeaderXClientName)

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")
		req presentations.GetPaymentRequest

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameInternalGetPayment),
			logger.Any(consts.LogFieldClientName, clientName),
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

	///*----------------------------
	//| STEP 2: Check cache
	//* ---------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState2))
	//
	//dataCache, err := ucase.getDataFromCache(ctx, paymentID, req.UserID)
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//	response.SetName(consts.ResponseInternalFailure)
	//
	//	lf = append(lf,
	//		logger.Any(lfState2Status, consts.LogStatusFailed),
	//		logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
	//	)
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedGetCache, err),
	//		lf...,
	//	)
	//	return
	//}
	//
	//if dataCache != nil {
	//	response.SetName(consts.ResponseSuccess)
	//	response.SetData(dataCache)
	//
	//	lf = append(lf,
	//		logger.Any(lfState2Status, consts.LogStatusFounded),
	//		logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
	//	)
	//
	//	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetPayment), lf...)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 3: Check DB
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	payment, err := ucase.paymentRepo.FindOnePayment(ctx, repositories.FindOnePaymentCriteria{ID: paymentID})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNamePayment, err),
			lf...,
		)
		return
	}

	if payment == nil {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNamePayment),
			lf...,
		)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	///*----------------------------
	//| STEP 3: Put data to cache
	//* ---------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState4))
	//
	paymentResp := dto.ConstructGetPaymentResponse(payment)
	//err = ucase.putDataToCache(ctx, req.UserID, &paymentResp, data.Config.Cache.InsertPaymentTTLSecond)
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//	response.SetName(consts.ResponseInternalFailure)
	//
	//	lf = append(lf,
	//		logger.Any(lfState4Status, consts.LogStatusFailed),
	//		logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
	//	)
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedSetCache, err),
	//		lf...,
	//	)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState4Status, consts.LogStatusSuccess))

	response.SetName(consts.ResponseSuccess)
	response.SetData(paymentResp)

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetPayment), lf...)
	return
}

func (ucase *internalGetPayment) getDataFromCache(ctx context.Context, id, userId string) (result *presentations.GetPaymentResponse, err error) {
	listByte, err := ucase.cacher.Get(ctx, helper.PaymentCacheKey(id, userId))
	if err != nil {
		return nil, err
	}

	if len(listByte) == 0 {
		return nil, nil
	}
	_ = json.Unmarshal(listByte, &result)
	return
}

func (ucase *internalGetPayment) putDataToCache(ctx context.Context, userId string, r *presentations.GetPaymentResponse, ttl int64) error {
	return ucase.cacher.Set(ctx, helper.PaymentCacheKey(r.ID, userId), util.DumpToString(r), time.Duration(ttl)*time.Second)
}
