/*
Copyright 2024 Adevinta
*/

package service

import (
	"context"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) ListIssues(ctx context.Context) ([]*api.Issue, error) {
	return s.db.ListIssues()
}
