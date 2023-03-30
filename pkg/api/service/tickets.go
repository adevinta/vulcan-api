/*
Copyright 2023 Adevinta
*/

package service

import (
	"context"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) CreateFindingTicket(ctx context.Context, ticket api.FindingTicketCreate) (*api.Ticket, error) {
	return s.vulcantrackerClient.CreateTicket(ctx, ticket)
}

func (s vulcanitoService) GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error) {
	return s.vulcantrackerClient.GetFindingTicket(ctx, findingID, teamID)
}
