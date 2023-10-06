/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"
	"slices"

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

	if !isValidFindingTransition(findingOverwrite.Status, findingOverwrite.StatusPrevious) {
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
		"OPEN":           {},
		"FIXED":          {},
		"EXPIRED":        {},
		"FALSE_POSITIVE": {},
	}

	_, existsInSet := validStatus[status]
	return existsInSet
}

func isValidFindingTransition(status, statusPrevious string) bool {
	// Valid transitions:
	//
	// OPEN           -> OPEN
	// OPEN           -> FALSE_POSITIVE
	// FALSE_POSITIVE -> OPEN
	// FALSE_POSITIVE -> FALSE_POSITIVE
	// FIXED          -> FALSE_POSITIVE
	if status == statusPrevious {
		return true
	}

	validTransitions := map[string][]string{
		"OPEN":           {"FALSE_POSITIVE"},
		"FALSE_POSITIVE": {"OPEN"},
		"FIXED":          {"FALSE_POSITIVE"},
	}
	return slices.Contains(validTransitions[statusPrevious], status)
}
