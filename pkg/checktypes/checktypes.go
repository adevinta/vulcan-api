/*
Copyright 2021 Adevinta
*/

package checktypes

import (
	"context"
	"net/http"

	"github.com/adevinta/vulcan-core-cli/vulcan-core/client"
)

// AssettypeInformer defines the methods needed by the Checkstypes struct
// to get the checktypes per assettype from vulcan-core.
type AssettypeInformer interface {
	IndexAssettypes(ctx context.Context, path string) (*http.Response, error)
	DecodeAssettypeCollection(resp *http.Response) (client.AssettypeCollection, error)
}

// Client provides information about the current checktypes that can be run
// in a scan.
type Client struct {
	informer AssettypeInformer
}

// New Creates a new client that provides information regarding the checktypes
// defined in vulcan core.
func New(informer AssettypeInformer) *Client {
	return &Client{informer: informer}
}

// ByAssettype returns a map where each key contains an assettype and each value
// the checks allowed to be executed for those asset types, e.g.,
// {{"Hostname":{"vulcan-nessus","vulcan-exposed,"vulcan-tls"}}.
func (c *Client) ByAssettype(ctx context.Context) (map[string][]string, error) {
	resp, err := c.informer.IndexAssettypes(ctx, client.IndexAssettypesPath())
	if err != nil {
		return nil, err
	}

	assettypes, err := c.informer.DecodeAssettypeCollection(resp)
	if err != nil {
		return nil, err
	}
	ret := map[string][]string{}
	for _, a := range assettypes {
		if a.Assettype == nil {
			a.Assettype = ptrToStr("")
		}
		if _, ok := ret[*a.Assettype]; !ok {
			ret[*a.Assettype] = []string{}
		}
		ret[*a.Assettype] = append(ret[*a.Assettype], a.Name...)
	}
	return ret, nil
}

func ptrToStr(in string) *string {
	return &in
}
