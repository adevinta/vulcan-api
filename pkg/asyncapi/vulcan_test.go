/*
Copyright 2022 Adevinta
*/

package asyncapi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var assetFixtures = map[string]AssetPayload{
	"Asset1": {
		Id:         "Asset1",
		Identifier: "example.com",
		AssetType:  (*AssetType)(strToPtr(AssetTypeDomainName)),
		Team: &Team{
			Id:          "Team1",
			Name:        "Team1",
			Description: "The one",
			Tag:         "tag1",
		},
	},
}

var findingFixtures = map[string]FindingPayload{
	"Finding1": {
		AffectedResource: "AffectedResource1",
		CurrentExposure:  10,
		Details:          "Details1",
		Id:               "FindingId1",
		ImpactDetails:    "ImpactDetails1",
		Issue: &Issue{
			CweId:       1,
			Description: "Description1",
			Id:          "IssueId1",
			Labels: []interface{}{[]string{
				"Label1",
				"Label2",
			}},
			Recommendations: []interface{}{[]string{
				"Recommendation1",
				"Recommendation2",
			}},
			ReferenceLinks: []interface{}{[]string{
				"ReferenceLink1",
				"ReferenceLink2",
			}},
			Summary: "Summary1",
		},
		Score: 7.0,
		Source: &Source{
			Component: "Component1",
			Id:        "SourceId1",
			Instance:  "SourceInstance1",
			Name:      "SourceName1",
			Options:   "SourceOptions1",
			Time:      "SourceTime1",
		},
		Status: "OPEN",
		Target: &Target{
			Id:         "TargetId1",
			Identifier: "TargetIdentifier1",
			Teams: []interface{}{[]string{
				"Team1",
				"Team2",
			}},
		},
		TotalExposure: 50,
	},
}

type nullLogger struct {
}

func (n nullLogger) Errorf(s string, params ...any) {
}

func (n nullLogger) Infof(s string, params ...any) {

}

func (n nullLogger) Debugf(s string, params ...any) {

}

type streamPayload struct {
	ID       string
	Entity   string
	Content  []byte
	Metadata map[string][]byte
}

type inMemStreamClient struct {
	payloads []streamPayload
}

func (i *inMemStreamClient) Push(entity string, id string, payload []byte, metadata map[string][]byte) error {
	i.payloads = append(i.payloads, streamPayload{
		ID:       id,
		Entity:   entity,
		Content:  payload,
		Metadata: metadata,
	})
	return nil
}

func TestVulcan_PushAsset(t *testing.T) {
	tests := []struct {
		name    string
		client  *inMemStreamClient
		logger  Logger
		asset   AssetPayload
		want    []streamPayload
		wantErr bool
	}{
		{
			name: "PushesAssets",
			client: &inMemStreamClient{
				payloads: []streamPayload{},
			},
			logger: nullLogger{},
			asset:  assetFixtures["Asset1"],
			want: []streamPayload{
				{
					ID:       assetFixtures["Asset1"].Team.Id + "/" + assetFixtures["Asset1"].Id,
					Entity:   AssetsEntityName,
					Content:  mustJSONMarshal(assetFixtures["Asset1"]),
					Metadata: metadata(assetFixtures["Asset1"]),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vulcan{
				client: tt.client,
				logger: tt.logger,
			}
			if err := v.PushAsset(tt.asset); (err != nil) != tt.wantErr {
				t.Errorf("Vulcan.PushAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
			sortOpts := cmpopts.SortSlices(func(a, b streamPayload) bool {
				return strings.Compare(a.ID, b.ID) < 0
			})
			got := tt.client.payloads
			diff := cmp.Diff(tt.want, got, sortOpts)
			if diff != "" {
				t.Fatalf("want!=got, diff: %s", diff)
			}
		})
	}
}

func TestVulcan_PushFinding(t *testing.T) {
	tests := []struct {
		name    string
		client  *inMemStreamClient
		logger  Logger
		finding FindingPayload
		want    []streamPayload
		wantErr bool
	}{
		{
			name: "PushesFindings",
			client: &inMemStreamClient{
				payloads: []streamPayload{},
			},
			logger:  nullLogger{},
			finding: findingFixtures["Finding1"],
			want: []streamPayload{
				{
					ID:      findingFixtures["Finding1"].Id,
					Entity:  FindingsEntityName,
					Content: mustJSONMarshal(findingFixtures["Finding1"]),
					Metadata: map[string][]byte{
						"version": []byte(Version),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Vulcan{
				client: tt.client,
				logger: tt.logger,
			}
			if err := v.PushFinding(tt.finding); (err != nil) != tt.wantErr {
				t.Errorf("Vulcan.PushFinding() error = %v, wantErr %v", err, tt.wantErr)
			}
			sortOpts := cmpopts.SortSlices(func(a, b streamPayload) bool {
				return strings.Compare(a.ID, b.ID) < 0
			})
			got := tt.client.payloads
			diff := cmp.Diff(tt.want, got, sortOpts)
			if diff != "" {
				t.Fatalf("want!=got, diff: %s", diff)
			}
		})
	}
}

func mustJSONMarshal(payload any) []byte {
	content, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return content
}

func strToPtr(v string) *string {
	return &v
}
