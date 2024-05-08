package presentations

type (
	CreatePaymentRequest struct {
		OrderID    string `json:"order_id"`
		CardNumber string `json:"card_number"`
		ExpMonth   string `json:"exp_month"`
		ExpYear    string `json:"exp_year"`
		CVV        string `json:"cvv"`
	}

	CreatePaymentResponse struct {
		ID string `json:"id"`
	}

	GetPaymentRequest struct {
		UserID string `url:"user_id,omitempty"`
	}

	GetPaymentResponse struct {
		ID        string `json:"id"`
		OrderID   string `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}
)
