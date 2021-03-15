/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"strings"

	"context"
	"reflect"

	"github.com/adevinta/errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	uuid "github.com/satori/go.uuid"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/global"
)

const (
	// path variables that allow
	// replacement.
	teamIDVar = "team_id"
	userIDVar = "user_id"

	// path variables that allow
	// global entities wildcards.
	groupIDVar   = "group_id"
	policyIDVar  = "policy_id"
	programIDVar = "program_id"
)

// ValidateUUIDs returns a middleware that inspects the request struct in
// search for ID parameters in the request path and validates their compliance
// with UUID format and/or some exceptions.
//
// We are using the "reflect" package to inspect the request struct, looking
// for fields that have `urlvar` tags with suffix `_id`, for example:
//
//   type UpdateTeamJSONRequest struct {
//	     ID          string `urlvar:"team_id"`
//	     Name        string `json:"name"`
//	     Description string `json:"description"`
//   }
//
// There are some special cases to UUID validation.
//	  - global entities:
//		Global entities are special values that can be specified by the client
//		and are expanded into other values before the service layer processes
//		the request (e.g.: periodic-full-scan).
//    - team_id urlvar values can be set with the team name:
//	    This middleware will inspect for that option and replace the team name
//	    for its UUID in the original request path.
//    - user_id urlvar values can be set with the user email:
//		This middleware will inspect for that option and replace the user email
//		for its UUID in the original request path.
func ValidateUUIDs(repo api.VulcanitoStore, globalEntities *global.Entities, logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			obj := reflect.TypeOf(request).Elem()

			// Iterate over struct fields.
			for i := 0; i < obj.NumField(); i++ {
				// Check if we have a `urlvar` tag value that ends up in `_id`.
				urlvar := obj.Field(i).Tag.Get("urlvar")
				if strings.HasSuffix(urlvar, "_id") {
					ID := reflect.ValueOf(request).Elem().Field(i).String()

					// If ID is a void string it means
					// it is not a used field for the endpoint.
					// If ID is a global entity, ignore UUID compliance.
					if ID == "" || isGlobalEntity(globalEntities, urlvar, ID) {
						continue
					}

					// Verify its compliance with UUID format.
					_, err := uuid.FromString(ID)
					if err != nil {
						var newID string

						// If it's not an UUID, check for team and user
						// special cases.
						if urlvar == teamIDVar {
							team, dberr := repo.FindTeamByName(ID)
							if dberr != nil {
								return nil, dberr
							}
							newID = team.ID
						} else if urlvar == userIDVar && strings.Contains(ID, "@") {
							user, dberr := repo.FindUserByEmail(ID)
							if dberr != nil {
								return nil, dberr
							}
							newID = user.ID
						}

						if newID != "" {
							// Update the original struct with the value found.
							reflect.ValueOf(request).Elem().Field(i).SetString(newID)
						} else {
							// Return malformed ID error.
							return nil, errors.Validation("ID is malformed")
						}
					}
				}
			}
			return next(ctx, request)
		}
	}
}

// isGlobalEntity compares given ID value with predefined global entities
// wildcards for the specified urlvar type.
// Returns true if input ID is a global entity for any type. Otherwise returns false.
func isGlobalEntity(globalEntities *global.Entities, urlvar, ID string) bool {
	var ok bool

	switch urlvar {
	case groupIDVar:
		_, ok = globalEntities.Groups()[ID]
	case policyIDVar:
		_, ok = globalEntities.Policies()[ID]
	case programIDVar:
		_, ok = globalEntities.Programs()[ID]
	}

	return ok
}
