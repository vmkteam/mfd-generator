package mfd

import "testing"

func Test_isStartedRow(t *testing.T) {
	tests := []struct {
		name string
		row  string
		want bool
	}{
		{
			name: "empty string",
			row:  "",
			want: false,
		},
		{
			name: "function",
			row:  "func test(){",
			want: true,
		},
		{
			name: "struct",
			row:  "type Test struct{",
			want: true,
		},
		{
			name: "row with method",
			row:  "func (t *test) Test(){",
			want: true,
		},
		{
			name: "func in comment",
			row:  "//func test(){",
			want: false,
		},
		{
			name: "struct in comment",
			row:  "//type Test struct{",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStartedRow(tt.row); got != tt.want {
				t.Errorf("isStartedFunctionOrStructRow() = %v, want %v", got, tt.want)
			}
		})
	}
}
