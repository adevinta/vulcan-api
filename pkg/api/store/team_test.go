/*
Copyright 2021 Adevinta
*/

package store

import (
	"encoding/json"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/lib/pq"

	apiErrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	baseModelFieldNames = []string{"ID", "CreatedAt", "UpdatedAt"}
	ignoreFieldsTeam    = cmpopts.IgnoreFields(api.Team{}, baseModelFieldNames...)
)

func TestStoreFindTeam(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	userteam := &api.UserTeam{UserID: "sdfsdf", TeamID: "sdfsdf"}
	userteams := []*api.UserTeam{userteam}
	tests := []struct {
		name    string
		teamID  string
		want    *api.Team
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			want: &api.Team{ID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Name:        "Foo Team",
				Description: "Foo foo...",
				UserTeam:    userteams,
				Tag:         "team:foo-team",
			},
			wantErr: nil,
		},
		{
			name:    "NotFound",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:    nil,
			wantErr: errors.New("record not found"),
		},
		{
			name:    "DatabaseErrorInvalidSyntax",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd",
			want:    nil,
			wantErr: errors.New(`pq: invalid input syntax for type uuid: "aaaaaaaa-bbbb-cccc-dddd"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeam(tt.teamID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.Team{}, append(baseModelFieldNames, "UserTeam")...)})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreFindTeamByProgram(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name      string
		programID string
		want      *api.Team
		wantErr   error
	}{
		{
			name:      "HappyPath",
			programID: "b75c2371-3a90-40dc-8831-98374506c80e",
			want:      &api.Team{ID: "93449fc4-6a84-4058-bac2-200e584ec435"},
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeamByProgram(tt.programID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.Team{}, append(baseModelFieldNames)...)})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreFindTeamsByUser(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		userID  string
		want    []*api.Team
		wantErr error
	}{
		{
			name:   "HappyPath",
			userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
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
					Tag:         "a.b.c.5d3e3f0bc169"},
			},
			wantErr: nil,
		},
		{
			name:    "UserWithoutTeams",
			userID:  "0585b0ce-e1f5-474b-a7c5-04e51673f8b4",
			want:    []*api.Team{},
			wantErr: nil,
		},
		{
			name:    "UserNotFound",
			userID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:    nil,
			wantErr: errors.New("record not found"),
		},
		{
			name:    "DatabaseErrorInvalidSyntax",
			userID:  "aaaaaaaa-bbbb-cccc-dddd",
			want:    nil,
			wantErr: errors.New(`pq: invalid input syntax for type uuid: "aaaaaaaa-bbbb-cccc-dddd"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeamsByUser(tt.userID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsTeam})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreFindTeamByIDForUser(t *testing.T) {
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
			teamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
			userID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
			want: &api.UserTeam{
				TeamID: "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				UserID: "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
				User: &api.User{
					ID:        "4a4bec34-8c1b-42c4-a6fb-2a2dbafc572e",
					Firstname: "Vulcan",
					Lastname:  "Team",
					Active:    common.Bool(true),
					Admin:     common.Bool(false),
					Observer:  common.Bool(false),
					Email:     "vulcan-team@vulcan.example.com",
					APIToken:  "3e666891f17cbb8defe642cd38eb9b7fd7ec0937e8ed5323e598fa983a35cbd6"},
				Role: "member",
			},
			wantErr: nil,
		},
		{
			name:    "NotFound",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			userID:  "aaaaaaaa-bbbb-cccc-dddd-ffffffffffff",
			want:    nil,
			wantErr: apiErrors.NotFound(`record not found`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeamByIDForUser(tt.teamID, tt.userID)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsTeam, ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreCreateTeam(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name       string
		emailOwner string
		team       *api.Team
		wantTeam   *api.Team
		wantErr    error
	}{
		{
			name:       "HappyPath",
			emailOwner: "vulcan-team@vulcan.example.com",
			team: &api.Team{
				Name:        "Create Team",
				Description: "Create this team...",
				Tag:         "1",
			},
			wantTeam: &api.Team{
				Name:        "Create Team",
				Description: "Create this team...",
				Tag:         "1",
			},
			wantErr: nil,
		},
		{
			name:       "UserNotExists",
			emailOwner: "not-exists@vulcan.example.com",
			team: &api.Team{
				Name:        "Create Team",
				Description: "Create this team...",
				Tag:         "2",
			},
			wantTeam: nil,
			wantErr:  errors.New(`record not found`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			team, err := testStoreLocal.CreateTeam(*tt.team, tt.emailOwner)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.wantTeam, team, cmp.Options{ignoreFieldsTeam})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreUpdateTeam(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name     string
		team     *api.Team
		wantTeam *api.Team
		wantErr  error
	}{
		{
			name: "HappyPath",
			team: &api.Team{
				ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				Name:        "Bar Team Updated",
				Description: "Bar bar...",
				Tag:         "a.b.c.5d3e3f0bc169"},
			wantTeam: &api.Team{
				ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
				Name:        "Bar Team Updated",
				Description: "Bar bar...",
				Tag:         "a.b.c.5d3e3f0bc169"},
			wantErr: nil,
		},
		{
			name: "TeamNotFound",
			team: &api.Team{
				ID:          "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				Name:        "Bar Team Updated",
				Description: "Bar bar..."},
			wantTeam: nil,
			wantErr:  errors.New("record not found"),
		},
		{
			name: "DatabaseError",
			team: &api.Team{
				ID:          "aaaaaaaa-bbbb-cccc-dddd",
				Name:        "Bar Team Updated",
				Description: "Bar bar..."},
			wantTeam: nil,
			wantErr:  errors.New(`pq: invalid input syntax for type uuid: "aaaaaaaa-bbbb-cccc-dddd"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			team, err := testStoreLocal.UpdateTeam(*tt.team)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.wantTeam, team, cmp.Options{ignoreFieldsTeam})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreDeleteTeam(t *testing.T) {
	localStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	testStoreLocal := localStore.(vulcanitoStore)
	defer testStoreLocal.Close()
	expCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	expUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	discoveryMergeTeamAssets := []api.Asset{}
	var discoveryMergeTeamID = "a14c7c65-66ab-4676-bcf6-0dea9719f5c6"
	err = testStoreLocal.Conn.
		Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Where("team_id = ?", discoveryMergeTeamID).
		Find(&discoveryMergeTeamAssets).Error

	tests := []struct {
		name         string
		teamID       string
		expOutbox    []expOutbox
		verifyOutBox func()
		wantErr      error
	}{
		{
			name:   "HappyPath",
			teamID: discoveryMergeTeamID,
			verifyOutBox: func() {
				expDeletedAllAssets := expOutbox{
					action: opDeleteTeam,
					dto: cdc.OpDeleteTeamDTO{
						Team: api.Team{
							ID:          discoveryMergeTeamID,
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &expCreatedAt,
							UpdatedAt:   &expUpdatedAt,
						},
					},
				}
				var gotOutbox []cdc.Outbox = make([]cdc.Outbox, 0)
				db := testStoreLocal
				err := db.Conn.Raw(`
					SELECT * FROM outbox
					ORDER BY created_at DESC`,
				).Scan(&gotOutbox).Error
				if err != nil {
					t.Fatalf("error verifying outbox: %v", err)
				}
				expOutboxLen := len(discoveryMergeTeamAssets) + 1
				if len(gotOutbox) != expOutboxLen {
					t.Fatalf("error verifying outbox, expected %d records got %d", expOutboxLen, len(gotOutbox))
				}
				ignoreFields := map[string][]string{}
				// We expect the last operation to be the delete all asset operation.
				diff := expDeletedAllAssets.Compare(gotOutbox[expOutboxLen-1], ignoreFields)
				if diff != "" {
					t.Fatalf("error verifying outbox, expected last operation != got last operation, diff:\n %s", diff)
				}
				gotOutbox = gotOutbox[:expOutboxLen-1]
				for _, exp := range discoveryMergeTeamAssets {
					var gotDTO *cdc.OpDeleteAssetDTO
					for z, a := range gotOutbox {
						if a.Operation != opDeleteAsset {
							fmtStr := "error verifying outbox, expected record in position to have operation %d to be: %s but is: %s"
							t.Fatalf(fmtStr, z, opDeleteAsset, a.Operation)
						}
						var dto cdc.OpDeleteAssetDTO
						err := json.Unmarshal(a.DTO, &dto)
						if err != nil {
							fmtStr := "error verifying outbox unmarshaling data from got outbox record: %s"
							t.Fatalf(fmtStr, string(a.DTO))
						}
						if dto.Asset.ID == exp.ID {
							gotDTO = &dto
							break
						}
					}
					if gotDTO == nil {
						fmtStr := "error verifying outbox expected asset %+v, not found in outbox"
						t.Fatalf(fmtStr, exp)
					}
					if gotDTO.DeleteAllAssetsOp != true {
						fmtStr := "error verifying outbox expected DeletedAllAssets to be true, but is false in: %+v"
						t.Fatalf(fmtStr, *gotDTO)
					}
					ignoreFieldsAsset := cmpopts.IgnoreFields(api.Asset{}, baseModelFieldNames...)
					diff := cmp.Diff(exp, gotDTO.Asset, ignoreFieldsAsset, ignoreFieldsTeam)
					if diff != "" {
						t.Fatalf("error verifying outbox, DTO's do not match.\nDiff:\n%v", diff)
					}
				}
			},
		},
		{
			name:    "TeamNotFound",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			wantErr: errors.New("record not found"),
		},
		{
			name:    "DatabaseError",
			teamID:  "aaaaaaaa-bbbb-cccc-dddd",
			wantErr: errors.New(`pq: invalid input syntax for type uuid: "aaaaaaaa-bbbb-cccc-dddd"`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// We need to clean the outbox before each test.
			err := testStoreLocal.Conn.Exec("DELETE FROM outbox").Error
			if err != nil {
				t.Fatalf("Error cleaning the outbox %+v", err)
			}
			err = testStoreLocal.DeleteTeam(tt.teamID)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatal(diff)
			}
			if tt.verifyOutBox != nil {
				tt.verifyOutBox()
			}
		})
	}
}

func TestStoreListTeams(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("testdata/TestStoreListTeams", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		want    []*api.Team
		wantErr error
	}{
		{
			name:    "HappyPath",
			wantErr: nil,
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
			got, err := testStoreLocal.ListTeams()
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsTeam})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreFindTeamsByTags(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		tags    []string
		want    []*api.Team
		wantErr error
	}{
		{
			name: "HappyPath",
			tags: []string{
				"team:foo-team",
				"a.b.c.5d3e3f0bc169",
			},
			want: []*api.Team{
				{
					ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Name:        "Foo Team",
					Description: "Foo foo...",
					Tag:         "team:foo-team",
				},
				{
					ID:          "d92e6a31-d889-425d-9a16-5d3e3f0bc169",
					Name:        "Bar Team",
					Description: "Bar bar...",
					Tag:         "a.b.c.5d3e3f0bc169",
				},
			},
			wantErr: nil,
		},

		{
			name:    "NoTeamsFound",
			tags:    []string{"inexistent"},
			want:    []*api.Team{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.FindTeamsByTags(tt.tags)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsTeam, ignoreFieldsUser})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func errToStr(err error) string {
	result := ""
	if err != nil {
		result = err.Error()
	}
	return result
}
