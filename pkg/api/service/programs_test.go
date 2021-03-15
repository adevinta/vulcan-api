/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	loggerProgram log.Logger
)

func init() {
	loggerProgram = log.NewLogfmtLogger(os.Stderr)
	loggerProgram = log.With(loggerProgram, "ts", log.DefaultTimestampUTC)
	loggerProgram = log.With(loggerProgram, "caller", log.DefaultCaller)

}

func TestVulcanitoService_CreateProgram(t *testing.T) {
	srv := vulcanitoService{
		db:     nil,
		logger: loggerProgram,
	}
	program := api.Program{}
	_, err := srv.CreateProgram(context.Background(), program, "")
	if err == nil {
		t.Error("Should return validation error if empty name")
	}
	expectedErrorMessage := "Key: 'Program.ProgramsGroupsPolicies' Error:Field validation for 'ProgramsGroupsPolicies' failed on the 'required' tag\nKey: 'Program.Name' Error:Field validation for 'Name' failed on the 'required' tag"
	diff := cmp.Diff(expectedErrorMessage, err.Error())
	if diff != "" {
		t.Errorf("Wrong error message, diff: %v", diff)
	}
}

func Test_vulcanitoService_CreateProgram(t *testing.T) {
	type fields struct {
		jwtConfig        jwt.Config
		db               api.VulcanitoStore
		logger           log.Logger
		programScheduler schedule.ScanScheduler
		scanEngineConfig scanengine.Config
	}
	type args struct {
		ctx     context.Context
		program api.Program
		team    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *api.Program
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := vulcanitoService{
				jwtConfig:        tt.fields.jwtConfig,
				db:               tt.fields.db,
				logger:           tt.fields.logger,
				programScheduler: tt.fields.programScheduler,
				scanEngineConfig: tt.fields.scanEngineConfig,
			}
			got, err := s.CreateProgram(tt.args.ctx, tt.args.program, tt.args.team)
			if (err != nil) != tt.wantErr {
				t.Errorf("vulcanitoService.CreateProgram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vulcanitoService.CreateProgram() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockScheduler is a do nothing mock implementation of scheduler to test happy path.
type mockScheduler struct{}

func (s *mockScheduler) CreateScanSchedule(programID, teamID, cronExpr string) error { return nil }
func (s *mockScheduler) GetScanScheduleByID(programID string) (string, error)        { return "", nil }
func (s *mockScheduler) DeleteScanSchedule(programID string) error                   { return nil }
func (s *mockScheduler) BulkCreateScanSchedules(schedules []schedule.ScanBulkSchedule) error {
	return nil
}

// invalidCronMockScheduler is a mock for scheduler interface which returns invalid cron error.
type cronErrMockScheduler struct{}

func (s *cronErrMockScheduler) CreateScanSchedule(programID, teamID, cronExpr string) error {
	return schedule.ErrInvalidCronExpr
}
func (s *cronErrMockScheduler) GetScanScheduleByID(programID string) (string, error) { return "", nil }
func (s *cronErrMockScheduler) DeleteScanSchedule(programID string) error            { return nil }
func (s *cronErrMockScheduler) BulkCreateScanSchedules(schedules []schedule.ScanBulkSchedule) error {
	return schedule.ErrInvalidCronExpr
}

// invalidPeriodMockScheduler is a mock for scheduler interface which returns invalid schedule period.
type periodErrMockScheduler struct{}

func (s *periodErrMockScheduler) CreateScanSchedule(programID, teamID, cronExpr string) error {
	return schedule.ErrInvalidSchedulePeriod
}
func (s *periodErrMockScheduler) GetScanScheduleByID(programID string) (string, error) { return "", nil }
func (s *periodErrMockScheduler) DeleteScanSchedule(programID string) error            { return nil }
func (s *periodErrMockScheduler) BulkCreateScanSchedules(schedules []schedule.ScanBulkSchedule) error {
	return schedule.ErrInvalidSchedulePeriod
}

func Test_vulcanitoService_ScheduleGlobalProgram(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	globalProgram := "periodic-full-scan"

	adminTrue := true
	adminFalse := false

	adminCtx := api.ContextWithUser(context.Background(), api.User{
		Admin: &adminTrue,
	})
	normalCtx := api.ContextWithUser(context.Background(), api.User{
		Admin: &adminFalse,
	})

	type fields struct {
		db               api.VulcanitoStore
		programScheduler schedule.ScanScheduler
	}
	type args struct {
		ctx       context.Context
		programID string
		cronExpr  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "should return error, not a global program",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				adminCtx,
				"someProgram",
				"*/2 * * * *",
			},
			wantErr: errors.Assertion("Program ID does not correspond to a global program"),
		},
		{
			name: "should return error, user is not admin",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				normalCtx,
				globalProgram,
				"*/2 * * * *",
			},
			wantErr: errors.Unauthorized("Can not schedule program"),
		},
		{
			name: "happy path",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				adminCtx,
				globalProgram,
				"*/2 * * * *",
			},
			wantErr: nil,
		},
		{
			name: "should return error, invalid cron expr",
			fields: fields{
				testStore,
				&cronErrMockScheduler{},
			},
			args: args{
				adminCtx,
				globalProgram,
				"1 2 3 4 5",
			},
			wantErr: errors.Assertion(schedule.ErrInvalidCronExpr),
		},
		{
			name: "should return error, invalid schedule period",
			fields: fields{
				testStore,
				&periodErrMockScheduler{},
			},
			args: args{
				adminCtx,
				globalProgram,
				"* * * * *",
			},
			wantErr: errors.Assertion(schedule.ErrInvalidSchedulePeriod),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := vulcanitoService{
				db:               tt.fields.db,
				programScheduler: tt.fields.programScheduler,
			}
			err := s.ScheduleGlobalProgram(tt.args.ctx, tt.args.programID, tt.args.cronExpr)
			if !reflect.DeepEqual(tt.wantErr, err) {
				t.Errorf("vulcanitoService.ScheduleGlobalProgram() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
		})
	}
}

func Test_vulcanitoService_CreateSchedule(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	validTeamID := "3c7c2963-6a03-4a25-a822-ebeb237db065"
	validProgramID := "1789b7b6-e8ed-49d9-a5a8-9ff9323593b6"

	invalidTeamID := "01010101-0101-0101-0101-010101010101"
	invalidProgramID := "01010101-0101-0101-0101-010101010101"

	ctx := context.Background()

	type fields struct {
		db               api.VulcanitoStore
		programScheduler schedule.ScanScheduler
	}
	type args struct {
		ctx       context.Context
		programID string
		teamID    string
		cronExpr  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "happy path",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				ctx,
				validProgramID,
				validTeamID,
				"*/2 * * * *",
			},
			wantErr: nil,
		},
		{
			name: "should return error, program not found due to invalid team ID",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				ctx,
				validProgramID,
				invalidTeamID,
				"*/2 * * * *",
			},
			wantErr: errors.NotFound("record not found"),
		},
		{
			name: "should return error, program not found due to invalid program ID",
			fields: fields{
				testStore,
				&mockScheduler{},
			},
			args: args{
				ctx,
				invalidProgramID,
				validTeamID,
				"*/2 * * * *",
			},
			wantErr: errors.NotFound("record not found"),
		},
		{
			name: "should return error, invalid cron expr",
			fields: fields{
				testStore,
				&cronErrMockScheduler{},
			},
			args: args{
				ctx,
				validProgramID,
				validTeamID,
				"1 2 3 4 5",
			},
			wantErr: errors.Assertion(schedule.ErrInvalidCronExpr),
		},
		{
			name: "should return error, invalid schedule period",
			fields: fields{
				testStore,
				&periodErrMockScheduler{},
			},
			args: args{
				ctx,
				validProgramID,
				validTeamID,
				"* * * * *",
			},
			wantErr: errors.Assertion(schedule.ErrInvalidSchedulePeriod),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := vulcanitoService{
				db:               tt.fields.db,
				programScheduler: tt.fields.programScheduler,
			}
			program, err := s.CreateSchedule(tt.args.ctx, tt.args.programID, tt.args.cronExpr, tt.args.teamID)
			if !reflect.DeepEqual(tt.wantErr, err) {
				t.Fatalf("vulcanitoService.CreateSchedule() error = '%v', wantErr = '%v'", err, tt.wantErr)
			}
			if err == nil && program.ID != tt.args.programID {
				t.Errorf("vulcanitoService.CreateSchedule() programID = '%s' expected = '%s'", program.ID, tt.args.programID)
			}
		})
	}
}
