/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"gopkg.in/go-playground/validator.v9"
)

func (s vulcanitoService) ListFindings(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	return s.vulndbClient.Findings(ctx, params, pagination)
}

func (s vulcanitoService) ListFindingsIssues(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsIssuesList, error) {
	return s.vulndbClient.FindingsIssues(ctx, params, pagination)
}

func (s vulcanitoService) ListFindingsByIssue(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	return s.vulndbClient.FindingsByIssue(ctx, params, pagination)
}

func (s vulcanitoService) ListFindingsTargets(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsTargetsList, error) {
	return s.vulndbClient.FindingsTargets(ctx, params, pagination)
}

func (s vulcanitoService) ListFindingsByTarget(ctx context.Context, params api.FindingsParams, pagination api.Pagination) (*api.FindingsList, error) {
	return s.vulndbClient.FindingsByTarget(ctx, params, pagination)
}

func (s vulcanitoService) ListFindingsLabels(ctx context.Context, params api.FindingsParams) (*api.FindingsLabels, error) {
	return s.vulndbClient.Labels(ctx, params)
}

func (s vulcanitoService) FindFinding(ctx context.Context, findingID string) (*api.Finding, error) {
	return s.vulndbClient.Finding(ctx, findingID)
}

func (s vulcanitoService) CreateFindingOverwrite(ctx context.Context, findingOverwrite api.FindingOverwrite) error {
	validationErr := validator.New().Struct(findingOverwrite)
	if validationErr != nil {
		return errors.Validation(validationErr)
	}

	if !isValidFindingStatus(findingOverwrite.Status) {
		return errors.Validation(fmt.Sprintf("Invalid status: '%s'", findingOverwrite.Status))
	}

	// Valid transitions:
	//
	// OPEN           -> OPEN
	// FALSE_POSITIVE -> OPEN
	// OPEN           -> FALSE_POSITIVE
	// FALSE_POSITIVE -> FALSE_POSITIVE
	if (findingOverwrite.StatusPrevious != "OPEN" && findingOverwrite.StatusPrevious != "FALSE_POSITIVE") ||
		(findingOverwrite.Status != "OPEN" && findingOverwrite.Status != "FALSE_POSITIVE") {
		return errors.Validation(fmt.Sprintf("Status transition not allowed: from '%s' to '%s'", findingOverwrite.StatusPrevious, findingOverwrite.Status))
	}

	return s.db.CreateFindingOverwrite(findingOverwrite)
}

func (s vulcanitoService) ListFindingOverwrites(ctx context.Context, findingID string) ([]*api.FindingOverwrite, error) {
	return s.db.ListFindingOverwrites(findingID)
}

func isValidFindingStatus(status string) bool {
	// Set of valid status type
	validStatus := map[string]struct{}{
		"OPEN":           struct{}{},
		"FIXED":          struct{}{},
		"EXPIRED":        struct{}{},
		"FALSE_POSITIVE": struct{}{},
	}

	_, existsInSet := validStatus[status]
	return existsInSet
}

func (s vulcanitoService) CreateFindingTicket(ctx context.Context, ticket api.FindingTicketCreate) (*api.Ticket, error) {
	return s.vulcantrackerClient.CreateTicket(ctx, ticket)
}

func (s vulcanitoService) GetFindingTicket(ctx context.Context, findingID, teamID string) (*api.Ticket, error) {
	return s.vulcantrackerClient.GetFindingTicket(ctx, findingID, teamID)
}
