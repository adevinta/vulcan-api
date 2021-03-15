/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/global"
	"github.com/adevinta/vulcan-api/pkg/schedule"
)

// Create program, creates a new program in the db and schedules it if needed.
func (s vulcanitoService) CreateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {
	validationErr := validator.New().Struct(program)
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}

	// Check at least one relation policyID-groupID was specified.
	if len(program.ProgramsGroupsPolicies) < 1 {
		return nil, errors.Validation("At least one policy group must be specified")
	}

	// Check that all the GroupPolicies have the policyID and groupID set.
	for _, p := range program.ProgramsGroupsPolicies {
		if p == nil {
			return nil, errors.Default("unexpected nil pointer")
		}
		if p.GroupID == "" || p.PolicyID == "" {
			return nil, errors.Validation("all the policy groups must specify a valid existing policy and group")
		}
	}

	return s.db.CreateProgram(program, teamID)
}

func (s vulcanitoService) CreateSchedule(ctx context.Context, programID string, cronExpr string, teamID string) (*api.Program, error) {
	p, err := s.db.FindProgram(programID, teamID)
	if err != nil {
		return nil, err
	}
	err = s.programScheduler.CreateScanSchedule(programID, teamID, cronExpr)
	if err != nil {
		if err == schedule.ErrInvalidCronExpr || err == schedule.ErrInvalidSchedulePeriod {
			return nil, errors.Assertion(err)
		}
		return nil, err
	}
	p.Cron = cronExpr
	return p, nil
}

// ScheduleGlobalProgram overrides the given global program cron for every team.
func (s vulcanitoService) ScheduleGlobalProgram(ctx context.Context, programID string, cronExpr string) error {
	if programID != global.PeriodicFullScan.ID {
		return errors.Assertion("Program ID does not correspond to a global program")
	}

	// Retrieve current authenticated user from context
	// and validate that has admin privileges
	currentUser, err := api.UserFromContext(ctx)
	if err != nil {
		_ = s.logger.Log(err.Error())
		return errors.Default(err)
	}
	if currentUser.Admin == nil || !*currentUser.Admin {
		return errors.Unauthorized("Can not schedule program")
	}

	// Build Bulk Schedule and schedule it
	teams, err := s.ListTeams(ctx)
	if err != nil {
		return err
	}

	schedules := make([]schedule.ScanBulkSchedule, len(teams))

	for i, t := range teams {
		bulkSchedule := schedule.ScanBulkSchedule{
			Str:       cronExpr,
			ProgramID: programID,
			TeamID:    t.ID,
			Overwrite: true,
		}
		schedules[i] = bulkSchedule
	}

	err = s.programScheduler.BulkCreateScanSchedules(schedules)
	if err == schedule.ErrInvalidCronExpr || err == schedule.ErrInvalidSchedulePeriod {
		return errors.Assertion(err)
	}
	return err
}

// DeleteSchedule deletes and schedules and returns the program information with the cron string updated to empty.
func (s vulcanitoService) DeleteSchedule(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	p, err := s.db.FindProgram(programID, teamID)
	if err != nil {
		return nil, err
	}
	err = s.programScheduler.DeleteScanSchedule(programID)
	if err != nil {
		return nil, err
	}
	p.Cron = ""
	return p, nil
}

func (s vulcanitoService) ListPrograms(ctx context.Context, teamID string) ([]*api.Program, error) {
	programs, err := s.db.ListPrograms(teamID)
	if err != nil {
		return nil, err
	}
	// TODO: This will could potentially be a problem because it makes
	// a request for every program of a team to know if a program is scheduled or not.
	// In future we should implement an endpoint in the scheduler that accepts a set of id's.
	for _, p := range programs {
		cron, err := s.programScheduler.GetScanScheduleByID(p.ID)
		if err != nil {
			_ = s.logger.Log("GetScheduleByIDError", fmt.Sprintf(`%v`, err), "ID", p.ID)
			if err.Error() != "ScheduleNotFound" {
				// We need to wrap the error because the scheduler package does no uses the spt-security/errors
				// package. All errors returned by the scheduler arriving here will be considered to be
				// Internal (status 500)
				return nil, errors.Default(err)
			}
			continue
		}
		p.Cron = cron
	}
	return programs, nil
}

func (s vulcanitoService) FindProgram(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	p, err := s.db.FindProgram(programID, teamID)
	if err != nil {
		return &api.Program{}, err
	}

	cron, err := s.programScheduler.GetScanScheduleByID(p.ID)
	if err != nil && err.Error() != "ScheduleNotFound" {
		return &api.Program{}, err
	}

	p.Cron = cron

	return p, nil
}

func (s vulcanitoService) UpdateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {

	// Check that all the GroupPolicies have the policyID and groupID set.
	for _, p := range program.ProgramsGroupsPolicies {
		if p == nil {
			return nil, errors.Default("unexpected nil pointer")
		}
		if p.GroupID == "" || p.PolicyID == "" {
			return nil, errors.Validation("all the policy groups must specify a valid existing policy and group")
		}
	}
	return s.db.UpdateProgram(program, teamID)
}

func (s vulcanitoService) DeleteProgram(ctx context.Context, program api.Program, teamID string) error {
	return s.db.DeleteProgram(program, teamID)
}
