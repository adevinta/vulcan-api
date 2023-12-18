/*
Copyright 2021 Adevinta
*/

package endpoint

import (
	"context"
	"fmt"
	"strings"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/common"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
)

type AssetRequest struct {
	ID                string     `json:"id" urlvar:"asset_id"`
	TeamID            string     `json:"team_id" urlvar:"team_id"`
	Type              string     `json:"type" validate:"required"`
	Identifier        string     `json:"identifier" validate:"required" urlquery:"identifier"`
	Options           *string    `json:"options,omitempty"`
	EnvironmentalCVSS *string    `json:"environmental_cvss,omitempty"`
	ROLFP             *api.ROLFP `json:"rolfp"`
	Scannable         *bool      `json:"scannable,omitempty"`
}

func (ar AssetRequest) NewAsset() *api.Asset {
	asset := api.Asset{
		TeamID:            ar.TeamID,
		Identifier:        ar.Identifier,
		AssetType:         &api.AssetType{Name: ar.Type},
		Options:           ar.Options,
		EnvironmentalCVSS: ar.EnvironmentalCVSS,
		ROLFP:             ar.ROLFP,
		Scannable:         ar.Scannable,
	}
	if ar.Scannable == nil {
		asset.Scannable = common.Bool(true)
	}
	return &asset
}

type AssetsListRequest struct {
	TeamID      string                  `json:"team_id" urlvar:"team_id"`
	Assets      []AssetRequest          `json:"assets"`
	Groups      []string                `json:"groups"`
	Annotations api.AssetAnnotationsMap `json:"annotations"`
}

type AssetWithAnnotationsRequest struct {
	AssetRequest

	Annotations api.AssetAnnotationsMap `json:"annotations"`
}

type DiscoveredAssetsRequest struct {
	TeamID    string                        `json:"team_id" urlvar:"team_id"`
	Assets    []AssetWithAnnotationsRequest `json:"assets"`
	GroupName string                        `json:"group_name"`
}

func makeListAssetsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*AssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		filterAsset := api.Asset{Identifier: req.Identifier}
		teamAssets, err := s.ListAssets(ctx, req.TeamID, filterAsset)
		if err != nil {
			return nil, err
		}

		response := []api.AssetResponse{}
		for _, asset := range teamAssets {
			resp := asset.ToResponse()
			response = append(response, resp)
		}

		return Ok{response}, nil
	}
}

// makeCreateAssetEndpoint returns an endpoint that creates new assets.
func makeCreateAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger, dnsHostnameValidation bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// We are expecting an assets list.
		requestBody, ok := request.(*AssetsListRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Iterate over the assets list, initializes each item and set the
		// team ID to be the same from request body.
		assets := []api.Asset{}
		for _, ar := range requestBody.Assets {
			asset := ar.NewAsset()
			asset.TeamID = requestBody.TeamID

			assets = append(assets, *asset)
		}

		// Iterate over groups list, initialize items and set team ID
		// to be the same from request body.
		groups := []api.Group{}
		for _, gr := range requestBody.Groups {
			group := api.Group{ID: gr, TeamID: requestBody.TeamID}
			groups = append(groups, group)
		}

		// Transform annotations from request to its API model format
		annotations := requestBody.Annotations.ToModel()

		// Ask for the service layer to create the assets.
		createdAssets, err := s.CreateAssets(ctx, assets, groups, annotations, dnsHostnameValidation)
		if err != nil {
			return nil, err
		}

		// Iterate over the list of created assets and assemble a response
		// array.
		response := []api.AssetResponse{}
		for _, createdAsset := range createdAssets {
			response = append(response, createdAsset.ToResponse())
		}

		return Created{response}, nil
	}
}

// makeCreateAssetMultiStatusEndpoint returns an endpoint that creates new assets.
func makeCreateAssetMultiStatusEndpoint(s api.VulcanitoService, logger kitlog.Logger, dnsHostnameValidation bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// We are expecting an assets list.
		requestBody, ok := request.(*AssetsListRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Iterate over the assets list, initializes each item and set the
		// team ID to be the same from request body.
		assets := []api.Asset{}
		for _, ar := range requestBody.Assets {
			asset := ar.NewAsset()
			asset.TeamID = requestBody.TeamID

			assets = append(assets, *asset)
		}

		// Iterate over groups list, initialize items and set team ID
		// to be the same from request body.
		groups := []api.Group{}
		for _, gr := range requestBody.Groups {
			group := api.Group{ID: gr, TeamID: requestBody.TeamID}
			groups = append(groups, group)
		}

		// Transform annotations from request to its API model format
		annotations := requestBody.Annotations.ToModel()

		// Ask for the service layer to create the assets.
		responses, err := s.CreateAssetsMultiStatus(ctx, assets, groups, annotations, dnsHostnameValidation)
		if err != nil {
			return nil, err
		}

		multiStatus := false
		for _, r := range responses {
			_, isError := r.Status.(error)
			if isError {
				multiStatus = true
				break
			}
		}

		if multiStatus {
			return MultiStatus{responses}, nil
		}

		return Created{responses}, nil
	}
}

// makeMergeDiscoveredAssetsEndpoint merges a list of assets into a discovery
// asset group, requested by a discovery service.
func makeMergeDiscoveredAssetsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*DiscoveredAssetsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		// Validate that the asset group name is one of the allowed ones.
		if !strings.HasSuffix(requestBody.GroupName, api.DiscoveredAssetsGroupSuffix) {
			return nil, errors.Validation("Asset group not allowed")
		}

		// Validate the assets list, initialize each item and set the
		// team ID to be the same from the request body.
		assets := []api.Asset{}
		for _, ar := range requestBody.Assets {
			if ar.Identifier == "" || ar.Type == "" {
				return nil, errors.Validation("Asset identifier and type are required for all the assets")
			}
			if !api.ValidAssetType(ar.Type) {
				return nil, errors.Validation(fmt.Errorf("Invalid asset type (%s) for asset (%v)", ar.Type, ar.Identifier))
			}

			asset := ar.NewAsset()
			asset.TeamID = requestBody.TeamID
			asset.AssetAnnotations = ar.Annotations.ToModel()

			assets = append(assets, *asset)
		}

		// Ask for the service layer to asynchronously merge the discovered assets.
		job, err := s.MergeDiscoveredAssetsAsync(ctx, requestBody.TeamID, assets, requestBody.GroupName)
		if err != nil {
			return nil, err
		}

		return Accepted{job.ToResponse()}, nil
	}
}

func makeFindAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		asset := api.Asset{
			ID:     requestBody.ID,
			TeamID: requestBody.TeamID,
		}
		assetr, err := s.FindAsset(ctx, asset)
		if err != nil {
			return nil, err
		}
		return Ok{assetr.ToResponse()}, nil
	}
}

func makeUpdateAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		asset := api.Asset{
			ID:                requestBody.ID,
			Identifier:        requestBody.Identifier,
			Options:           requestBody.Options,
			EnvironmentalCVSS: requestBody.EnvironmentalCVSS,
			Scannable:         requestBody.Scannable,
			TeamID:            requestBody.TeamID,
			ROLFP:             requestBody.ROLFP,
		}
		assetr, err := s.UpdateAsset(ctx, asset)
		if err != nil {
			return nil, err
		}
		return Ok{assetr.ToResponse()}, nil
	}
}

func makeDeleteAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		asset := api.Asset{
			ID:     requestBody.ID,
			TeamID: requestBody.TeamID,
		}
		err := s.DeleteAsset(ctx, asset)
		if err != nil {
			return nil, err
		}
		return NoContent{nil}, nil
	}
}

type ListGroupsRequest struct {
	TeamID string `json:"team_id" urlvar:"team_id"`
	Name   string `json:"name" urlquery:"name"`
}

func makeListGroupsEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*ListGroupsRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		groups, err := s.ListGroups(ctx, requestBody.TeamID, requestBody.Name)
		if err != nil {
			return nil, err
		}

		response := []api.GroupResponse{}
		for _, group := range groups {
			response = append(response, *group.ToResponse())
		}
		return Ok{response}, nil
	}
}

type AssetsGroupRequest struct {
	ID      string `json:"id" urlvar:"group_id"`
	TeamID  string `json:"team_id" urlvar:"team_id"`
	Name    string `json:"name"`
	Options string `json:"options"`
}

func makeCreateGroupEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetsGroupRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		assetGroup := api.Group{TeamID: requestBody.TeamID, Name: requestBody.Name, Options: requestBody.Options}
		assetsGroup, err := s.CreateGroup(ctx, assetGroup)
		if err != nil {
			return nil, err
		}
		return Created{assetsGroup.ToResponse()}, nil
	}
}

func makeUpdateGroupEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetsGroupRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		group := api.Group{
			ID:      requestBody.ID,
			TeamID:  requestBody.TeamID,
			Name:    requestBody.Name,
			Options: requestBody.Options,
		}
		updated, err := s.UpdateGroup(ctx, group)
		if err != nil {
			return nil, err
		}
		return Ok{updated.ToResponse()}, nil
	}

}

func makeDeleteGroupEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetsGroupRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		group := api.Group{ID: requestBody.ID, TeamID: requestBody.TeamID}
		err := s.DeleteGroup(ctx, group)
		if err != nil {
			return nil, errors.Delete(err)
		}
		return NoContent{nil}, nil
	}
}

func makeFindGroupEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*AssetsGroupRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		group := &api.Group{ID: requestBody.ID, TeamID: requestBody.TeamID}
		group, err := s.FindGroup(ctx, *group)
		if err != nil {
			return nil, err
		}
		return Ok{group.ToResponse()}, nil
	}
}

type GroupAssetRequest struct {
	GroupID string `json:"group_id" urlvar:"group_id"`
	AssetID string `json:"asset_id" urlvar:"asset_id"`
	TeamID  string `urlvar:"team_id"`
}

func makeGroupAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*GroupAssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		assetGroup := api.AssetGroup{
			AssetID: requestBody.AssetID,
			GroupID: requestBody.GroupID,
		}

		_, err := s.GroupAsset(ctx, assetGroup, requestBody.TeamID)
		if err != nil {
			return nil, err
		}

		asset, err := s.FindAsset(ctx, api.Asset{ID: requestBody.AssetID, TeamID: requestBody.TeamID})
		if err != nil {
			return nil, err
		}

		return Ok{asset.ToResponse()}, nil
	}
}

func makeUngroupAssetEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*GroupAssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}
		assetGroup := api.AssetGroup{
			AssetID: requestBody.AssetID,
			GroupID: requestBody.GroupID,
		}
		err := s.UngroupAsset(ctx, assetGroup, requestBody.TeamID)
		if err != nil {
			return nil, err
		}
		return NoContent{nil}, nil
	}
}

func makeListAssetGroupEndpoint(s api.VulcanitoService, logger kitlog.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		requestBody, ok := request.(*GroupAssetRequest)
		if !ok {
			return nil, errors.Assertion("Type assertion failed")
		}

		assetGroup := api.AssetGroup{GroupID: requestBody.GroupID}
		assets, err := s.ListAssetGroup(ctx, assetGroup, requestBody.TeamID)
		if err != nil {
			return nil, err
		}
		response := []api.AssetResponse{}
		for _, asset := range assets {
			if asset == nil {
				return nil, errors.Database(fmt.Sprintf("nil asset for group %v and team %v in the assets list: %+v", requestBody.GroupID, requestBody.TeamID, assets))
			}

			response = append(response, asset.ToResponse())
		}
		return Ok{response}, nil
	}
}
