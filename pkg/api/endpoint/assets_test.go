/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiErrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/service"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func TestMakeListAssetsEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req     interface{}
		name    string
		want    interface{}
		wantErr error
	}{
		{
			name: "HappyPath",
			req: &AssetRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			},
			want: Ok{
				Data: []api.AssetResponse{
					api.AssetResponse{
						ID:         "0f206826-14ec-4e85-a5a4-e2decdfbc193",
						Identifier: "foo1.vulcan.example.com",
						Scannable:  common.Bool(true),
						Options:    common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
						ROLFP:      api.DefaultROLFP,
						Groups:     []*api.GroupResponse{&api.GroupResponse{ID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4", Name: "Default"}},
					},
					api.AssetResponse{
						ID:         "283e773d-54b5-460a-91fe-f3dfca5838a6",
						Identifier: "foo1.vulcan.example.com",
						Scannable:  common.Bool(true),
						Options:    common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}},{"name":"vulcan-nessus","options":{"enabled":"false"}}]}`),
						ROLFP:      api.DefaultROLFP,
						Groups:     []*api.GroupResponse{&api.GroupResponse{ID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4", Name: "Default"}},
					},
					api.AssetResponse{
						ID:         "53ef6c94-0b07-4ba2-bc8c-6cef68c20ddb",
						Identifier: "foo2.vulcan.example.com",
						Scannable:  common.Bool(true),
						Options:    common.String(`{"checktype_options":[{"name":"vulcan-exposed-memcheck","options":{"https":"true","port":"11211"}}]}`),
						ROLFP:      api.DefaultROLFP,
						Groups:     []*api.GroupResponse{&api.GroupResponse{ID: "ab310d43-8cdf-4f65-9ee8-d1813a22bab4", Name: "Default"}},
					},
					api.AssetResponse{
						ID: "13376826-14ec-4e85-a5a4-e2decdfbc193",
						AssetType: api.AssetTypeResponse{
							ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
						},
						Identifier:        "foo3.vulcan.example.com",
						Options:           common.String("{}"),
						EnvironmentalCVSS: common.String("5"),
						Scannable:         common.Bool(true),
						ROLFP:             api.DefaultROLFP,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "TeamDoesNotExists",
			req: &AssetRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-aaaaaaaaaaaa",
			},
			want:    nil,
			wantErr: apiErrors.NotFound(`record not found`),
		},
		{
			name: "TeamWithoutAssets",
			req: &AssetRequest{
				TeamID: "2a5db64a-b4ad-4e1d-af01-152616d92e2d",
			},
			want: Ok{
				Data: []api.AssetResponse{},
			},
			wantErr: nil,
		},
		{
			name: "TeamIDMalformed",
			req: &AssetRequest{
				TeamID: "2a5db64a-xxxx",
			},
			want:    nil,
			wantErr: apiErrors.NotFound(`pq: invalid input syntax for type uuid: "2a5db64a-xxxx"`),
		},
		{
			name: "TeamIDMissing",
			req: &AssetRequest{
				TeamID: "",
			},
			want:    nil,
			wantErr: apiErrors.NotFound(`Team ID is empty`),
		},

		{
			name: "TypeAssertionError",
			req: AssetRequest{
				TeamID: "",
			},
			want:    nil,
			wantErr: apiErrors.NotFound(`Type assertion failed`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeListAssetsEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{ignoreFieldsAssetResponse})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestMakeCreateAssetsEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req            interface{}
		name           string
		want           interface{}
		wantErr        error
		wantClassified bool
	}{
		{
			name: "HappyPath",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Identifier:        "adevinta.com",
						Type:              "Hostname",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
					}}},

			want: Created{
				[]api.AssetResponse{
					{
						ID:                "eb8556b0-14ee-449f-b60c-e6876eec07d5",
						Identifier:        "adevinta.com",
						AssetType:         api.AssetTypeResponse{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             api.DefaultROLFP,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should be classified",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Identifier:        "www.adevinta.com",
						Type:              "Hostname",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
					}}},

			want: Created{
				[]api.AssetResponse{
					{
						ID:                "eb8556b0-14ee-449f-b60c-e6876eec07d5",
						Identifier:        "www.adevinta.com",
						AssetType:         api.AssetTypeResponse{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
					},
				},
			},
			wantErr:        nil,
			wantClassified: true,
		},
		{
			name: "Must have assets",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{
						Identifier:        "adevinta.com",
						Type:              "xxxxx",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
					}}},
			want:    nil,
			wantErr: apiErrors.Validation("[asset][adevinta.com][xxxxx] Asset type not found"),
		},
		{
			name: "IdentifierMissing",
			req: &AssetsListRequest{Assets: []AssetRequest{{
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier:        "",
				Type:              "hostname",
				Options:           common.String(`{}`),
				Scannable:         common.Bool(true),
				EnvironmentalCVSS: common.String("a.b.c.d"),
			}}},
			want:    nil,
			wantErr: apiErrors.NotFound("Key: 'Asset.TeamID' Error:Field validation for 'TeamID' failed on the 'required' tag\nKey: 'Asset.Identifier' Error:Field validation for 'Identifier' failed on the 'required' tag"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeCreateAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.AssetResponse{}, "ID", "AssetType", "ClassifiedAt")})
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			if tt.want != nil {
				assetResp := got.(Created).Data.([]api.AssetResponse)[0]
				if tt.wantClassified && assetResp.ClassifiedAt == nil {
					t.Fatal("Expected asset to be classified, but it was not")
				}
				if !tt.wantClassified && assetResp.ClassifiedAt != nil {
					t.Fatalf(("Expected asset to NOT be classified, but it was"))
				}
			}
		})
	}
}

func TestMakeCreateAssetsMultiStatusEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req            interface{}
		name           string
		want           interface{}
		wantErr        error
		wantClassified bool
	}{
		{
			name: "HappyPath",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Identifier:        "adevinta.com",
						Type:              "Hostname",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
					}}},

			want: Created{
				[]api.AssetCreationResponse{
					{
						ID:         "eb8556b0-14ee-449f-b60c-e6876eec07d5",
						Identifier: "adevinta.com",
						AssetType: api.AssetTypeResponse{
							ID:   "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
							Name: "Hostname"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             api.DefaultROLFP,
						Status: api.Status{
							Code: 201,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should be classified",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
						Identifier:        "www.adevinta.com",
						Type:              "Hostname",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
					}}},

			want: Created{
				[]api.AssetCreationResponse{
					{
						ID:         "eb8556b0-14ee-449f-b60c-e6876eec07d4",
						Identifier: "www.adevinta.com",
						AssetType: api.AssetTypeResponse{
							ID:   "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
							Name: "Hostname"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
						Status: api.Status{
							Code: 201,
						},
					},
				},
			},
			wantErr:        nil,
			wantClassified: true,
		},
		{
			name: "Must have assets",
			req: &AssetsListRequest{
				TeamID: "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Assets: []AssetRequest{
					{
						Identifier:        "adevinta.com",
						Type:              "xxxxx",
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
					}}},
			want: MultiStatus{
				[]api.AssetCreationResponse{
					{
						ID:                "???",
						Identifier:        "adevinta.com",
						AssetType:         api.AssetTypeResponse{Name: "xxxxx"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						Status: &apiErrors.ErrorStack{
							Errors: []apiErrors.Error{
								apiErrors.Error{
									Kind:           apiErrors.ErrValidation,
									Message:        "[asset][adevinta.com][xxxxx] Asset type not found",
									HTTPStatusCode: http.StatusUnprocessableEntity,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "IdentifierMissing",
			req: &AssetsListRequest{Assets: []AssetRequest{{
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Identifier:        "",
				Type:              "hostname",
				Options:           common.String(`{}`),
				Scannable:         common.Bool(true),
				EnvironmentalCVSS: common.String("a.b.c.d"),
			}}},
			want: MultiStatus{
				[]api.AssetCreationResponse{
					{
						ID:                "",
						Identifier:        "",
						AssetType:         api.AssetTypeResponse{Name: "hostname"},
						Options:           common.String(`{}`),
						Scannable:         common.Bool(true),
						EnvironmentalCVSS: common.String("a.b.c.d"),
						Status: &apiErrors.ErrorStack{
							Errors: []apiErrors.Error{
								apiErrors.Error{
									Kind:           apiErrors.ErrValidation,
									Message:        "Key: 'Asset.TeamID' Error:Field validation for 'TeamID' failed on the 'required' tag\nKey: 'Asset.Identifier' Error:Field validation for 'Identifier' failed on the 'required' tag",
									HTTPStatusCode: http.StatusUnprocessableEntity,
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeCreateAssetMultiStatusEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{
				cmpopts.IgnoreFields(api.AssetCreationResponse{}, "ID", "ClassifiedAt"),
				cmpopts.IgnoreFields(apiErrors.Error{}, "Kind"),
			})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}

			if tt.want != nil {
				var assetResp api.AssetCreationResponse
				if _, ok := got.(Created); ok {
					assetResp = got.(Created).Data.([]api.AssetCreationResponse)[0]
				} else {
					assetResp = got.(MultiStatus).Data.([]api.AssetCreationResponse)[0]
				}

				if tt.wantClassified && assetResp.ClassifiedAt == nil {
					t.Fatal("Expected asset to be classified, but it was not")
				}
				if !tt.wantClassified && assetResp.ClassifiedAt != nil {
					t.Fatalf(("Expected asset to NOT be classified, but it was"))
				}
			}
		})
	}
}

func TestMergeDiscoveredAssetEndpointValidation(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req     interface{}
		name    string
		want    interface{}
		wantErr error
	}{
		{
			name: "Doesn't fail if group ends with '-discovered-assets'",
			req: &DiscoveredAssetsRequest{
				TeamID: "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Assets: []AssetWithAnnotationsRequest{
					AssetWithAnnotationsRequest{
						AssetRequest: AssetRequest{
							Identifier:        "fancy.vulcan.example.com",
							Type:              "Hostname",
							Options:           common.String(`{}`),
							Scannable:         common.Bool(true),
							EnvironmentalCVSS: common.String("a.b.c.d"),
							ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
						},
					},
				},
				GroupName: "whatever-discovered-assets",
			},
			want:    NoContent{},
			wantErr: nil,
		},
		{
			name: "Fails if group doesn't end with '-discovered-assets'",
			req: &DiscoveredAssetsRequest{
				TeamID: "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Assets: []AssetWithAnnotationsRequest{
					AssetWithAnnotationsRequest{
						AssetRequest: AssetRequest{
							Identifier:        "fancy.vulcan.example.com",
							Type:              "Hostname",
							Options:           common.String(`{}`),
							Scannable:         common.Bool(true),
							EnvironmentalCVSS: common.String("a.b.c.d"),
							ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
						},
					},
				},
				GroupName: "Default",
			},
			wantErr: errors.New("Asset group not allowed"),
		},
		{
			name: "Fails if identifier is missing",
			req: &DiscoveredAssetsRequest{
				TeamID: "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Assets: []AssetWithAnnotationsRequest{
					AssetWithAnnotationsRequest{
						AssetRequest: AssetRequest{
							Type: "Hostname",
						},
					},
				},
				GroupName: "whatever-discovered-assets",
			},
			wantErr: errors.New("Asset identifier and type are required for all the assets"),
		},
		{
			name: "Fails if type is missing",
			req: &DiscoveredAssetsRequest{
				TeamID: "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Assets: []AssetWithAnnotationsRequest{
					AssetWithAnnotationsRequest{
						AssetRequest: AssetRequest{
							Identifier: "fancy.vulcan.example.com",
						},
					},
				},
				GroupName: "whatever-discovered-assets",
			},
			wantErr: errors.New("Asset identifier and type are required for all the assets"),
		},
		{
			name: "Fails if type is invalid",
			req: &DiscoveredAssetsRequest{
				TeamID: "ea686be5-be9b-473b-ab1b-621a4f575d51",
				Assets: []AssetWithAnnotationsRequest{
					AssetWithAnnotationsRequest{
						AssetRequest: AssetRequest{
							Identifier: "fancy.vulcan.example.com",
							Type:       "WRONG",
						},
					},
				},
				GroupName: "whatever-discovered-assets",
			},
			wantErr: errors.New("Invalid asset type (WRONG) for asset (fancy.vulcan.example.com)"),
		},
		{
			name: "Fails if more than one group matches the name",
			req: &DiscoveredAssetsRequest{
				TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
				GroupName: "coincident-discovered-assets",
			},
			wantErr: errors.New("more than one group matches the name coincident-discovered-assets"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestMergeDiscoveredAssetEndpointGroupCreation(t *testing.T) {
	const teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	oldAPIGroups, err := testService.ListGroups(context.Background(), teamID, "")
	if err != nil {
		t.Fatal(err)
	}

	var oldGroups []string
	for _, group := range oldAPIGroups {
		oldGroups = append(oldGroups, group.Name)
	}
	sort.Strings(oldGroups)

	tests := []struct {
		req     interface{}
		name    string
		want    interface{}
		wantErr error
	}{
		{
			name: "Group is not created it exists",
			req: &DiscoveredAssetsRequest{
				TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
				GroupName: "security-discovered-assets",
			},
			want:    oldGroups,
			wantErr: nil,
		},
		{
			name: "Group is created if it doesn't exist",
			req: &DiscoveredAssetsRequest{
				TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
				GroupName: "zzz-new-discovered-assets",
			},
			want:    append(oldGroups, "zzz-new-discovered-assets"),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err = makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			newAPIGroups, err := testService.ListGroups(context.Background(), teamID, "")
			if err != nil {
				t.Fatal(err)
			}

			var newGroups []string
			for _, group := range newAPIGroups {
				newGroups = append(newGroups, group.Name)
			}
			sort.Strings(newGroups)

			diff = cmp.Diff(tt.want, newGroups)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

// TestMergeDiscoveredAssetEndpointAssetsCreated checks that new assets are
// created, associated with the group, have the correct annotations, scannable
// field and other options.
func TestMergeDiscoveredAssetEndpointAssetsCreated(t *testing.T) {
	const (
		teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"
		// empty-discovered-assets
		groupID = "5296b879-cb7c-4372-bd65-c5a17152b10b"
	)

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	oldAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	oldAssets := make(map[string]bool)
	for _, asset := range oldAPIAssets {
		oldAssets[asset.ID] = true
	}

	req := &DiscoveredAssetsRequest{
		TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
		GroupName: "empty-discovered-assets",
		Assets: []AssetWithAnnotationsRequest{
			AssetWithAnnotationsRequest{
				AssetRequest: AssetRequest{
					Identifier:        "new.vulcan.example.com",
					Type:              "Hostname",
					Options:           common.String(`{}`),
					Scannable:         common.Bool(false),
					EnvironmentalCVSS: common.String("a.b.c.d"),
					ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
				},
				Annotations: api.AssetAnnotationsMap{
					"whateverkey": "whatevervalue",
				},
			},
		},
	}
	prefix := fmt.Sprintf("%s/empty", service.GenericAnnotationsPrefix)
	wantSize := len(oldAPIAssets) + 1
	wantAnnotations := api.AssetAnnotationsMap{
		fmt.Sprintf("%s/whateverkey", prefix): "whatevervalue",
	}
	wantROLFP := api.ROLFP{0, 0, 0, 0, 0, 1, false}
	wantCVSS := "a.b.c.d"

	_, err = makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	newAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	// Asset has been created
	if wantSize != len(newAPIAssets) {
		t.Fatalf("not all the assets were created: want(%d) got(%d)", wantSize, len(newAPIAssets))
	}

	for _, asset := range newAPIAssets {
		// Skip assets that were previously created.
		if oldAssets[asset.ID] {
			continue
		}

		// Check that the identifier and type match.
		if asset.Identifier != "new.vulcan.example.com" {
			t.Fatalf("asset identifier does not match: want(new.vulcan.example.com) got(%v)", asset.Identifier)
		}
		if asset.AssetType.Name != "Hostname" {
			t.Fatalf("asset type does not match: want(Hostname) got(%v)", asset.AssetType.Name)
		}

		// Check that annotations are correct.
		gotAnnotations := api.AssetAnnotations(asset.AssetAnnotations).ToMap()
		// Prefix set to "" because there shouldn't be annotations without
		// prefix either, as assets are new and have been created by the
		// discovery process.
		if !wantAnnotations.Matches(gotAnnotations, "") {
			t.Fatalf("asset annotations does not match: want(%v) got(%v)", wantAnnotations, gotAnnotations)
		}

		// Check that assets is not scannable.
		if *asset.Scannable {
			t.Error("asset shouldn't be scannable")
		}

		// Check other options.
		if *asset.Options != "{}" {
			t.Errorf("asset options: want({}) got(%v)", asset.Options)
		}
		if !cmp.Equal(*asset.ROLFP, wantROLFP) {
			t.Errorf("asset ROLFP: want(%v) got(%v)", wantROLFP, *asset.ROLFP)
		}
		if !cmp.Equal(*asset.EnvironmentalCVSS, wantCVSS) {
			t.Errorf("asset EnviromentalCVSS: want(%v) got(%v)", wantCVSS, *asset.EnvironmentalCVSS)
		}
	}
}

func TestMergeDiscoveredAssetEndpointAssetsAssociated(t *testing.T) {
	const (
		teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"
		// empty-discovered-assets
		groupID = "5296b879-cb7c-4372-bd65-c5a17152b10b"
		// default.vulcan.example.com (Hostname)
		assetID = "6ace4174-e704-4ea3-9c68-9096966a7e61"
	)

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	oldAPIGroupAssets, err := testService.ListAssetGroup(context.Background(), api.AssetGroup{GroupID: groupID}, teamID)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldAPIGroupAssets) != 0 {
		t.Fatalf("group is not empty: %d", len(oldAPIGroupAssets))
	}

	req := &DiscoveredAssetsRequest{
		TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
		GroupName: "empty-discovered-assets",
		Assets: []AssetWithAnnotationsRequest{
			AssetWithAnnotationsRequest{
				AssetRequest: AssetRequest{
					Identifier: "default.vulcan.example.com",
					Type:       "Hostname",
				},
			},
		},
	}

	_, err = makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	newAPIGroupAssets, err := testService.ListAssetGroup(context.Background(), api.AssetGroup{GroupID: groupID}, teamID)
	if err != nil {
		t.Fatal(err)
	}

	if len(newAPIGroupAssets) != 1 {
		t.Fatalf("group size is not correct: want(%d) got(%d)", 1, len(newAPIGroupAssets))
	}

	if id := newAPIGroupAssets[0].ID; id != assetID {
		t.Fatalf("asset ID does not match: want(%v) got(%v)", assetID, id)
	}
}

func TestMergeDiscoveredAssetEndpointAssetsUpdated(t *testing.T) {
	const (
		teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"
		// security-discovered-assets
		groupID = "1a893ae9-0340-48ff-a5ac-95408731c80b"
		// scannable.vulcan.example.com (Hostname)
		scannableID = "aeb51c5c-7732-444d-9519-55a5108809f9"
		// nonscannable.vulcan.example.com (Hostname)
		nonScannableID = "73e33dcb-d07c-41d1-bc32-80861b49941e"
	)

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	oldAPIGroupAssets, err := testService.ListAssetGroup(context.Background(), api.AssetGroup{GroupID: groupID}, teamID)
	if err != nil {
		t.Fatal(err)
	}

	if len(oldAPIGroupAssets) != 2 {
		t.Fatalf("group size is not correct: want(2) got(%d)", len(oldAPIGroupAssets))
	}

	for _, asset := range oldAPIGroupAssets {
		switch asset.ID {
		case scannableID:
			if !*asset.Scannable {
				t.Fatal("scannable asset is marked as non-scannable")
			}
		case nonScannableID:
			if *asset.Scannable {
				t.Fatal("nonscannable asset is marked as scannable")
			}

			annotations, err := testService.ListAssetAnnotations(context.Background(), teamID, asset.ID)
			if err != nil {
				t.Fatal(err)
			}
			annotationsMap := api.AssetAnnotations(annotations).ToMap()

			expectedAnnotations := api.AssetAnnotationsMap{
				"keywithoutprefix":                      "valuewithoutprefix",
				"autodiscovery/security/keytoupdate":    "valuetoupdate",
				"autodiscovery/security/keytonotupdate": "valuetonotupdate",
				"autodiscovery/security/keytodelete":    "valuetodelete",
			}
			if fmt.Sprintf("%v", expectedAnnotations) != fmt.Sprintf("%v", annotationsMap) {
				t.Fatalf("unexpected annotations: want(%v) got(%v)", expectedAnnotations, annotationsMap)
			}

		}
	}

	req := &DiscoveredAssetsRequest{
		TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
		GroupName: "security-discovered-assets",
		Assets: []AssetWithAnnotationsRequest{
			AssetWithAnnotationsRequest{
				AssetRequest: AssetRequest{
					Identifier: "scannable.vulcan.example.com",
					Type:       "Hostname",
					Scannable:  common.Bool(false),
				},
			},
			AssetWithAnnotationsRequest{
				AssetRequest: AssetRequest{
					Identifier: "nonscannable.vulcan.example.com",
					Type:       "Hostname",
					Scannable:  common.Bool(true),
				},
				Annotations: api.AssetAnnotationsMap{
					"keytoupdate":    "newvalue",
					"keytonotupdate": "valuetonotupdate",
				},
			},
		},
	}
	wantAnnotations := api.AssetAnnotationsMap{
		"keywithoutprefix":                      "valuewithoutprefix",
		"autodiscovery/security/keytoupdate":    "newvalue",
		"autodiscovery/security/keytonotupdate": "valuetonotupdate",
	}

	_, err = makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	newAPIGroupAssets, err := testService.ListAssetGroup(context.Background(), api.AssetGroup{GroupID: groupID}, teamID)
	if err != nil {
		t.Fatal(err)
	}

	if len(newAPIGroupAssets) != 2 {
		t.Fatalf("group size is not correct: want(%d) got(%d)", 2, len(newAPIGroupAssets))
	}

	for _, asset := range newAPIGroupAssets {
		switch asset.ID {
		case scannableID:
			if *asset.Scannable {
				t.Fatal("scannable asset is marked as scannable")
			}
		// Assets marked as non-scannable shouldn't be updated to scannable by
		// the discovery process, to avoid automatically marking assets as
		// scannable that were previously marked as non-scannable through the
		// UI.
		case nonScannableID:
			if *asset.Scannable {
				t.Fatal("nonscannable asset is marked as scannable")
			}

			annotations, err := testService.ListAssetAnnotations(context.Background(), teamID, asset.ID)
			if err != nil {
				t.Fatal(err)
			}

			annotationsMap := api.AssetAnnotations(annotations).ToMap()
			if fmt.Sprintf("%v", wantAnnotations) != fmt.Sprintf("%v", annotationsMap) {
				t.Fatalf("unexpected annotations: want(%v) got(%v)", wantAnnotations, annotationsMap)
			}

		default:
			t.Fatalf("unexpected asset in the group: asset ID (%v)", asset.ID)
		}
	}
}

func TestMergeDiscoveredAssetEndpointAssetsCleared(t *testing.T) {
	const (
		teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"
		// security-discovered-assets
		groupID = "1a893ae9-0340-48ff-a5ac-95408731c80b"
		// scannable.vulcan.example.com (Hostname)
		scannableID = "aeb51c5c-7732-444d-9519-55a5108809f9"
		// nonscannable.vulcan.example.com (Hostname)
		nonScannableID = "73e33dcb-d07c-41d1-bc32-80861b49941e"
	)

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	oldAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	for _, asset := range oldAPIAssets {
		switch asset.ID {
		case scannableID:
			if num := len(asset.AssetGroups); num != 1 {
				t.Fatalf("scannable asset does not belong to exactly one group: want(1) got(%v)", num)
			}
			if group := asset.AssetGroups[0].GroupID; group != groupID {
				t.Fatalf("scannable asset belongs to unexpected group: want(%v) got(%v)", groupID, group)
			}
		case nonScannableID:
			if num := len(asset.AssetGroups); num != 2 {
				t.Fatalf("non-scannable asset does not belong to exactly two groups: want(2) got(%v)", num)
			}

			annotations, err := testService.ListAssetAnnotations(context.Background(), teamID, asset.ID)
			if err != nil {
				t.Fatal(err)
			}
			annotationsMap := api.AssetAnnotations(annotations).ToMap()

			expectedAnnotations := api.AssetAnnotationsMap{
				"keywithoutprefix":                      "valuewithoutprefix",
				"autodiscovery/security/keytoupdate":    "valuetoupdate",
				"autodiscovery/security/keytonotupdate": "valuetonotupdate",
				"autodiscovery/security/keytodelete":    "valuetodelete",
			}
			if fmt.Sprintf("%v", expectedAnnotations) != fmt.Sprintf("%v", annotationsMap) {
				t.Fatalf("unexpected annotations: want(%v) got(%v)", expectedAnnotations, annotationsMap)
			}
		}
	}

	req := &DiscoveredAssetsRequest{
		TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
		GroupName: "security-discovered-assets",
		Assets:    []AssetWithAnnotationsRequest{},
	}
	wantAnnotations := api.AssetAnnotationsMap{
		"keywithoutprefix": "valuewithoutprefix",
	}

	_, err = makeMergeDiscoveredAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	newAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	for _, asset := range newAPIAssets {
		switch asset.ID {
		case scannableID:
			t.Fatal("scannable asset has not been deleted")
		case nonScannableID:
			if num := len(asset.AssetGroups); num != 1 {
				t.Fatalf("non-scannable asset does not belong to exactly one groups: want(1) got(%v)", num)
			}

			if group := asset.AssetGroups[0].GroupID; group == groupID {
				t.Fatalf("non-scannable asset has not been deassociated")
			}

			annotations, err := testService.ListAssetAnnotations(context.Background(), teamID, asset.ID)
			if err != nil {
				t.Fatal(err)
			}
			annotationsMap := api.AssetAnnotations(annotations).ToMap()

			if fmt.Sprintf("%v", wantAnnotations) != fmt.Sprintf("%v", annotationsMap) {
				t.Fatalf("unexpected annotations: want(%v) got(%v)", wantAnnotations, annotationsMap)
			}
		}
	}
}

func TestMakeUpdateAssetsEndpoint(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildTestService(testStore)

	tests := []struct {
		req            interface{}
		name           string
		want           interface{}
		wantErr        error
		wantClassified bool
	}{
		{
			name: "HappyPath",
			req: &AssetRequest{
				ID:                "13376826-14ec-4e85-a5a4-e2decdfbc193",
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Scannable:         common.Bool(false),
				EnvironmentalCVSS: common.String("10"),
			},
			want: Ok{api.AssetResponse{
				Identifier: "foo3.vulcan.example.com",
				AssetType: api.AssetTypeResponse{
					ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
				},
				Options:           common.String(`{}`),
				Scannable:         common.Bool(false),
				EnvironmentalCVSS: common.String("10"),
				ROLFP:             api.DefaultROLFP,
			}},
			wantErr:        nil,
			wantClassified: false,
		},
		{
			name: "RefreshROLFP",
			req: &AssetRequest{
				ID:                "13376826-14ec-4e85-a5a4-e2decdfbc193",
				TeamID:            "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
				Scannable:         common.Bool(false),
				EnvironmentalCVSS: common.String("10"),
				ROLFP: &api.ROLFP{
					Reputation: 0,
					Operation:  0,
					Legal:      0,
					Financial:  1,
					Personal:   0,
					Scope:      2,
					IsEmpty:    false,
				},
			},
			want: Ok{api.AssetResponse{
				Identifier: "foo3.vulcan.example.com",
				AssetType: api.AssetTypeResponse{
					ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
				},
				Options:           common.String(`{}`),
				Scannable:         common.Bool(false),
				EnvironmentalCVSS: common.String("10"),
				ROLFP: &api.ROLFP{
					Reputation: 0,
					Operation:  0,
					Legal:      0,
					Financial:  1,
					Personal:   0,
					Scope:      2,
					IsEmpty:    false,
				},
			}},
			wantErr:        nil,
			wantClassified: true,
		},
	}

	testTimeRef := time.Now()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeUpdateAssetEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{cmpopts.IgnoreFields(api.AssetResponse{}, "ID", "AssetType", "ClassifiedAt")})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}

			if tt.want != nil {
				assetResp := got.(Ok).Data.(api.AssetResponse)
				if tt.wantClassified {
					if assetResp.ClassifiedAt == nil || testTimeRef.After(*assetResp.ClassifiedAt) {
						t.Fatalf("ClassifiedAt timestamp has not been updated")
					}
				} else {
					if assetResp.ClassifiedAt != nil && testTimeRef.Before(*assetResp.ClassifiedAt) {
						t.Fatalf("ClassifiedAt timestamp was updated but it didn't have to")
					}
				}
			}
		})
	}
}
