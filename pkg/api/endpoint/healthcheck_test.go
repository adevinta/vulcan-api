/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestHealthcheckEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req     interface{}
		name    string
		want    interface{}
		wantErr error
	}{
		{
			name: "HealthcheckOK",
			req:  HealthcheckJSONRequest{},
			want: Ok{
				Data: api.HealthcheckResponse{
					Status: "OK",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeHealthcheckEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAssetResponse})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}
