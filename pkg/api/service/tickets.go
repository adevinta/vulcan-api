/*
Copyright 2023 Adevinta
*/

package service

import (
	"context"
	"fmt"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func isTrackerAllowed(ctx context.Context) bool {
	user, err := api.UserFromContext(ctx)
	return err == nil && (user.Admin != nil && *user.Admin) || (user.Observer != nil && *user.Observer)
}

// CreateFindingTicket requests the creation of a ticket in the ticket tracker
// with the values stored in the argument ticket.
func (s vulcanitoService) CreateFindingTicket(ctx context.Context, ticket api.FindingTicketCreate) (*api.Ticket, error) {
	if !isTrackerAllowed(ctx) {
		return nil, fmt.Errorf("unauthorized to create a ticket")
	}
	return s.vulcantrackerClient.CreateTicket(ctx, ticket)
}

// GetFindingTicket makes a request to vulcan tracker to find a ticket for a team.
func (s vulcanitoService) GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error) {
	if s.vulcantrackerClient == nil || !isTrackerAllowed(ctx) {
		return nil, nil
	}
	return s.vulcantrackerClient.GetFindingTicket(ctx, findingID, teamID)
}

// IsATeamOnboardedInVulcanTracker return if a team is onboarded in vulcan tracker.
func (s vulcanitoService) IsATeamOnboardedInVulcanTracker(ctx context.Context, teamID string, onboardedTeams []string) bool {
	if s.vulcantrackerClient == nil {
		return false
	}
	if !isTrackerAllowed(ctx) {
		return false
	}
	for _, team := range onboardedTeams {
		if team == teamID || team == "*" {
			return true
		}
	}
	return false
}
