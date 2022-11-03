/*
Copyright 2022 Adevinta
*/

package asyncapi

//go:generate sh -c "_gen/gen.sh ../../docs/asyncapi.yaml > models.go && go fmt models.go"

import (
	"encoding/json"
	"fmt"
	"strings"

	gokitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// LevelLogger implements the logger used by the [Vulcan] async API using an
// underlaying go-kit logger.
type LevelLogger struct {
	gokitlog.Logger
}

// Errorf logs an error message.
func (l LevelLogger) Errorf(s string, params ...any) {
	v := fmt.Sprintf(s, params...)
	level.Error(l.Logger).Log("log", v)
}

// Infof logs an information message.
func (l LevelLogger) Infof(s string, params ...any) {
	v := fmt.Sprintf(s, params...)
	level.Info(l.Logger).Log("log", v)
}

// Debugf logs a debug message.
func (l LevelLogger) Debugf(s string, params ...any) {
	v := fmt.Sprintf(s, params...)
	level.Debug(l.Logger).Log("log", v)
}

// AssetsEntityName defines the key for the assets entity used by an [EventStreamClient] to
// determine the topic where the assets are send.
const AssetsEntityName = "assets"

// Vulcan implements the asynchorus API of Vulcan.
type Vulcan struct {
	client EventStreamClient
	logger Logger
}

// EventStreamClient represent a client of an event stream system, like Kafka
// or AWS FIFO SQS queues.
type EventStreamClient interface {
	Push(entity string, id string, payload []byte, metadata map[string][]byte) error
}

// Logger defines the required methods to log info by the Vulcan async server.
type Logger interface {
	Errorf(string, ...any)
	Infof(string, ...any)
	Debugf(string, ...any)
}

// NewVulcan returns a Vulcan async server that uses the given
// [EventStreamClient] and [Logger].
func NewVulcan(client EventStreamClient, log Logger) *Vulcan {
	return &Vulcan{client, log}
}

// PushAsset publishes the state of an asset in the current point of time
// to the underlying [EventStreamClient].
func (v *Vulcan) PushAsset(asset AssetPayload) error {
	v.logger.Debugf("pushing asset %+v", asset)
	payload, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("error marshaling to json: %w", err)
	}
	// Even though the asset_id is always different for every asset, the PK of
	// an asset for the vulcan-api is the asset_id plus the team_id.
	id := strings.Join([]string{asset.Team.Id, asset.Id}, "/")
	metadata := metadata(asset)
	err = v.client.Push(AssetsEntityName, id, payload, metadata)
	if err != nil {
		return fmt.Errorf("error pushing asset %v: %w", asset, err)
	}
	v.logger.Debugf("asset pushed %+v", asset)
	return err
}

// DeleteAsset publishes an event to the underlying [EventStreamClient]
// indicating that an asset has been deleted.
func (v *Vulcan) DeleteAsset(asset AssetPayload) error {
	v.logger.Debugf("pushing asset deleted %+v", asset)
	// Even though the asset_id is always different for every asset, the PK of
	// an asset for the vulcan-api is the asset_id plus the team_id.
	id := strings.Join([]string{asset.Team.Id, asset.Id}, "/")
	metadata := metadata(asset)
	err := v.client.Push(AssetsEntityName, id, nil, metadata)
	if err != nil {
		return fmt.Errorf("error sending a delete asset event for the asset %v: %w", asset, err)
	}
	v.logger.Debugf("delete asset pushed %+v", asset)
	return err
}

// NullVulcan implements an Async Vulcan API interface that does not send the
// events to any [EventStreamClient]. It's intended to be used when the async
// API is disabled but other components still need to fullfill a dependency
// with the Vulcan Async Server.
type NullVulcan struct {
}

// DeleteAsset acepts an event indicating that an asset has been deleted and
// just ignores it.
func (v *NullVulcan) DeleteAsset(asset AssetPayload) error {
	return nil
}

// PushAsset acepts an event indicating that an asset has been modified or
// created and just ignores it.
func (v *NullVulcan) PushAsset(asset AssetPayload) error {
	return nil
}

func metadata(asset AssetPayload) map[string][]byte {
	// The asset type can't be nil.
	return map[string][]byte{
		"identifier": []byte(asset.Identifier),
		"type":       []byte(*asset.AssetType),
		"version":    []byte(Version),
	}
}
