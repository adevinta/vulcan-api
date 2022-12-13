// Code generated by goagen v1.4.3, DO NOT EDIT.
//
// API "Vulcan-API": assets Resource Client
//
// Command:
// $ main

package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CreateAssetsPath computes a request path to the create action of assets.
func CreateAssetsPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/assets", param0)
}

// Creates assets in bulk mode.
// This operation accepts an array of assets, an optional array of group identifiers, an optional map of annotations, and returns an array of successfully created assets.
// If no groups are specified, assets will be added to the team's Default group.
// If one of the specified assets already exists for the team but is currently not associated with the requested groups, the association is created.
// If for any reason, the creation of an asset fails, an error message will be returned referencing the failed asset and the entire operation will be rolled back.
// ---
// Valid asset types:
// - AWSAccount
// - DomainName
// - Hostname
// - IP
// - IPRange
// - DockerImage
// - WebAddress
// - GitRepository
// ---
// If the asset type is informed, then Vulcan will use that value to create the new asset.
// Otherwise, Vulcan will try to automatically discover the asset type.
// Notice that this may result in Vulcan creating more than one asset.
// For instance, an user trying to create an asset for "vulcan.example.com", without specifying the asset type, will end up with two assets created:
// - vulcan.example.com (DomainName) and
// - vulcan.example.com (Hostname).
func (c *Client) CreateAssets(ctx context.Context, path string, payload *CreateAssetPayload) (*http.Response, error) {
	req, err := c.NewCreateAssetsRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateAssetsRequest create the request corresponding to the create action endpoint of the assets resource.
func (c *Client) NewCreateAssetsRequest(ctx context.Context, path string, payload *CreateAssetPayload) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// CreateMultiStatusAssetsPath computes a request path to the createMultiStatus action of assets.
func CreateMultiStatusAssetsPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/assets/multistatus", param0)
}

// Creates assets in bulk mode (MultiStatus).
// This operation is similar to the "Create Assets in Bulk Mode", with 2 main differences:
// - This endpoint is not atomic. Each asset creation request will succeed or fail indenpendently of the other requests.
// - This endpoint will return an array of AssetResponse in the following way:
// · For each asset with specified type, returns an AssetResponse indicating the success or failure for its creation.
// · For each asset with no type specified and successfully created, returns one AssetResponse for each auto detected asset.
// · For each asset detected from the ones with no type indicated which their creation produced an error, returns one AssetResponse indicating the failure for its creation specifying its detected type.
// In the case of all assets being successfully created, this endpoint will return status code 201-Created.
// Otherwise, it will return a 207-MultiStatus code, indicating that at least one of the requested operations failed.
func (c *Client) CreateMultiStatusAssets(ctx context.Context, path string, payload *CreateAssetPayload) (*http.Response, error) {
	req, err := c.NewCreateMultiStatusAssetsRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateMultiStatusAssetsRequest create the request corresponding to the createMultiStatus action endpoint of the assets resource.
func (c *Client) NewCreateMultiStatusAssetsRequest(ctx context.Context, path string, payload *CreateAssetPayload) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// DeleteAssetsPath computes a request path to the delete action of assets.
func DeleteAssetsPath(teamID string, assetID string) string {
	param0 := teamID
	param1 := assetID

	return fmt.Sprintf("/api/v1/teams/%s/assets/%s", param0, param1)
}

// Delete an asset.
func (c *Client) DeleteAssets(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewDeleteAssetsRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeleteAssetsRequest create the request corresponding to the delete action endpoint of the assets resource.
func (c *Client) NewDeleteAssetsRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// DiscoverAssetsPath computes a request path to the discover action of assets.
func DiscoverAssetsPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/assets/discovery", param0)
}

// This endpoint receives a list of assets with embedded
// asset annotations, and the group name where to be added. It should be used by
// third-party asset discovery services to onboard the discovered assets into
// Vulcan. The provided list of assets will overwrite the assets previously
// present in the group, in a way that:
// - Assets that do not exist in the team will be created and associated to the
// group
// - Assets that were already existing in the team but not associated to the
// group will be associated
// - Existing assets where the scannable field or the annotations are different
// will be updated accordingly
// - Assets that were associated to the group and now are not present in the
// provided list will be de-associated from the group if they belong to any
// other group, or deleted otherwise
// Because of the latency of this operation the endpoint is asynchronous. It
// returns a 202-Accepted HTTP response with the Job information in the response
// body.
//
// The discovery group name must end with '-discovered-assets' to not mess with
// manually managed asset groups. Also the first part of the name should identify
// the discovery service using the endpoint, for example:
// serviceX-discovered-assets.
// Also be aware that the provided annotations may differ from the ones that will
// be stored, because they will include a prefix to not mess with any other
// annotations already present in the asset.
//
// Duplicated assets (same identifier and type) in the payload are ignored if all
// their attributes are matching. Otherwise they produce an error and the job is
// aborted.
func (c *Client) DiscoverAssets(ctx context.Context, path string, payload *DiscoveredAssetsPayload) (*http.Response, error) {
	req, err := c.NewDiscoverAssetsRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDiscoverAssetsRequest create the request corresponding to the discover action endpoint of the assets resource.
func (c *Client) NewDiscoverAssetsRequest(ctx context.Context, path string, payload *DiscoveredAssetsPayload) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ListAssetsPath computes a request path to the list action of assets.
func ListAssetsPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/assets", param0)
}

// List all assets from a team.
func (c *Client) ListAssets(ctx context.Context, path string, identifier *string) (*http.Response, error) {
	req, err := c.NewListAssetsRequest(ctx, path, identifier)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListAssetsRequest create the request corresponding to the list action endpoint of the assets resource.
func (c *Client) NewListAssetsRequest(ctx context.Context, path string, identifier *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if identifier != nil {
		values.Set("identifier", *identifier)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ShowAssetsPath computes a request path to the show action of assets.
func ShowAssetsPath(teamID string, assetID string) string {
	param0 := teamID
	param1 := assetID

	return fmt.Sprintf("/api/v1/teams/%s/assets/%s", param0, param1)
}

// Describe an asset.
func (c *Client) ShowAssets(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowAssetsRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowAssetsRequest create the request corresponding to the show action endpoint of the assets resource.
func (c *Client) NewShowAssetsRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// UpdateAssetsPath computes a request path to the update action of assets.
func UpdateAssetsPath(teamID string, assetID string) string {
	param0 := teamID
	param1 := assetID

	return fmt.Sprintf("/api/v1/teams/%s/assets/%s", param0, param1)
}

// Update an asset.
// Asset type and identifier can not be modified.
func (c *Client) UpdateAssets(ctx context.Context, path string, payload *AssetUpdatePayload) (*http.Response, error) {
	req, err := c.NewUpdateAssetsRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewUpdateAssetsRequest create the request corresponding to the update action endpoint of the assets resource.
func (c *Client) NewUpdateAssetsRequest(ctx context.Context, path string, payload *AssetUpdatePayload) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "https"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequestWithContext(ctx, "PATCH", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")
	if c.BearerSigner != nil {
		if err := c.BearerSigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
