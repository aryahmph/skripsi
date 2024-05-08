package order

type (
	BaseResponse struct {
		Name    string      `json:"name"`
		Message string      `json:"message"`
		Errors  interface{} `json:"errors,omitempty"`
	}

	GetOrderResponse struct {
		BaseResponse
		Data DataGetOrderResponse `json:"data"`
	}

	DataGetOrderResponse struct {
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

type (
	GetOrderRequest struct {
		ID         string `json:"id"`
		UserID     string `json:"user_id"`
		Status     string `json:"status"`
		ClientName string `json:"client_name"`
	}
)
