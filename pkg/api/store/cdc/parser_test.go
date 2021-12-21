/*
Copyright 2021 Adevinta
*/

package cdc

import (
	"context"
	"encoding/json"
	errs "errors"
	"testing"

	"github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"
	vulndb "github.com/adevinta/vulnerability-db-api/pkg/model"
)

const (
	errTestSetup = "err setting up test"
)

var (
	mockOpDeleteTeamData []byte
	mockOpDeleteTeamDTO  = OpDeleteTeamDTO{
		Team: api.Team{
			ID:  "t1",
			Tag: "mockDeleteTeamTag",
		},
	}

	mockOpCreateAssetData []byte
	mockOpCreateAssetDTO  = OpCreateAssetDTO{
		Asset: api.Asset{
			ID:         "a0",
			Identifier: "somehost.com",
			Team: &api.Team{
				Tag: "mockCreateAssetTag",
			},
		},
	}

	mockOpDeleteAssetData []byte
	mockOpDeleteAssetDTO  = OpDeleteAssetDTO{
		Asset: api.Asset{
			ID:         "a1",
			Identifier: "example.com",
			Team: &api.Team{
				Tag: "mockDeleteAssetTag",
			},
		},
		DupAssets: 0,
	}

	mockOpUpdateAssetData []byte
	mockOpUpdateAssetDTO  = OpUpdateAssetDTO{
		OldAsset: api.Asset{
			ID:         "aO",
			Identifier: "exampleNew.com",
			Team: &api.Team{
				Tag: "mockUpdateAssetTag",
			},
		},
		NewAsset: api.Asset{
			ID:         "aN",
			Identifier: "exampleOld.com",
			Team: &api.Team{
				Tag: "mockUpdateAssetTag",
			},
		},
		DupAssets: 0,
	}

	mockOpDeleteAllAssetsData []byte
	mockOpDeleteAllAssetsDTO  = OpDeleteAllAssetsDTO{
		Team: api.Team{
			ID:  "t2",
			Tag: "mockDeleteAllAssetsTag",
		},
	}

	mockOpFindingOverwriteData []byte
	mockOpFindingOverwriteDTO  = OpFindingOverwriteDTO{
		FindingOverwrite: api.FindingOverwrite{
			FindingID: "f1",
			Status:    "newstatus",
			Tag:       "mockFindingOverwriteTag",
		},
	}
)

type mockLoggr struct {
	log.Logger
	keyvals []interface{}
	ncalls  int
}

func (m *mockLoggr) Log(keyvals ...interface{}) error {
	m.keyvals = append(m.keyvals, keyvals...)
	return nil
}
func (m *mockLoggr) verifyErr(targetErr error) bool {
	for i, kv := range m.keyvals {
		if str, ok := kv.(string); ok && str == "error" {
			err, ok := m.keyvals[i+1].(error)
			return ok && errs.Is(err, targetErr)
		}
	}
	return false
}

type mockVulnDBClient struct {
	vulnerabilitydb.Client
	targetsF         func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error)
	createTargetF    func(ctx context.Context, payload api.CreateTarget) (*api.Target, error)
	deleteTagF       func(ctx context.Context, authTag, tag string) error
	deleteTargetTagF func(ctx context.Context, authTag, targetID, tag string) error
	updateFindingF   func(ctx context.Context, findingID string, payload *api.UpdateFinding, tag string) (*api.Finding, error)
}

func (m *mockVulnDBClient) Targets(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
	return m.targetsF(ctx, params, pagination)
}
func (m *mockVulnDBClient) CreateTarget(ctx context.Context, payload api.CreateTarget) (*api.Target, error) {
	return m.createTargetF(ctx, payload)
}
func (m *mockVulnDBClient) DeleteTag(ctx context.Context, authTag, tag string) error {
	return m.deleteTagF(ctx, authTag, tag)
}
func (m *mockVulnDBClient) DeleteTargetTag(ctx context.Context, authTag, targetID, tag string) error {
	return m.deleteTargetTagF(ctx, authTag, targetID, tag)
}
func (m *mockVulnDBClient) UpdateFinding(ctx context.Context, findingID string, payload *api.UpdateFinding, tag string) (*api.Finding, error) {
	return m.updateFindingF(ctx, findingID, payload, tag)
}

func init() {
	var err error
	mockOpDeleteTeamData, err = json.Marshal(mockOpDeleteTeamDTO)
	if err != nil {
		panic(errTestSetup)
	}
	mockOpCreateAssetData, err = json.Marshal(mockOpCreateAssetDTO)
	if err != nil {
		panic(errTestSetup)
	}
	mockOpDeleteAssetData, err = json.Marshal(mockOpDeleteAssetDTO)
	if err != nil {
		panic(errTestSetup)
	}
	mockOpUpdateAssetData, err = json.Marshal(mockOpUpdateAssetDTO)
	if err != nil {
		panic(errTestSetup)
	}
	mockOpDeleteAllAssetsData, err = json.Marshal(mockOpDeleteAllAssetsDTO)
	if err != nil {
		panic(errTestSetup)
	}
	mockOpFindingOverwriteData, err = json.Marshal(mockOpFindingOverwriteDTO)
	if err != nil {
		panic(errTestSetup)
	}
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name         string
		log          []Event
		vulnDBClient *mockVulnDBClient
		loggr        *mockLoggr
		wantNParsed  uint
		wantErr      error
	}{
		{
			name: "Happy path",
			log: []Event{
				Outbox{
					Operation: opDeleteTeam,
					DTO:       mockOpDeleteTeamData,
				},
				Outbox{
					Operation: opCreateAsset,
					DTO:       mockOpCreateAssetData,
				},
				Outbox{
					Operation: opDeleteAsset,
					DTO:       mockOpDeleteAssetData,
				},
				Outbox{
					Operation: opUpdateAsset,
					DTO:       mockOpUpdateAssetData,
				},
				Outbox{
					Operation: opDeleteAllAssets,
					DTO:       mockOpDeleteAllAssetsData,
				},
				Outbox{
					Operation: opFindingOverwrite,
					DTO:       mockOpFindingOverwriteData,
				},
			},
			vulnDBClient: &mockVulnDBClient{
				targetsF: func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
					return &api.TargetsList{Targets: []vulndb.Target{{}}}, nil
				},
				createTargetF: func(ctx context.Context, payload api.CreateTarget) (*api.Target, error) {
					var t = &api.Target{Target: vulndb.Target{
						ID:         "1",
						Identifier: payload.Identifier,
						Tags:       payload.Tags,
					}}
					return t, nil
				},
				deleteTagF: func(ctx context.Context, authTag, tag string) error {
					return nil
				},
				deleteTargetTagF: func(ctx context.Context, authTag, targetID, tag string) error {
					return nil
				},
				updateFindingF: func(ctx context.Context, findingID string, payload *api.UpdateFinding, tag string) (*api.Finding, error) {
					var f = &api.Finding{}
					return f, nil
				},
			},
			wantNParsed: 6,
		},
		{
			name: "Should return err unsupported action",
			log: []Event{
				Outbox{
					Operation: "unknownAction",
				},
			},
			loggr:       &mockLoggr{},
			wantNParsed: 0,
			wantErr:     errUnsupportedAction,
		},
		{
			name: "Should parse 1 and error 1 due to invalid data",
			log: []Event{
				Outbox{
					Operation: opDeleteTeam,
					DTO:       mockOpDeleteTeamData,
				},
				Outbox{
					Operation: opDeleteTeam,
					DTO:       []byte("someWrongData"),
				},
				Outbox{
					Operation: opDeleteAllAssets,
					DTO:       mockOpDeleteAllAssetsData,
				},
			},
			vulnDBClient: &mockVulnDBClient{
				deleteTagF: func(ctx context.Context, authTag, tag string) error {
					return nil
				},
			},
			loggr:       &mockLoggr{},
			wantNParsed: 1,
			wantErr:     errInvalidData,
		},
		{
			name: "Should parse 2 ignoring not found data from VulnDB",
			log: []Event{
				Outbox{
					Operation: opDeleteAsset,
					DTO:       mockOpDeleteAssetData,
				},
				Outbox{
					Operation: opDeleteAllAssets,
					DTO:       mockOpDeleteAllAssetsData,
				},
			},
			vulnDBClient: &mockVulnDBClient{
				deleteTagF: func(ctx context.Context, authTag, tag string) error {
					return errors.NotFound("not found")
				},
				targetsF: func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
					return &api.TargetsList{Targets: []vulndb.Target{}}, nil // Returning 0 targets
				},
			},
			loggr:       &mockLoggr{},
			wantNParsed: 2,
		},
		{
			name: "Should verify identifiermatching param and return multiple targets", // This should never happen, but just in case
			log: []Event{
				Outbox{
					Operation: opDeleteAsset,
					DTO:       mockOpDeleteAssetData,
				},
			},
			vulnDBClient: &mockVulnDBClient{
				targetsF: func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
					if !params.IdentifierMatch {
						return nil, errs.New("identifiermatch param should be set to true")
					}
					return &api.TargetsList{Targets: []vulndb.Target{{}, {}}}, nil // Returning 0 targets
				},
			},
			loggr:       &mockLoggr{},
			wantNParsed: 0,
			wantErr:     errTargetNotUnique,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewAsyncTxParser(tc.vulnDBClient, &api.JobsRunner{}, tc.loggr)
			nParsed := parser.Parse(tc.log)
			if nParsed != tc.wantNParsed {
				t.Fatalf("expected nParsed to be %d, but got %d", tc.wantNParsed, nParsed)
			}
		})
	}
}
