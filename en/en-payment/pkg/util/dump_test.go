// Package utils
// @author Daud Valentino
package util

import (
	"bytes"
	"strings"
	"testing"
)

func TestDumpToString(t *testing.T) {
	t.Parallel()
	if result := DumpToString("test"); result == "test" {
		t.Logf("expected: %v, got: %v", "test", result)
	} else {
		t.Fatalf("expected: %v, got: %v", "test", result)
	}

	if result := DumpToString(1); strings.Trim(result, "\n") == "1" {
		t.Logf("expected: %v, got: %s", "1", result)
	} else {
		t.Fatalf("expected: %v, got: %v", "1", result)
	}
}

func TestDebugPrint(t *testing.T) {
	DebugPrint("test print")
}

func TestToJSONByte(t *testing.T) {
	testCase := []struct {
		Input    interface{}
		Expected []byte
	}{
		{
			Input: struct {
				ID   uint64 `json:"id"`
				Name string `json:"name"`
			}{
				ID:   1,
				Name: "testcase1",
			},
			Expected: []byte{123, 34, 105, 100, 34, 58, 49, 44, 34, 110, 97, 109, 101, 34, 58, 34, 116, 101, 115, 116, 99, 97, 115, 101, 49, 34, 125, 10},
		},
		{
			Input:    "test",
			Expected: []byte{116, 101, 115, 116},
		},

		{
			Input:    []byte{116, 101, 115, 116},
			Expected: []byte{116, 101, 115, 116},
		},
	}

	for _, v := range testCase {

		r := ToJSONByte(v.Input)
		if string(v.Expected) == string(r) {
			t.Logf("expected %v, got %v", v.Expected, r)
		} else {
			t.Errorf("expected %v, got %v", v.Expected, r)
		}
	}
}

func TestToJSONByteCompact(t *testing.T) {
	t.Parallel()

	t.Log("test success")
	{
		b := []byte(`123`)
		expected := []byte{49, 50, 51}

		o := ToJSONByteCompact(b)

		if bytes.Compare(expected, o) != 0 {
			t.Errorf("expected %b, got %b", expected, o)
		}
	}

	t.Log("test error")
	{
		b := []byte(``)

		o := ToJSONByteCompact(b)

		if len(o) == 0 {
			t.Logf("expected [], got %v", o)
		}

	}
}
