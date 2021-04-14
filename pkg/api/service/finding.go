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

func (s vulcanitoService) FindFinding(ctx context.Context, findingID string) (*api.Finding, error) {
	return s.vulndbClient.Finding(ctx, findingID)
}

func (s vulcanitoService) CreateFindingOverride(ctx context.Context, findingOverride api.FindingOverride) error {
	validationErr := validator.New().Struct(findingOverride)
	if validationErr != nil {
		return errors.Validation(validationErr)
	}

	if !isValidFindingStatus(findingOverride.Status) {
		return errors.Validation(fmt.Sprintf("Invalid status: '%s'", findingOverride.Status))
	}

	// Valid transitions:
	//
	// OPEN           -> OPEN
	// FALSE_POSITIVE -> OPEN
	// OPEN           -> FALSE_POSITIVE
	// FALSE_POSITIVE -> FALSE_POSITIVE
	if (findingOverride.StatusPrevious != "OPEN" && findingOverride.StatusPrevious != "FALSE_POSITIVE") ||
		(findingOverride.Status != "OPEN" && findingOverride.Status != "FALSE_POSITIVE") {
		return errors.Validation(fmt.Sprintf("Status transition not allowed: from '%s' to '%s'", findingOverride.StatusPrevious, findingOverride.Status))
	}

	return s.db.CreateFindingOverride(findingOverride)
}

func (s vulcanitoService) ListFindingOverrides(ctx context.Context, findingID string) ([]*api.FindingOverride, error) {
	return s.db.ListFindingOverrides(findingID)
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
