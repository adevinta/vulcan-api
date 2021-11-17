/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"strings"

	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

type TeamMemberRequest struct {
	TeamID string `json:"team_id" urlvar:"team_id"`
	UserID string `json:"user_id" urlvar:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func makeListTeamMembersEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqBody, ok := request.(*TeamMemberRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		team, err := s.FindTeam(ctx, reqBody.TeamID)
		if err != nil {
			return nil, errors.NotFound("cannot find team")
		}
		response := []api.MemberResponse{}
		for _, teamMember := range team.UserTeam {
			response = append(response, *teamMember.ToResponse())
		}
		return Ok{response}, nil
	}
}

func makeFindTeamMemberEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamMemberRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		teamUser, err := s.FindTeamMember(ctx, requestBody.TeamID, requestBody.UserID)
		if err != nil {
			return nil, err
		}
		return Ok{teamUser.ToResponse()}, nil
	}
}

func makeCreateTeamMemberEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamMemberRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		teamUser := &api.UserTeam{}
		teamUser.UserID = requestBody.UserID
		teamUser.TeamID = requestBody.TeamID
		teamUser.User = &api.User{Email: strings.ToLower(requestBody.Email)}
		teamUser.Role = api.Role(requestBody.Role)
		teamUser, err := s.CreateTeamMember(ctx, *teamUser)
		if err != nil {
			return nil, err
		}
		return Created{teamUser.ToResponse()}, nil
	}
}

func makeUpdateTeamMemberEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamMemberRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		if !api.Role(requestBody.Role).Valid() {
			return nil, errors.Update("Not a valid role")
		}
		teamUser := &api.UserTeam{
			UserID: requestBody.UserID,
			TeamID: requestBody.TeamID,
			Role:   api.Role(requestBody.Role),
		}

		// Creates the team
		teamUserResp, err := s.UpdateTeamMember(ctx, *teamUser)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		return Ok{teamUserResp.ToResponse()}, nil
	}
}

func makeDeleteTeamMemberEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamMemberRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		err := s.DeleteTeamMember(ctx, requestBody.TeamID, requestBody.UserID)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		return NoContent{nil}, nil
	}
}
