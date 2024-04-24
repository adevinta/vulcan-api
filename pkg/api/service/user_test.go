/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/lib/pq"

	apiErrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/reports"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
	"github.com/adevinta/vulcan-api/pkg/testutil"
	"github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"
)

var (
	loggerUser               log.Logger
	ignoreUserDateFieldsOpts = cmpopts.IgnoreFields(api.User{}, datesFieldNames...)
	ignoreUserFieldsOpts     = cmpopts.IgnoreFields(api.User{}, append(datesFieldNames, "ID")...)
)

func init() {
	loggerUser = log.NewLogfmtLogger(os.Stderr)
	loggerUser = log.With(loggerUser, "ts", log.DefaultTimestampUTC)
	loggerUser = log.With(loggerUser, "caller", log.DefaultCaller)
}

type schedulerMock struct {
	schedule.ScanScheduler
}

type cgCatalogueMock struct {
	accounts map[string]string
}

func (c cgCatalogueMock) Name(providerID string) (string, error) {
	name := c.accounts[providerID]
	return name, nil
}

func buildUserVulcanitoSrv(s api.VulcanitoStore) api.VulcanitoService {
	return buildDefaultVulcanitoSrv(s, loggerUser)
}

func TestServiceCreateUser(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name       string
		user       *api.User
		srvBuilder VulcanitoServiceTestArgs
		want       *api.User
		wantErr    error
	}{
		{
			name: "HappyPath",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
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
			name: "UserInvalid",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			user: &api.User{
				Email:     "",
				Firstname: "New",
				Lastname:  "User",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			want:    nil,
			wantErr: apiErrors.Validation(`Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required' tag`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)
			got, err := testService.CreateUser(context.Background(), *tt.user)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, ignoreUserFieldsOpts)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestServiceUpdateUser(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		user       *api.User
		want       *api.User
		wantErr    error
	}{
		{
			name: "HappyPath",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
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
			name: "UserInvalid",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			user: &api.User{
				Email:     "",
				Firstname: "New",
				Lastname:  "User",
				Admin:     common.Bool(true),
				Observer:  common.Bool(true),
				Active:    common.Bool(true),
				APIToken:  "",
			},
			want:    nil,
			wantErr: apiErrors.Validation(`Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required' tag`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)
			got, err := testService.UpdateUser(context.Background(), *tt.user)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, ignoreUserDateFieldsOpts)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestServiceDeleteUser(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		userID     string
		wantErr    error
	}{
		{
			name: "HappyPath",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			userID:  "b58a941b-6c2f-402a-80ba-fbbf3c270dc6",
			wantErr: nil,
		},
		{
			name: "UserDoesNotExists",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			userID:  "b58a941b-6c2f-402a-80ba-aaaaaaaaaaaa",
			wantErr: apiErrors.NotFound(`User does not exists`),
		},
		{
			name: "Invalid",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			userID:  "",
			wantErr: apiErrors.Validation(`ID is empty`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)
			err = testService.DeleteUser(context.Background(), tt.userID)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestServiceFindUser(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		userID     string
		want       *api.User
		wantErr    error
	}{
		{
			name:   "HappyPath",
			userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
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
			name: "Invalid",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			userID:  "",
			want:    nil,
			wantErr: apiErrors.Validation(`ID is empty`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)
			got, err := testService.FindUser(context.Background(), tt.userID)
			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, ignoreUserDateFieldsOpts)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestServiceGenerateAPIToken(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	tests := []struct {
		signKey           string
		claim             string
		name              string
		authenticatedUser string
		activeUser        *bool
		adminUser         *bool
		Observer          *bool
		userID            string
		want              *api.Token
		wantErr           error
	}{
		{
			name:              "HappyPath",
			claim:             "email",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want: &api.Token{
				Email: "vulcan-team@vulcan.example.com",
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2MTU5OTU1NTgsInN1YiI6InZ1bGNhbi10ZWFtQHZ1bGNhbi5leGFtcGxlLmNvbSIsInR5cGUiOiJBUEkifQ.nRYZN3Yg-A59i5H4xz6KrvQcWYUq0FrpRtHlDq0Lmeg"},
			wantErr: nil,
		},
		{
			name:              "Missing email field on context",
			claim:             "",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want:              nil,
			wantErr:           apiErrors.Default(`type assertion failed when retrieving User from context`),
		},
		{
			name:              "MissingUserID",
			claim:             "email",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "",
			want:              nil,
			wantErr:           apiErrors.NotFound(`ID is empty`),
		},
		{
			name:              "UserDoesNotExists",
			claim:             "email",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "0585b0ce-e1f5-474b-a7c5-04e51673f666",
			want:              nil,
			wantErr:           apiErrors.NotFound(`User does not exists`),
		},
		{
			name:              "UserWithWrongID",
			claim:             "email",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "0585b0ce-e1f5-474b-a7c5",
			want:              nil,
			wantErr:           apiErrors.NotFound(`ID is malformed`),
		},
		{
			name:              "InvalidPermissions",
			claim:             "email",
			signKey:           "SUPERSECRETSIGNKEY",
			authenticatedUser: "vulcan-team@vulcan.example.com",
			activeUser:        common.Bool(true),
			adminUser:         common.Bool(false),
			Observer:          common.Bool(false),
			userID:            "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
			want:              nil,
			wantErr:           apiErrors.Forbidden(`Invalid permissions`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testServiceToken := New(loggerUser, testStore, jwt.NewJWTConfig(tt.signKey),
				scanengine.Config{Url: ""}, schedulerMock{}, reports.Config{},
				vulnerabilitydb.NewClient(nil, "", true), nil, nil, nil, cgCatalogueMock{},
				[]string{}, false)
			ctx := context.WithValue(context.Background(), tt.claim, api.User{Email: tt.authenticatedUser, Admin: tt.adminUser, Observer: tt.Observer, Active: tt.activeUser})
			got, err := testServiceToken.GenerateAPIToken(ctx, tt.userID)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			t1 := ""
			if tt.want != nil {
				t1 = tt.want.Token[:40]
			}

			t2 := ""
			if got != nil {
				t2 = got.Token[:40]
			}
			diff = cmp.Diff(t1, t2)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}
