/*
Copyright 2023 Adevinta
*/

package service

import (
	"context"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// CreateFindingTicket requests the creation of a ticket in the ticket tracker
// with the values stored in the argument ticket.
func (s vulcanitoService) CreateFindingTicket(ctx context.Context, ticket api.FindingTicketCreate) (*api.Ticket, error) {
	if s.vulcantrackerClient == nil {
		return nil, nil
	}
	return s.vulcantrackerClient.CreateTicket(ctx, ticket)
}

// GetFindingTicket makes a request to vulcan tracker to find a ticket for a team.
func (s vulcanitoService) GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error) {
	if s.vulcantrackerClient == nil {
		return nil, nil
	}
	return s.vulcantrackerClient.GetFindingTicket(ctx, findingID, teamID)
}
