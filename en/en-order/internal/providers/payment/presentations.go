package payment

type (
	BaseResponse struct {
		Name    string      `json:"name"`
		Message string      `json:"message"`
		Errors  interface{} `json:"errors,omitempty"`
	}

	GetPaymentResponse struct {
		BaseResponse
		Data DataGetPaymentResponse `json:"data"`
	}

	DataGetPaymentResponse struct {
		ID        string `json:"id"`
		OrderID   string `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}
)

type (
	GetPaymentRequest struct {
		ID         string `json:"id"`
		UserID     string `json:"user_id"`
		ClientName string `json:"client_name"`
	}
)
