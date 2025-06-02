/*
Copyright 2021 Adevinta
*/

package checktypes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Checktype struct {
	// List of the asset types that this checktype allows to be used against to
	Assets      []string  `form:"assets,omitempty" json:"assets,omitempty" yaml:"assets,omitempty" xml:"assets,omitempty"`
	Description string    `form:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty" xml:"description,omitempty"`
	Enabled     bool      `form:"enabled,omitempty" json:"enabled,omitempty" yaml:"enabled,omitempty" xml:"enabled,omitempty"`
	ID          uuid.UUID `form:"id,omitempty" json:"id,omitempty" yaml:"id,omitempty" xml:"id,omitempty"`
	// Image that needs to be pulled to run the Check of this type
	Image string `form:"image,omitempty" json:"image,omitempty" yaml:"image,omitempty" xml:"image,omitempty"`
	Name  string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
	// Default configuration options for the Checktype. It should be in JSON format
	Options map[string]any `form:"options,omitempty" json:"options,omitempty" yaml:"options,omitempty" xml:"options,omitempty"`
	// The queue name to be used by a check of this type
	QueueName string `form:"queue_name,omitempty" json:"queue_name,omitempty" yaml:"queue_name,omitempty" xml:"queue_name,omitempty"`
	// List of required vars that the agent must inject to a check using this checktype
	RequiredVars []string `form:"required_vars,omitempty" json:"required_vars,omitempty" yaml:"required_vars,omitempty" xml:"required_vars,omitempty"`
	// Specifies the maximum amount of time that the check should be running before it's killed
	Timeout int `form:"timeout,omitempty" json:"timeout,omitempty" yaml:"timeout,omitempty" xml:"timeout,omitempty"`
}

type AssettypeCollection []Assettype

type Assettype struct {
	Assettype string   `form:"assettype,omitempty" json:"assettype,omitempty" yaml:"assettype,omitempty" xml:"assettype,omitempty"`
	Name      []string `form:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
}

// Client provides information about the current checktypes that can be run
// in a scan.
type Client struct {
	URL string
	Cts []Checktype
}

type JSONChecktypes struct {
	Checktypes []Checktype `json:"checktypes"`
}

// New Creates a new client that provides information regarding the checktypes
// defined in vulcan core.
func New(URL string) *Client {
	return &Client{URL: URL}
}

func (c *Client) getChecktypes(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var ct JSONChecktypes
	if err := json.NewDecoder(resp.Body).Decode(&ct); err != nil {
		return err
	}
	c.Cts = ct.Checktypes
	return nil
}

// ByAssettype returns a map where each key contains an assettype and each value
// the checks allowed to be executed for those asset types, e.g.,
// {{"Hostname":{"vulcan-nessus","vulcan-exposed,"vulcan-tls"}}.
func (c *Client) ByAssettype(ctx context.Context) (map[string][]string, error) {
	err := c.getChecktypes(ctx)
	if err != nil {
		return nil, err
	}
	ret := map[string][]string{}
	for _, ct := range c.Cts {
		for _, asset := range ct.Assets {
			if _, ok := ret[asset]; !ok {
				ret[asset] = []string{}
			}
			ret[asset] = append(ret[asset], ct.Name)
		}
	}
	return ret, nil
}

func (c *Client) GetAssettypes(ctx context.Context) (AssettypeCollection, error) {
	m, err := c.ByAssettype(ctx)
	if err != nil {
		return nil, err
	}

	ret := AssettypeCollection{}
	for val, ct := range m {
		ret = append(ret, Assettype{
			Assettype: val,
			Name:      ct,
		})
	}
	return ret, nil
}

func (c *Client) GetChecktype(ctx context.Context, name string) (*Checktype, error) {
	err := c.getChecktypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, ct := range c.Cts {
		if ct.Name == name {
			return &ct, nil
		}
	}
	return nil, nil
}
