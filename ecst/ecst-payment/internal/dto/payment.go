package dto

import (
	"ecst-payment/internal/consts"
	"ecst-payment/internal/entity"
	"ecst-payment/internal/presentations"
)

func ConstructGetPaymentResponse(data *entity.Payment) presentations.GetPaymentResponse {
	return presentations.GetPaymentResponse{
		ID:        data.ID,
		OrderID:   data.OrderID,
		CreatedAt: data.CreatedAt.Format(consts.LayoutDateTimeFormat),
	}
}
