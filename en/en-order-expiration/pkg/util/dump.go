package util

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DumpToString interface to string
func DumpToString(v interface{}) string {
	str, ok := v.(string)
	if !ok {
		buff := &bytes.Buffer{}
		json.NewEncoder(buff).Encode(v)
		return buff.String()
	}

	return str
}

// DebugPrint for debug print on terminal
func DebugPrint(v ...interface{}) {
	for _, x := range v {
		fmt.Println(DumpToString(x))
	}
}

// Data To json Bytes
func ToJSONByte(v interface{}) []byte {

	switch v.(type) {
	case []byte:
		return v.([]byte)
	case string:
		return []byte(v.(string))
	}

	buff := &bytes.Buffer{}
	json.NewEncoder(buff).Encode(v)
	return buff.Bytes()
}

// ToJSONByteCompact for insignificant space characters elided
func ToJSONByteCompact(v interface{}) []byte {
	buff := bytes.Buffer{}
	b := ToJSONByte(v)

	json.Compact(&buff, b)
	return buff.Bytes()
}
