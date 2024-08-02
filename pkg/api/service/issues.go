/*
Copyright 2024 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) ListIssues(ctx context.Context, pagination api.Pagination) (*api.IssuesList, error) {
	return s.vulndbClient.Issues(ctx, pagination)
}
