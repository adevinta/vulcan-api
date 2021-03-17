/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiErrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	loggerTeams              log.Logger
	ignoreTeamDateFieldsOpts = cmpopts.IgnoreFields(api.Team{}, datesFieldNames...)
	sortTeamsOpts            = cmp.Transformer("Sort", func(in []*api.Team) []*api.Team {
		out := append([]*api.Team(nil), in...) // Copy input to avoid mutating it
		sort.Slice(out, func(i, j int) bool {
			less := strings.Compare(out[i].ID, out[j].ID)
			return less < 0
		})
		return out
	})
)

func init() {
	loggerTeams = log.NewLogfmtLogger(os.Stderr)
	loggerTeams = log.With(loggerTeams, "ts", log.DefaultTimestampUTC)
	loggerTeams = log.With(loggerTeams, "caller", log.DefaultCaller)
}

func buildTeamVulcanitoSrv(s api.VulcanitoStore) api.VulcanitoService {
	return buildDefaultVulcanitoSrv(s, loggerTeams)
}

func TestVulcanitoService_FindTeam(t *testing.T) {
	srv := vulcanitoService{
		db:     nil,
		logger: loggerTeams,
	}
	id := ""
	_, err := srv.FindTeam(context.Background(), id)
	if err == nil {
		t.Error("Should return validation error if empty name")
	}
	expectedErrorMessage := "ID is empty"
	diff := cmp.Diff(expectedErrorMessage, err.Error())
	if diff != "" {
		t.Errorf("Wrong error message, diff: %v", diff)
	}
}

type testFindTeamsByUserArgs struct {
	VulcanitoServiceTestArgs
	userID string
}

func TestFindTeamsByUser(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testFindTeamsByUserArgs
		want    []*api.Team
		wantErr bool
		err     error
	}{
		{
			name: "HappyPath",
			args: testFindTeamsByUserArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			},

			want: []*api.Team{
				&api.Team{
					ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Name:        "Foo Team",
					Description: "Foo foo...",
					Tag:         "team:foo-team",
				},
				&api.Team{
					ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
					Name:        "Bar Team",
					Description: "Bar bar...",
					Tag:         "a.b.c.5d3e3f0bc169",
				},
			},
		},
		{
			name: "NotFound",
			args: testFindTeamsByUserArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				userID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c7",
			},
			wantErr: true,
			err:     apiErrors.NotFound("record not found"),
		},
		{
			name: "MissingTeamID",
			args: testFindTeamsByUserArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				userID: "",
			},
			wantErr: true,
			err:     apiErrors.Validation(`ID is empty`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.FindTeamsByUser(tt.args.Ctx, tt.args.userID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("vulcanitoService.FindTeamsByUser() wantErr = true and no error returned")

				}
				if err.Error() != tt.err.Error() {
					t.Fatalf("vulcanitoService.FindTeamsByUser() wantErr %s, err %s", tt.err.Error(), err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(tt.want, got, ignoreTeamDateFieldsOpts)
			if diff != "" {
				t.Errorf("vulcanitoService.FindMember(). Diffs:%s", diff)
			}
		})
	}
}

type testCreateTeamArgs struct {
	VulcanitoServiceTestArgs
	team  api.Team
	email string
}

func TestCreateTeam(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testCreateTeamArgs
		want    *api.Team
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testCreateTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				team: api.Team{
					ID:          "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
					Name:        "TestTeam",
					Description: "Test team description",
					Tag:         "a:b:c:d",
				},
				email: "vulcan-team@vulcan.example.com",
			},

			want: &api.Team{
				ID:          "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
				Name:        "TestTeam",
				Description: "Test team description",
				Tag:         "a:b:c:d",
			},
		},
		{
			name: "TeamNameMissing",
			args: testCreateTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				team: api.Team{
					Name:        "",
					Description: "Create this team...",
					Tag:         "a:b:c:d",
				},
				email: "not-exists@vulcan.example.com",
			},
			wantErr: apiErrors.Validation(`Key: 'Team.Name' Error:Field validation for 'Name' failed on the 'required' tag`),
		},
		{
			name: "OwnerMissing",
			args: testCreateTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				team: api.Team{
					ID:          "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
					Name:        "a name",
					Description: "Test team description",
					Tag:         "a:b:c:d",
				},
				email: "",
			},
			wantErr: apiErrors.Validation(`Owner email is empty`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.CreateTeam(tt.args.Ctx, tt.args.team, tt.args.email)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("vulcanitoService.CreateTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got, ignoreTeamDateFieldsOpts)
			if diff != "" {
				t.Errorf("vulcanitoService.CreateTeam(). Diffs:%s", diff)
			}
		})
	}
}

type testUpdateTeamArgs struct {
	VulcanitoServiceTestArgs
	team api.Team
}

func TestUpdateTeam(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testUpdateTeamArgs
		want    *api.Team
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testUpdateTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				team: api.Team{
					ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
					Description: "Bar changed",
					Name:        "Bar changed",
					Tag:         "a.b.c.5d3e3f0bc169.changed",
				},
			},

			want: &api.Team{
				ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				Description: "Bar changed",
				Name:        "Bar changed",
				Tag:         "a.b.c.5d3e3f0bc169.changed",
			},
		},
		{
			name: "TeamNameMissing",
			args: testUpdateTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				team: api.Team{
					ID:  "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
					Tag: "a.b.c.5d3e3f0bc169",
				},
			},
			wantErr: apiErrors.Validation(`Key: 'Team.Name' Error:Field validation for 'Name' failed on the 'required' tag`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.UpdateTeam(tt.args.Ctx, tt.args.team)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("vulcanitoService.UpdateTeam() error = %v, wantErr %v", err, tt.wantErr)
			}

			diff := cmp.Diff(tt.want, got, ignoreTeamDateFieldsOpts)
			if diff != "" {
				t.Errorf("vulcanitoService.UpdateTeam(). Diffs:%s", diff)
			}
		})
	}
}

type testDeleteTeamArgs struct {
	VulcanitoServiceTestArgs
	teamID string
}

func TestDeleteTeam(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testDeleteTeamArgs
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testDeleteTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				teamID: "0ef82297-e7c7-4c46-a852-ae3ffbecc4bc",
			},
		},
		{
			name: "TeamIDEmpty",
			args: testDeleteTeamArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
				teamID: "",
			},
			wantErr: apiErrors.Validation(`ID is empty`),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			err = srv.DeleteTeam(tt.args.Ctx, tt.args.teamID)
			if err != nil {
				if (tt.wantErr != nil) && (tt.wantErr.Error() == err.Error()) {
					return
				}
				t.Fatalf("vulcanitoService.DeleteTeam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testListTeamsArgs struct {
	VulcanitoServiceTestArgs
}

func TestListTeamsArgs(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("testdata/TestListTeamsArgs", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testListTeamsArgs
		want    []*api.Team
		wantErr bool
	}{
		{
			name: "HappyPath",
			args: testListTeamsArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamVulcanitoSrv,
				},
			},

			want: []*api.Team{
				&api.Team{
					ID:   "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Name: "Foo Team", Description: "Foo foo...",
					Tag: "team:foo-team",
				},
				&api.Team{
					ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
					Name:        "Bar Team",
					Description: "Bar bar...", Tag: "a.b.c.5d3e3f0bc169",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.ListTeams(tt.args.Ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("vulcanitoService.ListTeams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(got, tt.want, sortTeamsOpts, ignoreTeamDateFieldsOpts)
			if diff != "" {
				t.Errorf("vulcanitoService.ListTeams(). Diffs: %s", diff)
			}
		})
	}
}
