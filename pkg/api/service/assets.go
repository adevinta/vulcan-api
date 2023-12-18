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
	"github.com/go-kit/kit/log/level"

	metrics "github.com/adevinta/vulcan-metrics-client"
	types "github.com/adevinta/vulcan-types"

	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/adevinta/vulcan-api/pkg/common"
)

// GenericAnnotationsPrefix defines a prefix to be added to each asset
// annotation provided to the discovery endpoint.
const GenericAnnotationsPrefix = "autodiscovery"

type asset struct {
	identifier string
	assetType  string
}

func (s vulcanitoService) ListAssets(ctx context.Context, teamID string, asset api.Asset) ([]*api.Asset, error) {
	if teamID == "" {
		return nil, errors.NotFound(`Team ID is empty`)
	}

	result, err := s.db.ListAssets(teamID, asset)
	if err != nil {
		_ = s.logger.Log("database error", err.Error())
	}
	return result, err
}

// CreateAssets receives an array of assets and creates them on store layer.
func (s vulcanitoService) CreateAssets(ctx context.Context, assets []api.Asset, groups []api.Group, annotations []*api.AssetAnnotation, dnsHostnameValidation bool) ([]api.Asset, error) {
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
			if !api.ValidAssetType(asset.AssetType.Name) {
				return nil, errors.Validation("Asset type not found", "asset", asset.Identifier, asset.AssetType.Name)
			}

			// Retrieve asset type from its name.
			assetTypeObj, err := s.GetAssetType(ctx, asset.AssetType.Name)
			if err != nil {
				return nil, err
			}

			asset.AssetTypeID = assetTypeObj.ID
			asset.AssetType = &api.AssetType{Name: assetTypeObj.Name}

			if err := asset.Validate(dnsHostnameValidation); err != nil {
				return nil, err
			}
			assetsToCreate = append(assetsToCreate, asset)
		} else {
			// Asset type NOT provided by the user in the request. Try to infere
			// asset type based on the identifier
			assetsDetected, err := s.detectAssets(ctx, asset, dnsHostnameValidation)
			if err != nil {
				return nil, errors.Validation(err, "asset", asset.Identifier, asset.AssetType.Name)
			}

			assetsToCreate = append(assetsToCreate, assetsDetected...)
		}
	}

	// Add Annotations and AWS Account alias (if needed).
	for i, a := range assetsToCreate {
		a.AssetAnnotations = annotations
		if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
			a.Alias = s.getAccountName(a.Identifier)
		}
		assetsToCreate[i] = a
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
func (s vulcanitoService) CreateAssetsMultiStatus(ctx context.Context, assets []api.Asset, groups []api.Group, annotations []*api.AssetAnnotation, dnsHostnameValidation bool) ([]api.AssetCreationResponse, error) {
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
			if !api.ValidAssetType(asset.AssetType.Name) {
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
			if err := asset.Validate(dnsHostnameValidation); err != nil {
				response.Status = err
				responses = append(responses, response)
				continue
			}

			assetGroup = []api.Asset{asset}

		} else {
			// If user did not specify the asset type, auto detect it.
			assetsDetected, err := s.detectAssets(ctx, asset, dnsHostnameValidation)
			if err != nil {
				response.Status = errors.Validation(err, "asset", asset.Identifier, asset.AssetType.Name)
				responses = append(responses, response)
				continue
			}
			assetGroup = assetsDetected
		}

		// For all AWSAccount assets that do not specify an Alias, try to
		// automatically fetch one.
		for i, a := range assetGroup {
			if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
				a.Alias = s.getAccountName(a.Identifier)
				assetGroup[i] = a
			}
		}

		// Request the asset creation to the store layer.
		// Each asset is created independently, even if they have been detected.
		// In case of failure the error is recorded as part of the response.
		for _, a := range assetGroup {
			a.AssetAnnotations = annotations
			assetCreated, err := s.db.CreateAsset(a, groups)
			if err != nil {
				response.Identifier = a.Identifier
				response.AssetType = a.AssetType.ToResponse()
				response.Status = err
				responses = append(responses, response)
				continue
			}

			// If the creation was successful, add the asset representation to the array of responses.
			responses = append(responses, api.AssetCreationResponse{
				ID:                assetCreated.ID,
				Identifier:        assetCreated.Identifier,
				AssetType:         assetCreated.AssetType.ToResponse(),
				Alias:             assetCreated.Alias,
				Options:           asset.Options,
				EnvironmentalCVSS: asset.EnvironmentalCVSS,
				Scannable:         asset.Scannable,
				ROLFP:             assetCreated.ROLFP,
				ClassifiedAt:      assetCreated.ClassifiedAt,
				Status: api.Status{
					Code: http.StatusCreated,
				},
			})
		}
	}

	return responses, nil
}

// MergeDiscoveredAssets receives an list of assets to merge with the existing
// assets of an auto-discovery group for a team.
func (s vulcanitoService) MergeDiscoveredAssets(ctx context.Context, teamID string, assets []api.Asset, groupName string) error {
	// Check if the group exists and otherwise create it. Also check that there
	// is no more than one match for the given group name.
	var group api.Group

	groups, err := s.db.ListGroups(teamID, groupName)
	switch {
	case err != nil:
		errMssg := fmt.Sprintf("unable to find group %s: %v", groupName, err)
		return errors.NotFound(errMssg)
	// No more than one group should be returned. This check is required
	// because the store layer is implemented using a LIKE filter.
	case len(groups) > 1:
		errMsg := fmt.Sprintf("more than one group matches the name %s", groupName)
		return errors.Validation(errMsg)
	// The group doesn't exist, create it.
	case len(groups) == 0:
		g := api.Group{
			TeamID: teamID,
			Name:   groupName,
		}
		res, err := s.CreateGroup(ctx, g)
		if err != nil {
			return errors.Database(err)
		}
		group = *res
	// There is exactly one matching group.
	default:
		// It shouldn't happen but checking to avoid possible nil pointer
		// dereferences.
		if groups[0] == nil {
			errMsg := fmt.Sprintf("unexpected nil pointer returned for the group %s", groupName)
			return errors.Database(errMsg)
		}
		group = *groups[0]
	}

	ops, err := s.calculateMergeOperations(ctx, teamID, assets, group)
	if err != nil {
		return err
	}

	mergedAssets := s.db.MergeAssets(ops)

	s.pushDiscoveryMetrics(assets, ops)
	return mergedAssets
}

// pushDiscoveryMetrics pushes metrics related to the discovery process.
func (s vulcanitoService) pushDiscoveryMetrics(assets []api.Asset, mergeOps api.AssetMergeOperations) {

	componentTag := "component:api"

	if len(mergeOps.Create) > 0 {
		createdMetric := metrics.Metric{
			Name:  "vulcan.discovery.created.count",
			Typ:   metrics.Count,
			Value: float64(len(mergeOps.Create)),
			Tags:  []string{componentTag},
		}
		s.metricsClient.Push(createdMetric)
	}

	skippedAssets := len(assets) -
		len(mergeOps.Create) -
		len(mergeOps.Assoc) -
		len(mergeOps.Update) -
		len(mergeOps.Del) -
		len(mergeOps.Deassoc)

	if skippedAssets > 0 {
		skippedMetric := metrics.Metric{
			Name:  "vulcan.discovery.skipped.count",
			Typ:   metrics.Count,
			Value: float64(skippedAssets),
			Tags:  []string{componentTag},
		}
		s.metricsClient.Push(skippedMetric)
	}

	if len(mergeOps.Update) > 0 {
		updatedMetric := metrics.Metric{
			Name:  "vulcan.discovery.updated.count",
			Typ:   metrics.Count,
			Value: float64(len(mergeOps.Update)),
			Tags:  []string{componentTag},
		}
		s.metricsClient.Push(updatedMetric)
	}

	if len(mergeOps.Del) > 0 {
		purgedMetric := metrics.Metric{
			Name:  "vulcan.discovery.purged.count",
			Typ:   metrics.Count,
			Value: float64(len(mergeOps.Del)),
			Tags:  []string{componentTag},
		}
		s.metricsClient.Push(purgedMetric)
	}

	if len(mergeOps.Deassoc) > 0 {
		dissociatedMetric := metrics.Metric{
			Name:  "vulcan.discovery.dissociated.count",
			Typ:   metrics.Count,
			Value: float64(len(mergeOps.Deassoc)),
			Tags:  []string{componentTag},
		}
		s.metricsClient.Push(dissociatedMetric)
	}

}

func (s vulcanitoService) calculateMergeOperations(ctx context.Context, teamID string, assets []api.Asset, group api.Group) (api.AssetMergeOperations, error) {
	ops := api.AssetMergeOperations{
		TeamID: teamID,
		Group:  group,
	}

	// NOTE: ListAssets is used as it's cheaper than execute several FindAsset
	// operations.
	allAssets, err := s.ListAssets(ctx, teamID, api.Asset{})
	if err != nil {
		return ops, err
	}
	allAssetsMap := make(map[string]*api.Asset)
	for _, a := range allAssets {
		if a.AssetType == nil {
			return ops, fmt.Errorf("missing values for team/asset (%v/%v)", teamID, a.ID)
		}
		key := fmt.Sprintf("%v-%v", a.Identifier, a.AssetType.Name)
		allAssetsMap[key] = a
	}

	// Create an index (identifier, type) for the old assets belonging to the
	// discovery group. The assets stored in the map are gathered from the
	// allAssetsMap because they include the information about the groups they
	// belong to.
	oldAssetsMap := make(map[string]*api.Asset)
	for _, ag := range group.AssetGroup {
		if ag.Asset == nil || ag.Asset.AssetType == nil {
			return ops, fmt.Errorf("missing values for team/asset/group (%v/%v/%v)", teamID, ag.AssetID, ag.GroupID)
		}
		key := fmt.Sprintf("%v-%v", ag.Asset.Identifier, ag.Asset.AssetType.Name)
		oldAssetsMap[key] = allAssetsMap[key]
	}

	// Prepend a prefix to the annotations so they can be merged without
	// messing with other annotations that assets might have.
	prefix := fmt.Sprintf("%s/%s", GenericAnnotationsPrefix, strings.TrimSuffix(group.Name, api.DiscoveredAssetsGroupSuffix))

	// Calculate assets to create, associate or update.
	dedupIdx := make(map[string]struct{})
	for _, a := range assets {
		key := fmt.Sprintf("%v-%v", a.Identifier, a.AssetType.Name)

		// If asset is duplicated (same identifier and type)
		// ignore it. Otherwise return error.
		if _, ok := dedupIdx[key]; ok {
			_ = level.Warn(s.logger).Log("Warning", "DuplicatedDiscoveryAsset",
				"identifier", a.Identifier, "type", a.AssetType.Name)
			continue
		}
		dedupIdx[key] = struct{}{}

		for _, aa := range a.AssetAnnotations {
			aa.Key = fmt.Sprintf("%s/%s", prefix, aa.Key)
		}

		old, okAll := allAssetsMap[key]

		// Asset is new. Create the asset and its annotations.
		if !okAll {
			// Retrieve asset type from its name.
			assetType, err := s.GetAssetType(ctx, a.AssetType.Name)
			if err != nil {
				return ops, fmt.Errorf("can not retrieve the asset type (%v)", a.AssetType.Name)
			}
			a.AssetTypeID = assetType.ID
			a.AssetType = &api.AssetType{Name: assetType.Name}

			// For all AWSAccount assets that do not specify an Alias, try to
			// automatically fetch one.
			if a.AssetType.Name == "AWSAccount" && a.Alias == "" {
				a.Alias = s.getAccountName(a.Identifier)
			}

			ops.Create = append(ops.Create, a)
			continue
		}

		_, okOld := oldAssetsMap[key]

		// Asset is not new but it wasn't associated to the group.
		if !okOld {
			ops.Assoc = append(ops.Assoc, *old)
		} else {
			// Asset was already existing in the group, so remove it from the old assets list to
			// later on identify stale assets (assets that were discovered in the
			// previous round but not in this one).
			delete(oldAssetsMap, key)
		}

		updatedAsset := api.Asset{
			ID:     old.ID,
			TeamID: old.TeamID,
		}
		// Check if the annotations or the scannable property needs to be
		// updated.
		updated := false

		// Only update the scannable field if it was scannable before and
		// the discovery tool marks the asset as non-scannable. Otherwise
		// an asset manually marked as non-scannable by the user might be
		// unintentionally scanned.
		if old.Scannable != nil && *old.Scannable && a.Scannable != nil && !*a.Scannable {
			updatedAsset.Scannable = a.Scannable
			updated = true
		}

		// Only update the annotations if they are different.
		oldAnnotations := api.AssetAnnotations(old.AssetAnnotations).ToMap()
		newAnnotations := api.AssetAnnotations(a.AssetAnnotations).ToMap()
		if !oldAnnotations.Matches(newAnnotations, prefix) {
			updatedAsset.AssetAnnotations = oldAnnotations.Merge(newAnnotations, prefix).ToModel()
			updated = true
		}

		if updated {
			ops.Update = append(ops.Update, updatedAsset)
		}
	}

	// Calculate assets to remove or deassociate. The oldAssetsMap should
	// contain only assets that were previously discovered but not in this
	// round.
	for _, old := range oldAssetsMap {
		del := true
		for _, g := range old.AssetGroups {
			// Only delete the asset if doesn't belong to more groups than
			// the auto-discovery group. Also remove annotations previously
			// added by the discovery service.
			if g.Group != nil && g.Group.Name != group.Name {
				aux := api.AssetAnnotations(old.AssetAnnotations).ToMap()
				old.AssetAnnotations = aux.Merge(api.AssetAnnotationsMap{}, prefix).ToModel()
				ops.Deassoc = append(ops.Deassoc, *old)
				del = false
				break
			}
		}

		if del {
			ops.Del = append(ops.Del, *old)
		}
	}

	return ops, nil
}

// MergeDiscoveredAssetsAsync stores the information necessary to perform the
// MergeDiscoveredAssets operation asynchronously.
func (s vulcanitoService) MergeDiscoveredAssetsAsync(ctx context.Context, teamID string, assets []api.Asset, groupName string) (*api.Job, error) {
	return s.db.MergeAssetsAsync(teamID, assets, groupName)
}

func (s vulcanitoService) detectAssets(ctx context.Context, asset api.Asset, dnsHostnameValidation bool) ([]api.Asset, error) {
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

		if !api.ValidAssetType(a.assetType) {
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
		if err = asset.Validate(dnsHostnameValidation); err != nil {
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
