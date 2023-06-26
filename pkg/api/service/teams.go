/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// isATeamOnboardedInVulcanTracker return if a team is onboarded in vulcan tracker.
func isATeamOnboardedInVulcanTracker(teamID string, onboardedTeams []string) bool {
	for _, team := range onboardedTeams {
		if team == teamID {
			return true
		}
	}
	return false
}

func (s vulcanitoService) CreateTeam(ctx context.Context, team api.Team, ownerEmail string) (*api.Team, error) {
	err := validator.New().Struct(team)
	if err != nil {
		return nil, errors.Validation(err)
	}

	if ownerEmail == "" {
		return nil, errors.Validation(`Owner email is empty`)
	}

	return s.db.CreateTeam(team, ownerEmail)
}

func (s vulcanitoService) UpdateTeam(ctx context.Context, team api.Team) (*api.Team, error) {
	err := validator.New().Struct(team)
	if err != nil {
		return nil, errors.Validation(err)
	}

	return s.db.UpdateTeam(team)
}

func (s vulcanitoService) FindTeam(ctx context.Context, id string) (*api.Team, error) {
	if id == "" {
		return nil, errors.Validation(`ID is empty`)
	}
	team, err := s.db.FindTeam(id)
	if err != nil {
		return team, err
	}
	if s.vulcantrackerClient == nil {
		team.UsingTracker = false
	} else {
		team.UsingTracker = isATeamOnboardedInVulcanTracker(id, s.allowedTrackerTeams)
	}
	return team, nil
}

func (s vulcanitoService) FindTeamByName(ctx context.Context, name string) (*api.Team, error) {
	if name == "" {
		return nil, errors.Validation(`Name is empty`)
	}
	return s.db.FindTeamByName(name)
}

func (s vulcanitoService) FindTeamByTag(ctx context.Context, tag string) (*api.Team, error) {
	if tag == "" {
		return nil, errors.Validation(`Tag is empty`)
	}
	team, err := s.db.FindTeamByTag(tag)
	if err != nil {
		return team, err
	}
	if s.vulcantrackerClient == nil {
		return team, nil
	}
	team.UsingTracker = isATeamOnboardedInVulcanTracker(team.ID, s.allowedTrackerTeams)

	return team, nil
}

func (s vulcanitoService) FindTeamsByTags(ctx context.Context, tags []string) ([]*api.Team, error) {
	if len(tags) == 0 {
		return nil, errors.Validation(`No tags specified`)
	}
	return s.db.FindTeamsByTags(tags)
}

func (s vulcanitoService) FindTeamsByUser(ctx context.Context, userID string) ([]*api.Team, error) {
	if userID == "" {
		return nil, errors.Validation(`ID is empty`)
	}
	return s.db.FindTeamsByUser(userID)
}

func (s vulcanitoService) DeleteTeam(ctx context.Context, id string) error {
	if id == "" {
		return errors.Validation(`ID is empty`)
	}
	return s.db.DeleteTeam(id)
}

func (s vulcanitoService) ListTeams(ctx context.Context) ([]*api.Team, error) {
	teams, err := s.db.ListTeams()
	if err != nil {
		return teams, err
	}
	if s.vulcantrackerClient == nil {
		return teams, nil
	}
	for _, team := range teams {
		team.UsingTracker = isATeamOnboardedInVulcanTracker(team.ID, s.allowedTrackerTeams)
	}
	return teams, nil
}
