/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/api"
	global "github.com/adevinta/vulcan-api/pkg/api/store/global"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	loggerProgram log.Logger
)

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

func init() {
	loggerProgram = log.NewLogfmtLogger(os.Stderr)
	loggerProgram = log.With(loggerProgram, "ts", log.DefaultTimestampUTC)
	loggerProgram = log.With(loggerProgram, "caller", log.DefaultCaller)

}

type MockGlobalStore struct {
	mux                sync.Mutex
	programsRepository map[string]global.Program
}

func (m *MockGlobalStore) Groups() map[string]global.Group {
	return nil
}

func (m *MockGlobalStore) Policies() map[string]global.Policy {
	return nil
}

func (m *MockGlobalStore) Programs() map[string]global.Program {
	m.mux.Lock()
	defer m.mux.Unlock()

	return m.programsRepository
}

type MockMetadataStore struct {
	mux                sync.Mutex
	programsRepository map[string]global.Program
}

func (m *MockMetadataStore) FindGlobalProgramMetadata(programID string, teamID string) (*api.GlobalProgramsMetadata, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if programID == "redcon-scan" {
		return &api.GlobalProgramsMetadata{
			Autosend: m.programsRepository[programID].DefaultMetadata.Autosend,
		}, nil
	}

	return nil, nil
}

func (m *MockMetadataStore) UpsertGlobalProgramMetadata(teamID, programID string, defaultAutosend bool, defaultCron string, autosend *bool, cron *string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if programID == "redcon-scan" {
		v := m.programsRepository[programID]
		v.DefaultMetadata.Autosend = autosend
		m.programsRepository[programID] = v
	}

	return nil
}

func (m *MockMetadataStore) DeleteProgramMetadata(program string) error {
	return nil
}

func (m *MockGlobalStore) Reports() map[string]global.Report {
	return nil
}

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

func Test_vulcanitoService_UpdateProgram(t *testing.T) {
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

	var False = false
	var True = true

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *api.Program
		wantErr error
	}{
		{
			name: "HappyPath",
			args: args{
				program: api.Program{
					ID:       "redcon-scan",
					Autosend: &True,
				},
			},
			want: &api.Program{
				ID:                     "redcon-scan",
				ProgramsGroupsPolicies: []*api.ProgramsGroupsPolicies{},
				Name:                   "redcon-scan",
				Disabled:               &True,
				Autosend:               &True,
				Global:                 &True,
			},
		},
		{
			name: "DisableAutoSend",
			args: args{
				program: api.Program{
					ID:       "redcon-scan",
					Autosend: &False,
				},
			},
			want: &api.Program{
				ID:                     "redcon-scan",
				ProgramsGroupsPolicies: []*api.ProgramsGroupsPolicies{},
				Name:                   "redcon-scan",
				Disabled:               &True,
				Autosend:               &False,
				Global:                 &True,
			},
		},
		{
			name: "SetName",
			args: args{
				program: api.Program{
					Name:     "bla",
					ID:       "redcon-scan",
					Autosend: &False,
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("only autosend field can be modified for global program"),
		},
		{
			name: "SetDisabled",
			args: args{
				program: api.Program{
					Disabled: &True,
					ID:       "redcon-scan",
					Autosend: &False,
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("only autosend field can be modified for global program"),
		},
	}
	var vTrue = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			programsRepository := map[string]global.Program{"redcon-scan": global.Program{
				Disabled: &vTrue,
			}}

			e := globalEntities{}
			e.store = &MockGlobalStore{
				programsRepository: programsRepository,
			}
			e.metadata = &MockMetadataStore{
				programsRepository: programsRepository,
			}
			e.scheduler = &globalScheduler{
				ScanScheduler: &MockScanScheduler{},
			}

			got, err := e.UpdateProgram(tt.args.ctx, tt.args.program, tt.args.team)
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
