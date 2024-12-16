package mBankStatementCSV

import (
	"testing"
)

func Test_movePointLeft(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", "0"},
		{"0", "0"},
		{"0.0", "0"},
		{"12345", "123.45"},
		{"12345.", "123.45"},
		{"12345.0", "123.45"},
		{"1234.5", "12.345"},
		{"123.45", "1.2345"},
		{"12.345", "0.12345"},
		{"1.2345", "0.012345"},
		{".12345", "0.0012345"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := movePointLeft(tt.s); got != tt.want {
				t.Errorf("movePointLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}
