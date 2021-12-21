/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiErrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
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

func TestMergeDiscoveredAssetsEndpoint(t *testing.T) {
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
			name: "Happy Path",
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
			want: Accepted{
				&api.JobResponse{
					TeamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
					Operation: "MergeDiscoveredAssets",
					Status:    api.JobStatusPending,
				},
			},
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeMergeDiscoveredAssetsEndpoint(testService, kitlog.NewNopLogger())(context.Background(), tt.req)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.want, got, cmp.Options{
				cmpopts.IgnoreFields(api.JobResponse{}, "ID"),
			})
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
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
