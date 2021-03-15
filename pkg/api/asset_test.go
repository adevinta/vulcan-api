/*
Copyright 2021 Adevinta
*/

package api

import "testing"

func TestROLFP_String(t *testing.T) {
	tests := []struct {
		name  string
		ROLFP *ROLFP
		want  string
	}{
		{
			name: "should return empty string if IsEmpty is true",
			ROLFP: &ROLFP{
				IsEmpty: true,
			},
			want: "",
		},
		{
			name: "should return properly convert to string when all the fields are set",
			ROLFP: &ROLFP{
				Reputation: 1,
				Operation:  1,
				Legal:      1,
				Financial:  1,
				Personal:   1,
				Scope:      2,
			},
			want: "R:1/O:1/L:1/F:1/P:1+S:2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.ROLFP.String(); got != tt.want {
				t.Errorf("ROLFP.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
