/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

type IssuesRequest struct {
	Page int `urlquery:"page"`
	Size int `urlquery:"size"`
}

func makeListIssuesEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// Only admin and observer users can query issues endpoint.
		authorized, err := isAuthorizedTagsParam(ctx)
		if err != nil {
			return nil, err
		}
		if !authorized {
			return nil, errors.Forbidden("User is not allowed to query issues")
		}

		r, ok := request.(*IssuesRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		pagination := api.Pagination{Page: r.Page, Size: r.Size}

		response, err = s.ListIssues(ctx, pagination)
		if err != nil {
			return nil, err
		}
		return Ok{response}, nil
	}
}
