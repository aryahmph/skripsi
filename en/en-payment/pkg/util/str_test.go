package util

import "testing"

func TestStringJoin(t *testing.T) {
	type args struct {
		elems   []string
		sep     string
		lastSep string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "#1",
			args: args{
				elems:   []string{"name", "email"},
				sep:     "=?,",
				lastSep: "=?",
			},
			want: "name=?,email=?",
		},
		{
			name: "#2",
			args: args{
				elems:   []string{"name"},
				sep:     "=?,",
				lastSep: "=?",
			},
			want: "name=?",
		},
		{
			name: "#3",
			args: args{
				elems:   []string{},
				sep:     "=?,",
				lastSep: "=?",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringJoin(tt.args.elems, tt.args.sep, tt.args.lastSep); got != tt.want {
				t.Errorf("StringJoin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubString(t *testing.T) {
	type args struct {
		input  string
		start  int
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "#1",
			args: args{
				input:  "",
				start:  1,
				length: 0,
			},
			want: "",
		},
		{
			name: "#2",
			args: args{
				input:  "to the moon",
				start:  0,
				length: 2,
			},
			want: "to",
		},
		{
			name: "#3",
			args: args{
				input:  "L",
				start:  0,
				length: 2,
			},
			want: "L",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubString(tt.args.input, tt.args.start, tt.args.length); got != tt.want {
				t.Errorf("SubString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringContains(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		result := StringContains("test", []string{"tes"})
		if result {
			t.Logf("expected '%v', got '%v", true, result)
		} else {
			t.Logf("expected '%v', got '%v", true, result)
		}
	})

	t.Run("Failed case", func(t *testing.T) {
		result := StringContains("test", []string{"boom"})
		if !result {
			t.Logf("expected '%v', got '%v", false, result)
		} else {
			t.Logf("expected '%v', got '%v", false, result)
		}
	})

}
