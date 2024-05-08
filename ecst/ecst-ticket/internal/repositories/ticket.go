package repositories

import (
	"context"
	"database/sql"
	"ecst-ticket/internal/entity"
	"ecst-ticket/pkg/postgres"
	"ecst-ticket/pkg/tracer"
	"ecst-ticket/pkg/util"
	"fmt"
	"strings"
)

type ticketRepository struct {
	db postgres.Adapter
}

func NewTicketRepository(db postgres.Adapter) TicketRepository {
	return &ticketRepository{db: db}
}

func (repo *ticketRepository) BulkInsertTicket(ctx context.Context, data *[]entity.Ticket) error {
	ctx = tracer.SpanStart(ctx, "repository.ticket.BulkInsertTicket")
	defer tracer.SpanFinish(ctx)

	q := "INSERT INTO tickets (id, ticket_group_id, code, category, price) VALUES %s"
	var valArgs []interface{}
	var colArgs []string

	counter := 1
	for _, v := range *data {
		colArgs = append(colArgs, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", counter, counter+1, counter+2, counter+3, counter+4))

		valArgs = append(valArgs, v.ID)
		valArgs = append(valArgs, v.TicketGroupID)
		valArgs = append(valArgs, v.Code)
		valArgs = append(valArgs, v.Category)
		valArgs = append(valArgs, v.Price)

		counter += 5
	}

	q = fmt.Sprintf(q, strings.Join(colArgs, ","))

	_, err := repo.db.Exec(ctx, q, valArgs...)
	return err
}

func (repo *ticketRepository) FindOneTicket(ctx context.Context, cri FindOneTicketCriteria) (*entity.Ticket, error) {
	ctx = tracer.SpanStart(ctx, "repository.ticket.FindOneTicket")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				t.id,
				ticket_group_id,
				COALESCE(order_id, '') AS order_id,
				code,
				category,
				price,
				version,
				t.created_at,
				t.updated_at
		FROM
		    	tickets t
		JOIN ticket_groups tg on tg.id = t.ticket_group_id
		WHERE
		    	t.deleted_at IS NULL
				AND tg.deleted_at IS NULL
				%s
	`
	q, args := cri.apply(q)

	var ticket entity.Ticket
	err := repo.db.FetchRow(ctx, &ticket, q, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return &ticket, nil
}

func (repo *ticketRepository) FindOneTicketGroup(ctx context.Context, id string) (*entity.TicketGroup, error) {
	ctx = tracer.SpanStart(ctx, "repository.ticket.FindOneTicketGroup")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				id,
				name,
				created_at,
				updated_at
		FROM
		    	ticket_groups
		WHERE
		    	deleted_at IS NULL
				AND id = $1
	`

	var ticketGroup entity.TicketGroup
	err := repo.db.FetchRow(ctx, &ticketGroup, q, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return &ticketGroup, nil
}

func (repo *ticketRepository) UpdateTicket(ctx context.Context, data entity.Ticket) error {
	ctx = tracer.SpanStart(ctx, "repository.ticket.UpdateTicket")
	defer tracer.SpanFinish(ctx)

	q := "UPDATE tickets SET %s WHERE id = $1"

	cols, vals, err := util.ToColumnsValues(data, "qb")
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	vals = append([]interface{}{data.ID}, vals...)

	q = fmt.Sprintf(q, util.ToUpdatePostgresColumn(cols, 2, true))

	_, err = repo.db.Exec(ctx, q, vals...)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	return nil
}

func (repo *ticketRepository) RemoveTicketOrderID(ctx context.Context, ticketID string) error {
	ctx = tracer.SpanStart(ctx, "repository.ticket.RemoveTicketOrderID")
	defer tracer.SpanFinish(ctx)

	q := "UPDATE tickets SET order_id = NULL, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1"

	_, err := repo.db.Exec(ctx, q, ticketID)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	return nil
}

func (repo *ticketRepository) ListTicketCategories(ctx context.Context, groupId string) ([]entity.TicketCategory, error) {
	ctx = tracer.SpanStart(ctx, "repository.ticket.ListTicketCategories")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT t.category, t.price, COUNT(*) total
		FROM tickets t
		WHERE deleted_at IS NULL
		  AND t.ticket_group_id = $1
		  AND t.order_id IS NULL
		GROUP BY t.category, t.price;
	`

	var ticketCategories []entity.TicketCategory
	err := repo.db.Fetch(ctx, &ticketCategories, q, groupId)
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return ticketCategories, nil
}

func (repo *ticketRepository) ListUnreservedTickets(ctx context.Context, groupId, category string) ([]string, error) {
	ctx = tracer.SpanStart(ctx, "repository.ticket.ListUnreservedTickets")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT id
		FROM tickets
		WHERE deleted_at IS NULL
		  AND category = $1
		  AND ticket_group_id = $2
		  AND order_id IS NULL
		LIMIT 400;
	`

	var ticketIDs []string
	err := repo.db.Fetch(ctx, &ticketIDs, q, category, groupId)
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return ticketIDs, nil
}
