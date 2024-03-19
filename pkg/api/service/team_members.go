/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"strings"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) FindTeamMember(ctx context.Context, teamID string, userID string) (*api.UserTeam, error) {
	if teamID == "" {
		return nil, errors.Validation(`Team ID is empty`)
	}
	if userID == "" {
		return nil, errors.Validation(`User ID is empty`)
	}
	return s.db.FindTeamMember(teamID, userID)
}

func (s vulcanitoService) CreateTeamMember(ctx context.Context, teamMember api.UserTeam) (*api.UserTeam, error) {
	if teamMember.User != nil && len(teamMember.User.Email) > 0 {
		user, err := s.db.FindUserByEmail(teamMember.User.Email)
		if errors.IsKind(err, errors.ErrNotFound) {
			newUser, createErr := s.db.CreateUser(api.User{
				Email: strings.ToLower(teamMember.User.Email),
			})
			if createErr != nil {
				return nil, errors.Create(createErr)
			}
			teamMember.UserID = newUser.ID
			teamMember.User = newUser
		} else {
			teamMember.UserID = user.ID
			teamMember.User = user
		}
	}

	if teamMember.Role == "" {
		teamMember.Role = api.Member
	}

	if !teamMember.Role.Valid() {
		return nil, errors.Validation(`Role is not valid`)
	}

	validationErr := validator.New().Struct(teamMember)
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}
	teamM, err := s.db.CreateTeamMember(teamMember)
	if err != nil {
		return nil, err
	}
	return teamM, nil
}

func (s vulcanitoService) UpdateTeamMember(ctx context.Context, teamMember api.UserTeam) (*api.UserTeam, error) {
	err := validator.New().Struct(teamMember)
	if err != nil {
		return nil, errors.Validation(err)
	}

	return s.db.UpdateTeamMember(teamMember)
}

func (s vulcanitoService) DeleteTeamMember(ctx context.Context, teamID string, userID string) error {
	if teamID == "" {
		return errors.Validation(`Team ID is empty`)
	}
	if userID == "" {
		return errors.Validation(`User ID is empty`)
	}

	return s.db.DeleteTeamMember(teamID, userID)
}
