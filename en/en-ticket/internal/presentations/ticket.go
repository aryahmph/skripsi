package presentations

type (
	CreateTicketsRequest struct {
		TicketGroupID string `json:"ticket_group_id"`
		Category      string `json:"category"`
		Price         int64  `json:"price"`
		Quantity      int64  `json:"quantity"`
	}
)

type (
	GetTicketResponse struct {
		ID            string `json:"id"`
		TicketGroupID string `json:"ticket_group_id"`
		OrderID       string `json:"order_id"`
		Code          string `json:"code"`
		Category      string `json:"category"`
		Price         int64  `json:"price"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}
)

type (
	ListTicketCategoriesResponse struct {
		Category string `json:"category"`
		Price    int64  `json:"price"`
		Total    int64  `json:"total"`
	}
)

type (
	ListUnreservedTicketsRequest struct {
		Category string `url:"category,omitmepty"`
	}
	
	ListUnreservedTicketsResponse struct {
		ID string `json:"id"`
	}
)
