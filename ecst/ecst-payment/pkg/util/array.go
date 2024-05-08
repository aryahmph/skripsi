// Package util
package util

import "reflect"

// InArray check if an element is exist in the array
func InArray(val interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				return true
			}
		}
	}
	return false
}

// InBetweenArray check if one of element in an array is exists in other array
func InBetweenArray(arrayComparator interface{}, arrayCompared interface{}) bool {
	if reflect.TypeOf(arrayComparator).Kind() != reflect.Slice {
		return false
	}

	s := reflect.ValueOf(arrayComparator)
	for i := 0; i < s.Len(); i++ {
		if InArray(s.Index(i).Interface(), arrayCompared) {
			return true
		}
	}

	return false
}
