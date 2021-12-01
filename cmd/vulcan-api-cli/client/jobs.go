// Code generated by goagen v1.4.3, DO NOT EDIT.
//
// API "Vulcan-API": jobs Resource Client
//
// Command:
// $ main

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ShowJobsPath computes a request path to the show action of jobs.
func ShowJobsPath(jobID string) string {
	param0 := jobID

	return fmt.Sprintf("/api/v1/jobs/%s", param0)
}

// Describes job status and results. The possible values for the status are:
// - 'PENDING': The job has been noted and is pending to be processed
// - 'RUNNING': The job is on execution
// - 'DONE': The job has finished, either successfully or unsuccesfully. Result.error needs to be processed to determine it
//
// The results field indicates if there was an error during the execution of the job, and otherwise can return data from the job execution
func (c *Client) ShowJobs(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowJobsRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowJobsRequest create the request corresponding to the show action endpoint of the jobs resource.
func (c *Client) NewShowJobsRequest(ctx context.Context, path string) (*http.Request, error) {
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