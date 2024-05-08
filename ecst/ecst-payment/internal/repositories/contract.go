package repositories

import (
	"context"
	"ecst-payment/internal/entity"
	"fmt"
	"strings"
)

type (
	PaymentRepository interface {
		InsertPayment(ctx context.Context, data entity.Payment) error
		FindOnePayment(ctx context.Context, criteria FindOnePaymentCriteria) (*entity.Payment, error)
	}

	OrderRepository interface {
		InsertOrder(ctx context.Context, data entity.Order) error
		FindOneOrder(ctx context.Context, cri FindOneOrderCriteria) (*entity.Order, error)
		UpdateOrder(ctx context.Context, data entity.Order) error
	}
)

type (
	FindOnePaymentCriteria struct {
		ID      string
		OrderID string
	}

	FindOneOrderCriteria struct {
		ID      string
		Status  string
		Version int32
	}
)

func (cri FindOnePaymentCriteria) apply(q string) (query string, args []interface{}) {
	var additionalQry strings.Builder
	args = make([]interface{}, 0)
	counter := 1

	if cri.ID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND id = $%d", counter))
		counter++
		args = append(args, cri.ID)
	}

	if cri.OrderID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND order_id = $%d", counter))
		counter++
		args = append(args, cri.OrderID)
	}

	return fmt.Sprintf(q, additionalQry.String()), args
}

func (cri FindOneOrderCriteria) apply(q string) (query string, args []interface{}) {
	var additionalQry strings.Builder
	args = make([]interface{}, 0)
	counter := 1

	if cri.ID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND id = $%d", counter))
		counter++
		args = append(args, cri.ID)
	}

	if cri.Status != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND status = $%d", counter))
		counter++
		args = append(args, cri.Status)
	}

	if cri.Version != 0 {
		additionalQry.WriteString(fmt.Sprintf(" AND version = $%d", counter))
		counter++
		args = append(args, cri.Version)
	}

	return fmt.Sprintf(q, additionalQry.String()), args
}
