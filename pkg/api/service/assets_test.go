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
	"testing"

	apierrors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/adevinta/vulcan-api/pkg/testutil"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "test.com",
					AssetTypeID: hostnameType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(""),
					ROLFP:       &api.ROLFP{IsEmpty: true},
					Alias:       "alias1",
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
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "nba.com",
					AssetTypeID: hostnameType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(""),
					ROLFP:       &api.ROLFP{IsEmpty: true},
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
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "127.0.0.1",
					AssetTypeID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:   &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "127.0.0.1/32",
					AssetTypeID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:   &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "127.0.0.2",
					AssetTypeID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40",
					AssetType:   &api.AssetType{ID: "d53a9a5a-70ca-4c71-9b0d-808b64dadc40", Name: "IP"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "localhost",
					AssetTypeID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					AssetType:   &api.AssetType{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "example.com",
					AssetTypeID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595",
					AssetType:   &api.AssetType{ID: "1937b564-bbc4-47f6-9722-b4a8c8ac0595", Name: "Hostname"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
				api.Asset{
					TeamID:      "3c7c2963-6a03-4a25-a822-ebeb237db065",
					Identifier:  "example.com",
					AssetTypeID: "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00",
					AssetType:   &api.AssetType{ID: "e2e4b23e-b72c-40a6-9f72-e6ade33a7b00", Name: "DomainName"},
					Options:     common.String(""),
					Scannable:   common.Bool(true),
					ROLFP:       &api.ROLFP{IsEmpty: true},
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
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "arn:aws:iam::123456789012:root",
					AssetTypeID: awsAccountType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(""),
					ROLFP:       &api.ROLFP{IsEmpty: true},
					Alias:       "alias1",
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
					TeamID:      "a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
					Identifier:  "arn:aws:iam::123456789011:root",
					AssetTypeID: awsAccountType.ID,
					Scannable:   common.Bool(true),
					Options:     common.String(""),
					ROLFP:       &api.ROLFP{IsEmpty: true},
				},
			},
			wantErr: nil,
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

			diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(api.Asset{}, "ID", "Team", "AssetType", "CreatedAt", "UpdatedAt", "ClassifiedAt"))
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
					AssetType:  api.AssetTypeResponse{},
					Options:    common.String(""),
					Scannable:  common.Bool(true),
					ROLFP:      &api.ROLFP{IsEmpty: true},
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
		db:     nil,
		logger: loggerAssets,
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
