/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/vulcan-api/pkg/api"
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

func (s vulcanitoService) UpdateFinding(ctx context.Context, findingID string, payload api.UpdateFinding, tag string) (*api.Finding, error) {
	return s.vulndbClient.UpdateFinding(ctx, findingID, &payload, tag)
}
