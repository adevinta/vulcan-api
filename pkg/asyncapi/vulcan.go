/*
Copyright 2022 Adevinta
*/

package asyncapi

import (
	"encoding/json"
	"fmt"
)

const assetsEntityName = "assets"

// Vulcan implements the asynchorus API of Vulcan.
type Vulcan struct {
	client EventStreamClient
	logger Logger
}

// EventStreamClient represent a client of an event stream system, like Kafka
// used by Vulcan to push the events of the its async API.
type EventStreamClient interface {
	Push(entity string, id string, payload []byte) error
}

// Logger defines the required methods to log info by the Vulcan async server.
type Logger interface {
	ErrorF(string, ...any)
	InfoF(string, ...any)
	DebugF(string, ...any)
}

// NewVulcan returns a Vulcan async server that uses the given Event stream
// client and logger.
func NewVulcan(client EventStreamClient, log Logger) Vulcan {
	return Vulcan{client, log}
}

// PushAsset publishes the state of an asset in the current point of time
// to the underlying event stream.
func (v *Vulcan) PushAsset(asset AssetPayload) error {
	v.logger.DebugF("pushing asset %+v", asset)
	payload, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("error marshaling to json: %w", err)
	}
	err = v.client.Push(assetsEntityName, asset.Id, payload)
	if err != nil {
		return fmt.Errorf("error sending pushing asset %v: %w", asset, err)
	}
	v.logger.DebugF("asset pushed %+v", asset)
	return err
}