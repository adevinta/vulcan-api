/*
Copyright 2021 Adevinta
*/

package store

import (
	"errors"
	"log"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

var (
	ignoreFieldsAsset       = cmpopts.IgnoreFields(api.Asset{}, append(baseModelFieldNames, "Team", "AssetType", "AssetTypeID", "ClassifiedAt")...)
	ignoreFieldsDisjoin     = cmpopts.IgnoreFields(api.Asset{}, "CreatedAt", "UpdatedAt", "Team", "AssetType")
	ignoreFieldsGroup       = cmpopts.IgnoreFields(api.Group{}, []string{"ID", "CreatedAt", "UpdatedAt", "Team", "AssetGroup"}...)
	ignoreFieldsAnnotations = cmpopts.IgnoreFields(api.AssetAnnotation{}, "Asset", "AssetID", "CreatedAt", "UpdatedAt")
)

func TestStoreListAssets(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	hostnameType, err := testStoreLocal.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}

	domainType, err := testStoreLocal.GetAssetType("domainname")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		teamID  string
		want    []*api.Asset
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			want: []*api.Asset{
				&api.Asset{
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "foo1.vulcan.example.com",
					AssetTypeID: hostnameType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:       api.DefaultROLFP,
					AssetGroups: []*api.AssetGroup{
						&api.AssetGroup{
							AssetID: "0f206826-14ec-4e85-a5a4-e2decdfbc193",
							GroupID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
							Group: &api.Group{
								ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
								TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
								Name:   "Default",
							},
						},
					},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				&api.Asset{
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "foo1.vulcan.example.com",
					AssetTypeID: domainType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:       api.DefaultROLFP,
					AssetGroups: []*api.AssetGroup{
						&api.AssetGroup{
							AssetID: "283e773d-54b5-460a-91fe-f3dfca5838a6",
							GroupID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
							Group: &api.Group{
								ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
								TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
								Name:   "Default",
							},
						},
					},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				&api.Asset{
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "foo2.vulcan.example.com",
					AssetTypeID: hostnameType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}}]}`),
					ROLFP:       api.DefaultROLFP,
					AssetGroups: []*api.AssetGroup{
						&api.AssetGroup{
							AssetID: "53ef6c94-0b07-4ba2-bc8c-6cef68c20ddb",
							GroupID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
							Group: &api.Group{
								ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
								TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
								Name:   "Default",
							},
						},
					},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				&api.Asset{
					TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:        "foo3.vulcan.example.com",
					AssetTypeID:       "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					Scannable:         common.Bool(true),
					Options:           common.String("{}"),
					EnvironmentalCVSS: common.String("5"),
					ROLFP:             api.DefaultROLFP,
					AssetGroups:       []*api.AssetGroup{},
					AssetAnnotations:  []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.ListAssets(tt.teamID, api.Asset{})
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAsset})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreCreateAssets(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	hostnameType, err := testStoreLocal.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		assets  []api.Asset
		groups  []api.Group
		want    []api.Asset
		wantErr error
	}{
		{
			name: "HappyPath",
			assets: []api.Asset{
				{
					TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:        "vulcan.example.com",
					AssetTypeID:       hostnameType.ID,
					EnvironmentalCVSS: common.String("c.v.s.s."),
					Scannable:         common.Bool(true),
					Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:             &api.ROLFP{IsEmpty: true},
					Alias:             "Alias1",
				}},
			groups: []api.Group{
				{
					ID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
				},
			},
			want: []api.Asset{
				{
					TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:        "vulcan.example.com",
					AssetTypeID:       hostnameType.ID,
					EnvironmentalCVSS: common.String("c.v.s.s."),
					Scannable:         common.Bool(true),
					Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:             &api.ROLFP{IsEmpty: true},
					Alias:             "Alias1",
					AssetAnnotations:  []*api.AssetAnnotation{},
				}},
			wantErr: nil,
		},
		{
			name: "NonExistentTeam",
			assets: []api.Asset{
				{
					TeamID:            "9f7a0c78-b752-4126-aa6d-0f286ada7b8f",
					Identifier:        "vulcan.example.com",
					AssetTypeID:       hostnameType.ID,
					EnvironmentalCVSS: common.String("c.v.s.s."),
					Scannable:         common.Bool(true),
					Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:             &api.ROLFP{IsEmpty: true},
				}},
			groups:  []api.Group{},
			want:    nil,
			wantErr: errors.New("[asset][vulcan.example.com][] record not found"),
		},
		{
			name: "WithAnnotations",
			assets: []api.Asset{
				{
					TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:        "asset_with_anontations.example.com",
					AssetTypeID:       hostnameType.ID,
					EnvironmentalCVSS: common.String("c.v.s.s."),
					Scannable:         common.Bool(true),
					Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:             &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{
						{
							Key:   "key1",
							Value: "value1",
						},
					},
				}},
			groups: []api.Group{},
			want: []api.Asset{
				{
					TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:        "asset_with_anontations.example.com",
					AssetTypeID:       hostnameType.ID,
					EnvironmentalCVSS: common.String("c.v.s.s."),
					Scannable:         common.Bool(true),
					Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
					ROLFP:             &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{
						{
							Key:   "key1",
							Value: "value1",
						},
					},
				}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.CreateAssets(tt.assets, tt.groups)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAsset}, cmp.Options{ignoreFieldsAnnotations})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestStoreCreateAsset(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	hostnameType, err := testStoreLocal.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}

	opts := `{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`
	expTeamCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	expTeamUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")

	tests := []struct {
		name      string
		asset     api.Asset
		groups    []api.Group
		want      *api.Asset
		wantErr   error
		expOutbox expOutbox
	}{
		{
			name: "HappyPath",
			asset: api.Asset{
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier:        "vulcan.example.com",
				AssetTypeID:       hostnameType.ID,
				EnvironmentalCVSS: common.String("c.v.s.s."),
				Scannable:         common.Bool(true),
				Options:           common.String(opts),
				ROLFP:             &api.ROLFP{IsEmpty: true},
				Alias:             "Alias1",
			},
			groups: []api.Group{
				{
					ID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
				},
			},
			want: &api.Asset{
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier:        "vulcan.example.com",
				AssetTypeID:       hostnameType.ID,
				EnvironmentalCVSS: common.String("c.v.s.s."),
				Scannable:         common.Bool(true),
				Options:           common.String(opts),
				ROLFP:             &api.ROLFP{IsEmpty: true},
				Alias:             "Alias1",
				AssetAnnotations:  []*api.AssetAnnotation{},
			},
			wantErr: nil,
			expOutbox: expOutbox{
				action: opCreateAsset,
				dto: cdc.OpCreateAssetDTO{
					Asset: api.Asset{
						TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Team: &api.Team{
							ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &expTeamCreatedAt,
							UpdatedAt:   &expTeamUpdatedAt,
						},
						Alias:      "Alias1",
						Identifier: "vulcan.example.com",
						AssetType: &api.AssetType{
							ID:   hostnameType.ID,
							Name: hostnameType.Name,
						},
						AssetTypeID:       hostnameType.ID,
						EnvironmentalCVSS: common.String("c.v.s.s."),
						ROLFP:             &api.ROLFP{IsEmpty: true},
						Scannable:         common.Bool(true),
						Options:           common.String(opts),
						AssetAnnotations:  []*api.AssetAnnotation{},
					},
				},
			},
		},
		{
			name: "NonExistentTeam",
			asset: api.Asset{
				TeamID:            "9f7a0c78-b752-4126-aa6d-0f286ada7b8f",
				Identifier:        "vulcan.example.com",
				AssetTypeID:       hostnameType.ID,
				EnvironmentalCVSS: common.String("c.v.s.s."),
				Scannable:         common.Bool(true),
				Options:           common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
				ROLFP:             &api.ROLFP{IsEmpty: true},
			},
			groups:  []api.Group{},
			want:    nil,
			wantErr: errors.New("[asset][vulcan.example.com][] record not found"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.CreateAsset(tt.asset, tt.groups)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAsset})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}

			if tt.wantErr == nil {
				ignoreFieldsOutbox := map[string][]string{"asset": {"id", "classified_at"}}
				verifyOutbox(t, testStoreLocal, tt.expOutbox, ignoreFieldsOutbox)
			}
		})
	}
}

func TestStoreUpdateAsset(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	testStoreLocal := testStore.(vulcanitoStore)
	defer testStoreLocal.Close()

	hpExpTeamCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	hpExpTeamUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	ndExpTeamCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2018-01-01 12:30:12")
	ndExpTeamUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2018-01-01 12:30:12")

	tests := []struct {
		name        string
		asset       api.Asset
		want        *api.Asset
		wantErr     error
		cleanOutbox bool
		expOutbox   expOutbox
	}{
		{
			name: "HappyPath",
			asset: api.Asset{
				ID:         "0f206826-14ec-4e85-a5a4-e2decdfbc193",
				TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier: "vulcan.example.bis.com",
			},
			want: &api.Asset{
				ID:         "0f206826-14ec-4e85-a5a4-e2decdfbc193",
				TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier: "vulcan.example.bis.com",
			},
			wantErr: nil,
			expOutbox: expOutbox{
				action: opUpdateAsset,
				dto: cdc.OpUpdateAssetDTO{
					OldAsset: api.Asset{
						ID:     "0f206826-14ec-4e85-a5a4-e2decdfbc193",
						TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Team: &api.Team{
							ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &hpExpTeamCreatedAt,
							UpdatedAt:   &hpExpTeamUpdatedAt,
						},
						Identifier: "foo1.vulcan.example.com",
					},
					NewAsset: api.Asset{
						ID:     "0f206826-14ec-4e85-a5a4-e2decdfbc193",
						TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Team: &api.Team{
							ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &hpExpTeamCreatedAt,
							UpdatedAt:   &hpExpTeamUpdatedAt,
						},
						Identifier: "vulcan.example.bis.com",
					},
					DupAssets: 1,
				},
			},
		},
		{
			name: "Should report no duplicates",
			asset: api.Asset{
				ID:         "49f90ed2-2f71-11e9-b210-d663bd873d93",
				TeamID:     "5125225e-4912-4464-b22e-e2542410c352",
				Identifier: "updated.vulcan.example.com",
			},
			want: &api.Asset{
				ID:         "49f90ed2-2f71-11e9-b210-d663bd873d93",
				TeamID:     "5125225e-4912-4464-b22e-e2542410c352",
				Identifier: "updated.vulcan.example.com",
			},
			wantErr: nil,
			expOutbox: expOutbox{
				action: opUpdateAsset,
				dto: cdc.OpUpdateAssetDTO{
					OldAsset: api.Asset{
						ID:     "49f90ed2-2f71-11e9-b210-d663bd873d93",
						TeamID: "5125225e-4912-4464-b22e-e2542410c352",
						Team: &api.Team{
							ID:          "5125225e-4912-4464-b22e-e2542410c352",
							Name:        "TeamWithAssetsDefaultSensitive",
							Description: "TeamWithAssetsDefaultSensitive",
							CreatedAt:   &ndExpTeamCreatedAt,
							UpdatedAt:   &ndExpTeamUpdatedAt,
						},
						Identifier: "noscan.vulcan.example.com",
					},
					NewAsset: api.Asset{
						ID:     "49f90ed2-2f71-11e9-b210-d663bd873d93",
						TeamID: "5125225e-4912-4464-b22e-e2542410c352",
						Team: &api.Team{
							ID:          "5125225e-4912-4464-b22e-e2542410c352",
							Name:        "TeamWithAssetsDefaultSensitive",
							Description: "TeamWithAssetsDefaultSensitive",
							CreatedAt:   &ndExpTeamCreatedAt,
							UpdatedAt:   &ndExpTeamUpdatedAt,
						},
						Identifier: "updated.vulcan.example.com",
					},
					DupAssets: 0,
				},
			},
		},
		{
			name: "UpdatesButNotDeletesAnnotations",
			asset: api.Asset{
				ID:         "73e33dcb-d07c-41d1-bc32-80861b49941e",
				Identifier: "nonscannable.vulcan.example.com",
				TeamID:     "ea686be5-be9b-473b-ab1b-621a4f575d51",
				AssetAnnotations: []*api.AssetAnnotation{
					{
						Key:   "newkey",
						Value: "newvalue",
					},
					{
						Key:   "autodiscovery/security/keytoupdate",
						Value: "updated",
					},
				},
			},
			want: &api.Asset{
				ID:         "73e33dcb-d07c-41d1-bc32-80861b49941e",
				TeamID:     "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Identifier: "nonscannable.vulcan.example.com",
				AssetAnnotations: []*api.AssetAnnotation{
					{
						Key:   "newkey",
						Value: "newvalue",
					},
					{
						Key:   "autodiscovery/security/keytoupdate",
						Value: "updated",
					},
				},
			},
			wantErr:     nil,
			cleanOutbox: true,
			expOutbox: expOutbox{
				action:     opUpdateAsset,
				notPresent: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanOutbox {
				if err := testStoreLocal.DeleteAllOutbox(); err != nil {
					t.Fatalf("error cleaning outbox: %+v", err)
				}
			}
			got, err := testStoreLocal.UpdateAsset(tt.asset)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			// UpdateAsset does not return the data related to asset being updated.
			// So we must do it by hand.
			annotations, err := testStoreLocal.ListAssetAnnotations(tt.asset.TeamID, tt.asset.ID)
			if err != nil {
				t.Fatalf("error getting asset annotations %+v", err)
			}
			got.AssetAnnotations = annotations
			trans := cmp.Transformer("SortAnnotations", func(in []*api.AssetAnnotation) []*api.AssetAnnotation {
				out := append([]*api.AssetAnnotation(nil), in...)
				sort.Slice(out, func(i, j int) bool {
					return strings.Compare(out[i].Key, out[j].Key) < 0
				})
				return out
			})

			diff := cmp.Diff(tt.want, got,
				trans,
				cmp.Options{ignoreFieldsAsset},
				cmp.Options{ignoreFieldsAnnotations},
			)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}

			if tt.wantErr == nil {
				ignoreFieldsOutbox := map[string][]string{
					"old_asset": {"alias", "asset_type", "asset_type_id", "environmental_cvss", "rolfp", "scannable", "classified_at", "options"},
					"new_asset": {"alias", "asset_type", "asset_type_id", "environmental_cvss", "rolfp", "scannable", "classified_at", "options"},
				}
				verifyOutbox(t, testStoreLocal, tt.expOutbox, ignoreFieldsOutbox)
			}
		})
	}
}

func TestVulcanitoStore_ListGroups(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name      string
		teamID    string
		groupName string
		want      []*api.Group
		wantErr   error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			want: []*api.Group{
				&api.Group{
					ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
					TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Name:   "Default",
				},
				&api.Group{
					ID:     "516099e5-7cb4-4624-8e6e-27af2de80872",
					TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Name:   "Sensitive",
				},
			},
			wantErr: nil,
		},
		{
			name:      "ShouldFilterByName",
			teamID:    "3C7C2963-6A03-4A25-A822-EBEB237DB065",
			groupName: "mygroup",
			want: []*api.Group{
				&api.Group{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Name:   "mygroup",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.ListGroups(tt.teamID, tt.groupName)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got,
				cmpopts.IgnoreFields(api.Group{}, []string{"CreatedAt", "UpdatedAt", "Team", "AssetGroup"}...))
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestVulcanitoStore_CreateGroup(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		group   *api.Group
		want    *api.Group
		wantErr error
	}{
		{
			name: "HappyPath",
			group: &api.Group{
				Name:   "my group",
				ID:     "b2411ffc-6441-4134-b6ed-dae27adeb6f9",
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
			want: &api.Group{
				Name:   "my group",
				ID:     "b2411ffc-6441-4134-b6ed-dae27adeb6f9",
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.CreateGroup(*tt.group)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsGroup})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestVulcanitoStore_DeleteGroup(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	group := api.Group{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065"}
	err = testStoreLocal.DeleteGroup(group)
	if err != nil {
		t.Error("Cannot delete group")
	}
}

func TestVulcanitoStore_UpdateGroup(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	group := &api.Group{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", Name: "new name", TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065"}
	group, err = testStoreLocal.UpdateGroup(*group)
	if err != nil {
		t.Error("Cannot update group")
	}
	if group.Name != "new name" {
		t.Error("Name was not updated correctly")
	}
}

func TestVulcanitoStore_FindGroup(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	group, err := testStoreLocal.FindGroup(api.Group{ID: "721d1c6b-f559-4c56-8ea5-ca1820173a3c", TeamID: "3C7C2963-6A03-4A25-A822-EBEB237DB065"})
	if err != nil {
		t.Error("Could not find group")
	}
	if group.Name != "mygroup" {
		t.Error("Did not find expected group")
	}
}

func TestStoreDeleteAsset(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	varTrue := true
	opts := `{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`
	expTeamCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	expTeamUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")

	tests := []struct {
		name      string
		asset     api.Asset
		expOutbox expOutbox
		wantErr   error
	}{
		{
			name: "HappyPath with duplicate asset",
			asset: api.Asset{
				ID:     "283e773d-54b5-460a-91fe-f3dfca5838a6",
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
			expOutbox: expOutbox{
				action: opDeleteAsset,
				dto: cdc.OpDeleteAssetDTO{
					Asset: api.Asset{
						ID:     "283e773d-54b5-460a-91fe-f3dfca5838a6",
						TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Team: &api.Team{
							ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &expTeamCreatedAt,
							UpdatedAt:   &expTeamUpdatedAt,
						},
						Identifier:  "foo1.vulcan.example.com",
						AssetTypeID: "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00",
						ROLFP:       api.DefaultROLFP,
						Scannable:   &varTrue,
						Options:     &opts,
					},
					DupAssets: 1,
				},
			},
		},
		{
			name: "HappyPath",
			asset: api.Asset{
				ID:     "0f206826-14ec-4e85-a5a4-e2decdfbc193",
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
			expOutbox: expOutbox{
				action: opDeleteAsset,
				dto: cdc.OpDeleteAssetDTO{
					Asset: api.Asset{
						ID:     "0f206826-14ec-4e85-a5a4-e2decdfbc193",
						TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Team: &api.Team{
							ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
							Name:        "Foo Team",
							Description: "Foo foo...",
							Tag:         "team:foo-team",
							CreatedAt:   &expTeamCreatedAt,
							UpdatedAt:   &expTeamUpdatedAt,
						},
						Identifier:  "foo1.vulcan.example.com",
						AssetTypeID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
						ROLFP:       api.DefaultROLFP,
						Scannable:   &varTrue,
						Options:     &opts,
					},
					DupAssets: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.DeleteAsset(tt.asset)
			if err != tt.wantErr {
				t.Fatal(err)
			}

			type Result struct {
				Count int
			}

			var result Result

			// do a raw query on database and ensure asset was deleted
			err = testStoreLocal.(vulcanitoStore).Conn.Raw(`
				SELECT count(*) FROM assets a WHERE id = ?`, tt.asset.ID).
				Scan(&result).Error
			if err != nil {
				t.Fatal(err)
			}

			if result.Count != 0 {
				t.Fatalf("Asset %v was not deleted", tt.asset)
			}

			verifyOutbox(t, testStoreLocal, tt.expOutbox, nil)
		})
	}
}

func TestStoreDeleteAllAssets(t *testing.T) {
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	expCreatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")
	expUpdatedAt, _ := time.Parse("2006-01-02 15:04:05", "2017-01-01 12:30:12")

	tests := []struct {
		name      string
		teamID    string
		expOutbox expOutbox
		wantErr   error
	}{
		{
			name:   "HappyPath",
			teamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			expOutbox: expOutbox{
				action: opDeleteAllAssets,
				dto: cdc.OpDeleteAllAssetsDTO{
					Team: api.Team{
						ID:          "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Name:        "Foo Team",
						Description: "Foo foo...",
						Tag:         "team:foo-team",
						CreatedAt:   &expCreatedAt,
						UpdatedAt:   &expUpdatedAt,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testStoreLocal.DeleteAllAssets("a14c7c65-66ab-4676-bcf6-0dea9719f5c6")
			if err != tt.wantErr {
				t.Fatal(err)
			}

			type Result struct {
				Count int
			}

			var result Result

			// do a raw query on database and ensure that there are no orphans asset group associations
			err = testStoreLocal.(vulcanitoStore).Conn.Raw(`SELECT count(*) FROM asset_group ag WHERE ag.asset_id not in (select a.id from assets a)`).Scan(&result).Error
			if err != nil {
				t.Fatal(err)
			}

			if result.Count != 0 {
				t.Fatalf("Number of orphan asset group associations left on database is different than zero: %d", result.Count)
			}

			verifyOutbox(t, testStoreLocal, tt.expOutbox, nil)
		})
	}
}

func Test_vulcanitoStore_DisjoinAssetsInGroups(t *testing.T) {
	type args struct {
		teamID        string
		inGroupID     string
		notInGroupIDs []string
	}
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name    string
		args    args
		want    []*api.Asset
		wantErr bool
	}{
		{
			name: "ReturnsDisjoinAssetGroups",
			args: args{
				inGroupID: "4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				notInGroupIDs: []string{
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
					"78ba0bb9-cce7-451d-8e17-ae6b3efbb788", // DisjoinB
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			want: []*api.Asset{
				&api.Asset{
					ID:                "b7c2e6bd-d63b-4bcc-8883-566aa3837c2d", // disjoin1.adevinta.com
					TeamID:            "d335c30c-944f-4ab0-9b43-cffccdfbd848",
					AssetTypeID:       "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					Identifier:        "disjoin1.adevinta.com",
					Options:           strToPtr(`{"opt":"val"}`),
					EnvironmentalCVSS: strToPtr("5"),
					Scannable:         boolToPtr(true),
					ROLFP:             api.DefaultROLFP,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.DisjoinAssetsInGroups(tt.args.teamID, tt.args.inGroupID, tt.args.notInGroupIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("vulcanitoStore.DisjoinAssetsInGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsDisjoin})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func Test_vulcanitoStore_DisjoinAssetsInGroups_Multiple(t *testing.T) {
	type args struct {
		teamID        string
		inGroupID     string
		notInGroupIDs []string
	}
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name      string
		args      args
		wantUUIDs []string
		wantErr   bool
	}{
		{
			name: "ReturnsBaseWithoutA&B",
			args: args{
				inGroupID: "4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				notInGroupIDs: []string{
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
					"78ba0bb9-cce7-451d-8e17-ae6b3efbb788", // DisjoinB
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: []string{
				"b7c2e6bd-d63b-4bcc-8883-566aa3837c2d", // disjoin1.adevinta.com
			},
		},
		{
			name: "ReturnsBaseWithoutA",
			args: args{
				inGroupID: "4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				notInGroupIDs: []string{
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: []string{
				"b7c2e6bd-d63b-4bcc-8883-566aa3837c2d", // disjoin1.adevinta.com
				"6a521ca7-490e-4789-a716-c2baca750884", // disjoin4.adevinta.com
			},
		},
		{
			name: "ReturnsGroupWithoutItself",
			args: args{
				inGroupID: "4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				notInGroupIDs: []string{
					"4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: nil,
		},
		{
			name: "ReturnsDisjointed",
			args: args{
				inGroupID: "a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				notInGroupIDs: []string{
					"78ba0bb9-cce7-451d-8e17-ae6b3efbb788", // DisjoinB
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: []string{
				"4e369e6b-a6bd-44f5-b0fc-690c063a240e", // disjoin2.adevinta.com
				"6c391632-89ea-4a99-9177-624f709351bb", // disjoin3.adevinta.com
			},
		},
		{
			name: "ReturnsGroupWithoutSuperset",
			args: args{
				inGroupID: "a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				notInGroupIDs: []string{
					"4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: nil,
		},
		{
			name: "ReturnsGroupWithoutVoid",
			args: args{
				inGroupID:     "a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				notInGroupIDs: nil,
				teamID:        "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: []string{
				"4e369e6b-a6bd-44f5-b0fc-690c063a240e", // disjoin2.adevinta.com
				"6c391632-89ea-4a99-9177-624f709351bb", // disjoin3.adevinta.com
			},
		},
		{
			name: "ReturnsGroupWithoutUnexistent",
			args: args{
				inGroupID:     "a4b93a4b-653e-4f01-92d2-bd360c50ae27",           // DisjoinA
				notInGroupIDs: []string{"cccccccc-aaaa-ffff-eeee-eeeeeeeeeeee"}, // Unexistent
				teamID:        "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantUUIDs: []string{
				"4e369e6b-a6bd-44f5-b0fc-690c063a240e", // disjoin2.adevinta.com
				"6c391632-89ea-4a99-9177-624f709351bb", // disjoin3.adevinta.com
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.DisjoinAssetsInGroups(tt.args.teamID, tt.args.inGroupID, tt.args.notInGroupIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("vulcanitoStore.DisjoinAssetsInGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var gotUUIDs []string
			for _, asset := range got {
				gotUUIDs = append(gotUUIDs, asset.ID)
			}
			if !cmp.Equal(tt.wantUUIDs, gotUUIDs) {
				t.Errorf("want: %v, got: %v\n", tt.wantUUIDs, gotUUIDs)
			}
		})
	}
}

func Test_vulcanitoStore_CountAssetsInGroups(t *testing.T) {
	type args struct {
		teamID   string
		groupIDs []string
	}
	testStoreLocal, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)
	}
	defer testStoreLocal.Close()

	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name: "CountAssetsSingle",
			args: args{
				groupIDs: []string{
					"4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantCount: 4,
		},
		{
			name: "CountAssetsMultiple",
			args: args{
				groupIDs: []string{
					"4123163c-6a78-43eb-8120-d129bcd0898a", // DisjoinBase
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantCount: 4,
		},
		{
			name: "CountAssetsDisjointed",
			args: args{
				groupIDs: []string{
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
					"78ba0bb9-cce7-451d-8e17-ae6b3efbb788", // DisjoinB
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantCount: 3,
		},
		{
			name: "CountAssetsNonExistentGroup",
			args: args{
				groupIDs: []string{
					"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", // Non-existent group
				},
				teamID: "d335c30c-944f-4ab0-9b43-cffccdfbd848",
			},
			wantCount: 0,
		},
		{
			name: "CountAssetsNonExistentTeam",
			args: args{
				groupIDs: []string{
					"a4b93a4b-653e-4f01-92d2-bd360c50ae27", // DisjoinA
				},
				teamID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", // Non-existent team
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStoreLocal.CountAssetsInGroups(tt.args.teamID, tt.args.groupIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("vulcanitoStore.CountAssetsInGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantCount {
				t.Errorf("want: %v, got: %v\n", tt.wantCount, got)
			}
		})
	}
}

func strToPtr(s string) *string {
	return &s
}

func boolToPtr(b bool) *bool {
	return &b
}
