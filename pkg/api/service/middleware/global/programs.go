/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	global "github.com/adevinta/vulcan-api/pkg/api/store/global"
)

func (e *globalEntities) CreateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {
	if _, ok := e.store.Programs()[program.ID]; ok {
		return nil, errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.CreateProgram(ctx, program, teamID)
}

func (e *globalEntities) ListPrograms(ctx context.Context, teamID string) ([]*api.Program, error) {
	programs, err := e.VulcanitoService.ListPrograms(ctx, teamID)
	if err != nil {
		return nil, err
	}
	globalPrograms := e.store.Programs()
	for n := range globalPrograms {
		program, err := e.FindProgram(ctx, n, teamID)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	return programs, nil
}

func (e *globalEntities) findPoliciesGroups(ctx context.Context, teamID string, policiesGroups []global.PolicyGroup) ([]*api.ProgramsGroupsPolicies, error) {
	res := []*api.ProgramsGroupsPolicies{}
	for _, pg := range policiesGroups {
		p, ok := e.store.Policies()[pg.Policy]
		if !ok {
			return nil, errors.Default(fmt.Sprintf("no global policy with name %s defined", pg.Policy))
		}
		g, ok := e.store.Groups()[pg.Group]
		if !ok {
			return nil, errors.Default(fmt.Sprintf("no global group with name %s defined", pg.Group))
		}
		policy, err := globalPolicyToPolicy(ctx, p)
		if err != nil {
			return nil, err
		}
		var group *api.Group
		if g.ShadowTeamGroup() == "" {
			group, err = globalGroupToGroup(teamID, g)

		} else {
			group, err = e.VulcanitoService.FindGroup(ctx, api.Group{
				TeamID: teamID,
				Name:   g.ShadowTeamGroup(),
			})
		}
		if err != nil {
			if errors.IsKind(err, errors.ErrNotFound) {
				continue
			}
			return nil, err
		}
		if g.ShadowTeamGroup() != "" {
			group.AssetGroup = filterNonScannableAssets(group.AssetGroup)
		}
		groupPolicy := &api.ProgramsGroupsPolicies{
			GroupID:  group.ID,
			Group:    group,
			PolicyID: policy.Name,
			Policy:   policy,
		}
		res = append(res, groupPolicy)
	}
	return res, nil
}

func (e *globalEntities) FindProgram(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	p, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.FindProgram(ctx, programID, teamID)
	}
	metadata, err := e.metadata.FindGlobalProgramMetadata(programID, teamID)
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}
	if err != nil && errors.IsKind(err, errors.ErrNotFound) {
		metadata = &p.DefaultMetadata
	}

	var autosend bool
	if metadata.Autosend != nil {
		autosend = *metadata.Autosend
	}

	var disabled bool
	if metadata.Disabled != nil {
		disabled = *metadata.Disabled
	}

	var cron = metadata.Cron

	name := p.Name
	if name == "" {
		name = programID
	}
	global := true
	program := &api.Program{
		Autosend: &autosend,
		Disabled: &disabled,
		ID:       programID,
		Name:     name,
		Cron:     cron,
		Global:   &global,
	}
	policyGroups, err := e.findPoliciesGroups(ctx, teamID, p.Policies)
	if err != nil {
		return nil, err
	}
	program.ProgramsGroupsPolicies = policyGroups
	return program, nil
}

func (e *globalEntities) UpdateProgram(ctx context.Context, program api.Program, teamID string) (*api.Program, error) {
	gp, ok := e.store.Programs()[program.ID]
	if !ok {
		return e.VulcanitoService.UpdateProgram(ctx, program, teamID)
	}
	// We allow to modify autosend and disabled flags.
	if (program.Autosend == nil && program.Disabled == nil) || program.Name != "" {
		return nil, errors.Validation("only autosend and disabled fields can be modified for global program")
	}
	defaultAutosend := false
	if gp.DefaultMetadata.Autosend != nil {
		defaultAutosend = *gp.DefaultMetadata.Autosend
	}
	defaultDisabled := false
	if gp.DefaultMetadata.Disabled != nil {
		defaultDisabled = *gp.DefaultMetadata.Disabled
	}
	err := e.metadata.UpsertGlobalProgramMetadata(teamID, program.ID, defaultAutosend, defaultDisabled, gp.DefaultMetadata.Cron, program.Autosend, program.Disabled, nil)
	if err != nil {
		return nil, err
	}
	ret, err := e.FindProgram(ctx, program.ID, program.TeamID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
func (e *globalEntities) DeleteProgram(ctx context.Context, program api.Program, teamID string) error {
	if _, ok := e.store.Programs()[program.ID]; ok {
		return errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.DeleteProgram(ctx, program, teamID)
}

func (e *globalEntities) CreateSchedule(ctx context.Context, programID string, cronExpr string, teamID string) (*api.Program, error) {
	gp, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.CreateSchedule(ctx, programID, cronExpr, teamID)
	}
	p, err := e.FindProgram(ctx, programID, teamID)
	if err != nil {
		return nil, err
	}
	defaultAutosend := false
	if gp.DefaultMetadata.Autosend != nil {
		defaultAutosend = *gp.DefaultMetadata.Autosend
	}
	defaultDisabled := false
	if gp.DefaultMetadata.Disabled != nil {
		defaultDisabled = *gp.DefaultMetadata.Disabled
	}
	err = e.metadata.UpsertGlobalProgramMetadata(teamID, programID, defaultAutosend, defaultDisabled, gp.DefaultMetadata.Cron, nil, nil, &cronExpr)
	if err != nil {
		return nil, err
	}
	if err := e.scheduler.CreateSchedule(programID, teamID, cronExpr); err != nil {
		return nil, err
	}
	p.Cron = cronExpr
	return p, nil
}

func (e *globalEntities) DeleteSchedule(ctx context.Context, programID string, teamID string) (*api.Program, error) {
	_, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.DeleteSchedule(ctx, programID, teamID)
	}
	p, err := e.FindProgram(ctx, programID, teamID)
	if err != nil {
		return nil, err
	}
	err = e.scheduler.DeleteSchedule(teamID, programID)
	if err != nil {
		if err.Error() == errSchedulerNotFound {
			return nil, errors.NotFound("no schedule for the program was found")
		}
		return nil, err
	}
	p.Cron = ""
	return p, nil
}

func filterNonScannableAssets(ag []*api.AssetGroup) []*api.AssetGroup {
	filtered := []*api.AssetGroup{}
	for _, a := range ag {
		a := a
		if a.Asset == nil {
			continue
		}
		if a.Asset.Scannable == nil || !*a.Asset.Scannable {
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}
