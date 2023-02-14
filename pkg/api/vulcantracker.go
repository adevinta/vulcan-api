/*
Copyright 2023 Adevinta
*/

package api

import (
	vulcantracker "github.com/adevinta/vulcan-tracker/pkg/model"
)

type FindingTicketCreate struct {
	FindingID   string `json:"finding_id" validate:"required"`
	TeamID      string `json:"team_id" validate:"required"`
	Summary     string `json:"summary" validate:"required"`
	Description string `json:"description"`
	URLTracker  string `json:"url_tracker"`
}

type FindingTicketCreateResponse struct {
	FindingID   string `json:"finding_id"`
	TeamID      string `json:"team_id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	URLTracker  string `json:"url_tracker"`
}

// Ticket represents the response data returned from the vulcan tracker service for
// the get ticket request.
type Ticket struct {
	Ticket vulcantracker.Ticket `json:"ticket"`
}

func (t Ticket) ToResponse() FindingTicketCreateResponse {
	output := FindingTicketCreateResponse{}

	output.TeamID = t.Ticket.TeamID
	output.FindingID = t.Ticket.FindingID
	output.Summary = t.Ticket.Summary
	output.Description = t.Ticket.Description
	output.URLTracker = t.Ticket.URLTracker

	return output
}
