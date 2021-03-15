/*
Copyright 2021 Adevinta
*/

package service

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestHealthcheckOk(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "HealthcheckOK",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()

			err = s.Healthcheck()

			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestHealthcheckServerDown(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "HealthcheckKO",
			wantErr: errors.New(`sql: database is closed`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
			if err != nil {
				t.Fatal(err)
			}
			s.Close()
			err = s.Healthcheck()
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}
