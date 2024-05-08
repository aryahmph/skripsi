// Package util
package util

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

const (
	letterBytes   = "0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits

	dateFormat     = `2006-01-02`
	dateTimeFormat = `2006-01-02 15:04:05`
)

// GenerateRandomNumberString generate random string number
func GenerateRandomNumberString(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}


// GenerateReferenceID generates reference ID
func GenerateReferenceID(prefix string) string {
	now := time.Now().Format("20060102030405")
	buff := bytes.NewBufferString(now)
	buff.WriteString(GenerateRandomNumberString(8))

	return fmt.Sprintf("%s%s", prefix, buff.String())
}

// GenerateAppID generates reference ID
func GenerateAppID(prefix string) string {
	now := time.Now().Format("20060102030405")
	buff := bytes.NewBufferString(now)
	buff.WriteString(GenerateRandomNumberString(6))

	return fmt.Sprintf("%s%s", prefix, buff.String())
}
