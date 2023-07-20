/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	apierrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/testutil"
	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type mockMetricsClient struct {
	metrics.Client
	metrics         []metrics.Metric
	expectedMetrics []metrics.Metric
}

func (c *mockMetricsClient) Push(metric metrics.Metric) {
	c.metrics = append(c.metrics, metric)
}

// Verify verifies the matching between mock client
// expected metrics and the actual pushed metrics.
func (c *mockMetricsClient) Verify() error {
	nMetrics := len(c.metrics)
	nExpectedMetrics := len(c.expectedMetrics)

	if nMetrics != nExpectedMetrics {
		return fmt.Errorf(
			"Number of metrics do not match: Expected %d, but got %d",
			nExpectedMetrics, nMetrics)
	}

	for _, m := range c.metrics {
		var found bool
		for _, em := range c.expectedMetrics {
			if reflect.DeepEqual(m, em) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Metrics do not match: Expected %v, but got %v",
				c.expectedMetrics, c.metrics)
		}
	}

	return nil
}

func TestGetTypesFromIdentifier(t *testing.T) {
	var tests = []struct {
		identifier          string
		expectedTypes       []string
		expectedIdentifiers []string
	}{
		{"arn:aws:iam::123456789012:root", []string{"AWSAccount"}, []string{"arn:aws:iam::123456789012:root"}},
		{"192.0.2.1", []string{"IP"}, []string{"192.0.2.1"}},
		{"192.0.2.1/32", []string{"IP"}, []string{"192.0.2.1"}},
		{"192.0.2.0/24", []string{"IPRange"}, []string{"192.0.2.0/24"}},
		{"vulcan.mpi-internal.com", []string{"DomainName"}, []string{"vulcan.mpi-internal.com"}},
		{"adevinta.com", []string{"Hostname", "DomainName"}, []string{"adevinta.com", "adevinta.com"}},
		{"www.adevinta.com", []string{"Hostname"}, []string{"www.adevinta.com"}},
		{"not.a.host.name", nil, nil},
		{"containers.adevinta.com/vulcan/application:5.5.2", []string{"DockerImage"}, []string{"containers.adevinta.com/vulcan/application:5.5.2"}},
		{"registry-1.docker.io/library/postgres:latest", []string{"DockerImage"}, []string{"registry-1.docker.io/library/postgres:latest"}},
		{"finntech/docker-elasticsearch-kubernetes", nil, nil},
		{"https://www.example.com", []string{"Hostname", "WebAddress"}, []string{"www.example.com", "https://www.example.com"}},
		{"registry-1.docker.io/artifact", []string{"DockerImage"}, []string{"registry-1.docker.io/artifact"}},
	}

	for _, tt := range tests {
		assets, _ := getTypesFromIdentifier(tt.identifier)

		var identifiers []string
		var types []string
		for _, a := range assets {
			identifiers = append(identifiers, a.identifier)
			types = append(types, a.assetType)
		}

		if !reflect.DeepEqual(tt.expectedIdentifiers, identifiers) {
			t.Fatalf("for identifier %s expected identifiers to be: %v\nbut got: %v",
				tt.identifier, tt.expectedIdentifiers, identifiers)
		}
		if !reflect.DeepEqual(tt.expectedTypes, types) {
			t.Fatalf("for identifier %s expected types to be: %v\nbut got: %v",
				tt.identifier, tt.expectedTypes, types)
		}
	}
}

func TestServiceListAssets(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	hostnameType, err := testStore.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}

	domainType, err := testStore.GetAssetType("domainname")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		teamID     string
		want       []*api.Asset
		wantErr    error
	}{
		{
			name: "HappyPath",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
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
					Options:           common.String("{}"),
					Scannable:         common.Bool(true),
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
			testService := tt.srvBuilder.BuildSrv(testStore)

			got, err := testService.ListAssets(context.Background(), tt.teamID, api.Asset{})
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.Asset{}, "ID", "Team", "AssetType", "CreatedAt"))
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func buildUserVulcanitoSrvWithAWSMock(mock AWSAccounts) VulcanitoServiceBuilder {
	return func(s api.VulcanitoStore) api.VulcanitoService {
		srv := buildDefaultVulcanitoSrv(s, loggerUser).(vulcanitoService)
		srv.awsAccounts = mock
		return srv
	}
}

func buildVulcanitoServiceWithMetricsClientMock(s api.VulcanitoStore, l log.Logger, m metrics.Client) api.VulcanitoService {
	srv := buildDefaultVulcanitoSrv(s, l).(vulcanitoService)
	srv.metricsClient = m
	return srv
}

func TestVulcanitoService_CreateAssets(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	hostnameType, err := testStore.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}

	awsAccountType, err := testStore.GetAssetType("AWSAccount")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		assets     []api.Asset
		groups     []api.Group
		want       []api.Asset
		wantErr    error
	}{
		{
			name: "Happy path, Default group",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "test.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Alias:      "alias1",
				},
			},
			groups: []api.Group{},
			want: []api.Asset{
				api.Asset{
					TeamID:           "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:       "test.com",
					AssetTypeID:      hostnameType.ID,
					Scannable:        common.Bool(true),
					Options:          common.String(""),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					Alias:            "alias1",
					AssetAnnotations: []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should add asset to custom group",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nba.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{
				api.Group{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want: []api.Asset{
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "nba.com",
					AssetTypeID:      hostnameType.ID,
					Scannable:        common.Bool(true),
					Options:          common.String(""),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should return error group does not belong to team",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nba.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
				},
			},
			groups: []api.Group{
				api.Group{
					ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want:    []api.Asset(nil),
			wantErr: errors.New("Unable to find group ab310d43-8cdf-4f65-9ee8-d1813a22bab4: record not found"),
		},
		{
			name: "Some assets have explicit asset type, others don't",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrv,
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "127.0.0.1",
					AssetType:  &api.AssetType{Name: "IP"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "127.0.0.1/32",
					AssetType:  &api.AssetType{Name: "IP"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "127.0.0.2",
					AssetType:  &api.AssetType{},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "localhost",
					AssetType:  &api.AssetType{},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "example.com",
					AssetType:  &api.AssetType{},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{},
			want: []api.Asset{
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "127.0.0.1",
					AssetTypeID:      "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:        &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "127.0.0.1/32",
					AssetTypeID:      "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:        &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "127.0.0.2",
					AssetTypeID:      "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:        &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "localhost",
					AssetTypeID:      "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					AssetType:        &api.AssetType{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "example.com",
					AssetTypeID:      "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					AssetType:        &api.AssetType{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
				api.Asset{
					TeamID:           "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:       "example.com",
					AssetTypeID:      "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00",
					AssetType:        &api.AssetType{ID: "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00", Name: "DomainName"},
					Options:          common.String(""),
					Scannable:        common.Bool(true),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},
		{
			name: "AWS Account happy path",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{
						"123456789012": "alias1",
					},
				}),
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "arn:aws:iam::123456789012:root",
					AssetType:  &api.AssetType{Name: "AWSAccount"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{},
			want: []api.Asset{
				{
					TeamID:           "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:       "arn:aws:iam::123456789012:root",
					AssetTypeID:      awsAccountType.ID,
					Scannable:        common.Bool(true),
					Options:          common.String(""),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					Alias:            "alias1",
					AssetAnnotations: []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},

		{
			name: "AWSAccAliasEmpty",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{
						"123456789012": "alias1",
					},
				}),
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "arn:aws:iam::123456789011:root",
					AssetType:  &api.AssetType{Name: "AWSAccount"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{},
			want: []api.Asset{
				{
					TeamID:           "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:       "arn:aws:iam::123456789011:root",
					AssetTypeID:      awsAccountType.ID,
					Scannable:        common.Bool(true),
					Options:          common.String(""),
					ROLFP:            &api.ROLFP{IsEmpty: true},
					AssetAnnotations: []*api.AssetAnnotation{},
				},
			},
			wantErr: nil,
		},
		{
			name: "Malformed ARN format",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{
						"123456789012": "alias1",
					},
				}),
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "arn:aws:iam:: 123456789011:root",
					AssetType:  &api.AssetType{Name: "AWSAccount"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups:  []api.Group{},
			want:    []api.Asset(nil),
			wantErr: errors.New("Identifier is not a valid AWSAccount"),
		},
		{
			name: "Malformed ARN format asset type not provided",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{
						"123456789012": "alias1",
					},
				}),
			},
			assets: []api.Asset{
				api.Asset{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "arn:aws:iam:: 123456789011:root",
					AssetType:  &api.AssetType{Name: ""},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups:  []api.Group{},
			want:    []api.Asset(nil),
			wantErr: errors.New("[asset][arn:aws:iam:: 123456789011:root][] Identifier is not a valid AWSAccount"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)

			got, err := testService.CreateAssets(context.Background(), tt.assets, tt.groups, []*api.AssetAnnotation{})
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			sortSlices := cmpopts.SortSlices(func(a, b api.Asset) bool {
				if a.TeamID != b.TeamID {
					return strings.Compare(a.TeamID, b.TeamID) < 0
				}
				if a.Identifier != b.Identifier {
					return strings.Compare(a.Identifier, b.Identifier) < 0
				}
				return strings.Compare(a.AssetTypeID, b.AssetTypeID) < 0
			})
			ignoreFields := cmpopts.IgnoreFields(api.Asset{}, "ID", "Team", "AssetType", "CreatedAt", "UpdatedAt", "ClassifiedAt")
			diff := cmp.Diff(tt.want, got, ignoreFields, sortSlices)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestVulcanitoService_CreateAssetsMultiStatus(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	hostnameType, err := testStore.GetAssetType("hostname")
	if err != nil {
		t.Fatal(err)
	}
	domainNameType, err := testStore.GetAssetType("DomainName")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		srvBuilder VulcanitoServiceTestArgs
		assets     []api.Asset
		groups     []api.Group
		want       []api.AssetCreationResponse
		wantErr    error
	}{
		{
			name: "Happy path, Default group",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{},
				}),
			},
			assets: []api.Asset{
				{
					TeamID:     "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier: "test.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{},
			want: []api.AssetCreationResponse{
				{
					Identifier: "test.com",
					AssetType:  hostnameType.ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should add asset to custom group",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{},
				}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nba.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{
				{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want: []api.AssetCreationResponse{
				{
					Identifier: "nba.com",
					AssetType:  hostnameType.ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should return error group does not belong to team",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{},
				}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nba.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Scannable:  common.Bool(true),
				},
			},
			groups: []api.Group{
				{
					ID:     "ab310d43-8cdf-4f65-9ee8-d1813a22bab4",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want:    []api.AssetCreationResponse(nil),
			wantErr: errors.New("Unable to find group ab310d43-8cdf-4f65-9ee8-d1813a22bab4: record not found"),
		},
		{
			name: "Ensure Asset Types are being tested",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{"asdasda": "alias1"},
				}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "adevinta.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "sfsdafdsfafsd",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "asdasda",
					Alias:      "alias1",
					AssetType:  &api.AssetType{Name: "AWSAccount"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "arn:aws:iam::123456789012:root",
					AssetType:  &api.AssetType{Name: "AWSAccount"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "1.2.3.4",
					AssetType:  &api.AssetType{Name: "IP"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "1.2.3",
					AssetType:  &api.AssetType{Name: "IP"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			want: []api.AssetCreationResponse{
				{
					Identifier: "adevinta.com",
					AssetType:  (&api.AssetType{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
				{
					Identifier: "sfsdafdsfafsd",
					AssetType:  (&api.AssetType{ID: "", Name: "Hostname"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: &apierrors.ErrorStack{
						Errors: []apierrors.Error{
							{
								Kind:           fmt.Errorf("Validation"),
								Message:        "Identifier is not a valid Hostname",
								HTTPStatusCode: 422,
							},
						},
					},
				},
				{
					Identifier: "asdasda",
					AssetType:  (&api.AssetType{Name: "AWSAccount"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Alias:      "alias1",
					Status: &apierrors.ErrorStack{
						Errors: []apierrors.Error{
							{
								Kind:           fmt.Errorf("Validation"),
								Message:        "Identifier is not a valid AWSAccount",
								HTTPStatusCode: 422,
							},
						},
					},
				},
				{
					Identifier: "arn:aws:iam::123456789012:root",
					AssetType:  (&api.AssetType{ID: "4347384a-88f8-11e8-9a94-a6cf71072f73", Name: "AWSAccount"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
				{
					Identifier: "1.2.3.4",
					AssetType:  (&api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
				{
					Identifier: "1.2.3",
					AssetType:  (&api.AssetType{Name: "IP"}).ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: &apierrors.ErrorStack{
						Errors: []apierrors.Error{
							{
								Kind:           fmt.Errorf("Validation"),
								Message:        "Identifier is not a valid IP",
								HTTPStatusCode: 422,
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should allow auto detect assets",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{},
				}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "adevinta.com",
					AssetType:  &api.AssetType{Name: ""},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{
				{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want: []api.AssetCreationResponse{
				{
					Identifier: "adevinta.com",
					AssetType:  hostnameType.ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
				{
					Identifier: "adevinta.com",
					AssetType:  domainNameType.ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should fail for auto detect asset",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{}}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "thisprobablydoesnotexistawejnsdgseqseg.com",
					AssetType:  &api.AssetType{Name: ""},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{
				{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want: []api.AssetCreationResponse{
				{
					Identifier: "thisprobablydoesnotexistawejnsdgseqseg.com",
					AssetType:  api.AssetTypeResponse{},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: &apierrors.ErrorStack{
						Errors: []apierrors.Error{
							{
								Kind:           fmt.Errorf("Validation"),
								Message:        "cannot parse asset type from identifier",
								HTTPStatusCode: 422,
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should fail for auto detect asset and insert typed asset",
			srvBuilder: VulcanitoServiceTestArgs{
				BuildSrv: buildUserVulcanitoSrvWithAWSMock(cgCatalogueMock{
					accounts: map[string]string{}}),
			},
			assets: []api.Asset{
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nfl.com",
					AssetType:  &api.AssetType{Name: ""},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
				{
					TeamID:     "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier: "nhl.com",
					AssetType:  &api.AssetType{Name: "Hostname"},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
				},
			},
			groups: []api.Group{
				{
					ID:     "721d1c6b-f559-4c56-8ea5-ca1820173a3c",
					TeamID: "3c7c2963-6a03-4a25-a822-ebeb237db065",
				},
			},
			want: []api.AssetCreationResponse{
				{
					Identifier: "nfl.com",
					AssetType: api.AssetTypeResponse{
						Name: "Hostname",
					},
					Options:   common.String(""),
					Scannable: common.Bool(true),
					ROLFP:     &api.ROLFP{IsEmpty: true},
					Status: &apierrors.ErrorStack{
						Errors: []apierrors.Error{
							{
								Kind:           fmt.Errorf("Duplicated record"),
								Message:        "pq: duplicate key value violates unique constraint \"unique_asset_group\"",
								HTTPStatusCode: 409,
							},
						},
					},
				},
				{
					Identifier: "nfl.com",
					AssetType: api.AssetTypeResponse{
						ID:   "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00",
						Name: "DomainName",
					},
					Options:   common.String(""),
					Scannable: common.Bool(true),
					ROLFP:     &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
				{
					Identifier: "nhl.com",
					AssetType:  hostnameType.ToResponse(),
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
					Status: api.Status{
						Code: 201,
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testService := tt.srvBuilder.BuildSrv(testStore)

			got, err := testService.CreateAssetsMultiStatus(context.Background(), tt.assets, tt.groups, []*api.AssetAnnotation{})
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}

			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.AssetCreationResponse{}, "ID", "Status", "ClassifiedAt"))
			if diff != "" {
				t.Errorf("%v\n", diff)
			}

			if len(tt.want) != len(got) {
				t.Fatalf("Array size is different: want %v, got %v\n", len(tt.want), len(got))
			}

			for i := 0; i < len(tt.want); i++ {
				wantStatus := -1
				wantMsg := ""

				gotStatus := -1
				gotMsg := ""

				if _, ok := got[i].Status.(*apierrors.ErrorStack); ok {
					gotStatus = got[i].Status.(*apierrors.ErrorStack).Errors[0].HTTPStatusCode
					gotMsg = got[i].Status.(*apierrors.ErrorStack).Errors[0].Message
				}

				if _, ok := got[i].Status.(api.Status); ok {
					gotStatus = got[i].Status.(api.Status).Code
				}

				if _, ok := tt.want[i].Status.(*apierrors.ErrorStack); ok {
					wantStatus = tt.want[i].Status.(*apierrors.ErrorStack).Errors[0].HTTPStatusCode
					wantMsg = tt.want[i].Status.(*apierrors.ErrorStack).Errors[0].Message
				}

				if _, ok := tt.want[i].Status.(api.Status); ok {
					wantStatus = tt.want[i].Status.(api.Status).Code
				}

				if wantStatus != gotStatus {
					t.Errorf("Status Code expected: %v, got: %v\n", wantStatus, gotStatus)
				}

				if wantMsg != gotMsg {
					t.Errorf("Status Message expected: %v, got: %v\n", wantMsg, gotMsg)
				}
			}
		})
	}
}

func TestMergeDiscoveredAssetsValidation(t *testing.T) {
	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildDefaultVulcanitoSrv(testStore, kitlog.NewNopLogger())

	tests := []struct {
		name      string
		teamID    string
		groupName string
		assets    []api.Asset
		wantErr   error
	}{
		{
			name:      "Fails if more than one group matches the name",
			teamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
			groupName: "coincident-discovered-assets",
			wantErr:   errors.New("more than one group matches the name coincident-discovered-assets"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testService.MergeDiscoveredAssets(context.Background(), tt.teamID, tt.assets, tt.groupName)
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(err))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}
}

func TestMergeDiscoveredAssetsGroupCreation(t *testing.T) {
	const teamID = "ea686be5-be9b-473b-ab1b-621a4f575d51"

	testStore, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer testStore.Close()

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), &mockMetricsClient{})

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
		name      string
		teamID    string
		groupName string
		assets    []api.Asset
		want      interface{}
		wantErr   error
	}{
		{
			name:      "Group is not created it exists",
			teamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
			groupName: "security-discovered-assets",
			want:      oldGroups,
			wantErr:   nil,
		},
		{
			name:      "Group is created if it doesn't exist",
			teamID:    "ea686be5-be9b-473b-ab1b-621a4f575d51",
			groupName: "zzz-new-discovered-assets",
			want:      append(oldGroups, "zzz-new-discovered-assets"),
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := testService.MergeDiscoveredAssets(context.Background(), tt.teamID, tt.assets, tt.groupName)
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

// TestMergeDiscoveredAssetsCreated checks that new assets are
// created, associated with the group, have the correct annotations, scannable
// field and other options.
func TestMergeDiscoveredAssetsCreated(t *testing.T) {
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

	expectedMetrics := []metrics.Metric{
		{
			Name:  "vulcan.discovery.created.count",
			Typ:   metrics.Count,
			Value: 1,
			Tags:  []string{"component:api"},
		},
	}

	mockClient := &mockMetricsClient{
		expectedMetrics: expectedMetrics,
	}

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), mockClient)

	oldAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	oldAssets := make(map[string]bool)
	for _, asset := range oldAPIAssets {
		oldAssets[asset.ID] = true
	}

	groupName := "empty-discovered-assets"
	assets := []api.Asset{
		{
			TeamID:            teamID,
			Identifier:        "new.vulcan.example.com",
			Options:           common.String(`{}`),
			Scannable:         common.Bool(false),
			EnvironmentalCVSS: common.String("a.b.c.d"),
			ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
			AssetAnnotations: []*api.AssetAnnotation{
				{
					Key:   "whateverkey",
					Value: "whatevervalue",
				},
			},
		},
	}

	prefix := fmt.Sprintf("%s/empty", GenericAnnotationsPrefix)
	wantSize := len(oldAPIAssets) + 1
	wantAnnotations := api.AssetAnnotationsMap{
		fmt.Sprintf("%s/whateverkey", prefix): "whatevervalue",
	}
	wantROLFP := api.ROLFP{0, 0, 0, 0, 0, 1, false}
	wantCVSS := "a.b.c.d"

	err = testService.MergeDiscoveredAssets(context.Background(), teamID, assets, groupName)
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

	if err := mockClient.Verify(); err != nil {
		t.Fatalf("Error verifying pushed metrics: %v", err)
	}
}

func TestMergeDiscoveredAssetsAssociated(t *testing.T) {
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

	expectedMetrics := []metrics.Metric{}

	mockClient := &mockMetricsClient{
		expectedMetrics: expectedMetrics,
	}

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), mockClient)

	oldAPIGroupAssets, err := testService.ListAssetGroup(context.Background(), api.AssetGroup{GroupID: groupID}, teamID)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldAPIGroupAssets) != 0 {
		t.Fatalf("group is not empty: %d", len(oldAPIGroupAssets))
	}

	groupName := "empty-discovered-assets"
	assets := []api.Asset{
		{
			TeamID:     teamID,
			Identifier: "default.vulcan.example.com",
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
		},
	}

	err = testService.MergeDiscoveredAssets(context.Background(), teamID, assets, groupName)
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

	if err := mockClient.Verify(); err != nil {
		t.Fatalf("Error verifying pushed metrics: %v", err)
	}
}

func TestMergeDiscoveredAssetsUpdated(t *testing.T) {
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

	expectedMetrics := []metrics.Metric{
		{
			Name:  "vulcan.discovery.updated.count",
			Typ:   metrics.Count,
			Value: 2,
			Tags:  []string{"component:api"},
		},
	}

	mockClient := &mockMetricsClient{
		expectedMetrics: expectedMetrics,
	}

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), mockClient)

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

	groupName := "security-discovered-assets"
	assets := []api.Asset{
		{
			TeamID:     teamID,
			Identifier: "scannable.vulcan.example.com",
			Scannable:  common.Bool(false),
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
		},
		{
			TeamID:     teamID,
			Identifier: "nonscannable.vulcan.example.com",
			Scannable:  common.Bool(true),
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
			AssetAnnotations: api.AssetAnnotationsMap{
				"keytoupdate":    "newvalue",
				"keytonotupdate": "valuetonotupdate",
			}.ToModel(),
		},
	}
	wantAnnotations := api.AssetAnnotationsMap{
		"keywithoutprefix":                      "valuewithoutprefix",
		"autodiscovery/security/keytoupdate":    "newvalue",
		"autodiscovery/security/keytonotupdate": "valuetonotupdate",
	}

	err = testService.MergeDiscoveredAssets(context.Background(), teamID, assets, groupName)
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

	if err := mockClient.Verify(); err != nil {
		t.Fatalf("Error verifying pushed metrics: %v", err)
	}
}

func TestMergeDiscoveredAssetsCleared(t *testing.T) {
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

	expectedMetrics := []metrics.Metric{
		{
			Name:  "vulcan.discovery.purged.count",
			Typ:   metrics.Count,
			Value: 1,
			Tags:  []string{"component:api"},
		},
		{
			Name:  "vulcan.discovery.dissociated.count",
			Typ:   metrics.Count,
			Value: 1,
			Tags:  []string{"component:api"},
		},
	}

	mockClient := &mockMetricsClient{
		expectedMetrics: expectedMetrics,
	}

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), mockClient)

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

	groupName := "security-discovered-assets"
	assets := []api.Asset{}
	wantAnnotations := api.AssetAnnotationsMap{
		"keywithoutprefix": "valuewithoutprefix",
	}

	err = testService.MergeDiscoveredAssets(context.Background(), teamID, assets, groupName)
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

	if err := mockClient.Verify(); err != nil {
		t.Fatalf("Error verifying pushed metrics: %v", err)
	}
}

func TestMergeDiscoveredAssetsDeduplicated(t *testing.T) {
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

	expectedMetrics := []metrics.Metric{
		{
			Name:  "vulcan.discovery.created.count",
			Typ:   metrics.Count,
			Value: 1,
			Tags:  []string{"component:api"},
		},
		{
			Name:  "vulcan.discovery.skipped.count",
			Typ:   metrics.Count,
			Value: 1,
			Tags:  []string{"component:api"},
		},
	}

	mockClient := &mockMetricsClient{
		expectedMetrics: expectedMetrics,
	}

	testService := buildVulcanitoServiceWithMetricsClientMock(testStore, kitlog.NewNopLogger(), mockClient)

	oldAPIAssets, err := testService.ListAssets(context.Background(), teamID, api.Asset{})
	if err != nil {
		t.Fatal(err)
	}

	oldAssets := make(map[string]bool)
	for _, asset := range oldAPIAssets {
		oldAssets[asset.ID] = true
	}

	groupName := "empty-discovered-assets"
	assets := []api.Asset{
		{
			TeamID:            teamID,
			Identifier:        "duplicated.vulcan.example.com",
			Options:           common.String(`{}`),
			Scannable:         common.Bool(false),
			EnvironmentalCVSS: common.String("a.b.c.d"),
			ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
		},
		{
			TeamID:            teamID,
			Identifier:        "duplicated.vulcan.example.com",
			Options:           common.String(`{}`),
			Scannable:         common.Bool(false),
			EnvironmentalCVSS: common.String("a.b.c.d"),
			ROLFP:             &api.ROLFP{0, 0, 0, 0, 0, 1, false},
			AssetType: &api.AssetType{
				Name: "Hostname",
			},
		},
	}

	wantSize := len(oldAPIAssets) + 1
	wantIdentifier := "duplicated.vulcan.example.com"
	wantType := "Hostname"

	err = testService.MergeDiscoveredAssets(context.Background(), teamID, assets, groupName)
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
		if asset.Identifier != wantIdentifier {
			t.Fatalf("asset identifier does not match: want(%s) got(%v)", wantIdentifier, asset.Identifier)
		}
		if asset.AssetType.Name != wantType {
			t.Fatalf("asset type does not match: want(%s) got(%v)", wantType, asset.AssetType.Name)
		}
	}

	if err := mockClient.Verify(); err != nil {
		t.Fatalf("Error verifying pushed metrics: %v", err)
	}
}

var (
	loggerAssets log.Logger
)

func init() {
	loggerAssets = log.NewLogfmtLogger(os.Stderr)
	loggerAssets = log.With(loggerAssets, "ts", log.DefaultTimestampUTC)
	loggerAssets = log.With(loggerAssets, "caller", log.DefaultCaller)

}

func TestVulcanitoService_CreateAssetsGroup(t *testing.T) {
	srv := vulcanitoService{
		db:            nil,
		logger:        loggerAssets,
		metricsClient: &mockMetricsClient{},
	}

	group := api.Group{}
	_, err := srv.CreateGroup(context.Background(), group)
	if err == nil {
		t.Error("Should return validation error if empty name")
	}
	expectedErrorMessage := "Key: 'Group.Name' Error:Field validation for 'Name' failed on the 'required' tag"
	diff := cmp.Diff(expectedErrorMessage, err.Error())
	if diff != "" {
		t.Errorf("Wrong error message, diff: %v", diff)
	}
}
