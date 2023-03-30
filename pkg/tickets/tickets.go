/*
Copyright 2023 Adevinta
*/
package tickets

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/adevinta/errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adevinta/vulcan-api/pkg/api"
)

const (
	ticketsPath       = "/%s/tickets"
	findingTicketPath = "/%s/tickets/findings/%s"

	authScheme = "TEAM team=%s"
	noAuth     = ""
)

// Client represents a vulcan tracker client.
type Client interface {
	CreateTicket(ctx context.Context, payload api.FindingTicketCreate) (*api.Ticket, error)
	GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a new tickets client with the given config and httpClient.
func NewClient(httpClient *http.Client, baseURL string, insecureTLS bool) Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureTLS, // nolint
				},
			},
		}
	}
	return &client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *client) performRequest(ctx context.Context, method, path, authTeam string, params map[string]string, payload []byte) ([]byte, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = BuildQueryFilter(params)

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	if authTeam != noAuth {
		req.Header.Set("Authorization", fmt.Sprintf(authScheme, authTeam))
	}

	if payload != nil {
		req.Header.Set("Content-type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if !IsHTTPStatusOk(resp.StatusCode) {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, ParseHTTPErr(resp.StatusCode, string(content))
	}

	return io.ReadAll(resp.Body)
}

// CreateTicket requests the creation of a ticket in the ticket tracker server configurated for the team.
func (c *client) CreateTicket(ctx context.Context, payload api.FindingTicketCreate) (*api.Ticket, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.performRequest(ctx, http.MethodPost, fmt.Sprintf(ticketsPath, payload.TeamID), noAuth, nil, data)
	if err != nil {
		return nil, err
	}

	var ticketResponse api.Ticket
	err = json.Unmarshal(resp, &ticketResponse)

	return &ticketResponse, err
}

// GetFindingTicket makes a request to vulcan tracker to find a ticket.
func (c *client) GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error) {
	path := fmt.Sprintf(findingTicketPath, teamID, findingID)

	resp, err := c.performRequest(ctx, http.MethodGet, path, noAuth, nil, nil)
	if err != nil {
		return nil, err
	}
	var ticketResponse api.Ticket
	err = json.Unmarshal(resp, &ticketResponse)

	return &ticketResponse, err
}

// IsHTTPStatusOk determines if a status code is an OK or not.
func IsHTTPStatusOk(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

// BuildQueryFilter builds the query params string to be added to a request.
func BuildQueryFilter(filters map[string]string) string {
	filterParts := []string{}
	for key, value := range filters {
		part := fmt.Sprintf("%s=%s", key, value)
		filterParts = append(filterParts, part)
	}
	return strings.Join(filterParts, "&")
}

// ParseHTTPErr wraps and transform an HTTP error into a custom error
// using github.com/adevinta/errors fro that
func ParseHTTPErr(statusCode int, mssg string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.Assertion(mssg)
	case http.StatusUnauthorized:
		return errors.Unauthorized(mssg)
	case http.StatusForbidden:
		return errors.Forbidden(mssg)
	case http.StatusNotFound:
		return errors.NotFound(mssg)
	case http.StatusMethodNotAllowed:
		return errors.MethodNotAllowed(mssg)
	case http.StatusUnprocessableEntity:
		return errors.Validation(mssg)
	default:
		return errors.Default(mssg)
	}
}
