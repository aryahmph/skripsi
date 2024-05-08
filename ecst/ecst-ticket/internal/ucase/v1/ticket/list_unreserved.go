package ticket

import (
	"context"
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/dto"
	"ecst-ticket/internal/entity"
	"ecst-ticket/internal/helper"
	"ecst-ticket/internal/presentations"
	"ecst-ticket/internal/repositories"
	"ecst-ticket/internal/ucase/contract"
	"ecst-ticket/pkg/cache"
	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/tracer"
	"ecst-ticket/pkg/util"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
)

type listUnreserved struct {
	cacher     cache.Cacher
	ticketRepo repositories.TicketRepository
}

func NewListUnreserved(cacher cache.Cacher, ticketRepo repositories.TicketRepository) contract.UseCase {
	return &listUnreserved{cacher: cacher, ticketRepo: ticketRepo}
}

func (ucase *listUnreserved) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		//lvState2       = consts.LogEventStateGetCache
		//lfState2Status = "state_2_get_list_from_cache_status"

		lvState3       = consts.LogEventStateFetchDB
		lfState3Status = "state_3_fetch_ticket_from_db_status"

		//lvState4       = consts.LogEventStatePutCache
		//lfState4Status = "state_4_put_ticket_data_to_cache_status"

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		groupID = mux.Vars(data.Request)["group_id"]

		userID    = data.Request.Header.Get(consts.HeaderXUserId)
		userEmail = data.Request.Header.Get(consts.HeaderXUserEmail)

		req presentations.ListUnreservedTicketsRequest

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameListUnreservedTickets),
			logger.Any(consts.LogFieldUserID, userID),
			logger.Any(consts.LogFieldUserEmail, userEmail),
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
	//dataCache, err := ucase.getDataFromCache(ctx, groupID, req.Category)
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
	//	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameListUnreservedTickets), lf...)
	//	return
	//}
	//
	//lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 3: Check DB
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	tickets, err := ucase.ticketRepo.ListUnreservedTickets(ctx, groupID, req.Category)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicket, err),
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
	ticketsResp := dto.ConstructListUnreservedTicketsResponse(tickets)
	//err = ucase.putDataToCache(ctx, groupID, req.Category, ticketsResp, data.Config.Cache.ListUnreservedTicketsTTLSecond)
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
	response.SetData(ticketsResp)

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameInternalGetTicket), lf...)
	return
}

func (ucase *listUnreserved) getDataFromCache(ctx context.Context, groupID, category string) (result []presentations.ListUnreservedTicketsResponse, err error) {
	listByte, err := ucase.cacher.Get(ctx, helper.ListUnreservedTicketsCacheKey(groupID, category))
	if err != nil {
		return nil, err
	}

	if len(listByte) == 0 {
		return nil, nil
	}
	_ = json.Unmarshal(listByte, &result)
	return
}

func (ucase *listUnreserved) putDataToCache(ctx context.Context, groupID, category string, r []presentations.ListUnreservedTicketsResponse, ttl int64) error {
	return ucase.cacher.Set(ctx, helper.ListUnreservedTicketsCacheKey(groupID, category), util.DumpToString(r), time.Duration(ttl)*time.Second)
}
