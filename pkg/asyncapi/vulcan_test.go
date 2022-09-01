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
	"Asset1": AssetPayload{
		Id: "Asset1",
		Team: &Team{
			Id:          "Team1",
			Name:        "Team1",
			Description: "The one",
			Tag:         "tag1",
		},
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
	ID      string
	Entity  string
	Content []byte
}

type inMemStreamClient struct {
	payloads []streamPayload
}

func (i *inMemStreamClient) Push(entity string, id string, payload []byte) error {
	i.payloads = append(i.payloads, streamPayload{
		ID:      id,
		Entity:  entity,
		Content: payload,
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
					ID:      assetFixtures["Asset1"].Team.Id + "/" + assetFixtures["Asset1"].Id,
					Entity:  AssetsEntityName,
					Content: mustJSONMarshal(assetFixtures["Asset1"]),
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

func mustJSONMarshal(assset AssetPayload) []byte {
	content, err := json.Marshal(assset)
	if err != nil {
		panic(err)
	}
	return content
}
