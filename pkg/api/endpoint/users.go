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
	"github.com/adevinta/vulcan-api/pkg/common"
)

type UserRequest struct {
	ID        string `json:"user_id" urlvar:"user_id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Active    *bool  `json:"active"`
	Admin     *bool  `json:"admin"`
	Observer  *bool  `json:"observer"`
}

func makeListUsersEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		users, err := s.ListUsers(ctx)
		if err != nil {
			return nil, err
		}

		elements := []*api.UserResponse{}
		for _, user := range users {
			elements = append(elements, user.ToResponse())
		}
		return Ok{elements}, nil
	}
}

func makeCreateUserEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		requestUser := api.User{
			Firstname: requestBody.FirstName,
			Lastname:  requestBody.LastName,
			Email:     requestBody.Email,
			Active:    requestBody.Active,
			// Admin attribute is set to false when creating a user.
			// This is set to avoid users creating other users with admin privileges.
			Admin:    common.Bool(false),
			Observer: common.Bool(false),
		}

		user, err := svc.CreateUser(ctx, requestUser)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		return Created{user.ToResponse()}, nil
	}
}

func makeUpdateUserEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		user, err := svc.FindUser(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		currentUser, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Default(err)
		}

		if (currentUser.Admin != nil) && !*currentUser.Admin && (currentUser.ID != user.ID) {
			return nil, errors.Forbidden("user can not modify other users")
		}

		user.ID = requestBody.ID
		user.Firstname = requestBody.FirstName
		user.Lastname = requestBody.LastName
		user.Active = requestBody.Active

		if (user.Admin != nil) && (requestBody.Admin != nil) && (*user.Admin != *requestBody.Admin) {
			if *currentUser.Admin {
				user.Admin = requestBody.Admin
			} else {
				return nil, errors.Forbidden("user can not modify admin attribute")
			}
		}

		if (user.Observer != nil) && (requestBody.Observer != nil) && (*user.Observer != *requestBody.Observer) {
			if *currentUser.Admin {
				user.Observer = requestBody.Observer
			} else {
				return nil, errors.Forbidden("user can not modify admin attribute")
			}
		}

		user, err = svc.UpdateUser(ctx, *user)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		return Ok{user.ToResponse()}, nil
	}
}

func makeFindUserEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		user, err := svc.FindUser(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal error
			return nil, err
		}

		return Ok{user.ToResponse()}, nil
	}
}

func makeDeleteUserEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		user, err := svc.FindUser(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal error
			return nil, err
		}

		currentUser, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Default(err)
		}

		if (currentUser.Admin != nil) && !*currentUser.Admin && (currentUser.ID != user.ID) {
			return nil, errors.Forbidden("user can not modify other users")
		}

		err = svc.DeleteUser(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal error
			return nil, err
		}

		return NoContent{nil}, nil
	}
}

func makeFindProfileEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		user, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Assertion("Type assertion failed")
		}

		return Ok{user.ToResponse()}, nil
	}
}

func makeGenerateAPITokenEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		token, err := s.GenerateAPIToken(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal error
			// TODO: consider also returning different error codes depending on the error, for example:
			// unauthorized when trying to generate a token for a user that is not you.
			return nil, err
		}

		return Created{token}, nil
	}
}

func makeFindTeamsByUserEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*UserRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		teams, err := s.FindTeamsByUser(ctx, requestBody.ID)
		if err != nil {
			return nil, err
		}

		teamList := []api.TeamResponse{}
		for _, team := range teams {
			teamList = append(teamList, *team.ToResponse())
		}

		return Ok{teamList}, nil
	}
}
