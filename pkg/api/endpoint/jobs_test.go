/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"errors"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestMakeFindJobEndpoint(t *testing.T) {
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
			name: "HappyPath1",
			req: &JobRequest{
				ID: "8ead6837-2967-42ad-9658-623c97c09d68",
			},
			want: Ok{
				Data: &api.JobResponse{
					ID:        "8ead6837-2967-42ad-9658-623c97c09d68",
					Operation: "OnboardDiscoveredAssets",
					Status:    "PENDING",
				},
			},
			wantErr: nil,
		},
		{
			name: "HappyPath2",
			req: &JobRequest{
				ID: "f63f0454-fd71-4f37-846a-507c9a1bb429",
			},
			want: Ok{
				Data: &api.JobResponse{
					ID:        "f63f0454-fd71-4f37-846a-507c9a1bb429",
					TeamID:    "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Operation: "OnboardDiscoveredAssets",
					Status:    "PENDING",
				},
			},
			wantErr: nil,
		},
		{
			name: "InvalidUUID",
			req: &JobRequest{
				ID: "1234",
			},
			wantErr: errors.New("ID is malformed"),
		},
		{
			name: "NotFound",
			req: &JobRequest{
				ID: "77f58c4b-7632-4e1b-8088-cb7241d148ae",
			},
			wantErr: errors.New("Job does not exist"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeFindJobEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}
