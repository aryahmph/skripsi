package repositories

import (
	"context"
	"database/sql"
	"ecst-order/internal/entity"
	"ecst-order/pkg/postgres"
	"ecst-order/pkg/tracer"
	"ecst-order/pkg/util"
	"fmt"
	"strings"
)

type ticketRepository struct {
	db postgres.Adapter
}

func NewTicketRepository(db postgres.Adapter) TicketRepository {
	return &ticketRepository{db: db}
}

func (repo *ticketRepository) InsertTicket(ctx context.Context, data entity.Ticket) error {
	ctx = tracer.SpanStart(ctx, "repository.ticket.InsertTicket")
	defer tracer.SpanFinish(ctx)

	q := "INSERT INTO tickets (%s) VALUES (%s)"
	cols, vals, err := util.ToColumnsValues(data, "db")
	if err != nil {
		return err
	}
	q = fmt.Sprintf(q, strings.Join(cols, ","), util.ToInsertPostgresValues(vals))

	_, err = repo.db.Exec(ctx, q, vals...)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	return nil
}

func (repo *ticketRepository) FindOneTicket(ctx context.Context, cri FindOneTicketCriteria) (*entity.Ticket, error) {
	ctx = tracer.SpanStart(ctx, "repository.ticket.FindOneTicket")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				id,
				code,
				category,
				price,
				version,
				created_at,
				updated_at
		FROM
		    	tickets
		WHERE
		    	deleted_at IS NULL
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

func (repo *ticketRepository) UpdateTicket(ctx context.Context, data entity.Ticket) error {
	ctx = tracer.SpanStart(ctx, "repository.ticket.UpdateTicket")
	defer tracer.SpanFinish(ctx)

	q := `
		UPDATE tickets
		SET
			code = $1,
			category = $2,
			price = $3,
			version = $4
		WHERE deleted_at IS NULL AND id = $5
	`
	_, err := repo.db.Exec(ctx, q, data.Code, data.Category, data.Price, data.Version, data.ID)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	return nil
}
