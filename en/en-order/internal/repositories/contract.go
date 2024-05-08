package repositories

import (
	"context"
	"en-order/internal/entity"
	"fmt"
	"strings"
)

type (
	OrderRepository interface {
		InsertOrder(ctx context.Context, data entity.Order) (bool, error)
		FindOneOrder(ctx context.Context, cri FindOneOrderCriteria) (*entity.Order, error)
		UpdateOrderStatus(ctx context.Context, id, status string, paymentID string) (isExist bool, err error)
	}
)

type (
	FindOneOrderCriteria struct {
		ID       string
		TicketID string
		UserID   string
		Status   string
		Statuses []string
	}
)

func (cri FindOneOrderCriteria) apply(q string) (query string, args []interface{}) {
	var additionalQry strings.Builder
	args = make([]interface{}, 0)
	counter := 1

	if cri.ID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND id = $%d", counter))
		counter++
		args = append(args, cri.ID)
	}

	if cri.TicketID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND ticket_id = $%d", counter))
		counter++
		args = append(args, cri.TicketID)
	}

	if cri.UserID != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND user_id = $%d", counter))
		counter++
		args = append(args, cri.UserID)
	}

	if cri.Status != "" {
		additionalQry.WriteString(fmt.Sprintf(" AND status = $%d", counter))
		counter++
		args = append(args, cri.Status)
	}

	if len(cri.Statuses) > 0 {
		var sb strings.Builder
		for _, s := range cri.Statuses {
			if sb.Len() > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("'%s'", s))
		}
		additionalQry.WriteString(fmt.Sprintf(" AND status IN (%s)", sb.String()))
	}

	return fmt.Sprintf(q, additionalQry.String()), args
}
