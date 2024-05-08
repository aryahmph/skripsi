// Package util
package util

import (
	"fmt"
	"strings"
)

func StringJoin(elems []string, sep, lastSep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%s%s",elems[0], lastSep)
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0])
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(s)
	}

	if lastSep != "" {
		b.WriteString(lastSep)
	}

	return b.String()
}



// StringContains check contain string
func StringContains(s string, contains []string) bool  {
	for i:=0; i< len(contains);i++ {
		if strings.Contains(strings.ToLower(s), strings.ToLower(contains[i])) {
			return true
		}
	}
	return false
}


// SubString substitute string
func SubString(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

