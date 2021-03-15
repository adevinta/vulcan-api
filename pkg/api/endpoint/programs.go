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

type ProgramRequest struct {
	ID           string               `json:"id" urlvar:"program_id"`
	Name         string               `json:"name"`
	PolicyGroups *ProgramPolicyGroups `json:"policy_groups"`
	Autosend     *bool                `json:"autosend"`
	Disabled     *bool                `json:"disabled"`
	TeamID       string               `json:"team_id" urlvar:"team_id"`
}

// ProgramPolicyGroups stores a slice with the list os pairs policyid groupid
// for a program.
type ProgramPolicyGroups []ProgramsPolicyGroup

// ToAPI returns the represetantion needed by vulcanito service for
// the slice of tuples of policies-groups associated to a program
func (p ProgramPolicyGroups) ToAPI() []*api.ProgramsGroupsPolicies {
	groupsPolicies := []*api.ProgramsGroupsPolicies{}
	for _, t := range p {
		groupsPolicies = append(groupsPolicies, &api.ProgramsGroupsPolicies{
			GroupID:  t.GroupID,
			PolicyID: t.PolicyID,
		})
	}
	return groupsPolicies
}

// ProgramsPolicyGroup holds the tuples (PolicyID,GroupID) that defines a
// a set of checktypes to be executed against a set of assets in a program.
type ProgramsPolicyGroup struct {
	PolicyID string `json:"policy_id" validate:"required"`
	GroupID  string `json:"group_id" validate:"required"`
}

// ScheduleRequest holds the payload required for the endpoint that schedules a program.
type ScheduleRequest struct {
	ID     string `json:"id" urlvar:"program_id"`
	TeamID string `json:"team_id" urlvar:"team_id"`
	Cron   string `json:"cron"`
}

// ScheduleGlobalRequest holds the payload required for the endpoint that schedules global programs.
type ScheduleGlobalRequest struct {
	ID   string `json:"id" urlvar:"program_id"`
	Cron string `json:"cron"`
}

func makeListProgramsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		programRequest, ok := request.(*ProgramRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		programs, err := s.ListPrograms(ctx, programRequest.TeamID)
		if err != nil {
			return nil, err
		}
		response := []api.ProgramResponse{}
		for _, program := range programs {
			response = append(response, *program.ToResponse())
		}
		return Ok{response}, nil
	}
}

func makeCreateProgramEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		programRequest, ok := request.(*ProgramRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		// Ensure there are no nil values that can cause a panic.
		if programRequest.PolicyGroups == nil {
			programRequest.PolicyGroups = &ProgramPolicyGroups{}
		}
		// When creating program if autosend is nil we assume false.
		var autosend bool
		if programRequest.Autosend != nil {
			autosend = *programRequest.Autosend
		}

		// When creating program if disabled is nil we assume false. (default value)
		var disabled = false
		if programRequest.Disabled != nil {
			disabled = *programRequest.Disabled
		}

		program := api.Program{
			Name:                   programRequest.Name,
			TeamID:                 programRequest.TeamID,
			Autosend:               &autosend,
			Disabled:               &disabled,
			ProgramsGroupsPolicies: programRequest.PolicyGroups.ToAPI(),
		}
		createdProgram, err := s.CreateProgram(ctx, program, programRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Created{createdProgram.ToResponse()}, nil
	}
}

func makeFindProgramEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		programRequest, ok := request.(*ProgramRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		found, err := s.FindProgram(ctx, programRequest.ID, programRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{found.ToResponse()}, nil
	}
}

// TODO: We are using the same struct ProgramRequest for creating a program
// and for updating a program. That is a problem because if the user does not set
// a field of the request after the update the program will have all those fields
// setted to the default go value which does not make sense.

func makeUpdateProgramEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		programRequest, ok := request.(*ProgramRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Ensure there are no nil values that can cause a panic.
		if programRequest.PolicyGroups == nil {
			programRequest.PolicyGroups = &ProgramPolicyGroups{}
		}
		program := api.Program{
			ID:                     programRequest.ID,
			Name:                   programRequest.Name,
			Autosend:               programRequest.Autosend,
			Disabled:               programRequest.Disabled,
			TeamID:                 programRequest.TeamID,
			ProgramsGroupsPolicies: programRequest.PolicyGroups.ToAPI(),
		}

		updated, err := s.UpdateProgram(ctx, program, programRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{updated.ToResponse()}, nil
	}
}

func makeDeleteProgramEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		programRequest, ok := request.(*ProgramRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		program := api.Program{
			ID: programRequest.ID,
		}
		err = s.DeleteProgram(ctx, program, programRequest.TeamID)
		if err != nil {
			return nil, err
		}
		return NoContent{nil}, nil
	}
}

func makeCreateScheduleEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*ScheduleRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		r.Cron = strings.TrimSpace(r.Cron)

		updated, err := s.CreateSchedule(ctx, r.ID, r.Cron, r.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{updated.ToResponse()}, nil
	}
}

func makeScheduleGlobalProgramEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*ScheduleGlobalRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		err = s.ScheduleGlobalProgram(ctx, r.ID, r.Cron)
		if err != nil {
			return nil, err
		}
		return Ok{}, nil
	}
}

func makeDeleteScheduleEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r, ok := request.(*ScheduleRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		updated, err := s.DeleteSchedule(ctx, r.ID, r.TeamID)
		if err != nil {
			return nil, err
		}
		return Ok{updated.ToResponse()}, nil
	}
}

func makeListProgramScansEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		scanRequest, ok := request.(*ListProgramScansRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		scans, err := s.ListScans(ctx, scanRequest.TeamID, scanRequest.ProgramID)
		if err != nil {
			return nil, err
		}
		response := []*api.ScanResponse{}
		for _, scan := range scans {
			response = append(response, scan.ToResponse())
		}
		return Ok{response}, nil
	}
}
