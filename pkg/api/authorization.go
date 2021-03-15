/*
Copyright 2021 Adevinta
*/

package api

import "context"

// AuthService defines the exposed functions of an authorization service.
type AuthService interface {
	AuthTenant(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error)
	AuthRol(ctx context.Context, tenant interface{}) (bool, error)
}
