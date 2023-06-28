/*
Copyright 2023 Adevinta
*/

package api

import (
	vulcantracker "github.com/adevinta/vulcan-tracker/pkg/model"
)

// FindingTicketCreate represents the data needed to create a ticket.
type FindingTicketCreate struct {
	FindingID   string `json:"finding_id" validate:"required"`
	TeamID      string `json:"team_id" validate:"required"`
	Summary     string `json:"summary" validate:"required"`
	Description string `json:"description"`
	URLTracker  string `json:"url_tracker"`
}

// FindingTicketCreateResponse represents a response when request a ticket
// creation.
type FindingTicketCreateResponse struct {
	URLTracker string `json:"url_tracker"`
}

// Ticket represents the response data returned from the vulcan tracker service for
// the get ticket request.
type Ticket struct {
	Ticket vulcantracker.Ticket `json:"ticket"`
}

// ToResponse transforms a ticket model into a response.
func (t Ticket) ToResponse() FindingTicketCreateResponse {
	output := FindingTicketCreateResponse{}
	output.URLTracker = t.Ticket.URLTracker

	return output
}
