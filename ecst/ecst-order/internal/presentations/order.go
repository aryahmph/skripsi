package presentations

type (
	CreateOrderRequest struct {
		TicketID string `json:"ticket_id"`
	}

	InternalGetOrderRequest struct {
		UserID string `url:"user_id,omitempty"`
		Status string `url:"status,omitempty"`
	}
)

type (
	CreateOrderResponse struct {
		ID string `json:"id"`
	}

	GetOrderResponse struct {
		ID        string `json:"id"`
		TicketID  string `json:"ticket_id"`
		UserID    string `json:"user_id"`
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
)
