package dto

import (
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/entity"
	"ecst-ticket/internal/presentations"
)

func ConstructGetTicketResponse(data *entity.Ticket) presentations.GetTicketResponse {
	return presentations.GetTicketResponse{
		ID:            data.ID,
		TicketGroupID: data.TicketGroupID,
		OrderID:       data.OrderID,
		Code:          data.Code,
		Category:      data.Category,
		Price:         data.Price,
		CreatedAt:     data.CreatedAt.Format(consts.LayoutDateTimeFormat),
		UpdatedAt:     data.UpdatedAt.Format(consts.LayoutDateTimeFormat),
	}
}

func ConstructListTicketCategoriesResponse(data []entity.TicketCategory) []presentations.ListTicketCategoriesResponse {
	result := make([]presentations.ListTicketCategoriesResponse, len(data))
	for i, v := range data {
		result[i] = presentations.ListTicketCategoriesResponse{
			Category: v.Category,
			Price:    v.Price,
			Total:    v.Total,
		}
	}
	return result
}

func ConstructListUnreservedTicketsResponse(data []string) []presentations.ListUnreservedTicketsResponse {
	result := make([]presentations.ListUnreservedTicketsResponse, len(data))
	for i, v := range data {
		result[i] = presentations.ListUnreservedTicketsResponse{
			ID: v,
		}
	}
	return result
}
