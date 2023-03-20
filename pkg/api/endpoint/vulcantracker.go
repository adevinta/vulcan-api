/*
Copyright 2023 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// FindingCreateTicketRequest represents a request to Vulcan Tracker that create a relationship
// between findings and tickets.
type FindingCreateTicketRequest struct {
	FindingID   string `json:"finding_id" urlvar:"finding_id"`
	TeamID      string `json:"team_id" urlvar:"team_id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func makeCreateFindingTicketEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*FindingCreateTicketRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		finding, err := s.FindFinding(ctx, r.FindingID)
		if err != nil {
			return nil, err
		}

		ticketCreate := api.FindingTicketCreate{
			FindingID:   r.FindingID,
			TeamID:      r.TeamID,
			Summary:     r.Summary,
			Description: r.Description,
		}

		if authorizedFindFindingRequest(finding.Finding.Target.Teams, r.TeamID) {
			ticket, err := s.CreateFindingTicket(ctx, ticketCreate)
			if err != nil {
				return nil, err
			}
			return Ok{ticket.ToResponse()}, nil
		}

		return Forbidden{}, nil
	}
}
