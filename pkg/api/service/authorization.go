/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"os"

	apiErrors "github.com/adevinta/errors"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/adevinta/vulcan-api/pkg/api"
)

var (
	errEmailNotFoundInCtx   = errors.New("Email not found in context")
	errUnexpectedTenantType = errors.New("unexpected tenant type")
	errMethodNotFoundInCtx  = errors.New("http method not found in context")
)

type authorization struct {
	db api.VulcanitoStore
}

// NewAuthorizationService creates a new instance of the authorization service.
// The service is not includes directly in the vulcan api service just because it's not directly exposed as a service
// to the end user but is injected to the authorization middleware in order to provide the logic it needs.
func NewAuthorizationService(db api.VulcanitoStore) api.AuthService {
	return &authorization{db: db}
}

// Returns team, permission and error
func (a *authorization) AuthTenant(ctx context.Context, request interface{}) (interface{}, bool, error) {

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	usr, err := api.UserFromContext(ctx)
	if err != nil {
		_ = logger.Log("authorization", err.Error())
		return nil, false, apiErrors.Unauthorized(errEmailNotFoundInCtx)
	}

	if usr.Admin != nil && *usr.Admin {
		_ = logger.Log("authorization", "user is admin")
		return nil, true, nil
	}

	if usr.Observer != nil && *usr.Observer {
		_ = logger.Log("authorization", "user is observer")
	}

	// Inspect the request struct to check if we have a field containing a Tag
	// equals `urlvar` equals "team_id". In positive case, then we assume that
	// the team ID is the same value as the field found
	teamID := ""
	obj := reflect.TypeOf(request).Elem()
	for i := 0; i < obj.NumField(); i++ {
		if obj.Field(i).Tag.Get("urlvar") == "team_id" {
			teamID = reflect.ValueOf(request).Elem().Field(i).String()
			break
		}
	}

	// If teamID is empty, request is for a global endpoint
	if teamID == "" {
		if usr.Observer != nil && *usr.Observer {
			// If request is for a global endpoint and user
			// is observer, authorize with member role
			return &api.UserTeam{Role: "member"}, false, nil
		}
		_ = logger.Log("authorization", "err with tenant id")
		return nil, false, err
	}

	t, err := a.db.FindTeamByIDForUser(teamID, usr.ID)
	if err != nil {
		_ = logger.Log("authorization", "FindTeamByIDForUser error")

		if usr.Observer != nil && *usr.Observer {
			ut := &api.UserTeam{}
			t, err := a.db.FindTeam(teamID)
			if err != nil {
				_ = logger.Log("authorization", "FindTeam error")

				if a.db.NotFoundError(err) {
					_ = logger.Log("authorization", "db not found error")
					return nil, false, nil
				}
				return nil, false, err
			}
			ut.Team = t
			ut.Role = "member"
			return ut, false, nil
		}

		if a.db.NotFoundError(err) {
			_ = logger.Log("authorization", "db not found error")
			return nil, false, nil
		}
		return nil, false, err
	}

	return t, false, nil
}

func (a *authorization) AuthRol(ctx context.Context, tenant interface{}) (bool, error) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	t, ok := tenant.(*api.UserTeam)
	if !ok {
		return false, apiErrors.Unauthorized(errUnexpectedTenantType)
	}
	// get the http method of the request.
	m, ok := ctx.Value(kithttp.ContextKeyRequestMethod).(string)
	if !ok {
		return false, errMethodNotFoundInCtx
	}

	// Deny authorization if we are not able to load the Role for this request
	if t == nil || !t.Role.Valid() {
		return false, nil
	}

	// For non owner roles profiles we only allow to perform get methods.
	if t.Role != api.Owner {
		_ = logger.Log("authorization", "not owner")
		if m != http.MethodGet {
			return false, nil
		}
	}
	// If we are here the user is authorized in the tenant (team) and:
	// The user is owner and is authorized to do whatever he wants on the tenant
	// or
	// The user is a member and is performing a get
	// in either case we must grant access.

	return true, nil
}
