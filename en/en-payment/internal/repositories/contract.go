package repositories

import (
	"context"
	"en-payment/internal/entity"
	"fmt"
	"strings"
)

type (
	PaymentRepository interface {
		InsertPayment(ctx context.Context, data entity.Payment) error
		FindOnePayment(ctx context.Context, criteria FindOnePaymentCriteria) (*entity.Payment, error)
	}
)

type (
	FindOnePaymentCriteria struct {
		ID      string
		OrderID string
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
