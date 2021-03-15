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

var (
	mockOpDeleteTeamData []byte
	mockOpDeleteTeamDTO  = OpDeleteTeamDTO{
		Team: api.Team{
			ID:  "t1",
			Tag: "mockDeleteTeamTag",
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

	mockOpDeleteAllAssetsData []byte
	mockOpDeleteAllAssetsDTO  = OpDeleteAllAssetsDTO{
		Team: api.Team{
			ID:  "t2",
			Tag: "mockDeleteAllAssetsTag",
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
	deleteTagF       func(ctx context.Context, authTag, tag string) error
	deleteTargetTagF func(ctx context.Context, authTag, targetID, tag string) error
}

func (m *mockVulnDBClient) Targets(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
	return m.targetsF(ctx, params, pagination)
}
func (m *mockVulnDBClient) DeleteTag(ctx context.Context, authTag, tag string) error {
	return m.deleteTagF(ctx, authTag, tag)
}
func (m *mockVulnDBClient) DeleteTargetTag(ctx context.Context, authTag, targetID, tag string) error {
	return m.deleteTargetTagF(ctx, authTag, targetID, tag)
}

func init() {
	var err error
	mockOpDeleteTeamData, err = json.Marshal(mockOpDeleteTeamDTO)
	if err != nil {
		panic("Err setting up test")
	}
	mockOpDeleteAssetData, err = json.Marshal(mockOpDeleteAssetDTO)
	if err != nil {
		panic("Err setting up test")
	}
	mockOpDeleteAllAssetsData, err = json.Marshal(mockOpDeleteAllAssetsDTO)
	if err != nil {
		panic("Err setting up test")
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
					Operation: opDeleteAsset,
					DTO:       mockOpDeleteAssetData,
				},
				Outbox{
					Operation: opDeleteAllAssets,
					DTO:       mockOpDeleteAllAssetsData,
				},
			},
			vulnDBClient: &mockVulnDBClient{
				targetsF: func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
					return &api.TargetsList{Targets: []vulndb.Target{{}}}, nil
				},
				deleteTagF: func(ctx context.Context, authTag, tag string) error {
					return nil
				},
				deleteTargetTagF: func(ctx context.Context, authTag, targetID, tag string) error {
					return nil
				},
			},
			wantNParsed: 3,
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
			parser := NewVulnDBTxParser(tc.vulnDBClient, tc.loggr)
			nParsed := parser.Parse(tc.log)
			if nParsed != tc.wantNParsed {
				t.Fatalf("expected nParsed to be %d, but got %d", tc.wantNParsed, nParsed)
			}
		})
	}
}
