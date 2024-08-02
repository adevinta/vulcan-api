/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func makeListIssuesEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		issues, err := s.ListIssues(ctx)
		if err != nil {
			return nil, err
		}

		elements := []*api.IssueResponse{}
		for _, issue := range issues {
			elements = append(elements, issue.ToResponse())
		}
		return Ok{elements}, nil
	}
}
