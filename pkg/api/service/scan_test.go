/*
Copyright 2021 Adevinta
*/

package service

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

type MockScanScheduler struct {
}

func (m *MockScanScheduler) CreateScanSchedule(programID, teamID, cronExpr string) error {
	return nil
}

func (m *MockScanScheduler) GetScanScheduleByID(programID string) (string, error) {
	return "", nil
}

func (m *MockScanScheduler) DeleteScanSchedule(programID string) error {
	return nil
}

func (m *MockScanScheduler) BulkCreateScanSchedules(schedules []schedule.ScanBulkSchedule) error {
	return nil
}

type testCreateScanArgs struct {
	VulcanitoServiceTestArgs
	teamID string
	scan   api.Scan
}

func TestCreateScan(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testCreateScanArgs
		want    *api.Scan
		wantErr error
	}{
		{
			name: "DisabledProgram",
			args: testCreateScanArgs{
				scan: api.Scan{
					ProgramID: "abd059d0-f81a-433b-be8a-ace3ebb6c926",
				},
				teamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065",
			},
			want:    nil,
			wantErr: fmt.Errorf("Program Disabled Program is disabled. [Program ID: abd059d0-f81a-433b-be8a-ace3ebb6c926]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := vulcanitoService{
				db:               testStore,
				programScheduler: &MockScanScheduler{},
			}

			got, err := srv.CreateScan(tt.args.Ctx, tt.args.scan, tt.args.teamID)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
			diff = cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.ProgramResponse{}, "ID")})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}
