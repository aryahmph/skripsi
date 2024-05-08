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

type paymentRepo struct {
	db postgres.Adapter
}

func NewPaymentRepo(db postgres.Adapter) PaymentRepository {
	return &paymentRepo{db: db}
}

func (repo *paymentRepo) FindOnePayment(ctx context.Context, cri FindOnePaymentCriteria) (*entity.Payment, error) {
	ctx = tracer.SpanStart(ctx, "repository.payment.FindOnePayment")
	defer tracer.SpanFinish(ctx)

	q := `
		SELECT
				id,
				order_id,
				created_at
		FROM
		    	payments
		WHERE
		    	deleted_at IS NULL
				%s
	`
	q, args := cri.apply(q)

	var payment entity.Payment
	err := repo.db.FetchRow(ctx, &payment, q, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		tracer.SpanError(ctx, err)
		return nil, err
	}

	return &payment, nil
}

func (repo *paymentRepo) InsertPayment(ctx context.Context, data entity.Payment) error {
	ctx = tracer.SpanStart(ctx, "repository.payment.InsertPayment")
	defer tracer.SpanFinish(ctx)

	cols, vals, err := util.ToColumnsValues(data, "db")
	if err != nil {
		return err
	}

	q := "INSERT INTO payments (%s) VALUES (%s)"
	q = fmt.Sprintf(q, strings.Join(cols, ","), util.ToInsertPostgresValues(vals))

	_, err = repo.db.Exec(ctx, q, vals...)
	if err != nil {
		tracer.SpanError(ctx, err)
		return err
	}
	return nil
}
