// Code generated by goagen v1.4.3, DO NOT EDIT.
//
// API "Vulcan-API": schedule Resource Client
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

// CreateSchedulePath computes a request path to the create action of schedule.
func CreateSchedulePath(teamID string, programID string) string {
	param0 := teamID
	param1 := programID

	return fmt.Sprintf("/api/v1/teams/%s/programs/%s/schedule", param0, param1)
}

// Create a new schedule.
func (c *Client) CreateSchedule(ctx context.Context, path string, payload *SchedulePayload) (*http.Response, error) {
	req, err := c.NewCreateScheduleRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateScheduleRequest create the request corresponding to the create action endpoint of the schedule resource.
func (c *Client) NewCreateScheduleRequest(ctx context.Context, path string, payload *SchedulePayload) (*http.Request, error) {
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

// DeleteSchedulePath computes a request path to the delete action of schedule.
func DeleteSchedulePath(teamID string, programID string) string {
	param0 := teamID
	param1 := programID

	return fmt.Sprintf("/api/v1/teams/%s/programs/%s/schedule", param0, param1)
}

// Delete a schedule.
func (c *Client) DeleteSchedule(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewDeleteScheduleRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeleteScheduleRequest create the request corresponding to the delete action endpoint of the schedule resource.
func (c *Client) NewDeleteScheduleRequest(ctx context.Context, path string) (*http.Request, error) {
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

// UpdateSchedulePath computes a request path to the update action of schedule.
func UpdateSchedulePath(teamID string, programID string) string {
	param0 := teamID
	param1 := programID

	return fmt.Sprintf("/api/v1/teams/%s/programs/%s/schedule", param0, param1)
}

// Update information about a schedule.
func (c *Client) UpdateSchedule(ctx context.Context, path string, payload *ScheduleUpdatePayload) (*http.Response, error) {
	req, err := c.NewUpdateScheduleRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewUpdateScheduleRequest create the request corresponding to the update action endpoint of the schedule resource.
func (c *Client) NewUpdateScheduleRequest(ctx context.Context, path string, payload *ScheduleUpdatePayload) (*http.Request, error) {
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
