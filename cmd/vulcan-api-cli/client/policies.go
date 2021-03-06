// Code generated by goagen v1.4.3, DO NOT EDIT.
//
// API "Vulcan-API": policies Resource Client
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

// CreatePoliciesPath computes a request path to the create action of policies.
func CreatePoliciesPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/policies", param0)
}

// Create a new policy.
func (c *Client) CreatePolicies(ctx context.Context, path string, payload *PolicyPayload) (*http.Response, error) {
	req, err := c.NewCreatePoliciesRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreatePoliciesRequest create the request corresponding to the create action endpoint of the policies resource.
func (c *Client) NewCreatePoliciesRequest(ctx context.Context, path string, payload *PolicyPayload) (*http.Request, error) {
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

// DeletePoliciesPath computes a request path to the delete action of policies.
func DeletePoliciesPath(teamID string, policyID string) string {
	param0 := teamID
	param1 := policyID

	return fmt.Sprintf("/api/v1/teams/%s/policies/%s", param0, param1)
}

// Delete a policy.
func (c *Client) DeletePolicies(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewDeletePoliciesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeletePoliciesRequest create the request corresponding to the delete action endpoint of the policies resource.
func (c *Client) NewDeletePoliciesRequest(ctx context.Context, path string) (*http.Request, error) {
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

// ListPoliciesPath computes a request path to the list action of policies.
func ListPoliciesPath(teamID string) string {
	param0 := teamID

	return fmt.Sprintf("/api/v1/teams/%s/policies", param0)
}

// List all policies from a team.
func (c *Client) ListPolicies(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewListPoliciesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListPoliciesRequest create the request corresponding to the list action endpoint of the policies resource.
func (c *Client) NewListPoliciesRequest(ctx context.Context, path string) (*http.Request, error) {
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

// ShowPoliciesPath computes a request path to the show action of policies.
func ShowPoliciesPath(teamID string, policyID string) string {
	param0 := teamID
	param1 := policyID

	return fmt.Sprintf("/api/v1/teams/%s/policies/%s", param0, param1)
}

// Show information about a policy.
func (c *Client) ShowPolicies(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowPoliciesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowPoliciesRequest create the request corresponding to the show action endpoint of the policies resource.
func (c *Client) NewShowPoliciesRequest(ctx context.Context, path string) (*http.Request, error) {
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

// UpdatePoliciesPath computes a request path to the update action of policies.
func UpdatePoliciesPath(teamID string, policyID string) string {
	param0 := teamID
	param1 := policyID

	return fmt.Sprintf("/api/v1/teams/%s/policies/%s", param0, param1)
}

// Update information about a policy.
func (c *Client) UpdatePolicies(ctx context.Context, path string, payload *PolicyUpdatePayload) (*http.Response, error) {
	req, err := c.NewUpdatePoliciesRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewUpdatePoliciesRequest create the request corresponding to the update action endpoint of the policies resource.
func (c *Client) NewUpdatePoliciesRequest(ctx context.Context, path string, payload *PolicyUpdatePayload) (*http.Request, error) {
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
