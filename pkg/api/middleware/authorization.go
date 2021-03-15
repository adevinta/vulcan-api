/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
)

// AuthorizeMiddleware implements the authorization over an endpoint.
type AuthorizeMiddleware interface {
	Authorize(next endpoint.Endpoint) endpoint.Endpoint
}

// Authorizer provides defines the funcs that the clients of the middleware must provide.
type Authorizer interface {
	// AuthTenant receives the context and returns an object representing the
	// tenant the user is trying to access to.
	// If the user is not authorized the function must return a nil tenant.
	// If the user is allowed to do anything on that tenant, e.g. a super admin,
	// the function must return true in the second parameter.
	AuthTenant(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error)

	// AuthRol grants or denies access to a resource depending on the rol of the user.
	AuthRol(ctx context.Context, tenant interface{}) (bool, error)
}

// Authorizer provides the middleware needed to perform authorization for api calls.
type authorizerMiddleware struct {
	auth   Authorizer
	logger log.Logger
}

// NewAuthorizationMiddleware creates a new authorization middleware using the provided authorizer.
func NewAuthorizationMiddleware(auth Authorizer, logger log.Logger) AuthorizeMiddleware {
	return &authorizerMiddleware{auth: auth, logger: logger}
}

// Authorize wires the authorization middleware over an endpoint
func (a *authorizerMiddleware) Authorize(next endpoint.Endpoint) endpoint.Endpoint {
	authorized := func(ctx context.Context, request interface{}) (interface{}, error) {
		tenant, skip, err := a.auth.AuthTenant(ctx, request)
		if err != nil {
			_ = a.logger.Log("AuthTenant error", err.Error())
			return nil, errors.Forbidden(err)
		}

		// Skip indicates that the user have access to everything so skip check the role.
		if skip {
			_ = a.logger.Log("Authorize", "skipping")
			return next(ctx, request)
		}

		// If the tennant is nil is because either the tenant does not exist or the user is not allowed
		// to access to it, in any case return access forbidden.
		if tenant == nil {
			_ = a.logger.Log("Authorize", "tenant is nil")
			return nil, errors.Forbidden("tenant is nil")
		}
		grant, err := a.auth.AuthRol(ctx, tenant)
		if err != nil {
			_ = a.logger.Log("Authorize", fmt.Sprintf("cannot auth tenant %v", tenant))
			return nil, errors.Forbidden(fmt.Sprintf("cannot auth tenant %v", tenant))
		}

		if !grant {
			_ = a.logger.Log("Authorize", "access not granted")
			return nil, errors.Forbidden("access not granted")
		}

		return next(ctx, request)
	}
	return authorized
}
