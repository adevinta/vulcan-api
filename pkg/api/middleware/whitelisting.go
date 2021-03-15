/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func NotWhitelisted(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			user, err := api.UserFromContext(ctx)
			if err != nil {
				return nil, errors.Default(err)
			}

			if (user.Admin != nil && *user.Admin) ||
				(user.Observer != nil && *user.Observer) {
				return next(ctx, request)
			}

			return nil, errors.Forbidden("not authorized to access this endpoint")
		}
	}
}
