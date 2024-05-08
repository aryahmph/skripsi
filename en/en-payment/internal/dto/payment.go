package dto

import (
	"en-payment/internal/consts"
	"en-payment/internal/entity"
	"en-payment/internal/presentations"
)

func ConstructGetPaymentResponse(data *entity.Payment) presentations.GetPaymentResponse {
	return presentations.GetPaymentResponse{
		ID:        data.ID,
		OrderID:   data.OrderID,
		CreatedAt: data.CreatedAt.Format(consts.LayoutDateTimeFormat),
	}
}
