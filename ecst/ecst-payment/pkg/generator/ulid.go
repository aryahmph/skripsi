package generator

import "github.com/oklog/ulid/v2"

func GenerateString() string {
	return ulid.Make().String()
}
