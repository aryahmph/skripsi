// Package util
// @author Daud Valentino
package util

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ToMap converts a struct to a map using the struct's tags.
//
// ToMap uses tags on struct fields to decide which fields to add to the
// returned map.
func StructToMap(src interface{}, tag string) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return out, fmt.Errorf("only accepted %s, got %s", reflect.Struct.String(), v.Kind().String())
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)

		tagsv := strings.Split(fi.Tag.Get(tag), ",")

		if tagsv[0] != "" && fi.PkgPath == "" {
			// skip if omitempty
			if (len(tagsv) > 1 && tagsv[1] == "omitempty") && IsEmptyValue(v.Field(i).Interface()) {
				continue
			}

			if isTime(v.Field(i)) {
				if v.Field(i).Interface().(time.Time).IsZero() && tagsv[1] == "omitempty" {
					continue
				}
			}

			// set key value of map interface output
			out[tagsv[0]] = v.Field(i).Interface()

		}
	}

	return out, nil
}

// ToColumnsValues iterate struct to separate key field and value
func ToColumnsValues(src interface{}, tag string) ([]string, []interface{}, error) {
	var columns []string
	var values []interface{}

	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("only accepted %s, got %s", reflect.Struct.String(), v.Kind().String())
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)

		tagsv := strings.Split(fi.Tag.Get(tag), ",")

		if tagsv[0] != "" && fi.PkgPath == "" {
			// skip if omitempty
			if (len(tagsv) > 1 && tagsv[1] == "omitempty") && IsEmptyValue(v.Field(i).Interface()) {
				continue
			}

			if isTime(v.Field(i)) {
				if v.Field(i).Interface().(time.Time).IsZero() && tagsv[1] == "omitempty" {
					continue
				}
			}

			// set value of string slice to value in struct field
			columns = append(columns, tagsv[0])

			// set value interface of value struct field
			values = append(values, v.Field(i).Interface())

		}
	}

	return columns, values, nil
}

func isTime(obj reflect.Value) bool {
	_, ok := obj.Interface().(time.Time)
	return ok
}

func SliceStructToBulkInsert(src interface{}, tag string) ([]string, []interface{}, []string, error) {
	var columns []string
	var replacers []string
	var values []interface{}

	v := reflect.Indirect(reflect.ValueOf(src))
	t := reflect.TypeOf(src)
	if t.Kind() == reflect.Ptr {
		//v = v.Elem()
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return columns, values, replacers, fmt.Errorf("only accepted %s, got %s", reflect.Struct.String(), t.Kind().String())
	}

	for i := 0; i < v.Len(); i++ {

		item := v.Index(i)
		if !item.IsValid() {
			continue
		}

		cols, val, err := ToColumnsValues(item.Interface(), tag)
		if err != nil {
			return columns, values, replacers, err
		}

		if len(columns) == 0 {
			columns = cols
		}

		pattern := fmt.Sprintf(`(%s)`, strings.TrimRight(strings.Repeat("?,", len(columns)), `,`))
		replacers = append(replacers, pattern)
		values = append(values, val...)
	}

	return columns, values, replacers, nil

}

func ToUpdatePostgresColumn(cols []string, start int, updateTimestamp bool) string {
	var sb strings.Builder

	length := len(cols)
	for i, v := range cols {
		if i != length-1 {
			sb.WriteString(fmt.Sprintf("%s=$%d, ", v, start))
		} else {
			sb.WriteString(fmt.Sprintf("%s=$%d", v, start))
		}
		start++
	}

	if updateTimestamp {
		sb.WriteString(", updated_at=CURRENT_TIMESTAMP")
	}

	return sb.String()
}

func ToInsertPostgresValues(cols []string) []string {
	colArgs := make([]string, len(cols))
	length := len(cols)
	for i, v := range cols {
		if i != length-1 {
			colArgs[i] = fmt.Sprintf("%s, ", v)
		} else {
			colArgs[i] = fmt.Sprintf("%s", v)
		}
	}

	return colArgs
}
