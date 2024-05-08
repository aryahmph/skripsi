package repositories

import (
	"context"
	"en-ticket/internal/entity"
	"fmt"
	"strings"
)

type (
	TicketRepository interface {
		BulkInsertTicket(ctx context.Context, data *[]entity.Ticket) error
		FindOneTicket(ctx context.Context, cri FindOneTicketCriteria) (*entity.Ticket, error)
		FindOneTicketGroup(ctx context.Context, id string) (*entity.TicketGroup, error)
		UpdateTicket(ctx context.Context, ticket entity.Ticket) error
		RemoveTicketOrderID(ctx context.Context, ticketID string) error
		ListTicketCategories(ctx context.Context, groupId string) ([]entity.TicketCategory, error)
		ListUnreservedTickets(ctx context.Context, groupId, category string) ([]string, error)
	}
)

type (
	FindOneTicketCriteria struct {
		ID      string
		OrderID string
		Code    string
		GroupID string
	}
)

func (cri FindOneTicketCriteria) apply(q string) (query string, args []interface{}) {
	var additionalQry strings.Builder
	counter := 1
	args = make([]interface{}, 0)

	if cri.ID != "" {
		additionalQry.WriteString(fmt.Sprintf("AND t.id = $%d ", counter))
		counter++
		args = append(args, cri.ID)
	}

	if cri.OrderID != "" {
		additionalQry.WriteString(fmt.Sprintf("AND t.order_id = $%d ", counter))
		counter++
		args = append(args, cri.OrderID)
	}

	if cri.Code != "" {
		additionalQry.WriteString(fmt.Sprintf("AND t.code = $%d ", counter))
		counter++
		args = append(args, cri.Code)
	}

	if cri.GroupID != "" {
		additionalQry.WriteString(fmt.Sprintf("AND t.ticket_group_id = $%d ", counter))
		counter++
		args = append(args, cri.GroupID)
	}

	return fmt.Sprintf(q, additionalQry.String()), args
}
