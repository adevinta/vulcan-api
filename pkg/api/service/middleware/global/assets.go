/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (e *globalEntities) ListAssets(ctx context.Context, teamID string, asset api.Asset) ([]*api.Asset, error) {
	// Map to store the relationship between asset ID and a slice of Global Group IDs.
	// ex: {"bcebfb2e-efad-4ffb-93a0-cae231fb473d":["default-global","sensitive-global"]}
	assetGlobalGroups := make(map[string][]string)

	// Map to store the result of a conversion of a global group to an API group.
	globalGroupsToAPIGroups := make(map[string]*api.Group)

	// We are going to iterate over each global group.
	for globalGroupID, grobalGroup := range e.store.Groups() {
		// Skip global groups that are shadowing regular API groups.
		if grobalGroup.ShadowTeamGroup() != "" {
			continue
		}

		// Convert a global group to API group.
		group, err := globalGroupToGroup(teamID, grobalGroup)
		if err != nil {
			return nil, errors.Default(err)
		}

		// Store the result of the conversion above into the map.
		globalGroupsToAPIGroups[globalGroupID] = group

		// Append the Global Group ID to the assets map.
		for _, assetGroup := range group.AssetGroup {
			assetGlobalGroups[assetGroup.AssetID] = append(assetGlobalGroups[assetGroup.AssetID], globalGroupID)
		}
		globalGroupsToAPIGroups[globalGroupID].AssetGroup = nil
	}

	// Retrieve the actual list of assets for the current team.
	assets, err := e.VulcanitoService.ListAssets(ctx, teamID, asset)
	if err != nil {
		return nil, errors.Default(err)
	}

	// For each asset, perform a lookup over the global group IDs associated with that asset ID.
	for _, asset := range assets {
		// For each global group ID, appends an AssetGroup entity to the array of AssetGroups.
		for _, groupID := range assetGlobalGroups[asset.ID] {
			asset.AssetGroups = append(asset.AssetGroups, &api.AssetGroup{
				Group:   globalGroupsToAPIGroups[groupID],
				GroupID: groupID,
				Asset:   asset,
				AssetID: asset.ID,
			})
		}
	}

	return assets, err
}
