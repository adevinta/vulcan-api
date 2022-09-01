/*
Copyright 2021 Adevinta
*/

package store

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func Test_sanitizeTableName(t *testing.T) {

	tests := []struct {
		name  string
		table string
		want  string
	}{
		{
			name:  "removesForbbidenChars",
			table: "(^~'test-^table^_name^~')",
			want:  "----test--table-_name----",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeTableName(tt.table); got != tt.want {
				t.Errorf("sanitizeTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}
