/*
Copyright 2021 Adevinta
*/

package client

import (
	"errors"

	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue/client/cgcatalogue"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue/client/none"
)

const (
	// CG identifies CloudGovernance client.
	CG = "CloudGovernance"
	// None identifies a dummy client.
	None = "None"
)

var (
	// ErrInvalidClientImpl indicates that the client implementation specified is not valid.
	ErrInvalidClientImpl = errors.New("invalid client implementation")
	// ErrInvalidClientConfig indicates that the given configuration is not valid for the specified client.
	ErrInvalidClientConfig = errors.New("invalid client configuration")
)

// AWSCatalogueAPIConfig represents the config params
// for an AWS catalogue API client.
type AWSCatalogueAPIConfig struct {
	URL           string
	Key           string
	Retries       int
	RetryInterval int
}

// NewClient represents a factory method to create a new AWSCatalogue client.
func NewClient(kind string, config interface{}) (awscatalogue.Client, error) {
	switch kind {
	case CG:
		cfg, ok := config.(AWSCatalogueAPIConfig)
		if !ok {
			return nil, ErrInvalidClientConfig
		}
		c := cgcatalogue.New(cfg.Key, cfg.URL, nil)
		return cgcatalogue.NewBackOffClient(c, cfg.RetryInterval, cfg.Retries), nil
	case None:
		return none.New(), nil
	default:
		return nil, ErrInvalidClientImpl
	}
}
