/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	global "github.com/adevinta/vulcan-api/pkg/api/store/global"
)

func (e *globalEntities) ListPolicies(ctx context.Context, teamID string) ([]*api.Policy, error) {
	policies, err := e.VulcanitoService.ListPolicies(ctx, teamID)
	if err != nil {

		if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
			return nil, err
		}
		policies = []*api.Policy{}
	}
	globalPolicies := e.store.Policies()
	for n, p := range globalPolicies {
		checktypes, err := p.Eval(ctx, e.globalPolicyConfig)
		if err != nil {
			return nil, err
		}
		policy := &api.Policy{
			ID:                n,
			Name:              n,
			TeamID:            teamID,
			ChecktypeSettings: checktypes,
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (e *globalEntities) FindPolicy(ctx context.Context, policyID string) (*api.Policy, error) {
	p, ok := e.store.Policies()[policyID]
	if !ok {
		return e.VulcanitoService.FindPolicy(ctx, policyID)
	}
	return globalPolicyToPolicy(ctx, e.globalPolicyConfig, p)
}

func (e *globalEntities) UpdatePolicy(ctx context.Context, policy api.Policy) (*api.Policy, error) {
	if _, ok := e.store.Policies()[policy.ID]; ok {
		return nil, errors.Forbidden(errEntityNotModifiable)
	}
	return e.VulcanitoService.UpdatePolicy(ctx, policy)
}

func (e *globalEntities) DeletePolicy(ctx context.Context, policy api.Policy) error {
	if _, ok := e.store.Policies()[policy.ID]; ok {
		return errors.Forbidden(errEntityNotModifiable)
	}
	return e.VulcanitoService.DeletePolicy(ctx, policy)
}

func (e *globalEntities) ListChecktypeSetting(ctx context.Context, policyID string) ([]*api.ChecktypeSetting, error) {
	p, ok := e.store.Policies()[policyID]
	if !ok {
		return e.VulcanitoService.ListChecktypeSetting(ctx, policyID)
	}
	return p.Eval(ctx, e.globalPolicyConfig)
}

func (e *globalEntities) FindChecktypeSetting(ctx context.Context, policyID, checktypeSettingID string) (*api.ChecktypeSetting, error) {
	p, ok := e.store.Policies()[policyID]
	if !ok {
		return e.VulcanitoService.FindChecktypeSetting(ctx, policyID, checktypeSettingID)
	}

	checktypes, err := p.Eval(ctx, e.globalPolicyConfig)
	if err != nil {
		return nil, err
	}
	// Find the checktype.
	var checktype *api.ChecktypeSetting
	for _, c := range checktypes {
		c := c
		if c.CheckTypeName == checktypeSettingID {
			checktype = c
			break
		}
	}
	if checktype == nil {
		return nil, errors.ErrNotFound
	}
	ret := &api.ChecktypeSetting{
		CheckTypeName: checktype.CheckTypeName,
		ID:            checktype.ID,
		Options:       checktype.Options,
	}
	return ret, nil
}

func (e *globalEntities) UpdateChecktypeSetting(ctx context.Context, checktypeSetting api.ChecktypeSetting) (*api.ChecktypeSetting, error) {
	if _, ok := e.store.Policies()[checktypeSetting.PolicyID]; ok {
		return nil, errors.Forbidden(errEntityNotModifiable)
	}
	return e.VulcanitoService.UpdateChecktypeSetting(ctx, checktypeSetting)
}

func globalPolicyToPolicy(ctx context.Context, gpc global.GlobalPolicyConfig, p global.Policy) (*api.Policy, error) {
	settings, err := p.Eval(ctx, gpc)
	if err != nil {
		return nil, err
	}
	var description *string
	if p.Description() != "" {
		aux := p.Description()
		description = &aux
	}
	ret := &api.Policy{
		ID:                p.Name(),
		Name:              p.Name(),
		ChecktypeSettings: settings,
		Description:       description,
	}
	return ret, nil
}
