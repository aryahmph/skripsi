package dto

import (
	"en-order/internal/consts"
	"en-order/internal/entity"
	"en-order/internal/presentations"
)

func ConstructGetOrderResponse(data *entity.Order) presentations.GetOrderResponse {
	return presentations.GetOrderResponse{
		ID:        data.ID,
		TicketID:  data.TicketID,
		UserID:    data.UserID,
		PaymentID: data.PaymentID,
		Status:    data.Status,
		Amount:    data.Amount,
		CreatedAt: data.CreatedAt.Format(consts.LayoutDateTimeFormat),
		UpdatedAt: data.UpdatedAt.Format(consts.LayoutDateTimeFormat),
	}
}
