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
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/saml"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	ignoreFieldsUser = cmpopts.IgnoreFields(api.User{}, baseModelFieldNames...)
)

func TestStoreCreateUser(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		user    *api.User
		want    *api.User
		wantErr error
	}{
		{
			name: "HappyPath",
			user: &api.User{
				Email:     "new-user@vulcan.example.com",
				Firstname: "New",
				Lastname:  "User",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			want: &api.User{
				Email:     "new-user@vulcan.example.com",
				Firstname: "New",
				Lastname:  "User",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			wantErr: nil,
		},
		{
			name: "AlreadyExists",
			user: &api.User{
				Email: "vulcan-team@vulcan.example.com",
			},
			wantErr: errors.New("User already exists"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.CreateUser(*tt.user)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreUpdateUser(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		user    *api.User
		want    *api.User
		wantErr error
	}{
		{
			name: "HappyPath",
			user: &api.User{
				ID:        "1123af8f-a9cd-48b1-8a0d-382d3cfb47c4",
				Email:     "update-user@vulcan.example.com",
				Firstname: "new name",
				Lastname:  "new lastname",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			want: &api.User{
				ID:        "1123af8f-a9cd-48b1-8a0d-382d3cfb47c4",
				Email:     "update-user@vulcan.example.com",
				Firstname: "new name",
				Lastname:  "new lastname",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			wantErr: nil,
		},
		{
			name: "UserDoesNotExists",
			user: &api.User{
				ID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			},
			wantErr: errors.New("User does not exists"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.UpdateUser(*tt.user)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreCreateUserIfNotExists(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		user    saml.UserData
		wantErr error
	}{
		{
			name: "HappyPath",
			user: saml.UserData{
				Email:     "saml-user@vulcan.example.com",
				FirstName: "saml",
				LastName:  "user",
				UserName:  "username"},
			wantErr: nil,
		},
		{
			name: "UserAlreadyExists",
			user: saml.UserData{
				Email: "vulcan-team@vulcan.example.com",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.CreateUserIfNotExists(tt.user)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
		})
	}
}

func TestStoreCreateRecipientsAsTeamMembers(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		teamID  string
		user    saml.UserData
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			user: saml.UserData{
				Email:     "saml-user@vulcan.example.com",
				FirstName: "saml",
				LastName:  "user",
				UserName:  "username"},
			wantErr: nil,
		},
		{
			name:   "UserAlreadyExists",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			user: saml.UserData{
				Email: "vulcan-team@vulcan.example.com",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.CreateUserIfNotExists(tt.user)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			team, err := testStoreLocal.FindTeam(tt.teamID)
			if err != nil {
				t.Fatal(err)
			}

			found := false
			for _, member := range team.UserTeam {
				if member.User.Email == tt.user.Email {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf("Recipient %s should be added as a member into team %s", tt.user.Email, tt.teamID)
			}
		})
	}
}

func TestStoreFindUserByID(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		userID  string
		want    *api.User
		wantErr error
	}{
		{
			name:   "HappyPath",
			userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want: &api.User{
				ID:        "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				Email:     "vulcan-team@vulcan.example.com",
				Firstname: "Vulcan",
				Lastname:  "Team",
				Admin:     common.Bool(false),
				Observer:  common.Bool(false),
				Active:    common.Bool(true),
				APIToken:  "3e666891f17cbb8defe642cd38eb9b7fd7ec0937e8ed5323e598fa983a35cbd6",
			},
			wantErr: nil,
		},
		{
			name:    "UserDoesNotExists",
			userID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			wantErr: errors.New("User does not exists"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindUserByID(tt.userID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreFindUserByEmail(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name      string
		userEmail string
		want      *api.User
		wantErr   error
	}{
		{
			name:      "HappyPath",
			userEmail: "vulcan-team@vulcan.example.com",
			want: &api.User{
				ID:        "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				Email:     "vulcan-team@vulcan.example.com",
				Firstname: "Vulcan",
				Lastname:  "Team",
				Admin:     common.Bool(false),
				Observer:  common.Bool(false),
				Active:    common.Bool(true),
				APIToken:  "3e666891f17cbb8defe642cd38eb9b7fd7ec0937e8ed5323e598fa983a35cbd6",
			},
			wantErr: nil,
		},
		{
			name:      "UserDoesNotExists",
			userEmail: "inexistent-user@vulcan.example.com",
			wantErr:   errors.New("User does not exists"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindUserByEmail(tt.userEmail)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreDeleteUser(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()
	type testCase struct {
		name    string
		userID  string
		wantErr error
	}
	tests := []testCase{
		{
			name:    "HappyPath",
			userID:  "b58a941b-6c2f-402a-80ba-fbbf3c270dc6",
			wantErr: nil,
		},
		{
			name:    "DeletesTeamsAssociations",
			userID:  "c2cbf3ee-1b2e-11e9-ab14-d663bd873d93",
			wantErr: nil,
		},
		{
			name:    "UserDoesNotExists",
			userID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			wantErr: errors.New("User does not exists"),
		},
		{
			name:    "InvalidUserID",
			userID:  "aaaaaaaa-bbbb-cccc-dddd",
			wantErr: errors.New("pq: invalid input syntax for type uuid: \"aaaaaaaa-bbbb-cccc-dddd\" (SQLSTATE 22P02)"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.DeleteUserByID(tt.userID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
				return
			}
			// If the test expected and error whe don't and we are here the test
			// is already finished.
			if tt.wantErr != nil {
				return
			}
			// If no error check that there the user does not exist and is not
			// associated with any team.
			_, err = testStoreLocal.FindUserByID(tt.userID)
			if err == nil {
				t.Errorf("user in defined in test was not deleted")
				return
			}
			if errToStr(err) != "User does not exists" {
				t.Errorf("error the user was not deleted %+v", err)
				return
			}

			teams, err := testStoreLocal.FindTeamsByUser(tt.userID)
			if errToStr(err) != "record not found" {
				t.Errorf("error checking there are no teams associated with the user: %+v", err)
				return
			}
			if len(teams) > 0 {
				t.Errorf("not all the user teams associations are deleted")
			}
		})
	}
}
