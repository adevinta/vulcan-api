/*
Copyright 2021 Adevinta
*/

package service

import (
	"os"
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
	loggerTeamMembers log.Logger
)

func init() {
	loggerTeamMembers = log.NewLogfmtLogger(os.Stderr)
	loggerTeamMembers = log.With(loggerTeamMembers, "ts", log.DefaultTimestampUTC)
	loggerTeamMembers = log.With(loggerTeamMembers, "caller", log.DefaultCaller)
}

func buildTeamMembersVulcanitoSrv(s api.VulcanitoStore) api.VulcanitoService {
	return buildDefaultVulcanitoSrv(s, loggerTeamMembers)
}

type testFindTeamMemberArgs struct {
	VulcanitoServiceTestArgs
	userID string
	teamID string
}

func TestFindTeamMember(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testFindTeamMemberArgs
		want    *api.UserTeam
		wantErr bool
		err     error
	}{
		{
			name: "HappyPath",
			args: testFindTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			},

			want: &api.UserTeam{
				UserID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				Role:   "member",
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
		},
		{
			name: "TeamIDisMissing",
			args: testFindTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "",
				userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			},
			wantErr: true,
			err:     apiErrors.Validation(`Team ID is empty`),
		},
		{
			name: "TeamIDisMissing",
			// teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			args: testFindTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				userID: "",
			},
			wantErr: true,
			err:     apiErrors.Validation(`User ID is empty`),
		},
		{
			name: "ErrNotFound",
			args: testFindTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c7",
				userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			},
			wantErr: true,
			err:     apiErrors.NotFound("record not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.FindTeamMember(tt.args.Ctx, tt.args.teamID, tt.args.userID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("vulcanitoService.FindTeamMember() wantErr = true and no error returned")

				}
				if err.Error() != tt.err.Error() {
					t.Fatalf("vulcanitoService.FindTeamMember() wantErr %s, err %s", tt.err.Error(), err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.UserTeam{}, "User", "Role", "Team", "CreatedAt", "UpdatedAt"))
			if diff != "" {
				t.Errorf("vulcanitoService.FindTeamMember(). Diffs:%s", diff)
			}
		})
	}
}

type testCreateTeamMembersArgs struct {
	VulcanitoServiceTestArgs
	teamMember api.UserTeam
}

func TestCreateTeamMember(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testCreateTeamMembersArgs
		want    *api.UserTeam
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testCreateTeamMembersArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					UserID: "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
					Role:   "member",
					TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
				},
			},

			want: &api.UserTeam{
				UserID: "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
				Role:   "member",
				TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
			},
		},
		{
			name: "MissingTeamID",
			args: testCreateTeamMembersArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					TeamID: "",
					UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
					Role:   "member",
				},
			},
			wantErr: apiErrors.Validation(`Key: 'UserTeam.TeamID' Error:Field validation for 'TeamID' failed on the 'required' tag`),
		},
		{
			name: "Without Role",
			args: testCreateTeamMembersArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
					UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
					Role:   "",
				},
			},
			want: &api.UserTeam{
				UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
				TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
				Role:   "member",
			},
			wantErr: nil,
		},
		{
			name: "Invalid Role",
			args: testCreateTeamMembersArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
					UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
					Role:   "INVALID",
				},
			},
			wantErr: apiErrors.Validation("Role is not valid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.CreateTeamMember(tt.args.Ctx, tt.args.teamMember)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("vulcanitoService.CreateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			diff = cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.UserTeam{}, "User", "Team", "CreatedAt", "UpdatedAt"))
			if diff != "" {
				t.Errorf("vulcanitoService.CreateTeamMember(). Diffs:%s", diff)
			}
		})
	}
}

type testUpdateTeamMemberArgs struct {
	VulcanitoServiceTestArgs
	teamMember api.UserTeam
}

func TestUpdateTeamMember(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testUpdateTeamMemberArgs
		want    *api.UserTeam
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testUpdateTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					UserID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
					Role:   "member",
					TeamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				},
			},

			want: &api.UserTeam{
				UserID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				Role:   "member",
				TeamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
			},
		},
		{
			name: "ValidationRule",
			args: testUpdateTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamMember: api.UserTeam{
					Role:   "member",
					TeamID: "2c99a8f7-1c87-4455-a9c9-82d1c0aee565",
				},
			},

			wantErr: apiErrors.Validation("Key: 'UserTeam.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			got, err := srv.UpdateTeamMember(tt.args.Ctx, tt.args.teamMember)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("vulcanitoService.CreateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != nil {
				got.Team = nil
				got.User = nil
			}

			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.UserTeam{}, "CreatedAt", "UpdatedAt"))
			if diff != "" {
				t.Errorf("vulcanitoService.CreateTeamMember(). Diffs:%s", diff)
			}
		})
	}
}

type testDeleteTeamMemberArgs struct {
	VulcanitoServiceTestArgs
	teamID string
	userID string
}

func TestDeleteTeamMember(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name    string
		args    testDeleteTeamMemberArgs
		wantErr error
	}{
		{
			name: "HappyPath",
			args: testDeleteTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			},
		},
		{
			name: "MissingTeamID",
			args: testDeleteTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "",
				userID: "1f862079-bd72-43d3-b09f-d855e01e67c5",
			},
			wantErr: apiErrors.Validation(`Team ID is empty`),
		},
		{
			name: "MissingUserID",
			args: testDeleteTeamMemberArgs{
				VulcanitoServiceTestArgs: VulcanitoServiceTestArgs{
					BuildSrv: buildTeamMembersVulcanitoSrv,
				},
				teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				userID: "",
			},
			wantErr: apiErrors.Validation(`User ID is empty`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.args.BuildSrv(testStore)
			err = srv.DeleteTeamMember(tt.args.Ctx, tt.args.teamID, tt.args.userID)
			if err != nil {
				if (tt.wantErr != nil) && (tt.wantErr.Error() == err.Error()) {
					return
				}
				t.Fatalf("vulcanitoService.CreateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
