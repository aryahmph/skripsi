// Package util
package util

import "testing"

func TestInArray(t *testing.T) {
	t.Parallel()
	testCase := []struct {
		Case       int
		Value      interface{}
		Collection interface{}
		Valid      bool
	}{
		{
			Case:  1,
			Value: "test",
			Collection: []string{
				"test",
				"test123",
			},
			Valid: true,
		},
		{
			Case:  2,
			Value: 2,
			Collection: []string{
				"test",
				"test123",
			},
			Valid: false,
		},
		{
			Case:       3,
			Value:      int64(3),
			Collection: []int64{1, 2, 3},
			Valid:      true,
		},

		{
			Case:  4,
			Value: "case-4",
			Collection: map[string]string{
				"case-4": "case-4",
			},
			Valid: false,
		},
	}

	for _, v := range testCase {

		valid := InArray(v.Value, v.Collection)

		if valid == v.Valid {
			t.Logf("scenario #%v exptected %v, got %v", v.Case, v.Valid, valid)
		}

		if valid != v.Valid {
			t.Errorf("scenario #%v exptected %v, got %v", v.Case, v.Valid, valid)
		}
	}
}

func TestInBetweenArray(t *testing.T) {
	t.Parallel()
	testCase := []struct {
		Case       int
		Value      interface{}
		Collection interface{}
		Valid      bool
	}{
		{
			Case: 1,
			Value: []string{
				"test111",
				"test123",
				"test-test",
			},
			Collection: []string{
				"test",
				"test123",
			},
			Valid: true,
		},
		{
			Case:  2,
			Value: 2,
			Collection: []string{
				"test",
				"test123",
			},
			Valid: false,
		},
		{
			Case:       3,
			Value:      []int64{1, 6, 4, 5},
			Collection: []int64{1, 2, 3},
			Valid:      true,
		},

		{
			Case:  4,
			Value: []string{"case-7", "case-3"},
			Collection: map[string]string{
				"case-4": "case-4",
			},
			Valid: false,
		},
	}

	for _, v := range testCase {

		valid := InBetweenArray(v.Value, v.Collection)

		if valid == v.Valid {
			t.Logf("scenario #%v exptected %v, got %v", v.Case, v.Valid, valid)
		}

		if valid != v.Valid {
			t.Errorf("scenario #%v exptected %v, got %v", v.Case, v.Valid, valid)
		}
	}
}
