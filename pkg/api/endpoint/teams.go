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

type TeamRequest struct {
	ID          string `urlvar:"team_id"`
	Name        string `json:"name"`
	Tag         string `json:"tag" urlquery:"tag"`
	Description string `json:"description"`
}

type TeamUpdateRequest struct {
	ID          string  `urlvar:"team_id"`
	Name        *string `json:"name"`
	Tag         *string `json:"tag"`
	Description *string `json:"description"`
}

func makeListTeamsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r, ok := request.(*TeamRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		var teams []*api.Team
		if r.Tag != "" {
			team, err := s.FindTeamByTag(ctx, r.Tag)
			if err != nil {
				return nil, err
			}
			teams = append(teams, team)
		} else {
			var err error
			teams, err = s.ListTeams(ctx)
			if err != nil {
				return nil, err
			}
		}

		elements := []*api.TeamResponse{}
		for _, team := range teams {
			elements = append(elements, team.ToResponse())
		}
		return Ok{elements}, nil
	}
}

// Creates an endpoint for creating new teams
// It receives an api.Team in the request payload and then store it on the
// database. Also, it includes the current authenticated team as a team owner
// It also creates the Default group for the team
func makeCreateTeamEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		requestBody, ok := request.(*TeamRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		user, err := api.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Default(err)
		}
		email := strings.ToLower(user.Email)

		team := &api.Team{
			Name:        requestBody.Name,
			Description: requestBody.Description,
			Tag:         requestBody.Tag,
		}

		// Creates the team
		team, err = s.CreateTeam(ctx, *team, email)
		if err != nil {
			return nil, err
		}

		_, err = s.FindGroup(ctx, api.Group{TeamID: team.ID, Name: "Default"})
		if errors.IsKind(err, errors.ErrNotFound) {
			_, err := s.CreateGroup(ctx, api.Group{TeamID: team.ID, Name: "Default"})
			if err != nil {
				return nil, err
			}
		}

		return Created{team.ToResponse()}, nil
	}
}

func makeFindTeamEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		team, err := s.FindTeam(ctx, requestBody.ID)
		if err != nil {
			return nil, err
		}
		return Ok{team.ToResponse()}, nil
	}
}

func makeUpdateTeamEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamUpdateRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		team, err := s.FindTeam(ctx, requestBody.ID)
		if err != nil {
			return nil, err
		}

		if requestBody.Name != nil {
			team.Name = *requestBody.Name
		}

		if requestBody.Description != nil {
			team.Description = *requestBody.Description
		}

		if requestBody.Tag != nil {
			team.Tag = *requestBody.Tag
		}

		team, err = s.UpdateTeam(ctx, *team)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}
		return Ok{team.ToResponse()}, nil
	}
}

func makeDeleteTeamEndpoint(svc api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*TeamRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		err := svc.DeleteTeam(ctx, requestBody.ID)
		if err != nil {
			//TODO: log internal err
			return nil, err
		}

		return NoContent{nil}, nil
	}
}
