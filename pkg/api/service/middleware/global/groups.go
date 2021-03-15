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

func (e *globalEntities) ListAssetGroup(ctx context.Context, a api.AssetGroup, teamID string) ([]*api.Asset, error) {
	// The global entities have precedence.
	g, ok := e.store.Groups()[a.GroupID]
	if !ok {
		return e.VulcanitoService.ListAssetGroup(ctx, a, teamID)
	}
	// if the global group is shadowing a team group return not found
	// as the global shadowing global groups are not queryable.
	if g.ShadowTeamGroup() != "" {
		return nil, errors.ErrNotFound
	}
	return g.Eval(teamID)
}

func (e *globalEntities) ListGroups(ctx context.Context, teamID, groupName string) ([]*api.Group, error) {
	groups, err := e.VulcanitoService.ListGroups(ctx, teamID, groupName)
	if err != nil {
		if !errors.IsKind(err, errors.ErrNotFound) {
			return nil, err
		}
		groups = []*api.Group{}
	}

	global := e.store.Groups()
	// TODO: Because how the response is defined in:
	// pkg/api/assets_group.go#Group.ToResponse() we must load all the assets for
	// a group in memory only for returning the count. We should modify the type
	// api.Group to allow to include a field that contains only the count.
	for _, g := range global {
		// Global shadowing groups are not returned
		if g.ShadowTeamGroup() != "" {
			continue
		}
		g, err := globalGroupToGroup(teamID, g)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (e *globalEntities) FindGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	// Global entities take precedence.
	groups := e.store.Groups()
	g, ok := groups[group.ID]
	if !ok {
		return e.VulcanitoService.FindGroup(ctx, group)
	}
	if g.ShadowTeamGroup() != "" {
		return nil, errors.NotFound(errors.ErrNotFound)
	}
	return globalGroupToGroup(group.TeamID, g)
}

func (e *globalEntities) UpdateGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	if _, ok := e.store.Groups()[group.ID]; ok {
		return nil, errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.UpdateGroup(ctx, group)
}

func (e *globalEntities) DeleteGroup(ctx context.Context, group api.Group) error {
	if _, ok := e.store.Groups()[group.ID]; ok {
		return errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.DeleteGroup(ctx, group)
}

func (e *globalEntities) GroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) (*api.AssetGroup, error) {
	if _, ok := e.store.Groups()[assetGroup.GroupID]; ok {
		return nil, errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.GroupAsset(ctx, assetGroup, teamID)
}

func (e *globalEntities) UngroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) error {
	if _, ok := e.store.Groups()[assetGroup.GroupID]; ok {
		return errors.MethodNotAllowed(errEntityNotModifiable)
	}
	return e.VulcanitoService.UngroupAsset(ctx, assetGroup, teamID)
}

func globalGroupToGroup(teamID string, g global.Group) (*api.Group, error) {
	// TODO: Because how the response is defined in:
	// pkg/api/assets_group.go#Group.ToResponse() we must load all the assets
	// for a group in memory only for returning the count. We should modify the
	// type api.Group to allow to include a field that contains only the count.
	assets, err := g.Eval(teamID)
	if err != nil {
		return nil, err
	}
	var description *string
	if g.Description() != "" {
		aux := g.Description()
		description = &aux
	}
	groupResponse := &api.Group{
		ID:          g.Name(),
		Name:        g.Name(),
		Options:     g.Options(),
		Description: description,
		TeamID:      teamID,
	}
	assetGroup := []*api.AssetGroup{}
	for _, a := range assets {
		a := a
		assetGroup = append(assetGroup, &api.AssetGroup{
			AssetID: a.ID,
			Asset:   a,
		})
	}
	groupResponse.AssetGroup = assetGroup
	return groupResponse, nil
}
