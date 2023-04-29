/*
	deleteTargetTagF func(ctx context.Context, authTeam, targetID, tag string) error
Copyright 2021 Adevinta
*/

package cdc

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/asyncapi"
	"github.com/adevinta/vulcan-api/pkg/asyncapi/kafka"
	"github.com/adevinta/vulcan-api/pkg/testutil"
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
			AssetType: &api.AssetType{
				Name: "DomainName",
			},
			Team: &api.Team{
				ID:  "t1",
				Tag: "mockCreateAssetTag",
			},
		},
	}

	mockOpDeleteAssetData []byte
	mockOpDeleteAssetDTO  = OpDeleteAssetDTO{
		Asset: api.Asset{
			ID:         "a1",
			Identifier: "example.com",
			AssetType: &api.AssetType{
				Name: "DomainName",
			},
			Team: &api.Team{
				ID: "t1",
			},
		},
		DupAssets: 0,
	}

	mockOpUpdateAssetData []byte
	mockOpUpdateAssetDTO  = OpUpdateAssetDTO{
		OldAsset: api.Asset{
			ID:         "aO",
			Identifier: "exampleNew.com",
			AssetType: &api.AssetType{
				Name: "DomainName",
			},
			Team: &api.Team{
				ID:  "t1",
				Tag: "mockUpdateAssetTag",
			},
		},
		NewAsset: api.Asset{
			ID:         "aN",
			Identifier: "exampleOld.com",
			AssetType: &api.AssetType{
				Name: "DomainName",
			},
			Team: &api.Team{
				ID:  "t1",
				Tag: "mockUpdateAssetTag",
			},
		},
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
			TeamID:    "mockFindingOverwriteTeamID",
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
	targetsF          func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error)
	createTargetF     func(ctx context.Context, payload api.CreateTarget) (*api.Target, error)
	deleteTeamF       func(ctx context.Context, authTeam, teamID string) error
	deleteTeamTagF    func(ctx context.Context, authTeam, teamID, tag string) error
	deleteTargetTeamF func(ctx context.Context, authTeam, targetID, teamID string) error
	deleteTargetTagF  func(ctx context.Context, authTeam, targetID, tag string) error
	getFindingF       func(ctx context.Context, findingID string) (*api.Finding, error)
	updateFindingF    func(ctx context.Context, findingID string, payload *api.UpdateFinding, tag string) (*api.Finding, error)
}

func (m *mockVulnDBClient) Targets(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
	return m.targetsF(ctx, params, pagination)
}
func (m *mockVulnDBClient) CreateTarget(ctx context.Context, payload api.CreateTarget) (*api.Target, error) {
	return m.createTargetF(ctx, payload)
}
func (m *mockVulnDBClient) DeleteTeam(ctx context.Context, authTeam, teamID string) error {
	return m.deleteTeamF(ctx, authTeam, teamID)
}
func (m *mockVulnDBClient) DeleteTeamTag(ctx context.Context, authTeam, teamID, tag string) error {
	return m.deleteTeamTagF(ctx, authTeam, teamID, tag)
}
func (m *mockVulnDBClient) DeleteTargetTeam(ctx context.Context, authTeam, targetID, teamID string) error {
	return m.deleteTargetTeamF(ctx, authTeam, targetID, teamID)
}
func (m *mockVulnDBClient) DeleteTargetTag(ctx context.Context, authTeam, targetID, tag string) error {
	return m.deleteTargetTagF(ctx, authTeam, targetID, tag)
}
func (m *mockVulnDBClient) Finding(ctx context.Context, findingID string) (*api.Finding, error) {
	return m.getFindingF(ctx, findingID)
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
		name              string
		log               []Event
		vulnDBClient      *mockVulnDBClient
		asyncAPI          func() (*asyncapi.Vulcan, kafka.Client, error)
		loggr             *mockLoggr
		wantNParsed       uint
		wantAsyncAssets   []testutil.AssetTopicData
		wantAsyncFindings []testutil.FindingTopicData
		wantErr           error
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
						Teams:      payload.Teams,
					}}
					return t, nil
				},
				deleteTeamF: func(ctx context.Context, authTeam, teamID string) error {
					return nil
				},
				deleteTeamTagF: func(ctx context.Context, authTeam, teamID, tag string) error {
					return nil
				},
				deleteTargetTeamF: func(ctx context.Context, authTeam, targetID, teamID string) error {
					return nil
				},
				deleteTargetTagF: func(ctx context.Context, authTeam, targetID, tag string) error {
					return nil
				},
				getFindingF: func(ctx context.Context, findingID string) (*api.Finding, error) {
					return &api.Finding{
						Finding: vulndb.FindingExpanded{
							Finding: vulndb.Finding{ID: "1"}},
					}, nil
				},
				updateFindingF: func(ctx context.Context, findingID string, payload *api.UpdateFinding, tag string) (*api.Finding, error) {
					var f = &api.Finding{}
					return f, nil
				},
			},
			asyncAPI: newTestAsyncAPI,
			loggr:    &mockLoggr{},
			wantAsyncAssets: []testutil.AssetTopicData{
				{
					Payload: asyncapi.AssetPayload{

						Id:         "a0",
						Identifier: "somehost.com",
						AssetType:  (*asyncapi.AssetType)(strToPtr(asyncapi.AssetTypeDomainName)),
						Team: &asyncapi.Team{
							Id:  "t1",
							Tag: "mockCreateAssetTag",
						},
					},
					Headers: map[string][]byte{
						"identifier": []byte("somehost.com"),
						"type":       []byte(asyncapi.AssetTypeDomainName),
						"version":    []byte(asyncapi.Version),
					},
				},
				// Tombstone of the asset deleted.
				{
					Headers: map[string][]byte{
						"identifier": []byte("example.com"),
						"type":       []byte(asyncapi.AssetTypeDomainName),
						"version":    []byte(asyncapi.Version),
					},
				},
				{
					Payload: asyncapi.AssetPayload{
						Id:         "aN",
						Identifier: "exampleOld.com",
						AssetType:  (*asyncapi.AssetType)(strToPtr(asyncapi.AssetTypeDomainName)),
						Team: &asyncapi.Team{
							Id:  "t1",
							Tag: "mockUpdateAssetTag",
						},
					},
					Headers: map[string][]byte{
						"identifier": []byte("exampleOld.com"),
						"type":       []byte(asyncapi.AssetTypeDomainName),
						"version":    []byte(asyncapi.Version),
					},
				},
			},
			wantAsyncFindings: []testutil.FindingTopicData{
				{
					Payload: asyncapi.FindingPayload{
						Id: "1",
						Issue: &asyncapi.Issue{
							Recommendations: []any{nil},
							ReferenceLinks:  []any{nil},
							Labels:          []any{nil},
						},
						Source: &asyncapi.Source{},
						Target: &asyncapi.Target{
							Teams: []any{nil},
						},
						Resources: []any{nil},
					},
					Headers: map[string][]byte{
						"version": []byte(asyncapi.Version),
					},
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
			asyncAPI:    newTestAsyncAPI,
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
				deleteTeamF: func(ctx context.Context, authTeam, teamID string) error {
					return nil
				},
				deleteTeamTagF: func(ctx context.Context, authTeam, teamID, tag string) error {
					return nil
				},
			},
			asyncAPI:    newTestAsyncAPI,
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
				deleteTeamF: func(ctx context.Context, authTeam, teamID string) error {
					return errors.NotFound("not found")
				},
				deleteTeamTagF: func(ctx context.Context, authTeam, teamID, tag string) error {
					return errors.NotFound("not found")
				},
				targetsF: func(ctx context.Context, params api.TargetsParams, pagination api.Pagination) (*api.TargetsList, error) {
					return &api.TargetsList{Targets: []vulndb.Target{}}, nil // Returning 0 targets
				},
			},
			asyncAPI: newTestAsyncAPI,
			loggr:    &mockLoggr{},
			wantAsyncAssets: []testutil.AssetTopicData{
				// Tombstone of the asset deleted.
				{
					Headers: map[string][]byte{
						"identifier": []byte("example.com"),
						"type":       []byte(asyncapi.AssetTypeDomainName),
						"version":    []byte(asyncapi.Version),
					},
				},
			},
			wantNParsed: 2,
		},
		{
			name: "Should verify identifier matching param and return multiple targets", // This should never happen, but just in case
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
			asyncAPI: newTestAsyncAPI,
			loggr:    &mockLoggr{},
			wantAsyncAssets: []testutil.AssetTopicData{
				// Tombstone of the asset deleted.
				{
					Headers: map[string][]byte{
						"identifier": []byte("example.com"),
						"type":       []byte(asyncapi.AssetTypeDomainName),
						"version":    []byte(asyncapi.Version),
					},
				},
			},
			wantNParsed: 0,
			wantErr:     errTargetNotUnique,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asyncAPI, kclient, err := tc.asyncAPI()
			if err != nil {
				t.Fatalf("error creating the Async API: %v", err)
			}
			parser := NewAsyncTxParser(tc.vulnDBClient, &api.JobsRunner{}, asyncAPI, tc.loggr)
			nParsed := parser.Parse(tc.log)
			if nParsed != tc.wantNParsed {
				t.Fatalf("expected nParsed to be %d, but got %d", tc.wantNParsed, nParsed)
			}

			// Verify async assets
			topic := kclient.Topics[asyncapi.AssetsEntityName]
			gotAssets, err := testutil.ReadAllAssetsTopic(topic)
			if err != nil {
				t.Fatalf("error reading assets from kafka %v", err)
			}
			wantAssets := tc.wantAsyncAssets
			sortSlices := cmpopts.SortSlices(func(a, b testutil.AssetTopicData) bool {
				return strings.Compare(a.Payload.Id, b.Payload.Id) < 0
			})
			diff := cmp.Diff(wantAssets, gotAssets, sortSlices)
			if diff != "" {
				t.Fatalf("want!=got, diff: %s", diff)
			}

			// Verify async findings
			topic = kclient.Topics[asyncapi.FindingsEntityName]
			gotFindings, err := testutil.ReadAllFindingsTopic(topic)
			if err != nil {
				t.Fatalf("error reading findings from kafka %v", err)
			}
			wantFindings := tc.wantAsyncFindings
			sortSlices = cmpopts.SortSlices(func(a, b testutil.FindingTopicData) bool {
				return strings.Compare(a.Payload.Id, b.Payload.Id) < 0
			})
			diff = cmp.Diff(wantFindings, gotFindings, sortSlices)
			if diff != "" {
				t.Fatalf("want!=got, diff: %s", diff)
			}
		})
	}
}

type nullLogger struct {
}

func (n nullLogger) Errorf(s string, params ...any) {
}

func (n nullLogger) Infof(s string, params ...any) {

}

func (n nullLogger) Debugf(s string, params ...any) {

}

func newTestAsyncAPI() (*asyncapi.Vulcan, kafka.Client, error) {
	topics := map[string]string{
		asyncapi.AssetsEntityName:   "assets",
		asyncapi.FindingsEntityName: "findings",
	}
	testTopics, err := testutil.PrepareKafka(topics)
	if err != nil {
		return nil, kafka.Client{}, fmt.Errorf("error creating test topics: %v", err)
	}
	kclient, err := kafka.NewClient("", "", testutil.KafkaTestBroker, testTopics)
	if err != nil {
		return nil, kafka.Client{}, fmt.Errorf("error creating the kafka client: %v", err)
	}
	vulcan := asyncapi.NewVulcan(&kclient, nullLogger{})
	return vulcan, kclient, nil
}

func strToPtr(s string) *string {
	return &s
}
