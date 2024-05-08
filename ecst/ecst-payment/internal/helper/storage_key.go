package helper

import (
	"ecst-payment/internal/consts"
	"fmt"
)

func PaymentCacheKey(id, userId string) string {
	return fmt.Sprintf("%s:payment:id:%s:user_id:%s", consts.CacheKeyPaymentService, id, userId)
}
