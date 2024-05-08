package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"ecst-payment/internal/entity"

	"ecst-payment/pkg/postgres"
	"ecst-payment/pkg/tracer"
	"ecst-payment/pkg/util"
)

type orderRepository struct {
	db postgres.Adapter
}

func NewOrderRepository(db postgres.Adapter) OrderRepository {
	return &orderRepository{db: db}
}

func (repo *orderRepository) InsertOrder(ctx context.Context, data entity.Order) error {
	ctx = tracer.SpanStart(ctx, "repository.order.InsertOrder")
	defer tracer.SpanFinish(ctx)

	cols, vals, err := util.ToColumnsValues(data, "db")
	if err != nil {
		return err
	}

	q := "INSERT INTO orders (%s) VALUES (%s)"
	q = fmt.Sprintf(q, strings.Join(cols, ","), util.ToInsertPostgresValues(vals))

	_, err = repo.db.Exec(ctx, q, vals...)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}
	return nil
}

func (repo *orderRepository) FindOneOrder(ctx context.Context, cri FindOneOrderCriteria) (*entity.Order, error) {
	ctx = tracer.SpanStart(ctx, "repository.order.FindOneOrder")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				id,
				status,
				amount,
				version,
				created_at,
				updated_at
		FROM
		    	orders
		WHERE
		    	deleted_at IS NULL
				%s
	`

	q, args := cri.apply(q)

	var order entity.Order
	err := repo.db.FetchRow(ctx, &order, q, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return &order, nil
}

func (repo *orderRepository) UpdateOrder(ctx context.Context, data entity.Order) error {
	ctx = tracer.SpanStart(ctx, "repository.order.UpdateOrder")
	defer tracer.SpanFinish(ctx)

	q := `
		UPDATE
			orders
		SET
			status = $1,
			version = version + 1
		WHERE
		    deleted_at IS NULL AND
			id = $2 AND
			version = $3
	`

	_, err := repo.db.Exec(ctx, q, data.Status, data.ID, data.Version-1)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}

	return nil
}
