package order

import (
	"context"
	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/internal/dto"
	"en-order/internal/entity"
	"en-order/internal/helper"
	"en-order/internal/presentations"
	"en-order/internal/repositories"
	ucaseContract "en-order/internal/ucase/contract"
	"en-order/pkg/cache"
	"en-order/pkg/logger"
	"en-order/pkg/tracer"
	"en-order/pkg/util"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
)

type internalGetOrder struct {
	cacher    cache.Cacher
	orderRepo repositories.OrderRepository
}

func NewInternalGetOrder(cacher cache.Cacher, orderRepo repositories.OrderRepository) ucaseContract.UseCase {
	return &internalGetOrder{cacher: cacher, orderRepo: orderRepo}
}

func (ucase *internalGetOrder) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		//lvState2       = consts.LogEventStateGetCache
		//lfState2Status = "state_2_get_order_from_cache_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_fetch_order_from_db_status"

		//lvState4       = consts.LogEventStatePutCache
		//lfState4Status = "state_4_put_order_data_to_cache_status"

		orderID    = mux.Vars(data.Request)["id"]
		clientName = data.Request.Header.Get(consts.HeaderXClientName)

		req presentations.InternalGetOrderRequest

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameInternalGetOrder),
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
	//dataCache, err := ucase.getDataFromCache(ctx, orderID)
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
	//	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetOrder), lf...)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 3: Check DB
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	order, err := ucase.orderRepo.FindOneOrder(ctx, repositories.FindOneOrderCriteria{
		ID:     orderID,
		UserID: req.UserID,
		Status: req.Status,
	})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameOrder, err),
			lf...,
		)
		return
	}

	if order == nil {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameOrder),
			lf...,
		)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	///*----------------------------
	//| STEP 4: Put data to cache
	//* ---------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState4))
	//
	orderResp := dto.ConstructGetOrderResponse(order)
	//err = ucase.putDataToCache(ctx, &orderResp, data.Config.Cache.InsertOrderTTLSecond)
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
	response.SetData(orderResp)

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetOrder), lf...)
	return
}

func (ucase *internalGetOrder) getDataFromCache(ctx context.Context, id string) (result *presentations.GetOrderResponse, err error) {
	listByte, err := ucase.cacher.Get(ctx, helper.OrderCacheKey(id))
	if err != nil {
		return nil, err
	}

	if len(listByte) == 0 {
		return nil, nil
	}
	_ = json.Unmarshal(listByte, &result)
	return
}

func (ucase *internalGetOrder) putDataToCache(ctx context.Context, r *presentations.GetOrderResponse, ttl int64) error {
	return ucase.cacher.Set(ctx, helper.OrderCacheKey(r.ID), util.DumpToString(r), time.Duration(ttl)*time.Second)
}
