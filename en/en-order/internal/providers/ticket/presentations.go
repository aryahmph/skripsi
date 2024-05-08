package ticket

type (
	BaseResponse struct {
		Name    string      `json:"name"`
		Message string      `json:"message"`
		Errors  interface{} `json:"errors,omitempty"`
	}

	GetTicketResponse struct {
		BaseResponse
		Data DataGetTicketResponse `json:"data"`
	}

	DataGetTicketResponse struct {
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
	GetTicketRequest struct {
		ID         string `json:"id"`
		ClientName string `json:"client_name"`
	}
)
