package entity

import "time"

var TableNameTicket = "tickets"

type Ticket struct {
	ID        string    `db:"id" qb:"id,omitempty"`
	Code      string    `db:"code"`
	Category  string    `db:"category"`
	Price     int64     `db:"price"`
	Version   int32     `db:"version"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`
	DeletedAt time.Time `db:"deleted_at,omitempty"`
}
