/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	goerrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adevinta/errors"

	types "github.com/adevinta/vulcan-types"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/adevinta/vulcan-api/pkg/common"
)

type asset struct {
	identifier string
	assetType  string
}

func (s vulcanitoService) ListAssets(ctx context.Context, teamID string) ([]*api.Asset, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}

	result, err := s.db.ListAssets(teamID)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

// CreateAssets receives an array of assets and creates them on store layer.
func (s vulcanitoService) CreateAssets(ctx context.Context, assets []api.Asset, groups []api.Group) ([]api.Asset, error) {
	assetsToCreate := []api.Asset{}

	// If no group is specified, add Default group data to groups list.
	if len(groups) == 0 && len(assets) > 0 {
		groups = append(groups, api.Group{TeamID: assets[0].TeamID, Name: "Default"})
	}

	// Validate groups and their belonging to the specified team ID.
	for i, g := range groups {
		group, err := s.db.FindGroupInfo(g)
		if err != nil {
			errMssg := fmt.Sprintf("Unable to find group %s: %v", g.ID, err)
			return nil, errors.NotFound(errMssg)
		}
		groups[i] = *group
	}

	// Iterate over the assets list and identify assets to be created.
	// The reason for this is that the user can request the creation of assets
	// without specifying the asset type. In this case, the API will try to
	// automatically detect the asset type.
	for _, asset := range assets {
		asset := asset

		// Asset type provided by the user in the request.
		if asset.AssetType != nil && asset.AssetType.Name != "" {
			if !validAssetType(asset.AssetType.Name) {
				return nil, errors.Validation("Asset type not found", "asset", asset.Identifier, asset.AssetType.Name)
			}

			// Retrieve asset type from its name.
			assetTypeObj, err := s.GetAssetType(ctx, asset.AssetType.Name)
			if err != nil {
				return nil, err
			}

			asset.AssetTypeID = assetTypeObj.ID
			asset.AssetType = &api.AssetType{Name: assetTypeObj.Name}

			if err := asset.Validate(); err != nil {
				return nil, err
			}
			assetsToCreate = append(assetsToCreate, asset)
		} else {
			// Asset type NOT provided by the user in the request. Try to infere
			// asset type based on the identifier
			assetsDetected, err := s.detectAssets(ctx, asset)
			if err != nil {
				return nil, errors.Validation(err, "asset", asset.Identifier, asset.AssetType.Name)
			}

			assetsToCreate = append(assetsToCreate, assetsDetected...)
		}
	}

	// For all AWSAccount assets that do not specify an Alias, try to
	// automatically fetch one
	for i, a := range assetsToCreate {
		if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
			a.Alias = s.getAccountName(a.Identifier)
			assetsToCreate[i] = a
		}
	}
	return s.db.CreateAssets(assetsToCreate, groups)
}

func (s vulcanitoService) getAccountName(identifier string) string {
	// Get the identifier of the account.
	// arn:aws:iam::552016233890:root .
	id := strings.Replace(identifier, "arn:aws:iam::", "", -1)
	// 552016233890:root
	id = strings.Replace(id, ":root", "", -1)
	id = strings.Trim(id, " ")
	name, err := s.awsAccounts.Name(id)
	if err != nil {
		name = identifier
		if goerrors.Is(err, awscatalogue.ErrAccountNotFound) {
			s.logger.Log("ErrorAWSAccountNotFound", identifier)
		} else {
			s.logger.Log("ErrorGettingAWSAccountName", err, "identifier", identifier)
		}
	}
	return name
}

// CreateAssetsMultiStatus receives an array of assets and request their creation to the store layer.
// Also, this method will associate the assets with the specified groups.
// It returns an array containing one response per request, in the same order as in the original request.
func (s vulcanitoService) CreateAssetsMultiStatus(ctx context.Context, assets []api.Asset, groups []api.Group) ([]api.AssetCreationResponse, error) {
	responses := []api.AssetCreationResponse{}

	// If no group is specified, add Default group data to groups list.
	if len(groups) == 0 && len(assets) > 0 {
		groups = append(groups, api.Group{TeamID: assets[0].TeamID, Name: "Default"})
	}

	// Validate groups and their belonging to the specified team ID.
	for i, g := range groups {
		group, err := s.db.FindGroupInfo(g)
		if err != nil {
			errMssg := fmt.Sprintf("Unable to find group %s: %v", g.ID, err)
			return nil, errors.NotFound(errMssg)
		}
		groups[i] = *group
	}

	// Iterate over the assets list and request their creation to the store layer.
	for _, asset := range assets {
		asset := asset

		// Prepare response
		response := api.AssetCreationResponse{
			Identifier:        asset.Identifier,
			AssetType:         asset.AssetType.ToResponse(),
			Options:           asset.Options,
			EnvironmentalCVSS: asset.EnvironmentalCVSS,
			ROLFP:             asset.ROLFP,
			Scannable:         asset.Scannable,
			ClassifiedAt:      asset.ClassifiedAt,
			Alias:             asset.Alias,
			Status:            nil,
		}

		// AssetGroup holds information for the assets to be created.
		// 	· For each input asset, if type was specified and correct
		// 	  it holds an asset array with that single asset as entry.
		//	· For each input asset, if type was not specified, it holds
		// 	  the list of "auto detected" assets for it.
		var assetGroup []api.Asset

		// If user specified the asset type.
		if asset.AssetType != nil && asset.AssetType.Name != "" {
			// If the asset type is invalid, abort the asset creation.
			if !validAssetType(asset.AssetType.Name) {
				response.Status = errors.Validation("Asset type not found", "asset", asset.Identifier, asset.AssetType.Name)
				responses = append(responses, response)
				continue
			}

			// Retrieve asset type from its name.
			assetTypeObj, err := s.GetAssetType(ctx, asset.AssetType.Name)
			if err != nil {
				// If there is an error retrieiving the asset type information, abort the asset creation.
				response.Status = err
				responses = append(responses, response)
				continue
			}
			asset.AssetTypeID = assetTypeObj.ID
			asset.AssetType = &api.AssetType{Name: assetTypeObj.Name}

			// If the asset is invalid, abort the asset creation.
			if err := asset.Validate(); err != nil {
				response.Status = err
				responses = append(responses, response)
				continue
			}

			assetGroup = []api.Asset{asset}

		} else {
			// If user did not specify the asset type, auto detect it.
			assetsDetected, err := s.detectAssets(ctx, asset)
			if err != nil {
				response.Status = errors.Validation(err, "asset", asset.Identifier, asset.AssetType.Name)
				responses = append(responses, response)
				continue
			}
			assetGroup = assetsDetected
		}

		// For all AWSAccount assets that do not specify an Alias, try to
		// automatically fetch one
		for i, a := range assetGroup {
			if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
				a.Alias = s.getAccountName(a.Identifier)
				assetGroup[i] = a
			}
		}

		// Request the asset creation to the store layer.
		assetsCreated, err := s.db.CreateAssets(assetGroup, groups)
		if err != nil {
			// If assets group atomic creation failed, return error
			// for the asset specified by user, not for the ones that
			// might be auto detected from it if no type was specified.
			response.Status = err
			responses = append(responses, response)
			continue
		}

		// If the creation was successful, add the asset representation to the array of responses.
		for _, ac := range assetsCreated {
			responses = append(responses, api.AssetCreationResponse{
				ID:                ac.ID,
				Identifier:        ac.Identifier,
				AssetType:         ac.AssetType.ToResponse(),
				Alias:             ac.Alias,
				Options:           asset.Options,
				EnvironmentalCVSS: asset.EnvironmentalCVSS,
				Scannable:         asset.Scannable,
				ROLFP:             ac.ROLFP,
				ClassifiedAt:      ac.ClassifiedAt,
				Status: api.Status{
					Code: http.StatusCreated,
				},
			})
		}
	}

	return responses, nil
}

func (s vulcanitoService) detectAssets(ctx context.Context, asset api.Asset) ([]api.Asset, error) {
	assets, err := getTypesFromIdentifier(asset.Identifier)
	if err != nil {
		return nil, err
	}

	if len(assets) == 0 {
		return nil, errors.Validation("cannot parse asset type from identifier")
	}

	apiAssets := []api.Asset{}
	for _, a := range assets {
		asset := asset

		if !validAssetType(a.assetType) {
			return nil, errors.Default("invalid asset type returned by auto-detection routine")
		}

		// Retrieve asset type from its name.
		assetTypeObj, err := s.GetAssetType(ctx, a.assetType)
		if err != nil {
			return nil, err
		}

		// Sometimes the identifier needs to be overwritten.  For example in
		// cases like 127.0.0.1/32 it will be overwritten by 127.0.0.1, because
		// it must be added as an IP address.
		asset.Identifier = a.identifier
		asset.AssetTypeID = assetTypeObj.ID
		asset.AssetType = &api.AssetType{Name: assetTypeObj.Name}

		// In case the asset is an AWS Account an the Alias is not filled
		// try to get the name of the account and set the alias to it.
		if asset.AssetType.Name == "AWSAccount" && asset.Alias == "" {
			asset.Alias = s.getAccountName(asset.Identifier)
		}

		// Validate asset model before appending it to the results
		if err = asset.Validate(); err != nil {
			return nil, err
		}

		apiAssets = append(apiAssets, asset)
	}

	return apiAssets, nil
}

func (s vulcanitoService) GetAssetType(ctx context.Context, assetTypeName string) (*api.AssetType, error) {
	return s.db.GetAssetType(assetTypeName)
}

func (s vulcanitoService) FindAsset(ctx context.Context, asset api.Asset) (*api.Asset, error) {
	return s.db.FindAsset(asset.TeamID, asset.ID)
}

func (s vulcanitoService) UpdateAsset(ctx context.Context, asset api.Asset) (*api.Asset, error) {
	if !common.IsStringEmpty(asset.Options) && !common.IsValidJSON(asset.Options) {
		return nil, errors.Validation("asset.options is not valid json")
	}

	// If ROLFP is set, then refresh
	// classified_at timestamp.
	if asset.ROLFP != nil {
		now := time.Now()
		asset.ClassifiedAt = &now
	}

	updated, err := s.db.UpdateAsset(asset)
	if err != nil {
		return nil, err
	}
	return s.db.FindAsset(asset.TeamID, updated.ID)
}

func (s vulcanitoService) DeleteAsset(ctx context.Context, asset api.Asset) error {
	return s.db.DeleteAsset(asset)
}

func (s vulcanitoService) DeleteAllAssets(ctx context.Context, teamID string) error {
	return s.db.DeleteAllAssets(teamID)
}

func (s vulcanitoService) CreateGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	if err := group.Validate(); err != nil {
		return nil, errors.Validation(err)
	}
	return s.db.CreateGroup(group)
}

func (s vulcanitoService) ListGroups(ctx context.Context, teamID, groupName string) ([]*api.Group, error) {
	return s.db.ListGroups(teamID, groupName)
}

func (s vulcanitoService) UpdateGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	if !common.IsStringEmpty(&group.Options) && !common.IsValidJSON(&group.Options) {
		return nil, errors.Validation("group.options needs to be valid json")
	}
	foundGroup, err := s.db.FindGroup(api.Group{ID: group.ID})
	if err != nil {
		return nil, err
	}
	if foundGroup.Name == "Default" {
		return nil, errors.Update("Cannot update Default group")
	}
	updated, err := s.db.UpdateGroup(group)
	if err != nil {
		return nil, err
	}
	return s.db.FindGroup(*updated)
}

func (s vulcanitoService) DeleteGroup(ctx context.Context, group api.Group) error {
	foundGroup, err := s.db.FindGroup(group)
	if err != nil {
		return err
	}

	if foundGroup.Name == "Default" {
		return errors.Delete("Cannot delete default group")
	}

	return s.db.DeleteGroup(group)
}

func (s vulcanitoService) FindGroup(ctx context.Context, group api.Group) (*api.Group, error) {
	return s.db.FindGroup(group)
}

func (s vulcanitoService) GroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) (*api.AssetGroup, error) {
	return s.db.GroupAsset(assetGroup, teamID)
}

func (s vulcanitoService) ListAssetGroup(ctx context.Context, assetGroup api.AssetGroup, teamID string) ([]*api.Asset, error) {
	res, err := s.db.ListAssetGroup(assetGroup, teamID)
	if err != nil {
		return nil, err
	}

	assets := []*api.Asset{}
	for _, item := range res {
		assets = append(assets, item.Asset)
	}
	return assets, nil
}

func (s vulcanitoService) UngroupAsset(ctx context.Context, assetGroup api.AssetGroup, teamID string) error {
	return s.db.UngroupAssets(assetGroup, teamID)
}

func getTypesFromIdentifier(identifier string) ([]asset, error) {
	a := asset{
		identifier: identifier,
	}

	if types.IsAWSARN(identifier) {
		a.assetType = "AWSAccount"
		return []asset{a}, nil
	}

	if types.IsDockerImage(identifier) {
		a.assetType = "DockerImage"
		return []asset{a}, nil
	}

	if types.IsGitRepository(identifier) {
		a.assetType = "GitRepository"
		return []asset{a}, nil
	}

	if types.IsIP(identifier) {
		a.assetType = "IP"
		return []asset{a}, nil
	}

	if types.IsCIDR(identifier) {
		a.assetType = "IPRange"

		// In case the CIDR has a /32 mask, remove the mask
		// and add the asset as an IP.
		if types.IsHost(identifier) {
			a.identifier = strings.TrimSuffix(identifier, "/32")
			a.assetType = "IP"
		}

		return []asset{a}, nil
	}

	var assets []asset

	isWeb := false
	if types.IsURL(identifier) {
		isWeb = true

		// From a URL like https://adevinta.com not only a WebAddress
		// type can be extracted, also a hostname (adevinta.com) and
		// potentially a domain name.
		u, err := url.ParseRequestURI(identifier)
		if err != nil {
			return nil, err
		}
		identifier = u.Hostname() // Overwrite identifier to check for hostname and domain.
	}

	if types.IsHostname(identifier) {
		h := asset{
			identifier: identifier,
			assetType:  "Hostname",
		}
		assets = append(assets, h)

		// Add WebAddress type only for URLs with valid hostnames.
		if isWeb {
			// At this point a.identifier contains the original identifier,
			// not the overwritten identifier.
			a.assetType = "WebAddress"
			assets = append(assets, a)
		}
	}

	ok, err := types.IsDomainName(identifier)
	if err != nil {
		return nil, fmt.Errorf("can not guess if the asset is a domain: %v", err)
	}
	if ok {
		d := asset{
			identifier: identifier,
			assetType:  "DomainName",
		}
		assets = append(assets, d)
	}

	return assets, nil
}

func validAssetType(assetType string) bool {
	valid := []string{"AWSAccount", "IP", "IPRange", "DomainName", "Hostname", "DockerImage", "WebAddress", "GitRepository"}
	for _, a := range valid {
		if strings.EqualFold(a, assetType) {
			return true
		}
	}
	return false
}
