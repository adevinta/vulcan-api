/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/lib/pq"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	ignoreJobsDateFieldsOpts = cmpopts.IgnoreFields(api.Job{}, datesFieldNames...)
)

func TestServiceFindJob(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		ID      string
		name    string
		want    interface{}
		wantErr error
	}{
		{
			name: "HappyPath1",
			ID:   "8ead6837-2967-42ad-9658-623c97c09d68",
			want: &api.Job{
				ID:        "8ead6837-2967-42ad-9658-623c97c09d68",
				Operation: "OnboardDiscoveredAssets",
				Status:    "PENDING",
			},
			wantErr: nil,
		},
		{
			name: "HappyPath2",
			ID:   "f63f0454-fd71-4f37-846a-507c9a1bb429",
			want: &api.Job{
				ID:        "f63f0454-fd71-4f37-846a-507c9a1bb429",
				TeamID:    "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Operation: "OnboardDiscoveredAssets",
				Status:    "PENDING",
			},
			wantErr: nil,
		},
		/*
			{
				name:    "InvalidUUID",
				ID:      "1234",
				want:    nil,
				wantErr: errors.New("ID is malformed"),
			},
			{
				name: "NotFound",
				ID:   "77f58c4b-7632-4e1b-8088-cb7241d148ae",
				//want:    nil,
				wantErr: errors.New("Job does not exists"),
			},
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := vulcanitoService{
				db: testStore,
			}
			got, err := s.FindJob(context.Background(), tt.ID)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(nil, nil, ignoreJobsDateFieldsOpts)
			t.Logf("diff(%v)", diff)

			diff = cmp.Diff(tt.want, got, ignoreJobsDateFieldsOpts)
			if diff != "" {
				t.Logf("want(%+v),got(%+v)", tt.want, got)
				t.Errorf("%v\n", diff)
			}
		})
	}
}
