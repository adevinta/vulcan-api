/*
Copyright 2021 Adevinta
*/

package store

import (
	"errors"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	ignoreFieldsUserTeam = cmpopts.IgnoreFields(api.UserTeam{}, "User", "Role", "Team", "CreatedAt", "UpdatedAt")
)

func TestStoreFindTeamMember(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		teamID  string
		userID  string
		want    *api.UserTeam
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want: &api.UserTeam{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				UserID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				Role:   "member"},
			wantErr: nil,
		},
		{
			name:    "NotFound",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			userID:  "aaaaaaaa-bbbb-cccc-dddd-ffffffffffff",
			want:    nil,
			wantErr: errors.New("record not found"),
		},
		{
			name:    "DatabaseErrorInvalidSyntax",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd",
			userID:  "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want:    nil,
			wantErr: errors.New(`pq: invalid input syntax for uuid: "aaaaaaaa-bbbb-cccc-dddd"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeamMember(tt.teamID, tt.userID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsUserTeam})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreCreateTeamMember(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name         string
		teamID       string
		userID       string
		role         string
		wantTeamUser *api.UserTeam
		wantErr      error
	}{
		{
			name:   "HappyPath",
			teamID: "42e5f970-d104-437f-b356-32a74e2cfd5a",
			userID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
			role:   "member",
			wantTeamUser: &api.UserTeam{
				TeamID: "42e5f970-d104-437f-b356-32a74e2cfd5a",
				UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
				Role:   "member",
			},
			wantErr: nil,
		},
		{
			name:         "IsAlreadyAMember",
			teamID:       "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			userID:       "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			role:         "member",
			wantTeamUser: nil,
			wantErr:      errors.New("User is already a member of this team"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			teamMember, err := testStoreLocal.CreateTeamMember(api.UserTeam{
				TeamID: tt.teamID,
				UserID: tt.userID,
				Role:   api.Role(tt.role)})
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.wantTeamUser, teamMember, cmp.Options{ignoreFieldsUserTeam})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestUpdateTeamMember(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name         string
		teamUser     *api.UserTeam
		wantTeamUser *api.UserTeam
		wantErr      error
	}{
		{
			name: "HappyPath",
			teamUser: &api.UserTeam{
				TeamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
				Role:   "member",
			},
			wantTeamUser: &api.UserTeam{
				TeamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
				Role:   "member",
			},
			wantErr: nil,
		},
		{
			name: "NotAMember",
			teamUser: &api.UserTeam{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				UserID: "175b48c6-043b-48cf-a7c5-1934f6f50c7a",
				Role:   "member",
			},
			wantTeamUser: nil,
			wantErr:      errors.New(`User is not a member of this team`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			teamMember, err := testStoreLocal.UpdateTeamMember(*tt.teamUser)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.wantTeamUser, teamMember, cmp.Options{ignoreFieldsUserTeam})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestDeleteTeamMember(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		teamID  string
		userID  string
		wantErr error
	}{
		{
			name:    "HappyPath",
			teamID:  "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
			userID:  "1f862079-bd72-43d3-b09f-d855e01e67c5",
			wantErr: nil,
		},
		{
			name:    "NotAMember",
			teamID:  "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			userID:  "1f862079-bd72-43d3-b09f-d855e01e67c5",
			wantErr: errors.New(`User is not a member of this team`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.DeleteTeamMember(tt.teamID, tt.userID)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}
