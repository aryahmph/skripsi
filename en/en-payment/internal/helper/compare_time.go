package helper

import (
	"fmt"
	"time"
	
	"en-payment/internal/consts"
)

// IsValidExecutionTime validate execution time
func IsValidExecutionTime(eet time.Duration, createdTime, nowTime string) (bool, error) {
	it, err := time.Parse(consts.LayoutDateTimeFormat, createdTime)

	if err != nil {
		return false, fmt.Errorf("initiated time error %v ", err)
	}

	expire := it.Add(eet)

	nt, err := time.Parse(consts.LayoutDateTimeFormat, nowTime)
	if err != nil {
		return false, fmt.Errorf("time now error %v ", err)
	}

	if nt.Unix() >= expire.Unix() {
		return false, nil
	}

	return true, nil
}
