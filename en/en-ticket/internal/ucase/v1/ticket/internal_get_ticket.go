package ticket

import (
	"context"
	"en-ticket/internal/appctx"
	"en-ticket/internal/consts"
	"en-ticket/internal/dto"
	"en-ticket/internal/entity"
	"en-ticket/internal/helper"
	"en-ticket/internal/presentations"
	repository "en-ticket/internal/repositories"
	ucase "en-ticket/internal/ucase/contract"
	"en-ticket/pkg/cache"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/tracer"
	"en-ticket/pkg/util"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
)

type internalGetTicket struct {
	ticketRepo repository.TicketRepository
	cacher     cache.Cacher
}

func NewInternalGetTicket(ticketRepo repository.TicketRepository, cacher cache.Cacher) ucase.UseCase {
	return &internalGetTicket{ticketRepo: ticketRepo, cacher: cacher}
}

func (ucase *internalGetTicket) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		//lvState1       = consts.LogEventStateGetCache
		//lfState1Status = "state_1_get_ticket_from_cache_status"

		lvState2       = consts.LogEventStateFetchDB
		lfState2Status = "state_2_fetch_ticket_from_db_status"

		//lvState3       = consts.LogEventStatePutCache
		//lfState3Status = "state_3_put_ticket_data_to_cache_status"

		ticketID   = mux.Vars(data.Request)["id"]
		clientName = data.Request.Header.Get(consts.HeaderXClientName)

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameInternalGetTicket),
			logger.Any(consts.LogFieldClientName, clientName),
		}
	)

	defer tracer.SpanFinish(ctx)

	///*----------------------------
	//| STEP 1: Check cache
	//* ---------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState1))
	//
	//dataCache, err := ucase.getDataFromCache(ctx, ticketID)
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//	response.SetName(consts.ResponseInternalFailure)
	//
	//	lf = append(lf,
	//		logger.Any(lfState1Status, consts.LogStatusFailed),
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
	//		logger.Any(lfState1Status, consts.LogStatusFounded),
	//		logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
	//	)
	//
	//	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetTicket), lf...)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState1Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 2: Check DB
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState2))

	ticket, err := ucase.ticketRepo.FindOneTicket(ctx, repository.FindOneTicketCriteria{ID: ticketID})
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState2Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicket, err),
			lf...,
		)
		return
	}

	if ticket == nil {
		response.SetName(consts.ResponseDataNotFound)

		lf = append(lf,
			logger.Any(lfState2Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.WarnWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameTicket),
			lf...,
		)
		return
	}

	lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	///*----------------------------
	//| STEP 2-3: Put data to cache
	//* ---------------------------*/
	//lf = append(lf, logger.Any(consts.LogFieldState, lvState3))
	//
	ticketResp := dto.ConstructGetTicketResponse(ticket)
	//err = ucase.putDataToCache(ctx, &ticketResp, data.Config.Cache.InsertTicketTTLSecond)
	//if err != nil {
	//	tracer.SpanError(ctx, err)
	//	response.SetName(consts.ResponseInternalFailure)
	//
	//	lf = append(lf,
	//		logger.Any(lfState3Status, consts.LogStatusFailed),
	//		logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
	//	)
	//	logger.ErrorWithContext(ctx,
	//		logger.SetMessageFormat(consts.LogMessageFailedSetCache, err),
	//		lf...,
	//	)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	response.SetName(consts.ResponseSuccess)
	response.SetData(ticketResp)

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetTicket), lf...)
	return
}

func (ucase *internalGetTicket) getDataFromCache(ctx context.Context, id string) (result *presentations.GetTicketResponse, err error) {
	listByte, err := ucase.cacher.Get(ctx, helper.TicketCacheKey(id))
	if err != nil {
		return nil, err
	}

	if len(listByte) == 0 {
		return nil, nil
	}
	_ = json.Unmarshal(listByte, &result)
	return
}

func (ucase *internalGetTicket) putDataToCache(ctx context.Context, r *presentations.GetTicketResponse, ttl int64) error {
	return ucase.cacher.Set(ctx, helper.TicketCacheKey(r.ID), util.DumpToString(r), time.Duration(ttl)*time.Second)
}
