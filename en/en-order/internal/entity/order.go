package entity

import "time"

var (
	TableNameOrder = "orders"
)

type (
	Order struct {
		ID        string    `db:"id" qb:"id,omitempty"`
		TicketID  string    `db:"ticket_id" qb:"ticket_id,omitempty"`
		UserID    string    `db:"user_id" qb:"user_id,omitempty"`
		PaymentID string    `db:"payment_id,omitempty" qb:"payment_id,omitempty"`
		Status    string    `db:"status" qb:"status,omitempty"`
		Amount    int64     `db:"amount" qb:"amount,omitempty"`
		CreatedAt time.Time `db:"created_at,omitempty" qb:"created_at,omitempty"`
		UpdatedAt time.Time `db:"updated_at,omitempty" qb:"updated_at,omitempty"`
		DeletedAt time.Time `db:"deleted_at,omitempty" qb:"deleted_at,omitempty"`
	}
)
