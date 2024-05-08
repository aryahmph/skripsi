package helper

import (
	"en-ticket/internal/consts"
	"fmt"
)

func TicketCacheKey(id string) string {
	return fmt.Sprintf("%s:ticket:id:%s", consts.CacheKeyTicketService, id)
}

func ListTicketCategoriesCacheKey(groupID string) string {
	return fmt.Sprintf("%s:list_ticket_categories:group_id:%s", consts.CacheKeyTicketService, groupID)
}
func ListUnreservedTicketsCacheKey(groupID, category string) string {
	return fmt.Sprintf("%s:list_unreserved_tickets:group_id:%s:category:%s", consts.CacheKeyTicketService, groupID, category)
}
