package ticket

import (
	"context"
	"en-ticket/internal/appctx"
	"en-ticket/internal/consts"
	"en-ticket/internal/dto"
	"en-ticket/internal/entity"
	"en-ticket/internal/helper"
	"en-ticket/internal/presentations"
	"en-ticket/internal/repositories"
	"en-ticket/internal/ucase/contract"
	"en-ticket/pkg/cache"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/tracer"
	"en-ticket/pkg/util"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
)

type listCategories struct {
	cacher     cache.Cacher
	ticketRepo repositories.TicketRepository
}

func NewListCategories(cacher cache.Cacher, ticketRepo repositories.TicketRepository) contract.UseCase {
	return &listCategories{cacher: cacher, ticketRepo: ticketRepo}
}

func (ucase *listCategories) Serve(data *appctx.Data) (response appctx.Response) {
	var (
		lvState1       = consts.LogEventStateGetCache
		lfState1Status = "state_1_get_list_from_cache_status"

		lvState2       = consts.LogEventStateFetchDB
		lfState2Status = "state_2_fetch_ticket_from_db_status"

		lvState3       = consts.LogEventStatePutCache
		lfState3Status = "state_3_put_ticket_data_to_cache_status"

		ctx = tracer.SpanStart(data.Request.Context(), "Serve")

		groupID = mux.Vars(data.Request)["group_id"]

		userID    = data.Request.Header.Get(consts.HeaderXUserId)
		userEmail = data.Request.Header.Get(consts.HeaderXUserEmail)

		lf = []logger.Field{
			logger.EventName(consts.LogEventNameListTicketCategories),
			logger.Any(consts.LogFieldUserID, userID),
			logger.Any(consts.LogFieldUserEmail, userEmail),
		}
	)

	defer tracer.SpanFinish(ctx)

	/*----------------------------
	| STEP 1: Check cache
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState1))

	dataCache, err := ucase.getDataFromCache(ctx, groupID)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState1Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedGetCache, err),
			lf...,
		)
		return
	}

	if dataCache != nil {
		response.SetName(consts.ResponseSuccess)
		response.SetData(dataCache)

		lf = append(lf,
			logger.Any(lfState1Status, consts.LogStatusFounded),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)

		logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameListTicketCategories), lf...)
		return
	}

	lf = append(lf, logger.Any(lfState1Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 2: Check DB
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState2))

	categories, err := ucase.ticketRepo.ListTicketCategories(ctx, groupID)
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

	lf = append(lf, logger.Any(lfState2Status, consts.LogStatusSuccess))

	/*----------------------------
	| STEP 3: Put data to cache
	* ---------------------------*/
	lf = append(lf, logger.Any(consts.LogFieldState, lvState3))

	categoriesResp := dto.ConstructListTicketCategoriesResponse(categories)
	err = ucase.putDataToCache(ctx, groupID, categoriesResp, data.Config.Cache.ListTicketCategoriesTTLSecond)
	if err != nil {
		tracer.SpanError(ctx, err)
		response.SetName(consts.ResponseInternalFailure)

		lf = append(lf,
			logger.Any(lfState3Status, consts.LogStatusFailed),
			logger.EventOutputHttp(response.GetCode(), response, util.DumpToString(response)),
		)
		logger.ErrorWithContext(ctx,
			logger.SetMessageFormat(consts.LogMessageFailedSetCache, err),
			lf...,
		)
		return
	}

	lf = append(lf, logger.Any(lfState3Status, consts.LogStatusSuccess))

	response.SetName(consts.ResponseSuccess)
	response.SetData(categoriesResp)

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameListTicketCategories), lf...)
	return
}

func (ucase *listCategories) getDataFromCache(ctx context.Context, id string) (result []presentations.ListTicketCategoriesResponse, err error) {
	listByte, err := ucase.cacher.Get(ctx, helper.ListTicketCategoriesCacheKey(id))
	if err != nil {
		return nil, err
	}

	if len(listByte) == 0 {
		return nil, nil
	}
	_ = json.Unmarshal(listByte, &result)
	return
}

func (ucase *listCategories) putDataToCache(ctx context.Context, groupID string, r []presentations.ListTicketCategoriesResponse, ttl int64) error {
	return ucase.cacher.Set(ctx, helper.ListTicketCategoriesCacheKey(groupID), util.DumpToString(r), time.Duration(ttl)*time.Second)
}
