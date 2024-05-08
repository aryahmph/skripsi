package presentations

type ListTicketsCache struct {
	Tickets []ListTicketData `json:"tickets"`
}

type ListTicketData struct {
	ID            string `json:"id"`
	TicketGroupID string `json:"ticket_group_id"`
	Code          string `json:"code"`
	Category      string `json:"category"`
	Price         int64  `json:"price"`
}
