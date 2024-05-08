package worker

import (
	"context"
	"en-ticket/internal/appctx"
	"en-ticket/internal/bootstrap"
	"en-ticket/internal/consts"
	"en-ticket/internal/entity"
	"en-ticket/internal/presentations"
	"en-ticket/internal/repositories"
	"en-ticket/pkg/generator"
	"en-ticket/pkg/postgres"
	"encoding/json"
	"fmt"

	"en-ticket/pkg/logger"
)

func RunWorkerCreateTickets(ctx context.Context, args []string) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Load config error %v", err), logger.EventName("InitiateConfig"))
	}

	bootstrap.RegistryLogger(cfg)
	var db postgres.Adapter

	if cfg.App.IsSingle {
		db = bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	} else {
		db = bootstrap.RegistryPostgresDB(cfg.DBWrite, cfg.DBRead, cfg.App.Timezone)
	}

	ticketRepo := repositories.NewTicketRepository(db)

	dataJson := args[0]
	var data presentations.CreateTicketsRequest
	err = json.Unmarshal([]byte(dataJson), &data)
	if err != nil {
		logger.Fatal(err,
			logger.Any("data", dataJson),
			logger.EventName(consts.LogEventNameCreateTickets),
		)
	}

	tg, err := ticketRepo.FindOneTicketGroup(ctx, data.TicketGroupID)
	if err != nil {
		logger.Fatal(err,
			logger.Any("data", dataJson),
			logger.EventName(consts.LogEventNameCreateTickets),
		)
		logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedFetchDB, entity.TableNameTicketGroup, err))
		return
	}

	if tg == nil {
		logger.WarnWithContext(ctx, logger.SetMessageFormat(consts.LogMessageDBNotFound, entity.TableNameTicketGroup))
		return
	}

	tickets := make([]entity.Ticket, data.Quantity)
	for i := int64(0); i < data.Quantity; i++ {
		id := generator.GenerateString()
		tickets[i] = entity.Ticket{
			ID:            id,
			TicketGroupID: data.TicketGroupID,
			Category:      data.Category,
			Price:         data.Price,
			Code:          fmt.Sprintf("%s-%d", data.Category, i+1),
		}

	}

	for i := 0; i < len(tickets); i += 5000 {
		end := i + 5000
		if end > len(tickets) {
			end = len(tickets)
		}

		x := tickets[i:end]
		err = ticketRepo.BulkInsertTicket(ctx, &x)
		if err != nil {
			logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedInsertDB, entity.TableNameTicket, err))
			return
		}
	}

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreateTickets))
}
