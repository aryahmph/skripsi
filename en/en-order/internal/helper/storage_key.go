package helper

import (
	"en-order/internal/consts"
	"fmt"
)

func OrderCacheKey(id string) string {
	return fmt.Sprintf("%s:order:id:%s", consts.CacheKeyOrderService, id)
}
