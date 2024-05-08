package worker

import (
	"context"
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/bootstrap"
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/entity"
	"ecst-ticket/internal/presentations"
	"ecst-ticket/internal/repositories"
	"ecst-ticket/pkg/generator"
	"ecst-ticket/pkg/kafka"
	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"sync"
	"time"
)

func RunWorkerCreateTickets(ctx context.Context, args []string) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Load config error %v", err), logger.EventName("InitiateConfig"))
	}

	bootstrap.RegistryLogger(cfg)

	db := bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	kp := bootstrap.RegistryKafkaProducer(cfg)

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

	var wg sync.WaitGroup
	messageCh := make(chan presentations.KafkaMessageBase, len(tickets))

	for _, ticket := range tickets {
		wg.Add(1)
		go func(cfg *appctx.Config, kp kafka.Producer, ticket entity.Ticket) {
			defer wg.Done()

			message := presentations.KafkaMessageBase{
				Source: presentations.KafkaMessageBaseSource{
					AppName: cfg.App.AppName,
					AppEnv:  cfg.App.Env,
				},
				Check: &presentations.KafkaMessageChecker{
					InitiateTime: time.Now().Format(consts.LayoutDateTimeFormat),
					ServiceOrigin: presentations.KafkaMessageOriginService{
						ServiceName: cfg.App.AppName,
						TargetTopic: cfg.KafkaTopics.TopicCreateTicket,
					},
					Count:      0,
					NextSecond: cast.ToUint(cfg.KafkaNextSecond.NextCreateTicket),
					MaxSecond:  cast.ToUint(cfg.KafkaEETSecond.EETCreateTicket),
				},
				CreatedAt: time.Now().Format(consts.LayoutDateTimeFormat),
				Payload: &presentations.KafkaMessageCreateTicketPayload{
					ID:       ticket.ID,
					Code:     ticket.Code,
					Category: ticket.Category,
					Price:    ticket.Price,
					Version:  1,
				},
			}

			messageCh <- message
		}(cfg, kp, ticket)
	}

	go func() {
		wg.Wait()
		close(messageCh)
	}()

	for message := range messageCh {
		wg.Add(1)
		go func(cfg *appctx.Config, kp kafka.Producer, message presentations.KafkaMessageBase) {
			defer wg.Done()

			km := kafka.MessageContext{
				Value:   util.DumpToString(message),
				Topic:   cfg.KafkaTopics.TopicCreateTicket,
				Verbose: true,
			}

			err := kp.Publish(ctx, &km)
			if err != nil {
				logger.ErrorWithContext(ctx, logger.SetMessageFormat(consts.LogMessageFailedPublishMessage, err))
			}
		}(cfg, kp, message)
	}

	wg.Wait()

	logger.InfoWithContext(ctx, logger.SetMessageFormat(consts.LogMessageSuccess, consts.LogEventNameCreateTickets))
}
