/*
Copyright 2021 Adevinta
*/

package global

import (
	"encoding/json"
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func init() {
	// All the global groups must be defined here by calling the registerGroup function.
	registerGroup(&DefaultGroup{&globalGroup{}})
	registerGroup(&SensitiveGroup{&globalGroup{}})
	registerGroup(&RedconGroup{&globalGroup{}})
	registerGroup(&WebScanningGroup{&globalGroup{}})
	registerGroup(&CPGroup{&globalGroup{}})
}

// DefaultGroup resolves all the assets present
type DefaultGroup struct {
	*globalGroup
}

// Name returns the name of the group.
func (d *DefaultGroup) Name() string {
	return "default-global"
}

// Description returns a meanfull explanation of the group.
func (d *DefaultGroup) Description() string {
	return "This global group contains all the assets that are in the default " +
		"group of your team and not in the sensitive group. Assets in this group " +
		"will be scanned using the default global policy."
}

// Eval returns the current assets of a team belinging to this group.
func (d *DefaultGroup) Eval(teamID string) ([]*api.Asset, error) {
	// Find the group id of the Default group of the team.
	dg, err := d.Store.FindGroupInfo(api.Group{
		Name:   "Default",
		TeamID: teamID,
	})
	if err != nil {
		if errors.IsRootOfKind(err, errors.ErrNotFound) {
			return []*api.Asset{}, nil
		}
		return nil, err
	}

	sg, err := d.Store.FindGroupInfo(api.Group{
		Name:   "Sensitive",
		TeamID: teamID,
	})

	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}

	if err == nil {
		assets, err := d.Store.DisjoinAssetsInGroups(teamID, dg.ID, []string{sg.ID})
		if err != nil {
			return nil, err
		}
		for _, a := range assets {
			a := a
			err = addSecurityLevel(a)
			if err != nil {
				return nil, err
			}
		}
		return assets, nil
	}
	assetGroups, err := d.Store.FindGroup(api.Group{
		Name:   "Default",
		TeamID: teamID,
	})
	if err != nil {
		return nil, err
	}
	assets := []*api.Asset{}
	for _, a := range assetGroups.AssetGroup {
		aa := a.Asset
		err = addSecurityLevel(aa)
		if err != nil {
			return nil, err
		}
		assets = append(assets, aa)
	}
	return assets, nil
}

func addSecurityLevel(a *api.Asset) error {
	options := map[string]interface{}{}
	if a.Options != nil && *a.Options != "" {
		// By now we consider that if the options in the db are not empty and
		// not null they are valid json. If that's not the case the scan will
		// fail in any case as the scan engine has also to unmarshal them.
		err := json.Unmarshal([]byte(*a.Options), &options)
		if err != nil {
			return err
		}
	}
	// The default group adds automatically the roflp level as an option to the asset.
	options["security_level"] = a.ROLFP.Level()
	content, err := json.Marshal(options)
	if err != nil {
		return err
	}
	opts := string(content)
	a.Options = &opts
	return nil
}

// SensitiveGroup global group shadows the sensitive concrete group of a team.
type SensitiveGroup struct {
	*globalGroup
}

// Name returns the name of the group.
func (d *SensitiveGroup) Name() string {
	return "sensitive-global"
}

func (d *SensitiveGroup) Description() string {
	return `Assets in the sensitive group of the team`
}

func (g *SensitiveGroup) ShadowTeamGroup() string {
	return "Sensitive"
}

// RedconGroup resolves the assets detected by Redcon excluding those present
// in the Default and Sensitive groups.
type RedconGroup struct {
	*globalGroup
}

// Name returns the name of the group.
func (g *RedconGroup) Name() string {
	return "redcon-global"
}

// Description returns a meaningful explanation of the group.
func (g *RedconGroup) Description() string {
	return `This global group contains the assets detected by Redcon excluding those present in the Default and Sensitive groups.`
}

// Eval returns the current assets of a team belonging to this group.
func (g *RedconGroup) Eval(teamID string) ([]*api.Asset, error) {
	// Find the group id of the Redcon group of the team.
	rg, err := g.Store.FindGroupInfo(api.Group{
		Name:   api.DiscoveredAssetsGroupName,
		TeamID: teamID,
	})
	if err != nil {
		if errors.IsRootOfKind(err, errors.ErrNotFound) {
			return []*api.Asset{}, nil
		}
		return nil, err
	}

	var excluded []string

	dg, err := g.Store.FindGroupInfo(api.Group{
		Name:   "Default",
		TeamID: teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("Default group not found: %w", err)
	}

	excluded = append(excluded, dg.ID)

	sg, err := g.Store.FindGroupInfo(api.Group{
		Name:   "Sensitive",
		TeamID: teamID,
	})
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}

	if err == nil {
		excluded = append(excluded, sg.ID)
	}

	return g.Store.DisjoinAssetsInGroups(teamID, rg.ID, excluded)
}

// WebScanning global group contains the assets which will be scanned by web scanners.
type WebScanningGroup struct {
	*globalGroup
}

// Name returns the name of the group.
func (d *WebScanningGroup) Name() string {
	return "web-scanning-global"
}

func (d *WebScanningGroup) Description() string {
	return `assets scanned by web scanners`
}

func (g *WebScanningGroup) ShadowTeamGroup() string {
	return ""
}

// Eval returns the current assets of a team belonging to this group.
func (g *WebScanningGroup) Eval(teamID string) ([]*api.Asset, error) {
	// Find the group id of the WebScanning group of the team.
	wsg, err := g.Store.FindGroupInfo(api.Group{
		Name:   api.WebScanningAssetsGroupName,
		TeamID: teamID,
	})
	if err != nil {
		if errors.IsRootOfKind(err, errors.ErrNotFound) {
			return []*api.Asset{}, nil
		}
		return nil, err
	}

	var excluded []string

	return g.Store.DisjoinAssetsInGroups(teamID, wsg.ID, excluded)
}

// CPGroup resolves the assets detected by CP excluding those present
// in the Default, Sensitive and Redcon groups.
type CPGroup struct {
	*globalGroup
}

// Name returns the name of the group.
func (g *CPGroup) Name() string {
	return "cp-global"
}

// Description returns a meaningful explanation of the group.
func (g *CPGroup) Description() string {
	return "This global group contains the Common Platform assets as discovered " +
		"by the CP, excluding those present in the Default, Sensitive or Redcon " +
		"groups."
}

// Eval returns the current assets of a team belonging to this group.
func (g *CPGroup) Eval(teamID string) ([]*api.Asset, error) {
	// Find the group id of the CP group of the team.
	cpg, err := g.Store.FindGroupInfo(api.Group{
		Name:   api.CPDiscoveredAssetsGroupName,
		TeamID: teamID,
	})
	if err != nil {
		if errors.IsRootOfKind(err, errors.ErrNotFound) {
			return []*api.Asset{}, nil
		}
		return nil, err
	}

	var excluded []string

	dg, err := g.Store.FindGroupInfo(api.Group{
		Name:   "Default",
		TeamID: teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("Default group not found: %w", err)
	}
	excluded = append(excluded, dg.ID)

	sg, err := g.Store.FindGroupInfo(api.Group{
		Name:   "Sensitive",
		TeamID: teamID,
	})
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		excluded = append(excluded, sg.ID)
	}

	rg, err := g.Store.FindGroupInfo(api.Group{
		Name:   api.DiscoveredAssetsGroupName,
		TeamID: teamID,
	})
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		return nil, err
	}
	if err == nil {
		excluded = append(excluded, rg.ID)
	}

	return g.Store.DisjoinAssetsInGroups(teamID, cpg.ID, excluded)
}
