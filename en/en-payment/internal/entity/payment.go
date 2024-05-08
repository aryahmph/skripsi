package entity

import "time"

var (
	TableNamePayment = "payments"
)

type (
	Payment struct {
		ID        string    `db:"id" qb:"id,omitempty"`
		OrderID   string    `db:"order_id" qb:"order_id,omitempty"`
		CreatedAt time.Time `db:"created_at,omitempty" qb:"created_at,omitempty"`
	}
)
