package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"en-order/internal/entity"

	"en-order/pkg/postgres"
	"en-order/pkg/tracer"
	"en-order/pkg/util"
)

type orderRepository struct {
	db postgres.Adapter
}

func NewOrderRepository(db postgres.Adapter) OrderRepository {
	return &orderRepository{db: db}
}

func (repo orderRepository) InsertOrder(ctx context.Context, data entity.Order) (isExist bool, err error) {
	ctx = tracer.SpanStart(ctx, "repository.order.InsertOrder")
	defer tracer.SpanFinish(ctx)

	// Prepare insert query
	cols, vals, err := util.ToColumnsValues(data, "db")
	if err != nil {
		return isExist, err
	}

	q := "INSERT INTO orders (%s) VALUES (%s)"
	q = fmt.Sprintf(q, strings.Join(cols, ","), util.ToInsertPostgresValues(vals))

	// Insert order
	_, err = repo.db.Exec(ctx, q, vals...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return true, nil
		}
		tracer.SpanError(ctx, err)
		return false, err
	}

	return false, err
}

func (repo orderRepository) FindOneOrder(ctx context.Context, cri FindOneOrderCriteria) (*entity.Order, error) {
	ctx = tracer.SpanStart(ctx, "repository.order.FindOneOrder")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				id,
				ticket_id,
				user_id,
				COALESCE(payment_id, '') payment_id,
				status,
				amount,
				created_at,
				updated_at
		FROM
		    	orders
		WHERE
		    	deleted_at IS NULL
				%s
	`
	q, args := cri.apply(q)

	//fmt.Println(q, args)

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

func (repo orderRepository) UpdateOrderStatus(ctx context.Context, id, status, paymentID string) (isExist bool, err error) {
	ctx = tracer.SpanStart(ctx, "repository.order.UpdateOrderStatus")
	defer tracer.SpanFinish(ctx)

	q1 := "SELECT id FROM orders WHERE id = $1 AND status = 'created' FOR UPDATE"
	q2 := "UPDATE orders SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1"

	if paymentID != "" {
		q2 = "UPDATE orders SET status = $2, payment_id = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
	}

	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		tracer.SpanError(ctx, err)
		return false, err
	}

	var rid string
	err = tx.QueryRowContext(ctx, q1, id).Scan(&rid)
	if err == sql.ErrNoRows {
		tx.Commit()
		return false, nil
	}
	if err != nil {
		tx.Rollback()
		tracer.SpanError(ctx, err)
		return false, err
	}

	if paymentID != "" {
		_, err = tx.ExecContext(ctx, q2, id, status, paymentID)
		if err != nil {
			tx.Rollback()
			tracer.SpanError(ctx, err)
			return false, err
		}
	} else {
		_, err = tx.ExecContext(ctx, q2, id, status)
		if err != nil {
			tx.Rollback()
			tracer.SpanError(ctx, err)
			return false, err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		tracer.SpanError(ctx, err)
		return false, err
	}

	return true, nil
}
