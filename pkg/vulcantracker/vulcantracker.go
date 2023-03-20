/*
Copyright 2023 Adevinta
*/
package vulcantracker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/common"
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
	IsATeamOnboardedInVulcanTracker(ctx context.Context, teamID string) bool // feature flag.
}

type client struct {
	baseURL        string
	httpClient     *http.Client
	onboardedTeams []string // feature flag.
}

// NewClient returns a new vulcantracker client with the given config and httpClient.
func NewClient(httpClient *http.Client, baseURL string, insecureTLS bool, onboardedTeams []string) Client {
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
		httpClient:     httpClient,
		baseURL:        baseURL,
		onboardedTeams: onboardedTeams,
	}
}

func (c *client) performRequest(ctx context.Context, method, path, authTeam string, params map[string]string, payload []byte) ([]byte, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = common.BuildQueryFilter(params)

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

	if !common.IsHttpStatusOk(resp.StatusCode) {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, common.ParseHttpErr(resp.StatusCode, string(content))
	}

	return io.ReadAll(resp.Body)
}

// IsATeamOnboardedInVulcanTracker return if a team is onboarded in vulcan tracker.
func (c *client) IsATeamOnboardedInVulcanTracker(ctx context.Context, teamID string) bool {
	for _, team := range c.onboardedTeams {
		if team == teamID {
			return true
		}
	}
	return false
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
