package entity

import "time"

var (
	TableNameTicket      = "tickets"
	TableNameTicketGroup = "ticket_groups"
)

type (
	TicketGroup struct {
		ID        string    `db:"id"`
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at,omitempty"`
		UpdatedAt time.Time `db:"updated_at,omitempty"`
		DeletedAt time.Time `db:"deleted_at,omitempty"`
	}

	Ticket struct {
		ID            string    `db:"id" qb:"id,omitempty"`
		TicketGroupID string    `db:"ticket_group_id" qb:"ticket_group_id,omitempty"`
		OrderID       string    `db:"order_id,omitempty" qb:"order_id,omitempty"`
		Code          string    `db:"code" qb:"code,omitempty"`
		Category      string    `db:"category" qb:"category,omitempty"`
		Price         int64     `db:"price" qb:"price,omitempty"`
		CreatedAt     time.Time `db:"created_at,omitempty" qb:"created_at,omitempty"`
		UpdatedAt     time.Time `db:"updated_at,omitempty" qb:"updated_at,omitempty"`
		DeletedAt     time.Time `db:"deleted_at,omitempty" qb:"deleted_at,omitempty"`
	}
)

type (
	TicketCategory struct {
		Category string `db:"category"`
		Price    int64  `db:"price"`
		Total    int64  `db:"total"`
	}
)
