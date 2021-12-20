/*
Copyright 2021 Adevinta
*/

package store

import (
	"encoding/json"
	errs "errors"

	errors "github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/api/store/cdc"
	"github.com/jinzhu/gorm"
)

const (
	// operations
	opDeleteTeam            = "DeleteTeam"
	opCreateAsset           = "CreateAsset"
	opDeleteAsset           = "DeleteAsset"
	opUpdateAsset           = "UpdateAsset"
	opDeleteAllAssets       = "DeleteAllAssets"
	opFindingOverwrite      = "FindingOverwrite"
	opMergeDiscoveredAssets = "MergeDiscoveredAssets"
)

var (
	errInvalidParams   = errs.New("invalid parameters")
	errUnimplementedOp = errs.New("operation not implemented")
)

func (db vulcanitoStore) pushToOutbox(tx *gorm.DB, op string, data ...interface{}) error {
	var buildFunc func(*gorm.DB, ...interface{}) (interface{}, error)
	switch op {
	case opDeleteTeam:
		buildFunc = db.buildDeleteTeamDTO
	case opCreateAsset:
		buildFunc = db.buildCreateAssetDTO
	case opDeleteAsset:
		buildFunc = db.buildDeleteAssetDTO
	case opUpdateAsset:
		buildFunc = db.buildUpdateAssetDTO
	case opDeleteAllAssets:
		buildFunc = db.buildDeleteAllAssetsDTO
	case opFindingOverwrite:
		buildFunc = db.buildFindingOverwriteDTO
	case opMergeDiscoveredAssets:
		buildFunc = db.buildMergeDiscoveredAssetsDTO
	default:
		return errUnimplementedOp
	}

	dto, err := buildFunc(tx, data...)
	if err != nil {
		return db.logError(errors.Default(err))
	}

	dtoData, err := json.Marshal(dto)
	if err != nil {
		return db.logError(errors.Default(err))
	}

	return db.insertIntoOutbox(tx, cdc.Outbox{
		Operation: op,
		SchemaVer: cdc.OutboxVersion,
		DTO:       dtoData,
	})
}

// buildDeleteTeamDTO builds a DeleteTeam action DTO for outbox.
// Expected input:
//	- api.Team
func (db vulcanitoStore) buildDeleteTeamDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	team, ok := data[0].(api.Team)
	if !ok {
		return nil, errInvalidParams
	}

	// Don't store unnecessary data
	team.Assets = nil
	team.UserTeam = nil
	team.Groups = nil

	return cdc.OpDeleteTeamDTO{Team: team}, nil
}

// buildCreateAssetDTO builds a CreateAsset action DTO for outbox.
// Expected input:
//	- api.Asset
func (db vulcanitoStore) buildCreateAssetDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	asset, ok := data[0].(api.Asset)
	if !ok || asset.Team == nil {
		return nil, errInvalidParams
	}

	// Don't store unnecessary data
	asset.AssetGroups = nil

	return cdc.OpCreateAssetDTO{Asset: asset}, nil
}

// buildDeleteAssetDTO builds a DeleteAsset action DTO for outbox.
// Expected input:
//	- api.Asset
func (db vulcanitoStore) buildDeleteAssetDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	asset, ok := data[0].(api.Asset)
	if !ok || asset.Team == nil {
		return nil, errInvalidParams
	}

	// Because multiple assets can have the same identifier, even
	// for the same team, we have to count how many duplicates
	// are for the given asset identifier and its associated team.
	dupAssets, err := db.countTeamAssetsByIdentifier(asset.TeamID, asset.Identifier)
	if err != nil {
		return nil, err
	}
	// Do not count the one that will be deleted in this tx.
	dupAssets--

	// Don't store unnecessary data
	asset.AssetGroups = nil

	return cdc.OpDeleteAssetDTO{Asset: asset, DupAssets: dupAssets}, nil
}

// buildUpdateAssetDTO builds a UpdateAsset action DTO for outbox.
// This action should only be triggered when the asset update operation
// changes the asset's identifier.
// Expected input:
//	- api.Asset (Old Asset)
//  - api.Asset (New Asset)
func (db vulcanitoStore) buildUpdateAssetDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 2 {
		return nil, errInvalidParams
	}
	oldAsset, ok := data[0].(api.Asset)
	if !ok || oldAsset.Team == nil {
		return nil, errInvalidParams
	}
	newAsset, ok := data[1].(api.Asset)
	if !ok {
		return nil, errInvalidParams
	}

	// If team data is not filled for new
	// asset, copy it from old asset
	if newAsset.Team == nil {
		newAsset.Team = oldAsset.Team
	}

	// The data that we need to store for old asset is the same as for an
	// asset delete operation, because we have to know if the identifier
	// has duplicates or not in order to remove the association from the
	// Vulnerability DB or not.
	dto, err := db.buildDeleteAssetDTO(tx, oldAsset)
	if err != nil {
		return nil, err
	}

	delAssetDTO, ok := dto.(cdc.OpDeleteAssetDTO)
	if !ok {
		return nil, errs.New("error building intermediate DeleteAssetDTO for outbox UpdateAsset")
	}

	return cdc.OpUpdateAssetDTO{OldAsset: delAssetDTO.Asset, NewAsset: newAsset, DupAssets: delAssetDTO.DupAssets}, nil
}

// buildDeleteAllAssetsDTO builds a DeleteAllAssets action DTO for outbox.
// Expected input:
//	- teamID string
func (db vulcanitoStore) buildDeleteAllAssetsDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	teamID, ok := data[0].(string)
	if !ok {
		return nil, errInvalidParams
	}
	team, err := db.FindTeam(teamID)
	if err != nil {
		return nil, err
	}

	// Don't store unnecessary data
	team.Assets = nil
	team.UserTeam = nil
	team.Groups = nil

	return cdc.OpDeleteAllAssetsDTO{Team: *team}, nil
}

// buildFindingOverwriteDTO builds a FindingOverwrite action DTO for outbox.
// Expected input:
//	- api.FindingOverwrite
func (db vulcanitoStore) buildFindingOverwriteDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	findingOverwrite, ok := data[0].(api.FindingOverwrite)
	if !ok {
		return nil, errInvalidParams
	}

	return cdc.OpFindingOverwriteDTO{FindingOverwrite: findingOverwrite}, nil
}

func (db vulcanitoStore) insertIntoOutbox(tx *gorm.DB, outbox cdc.Outbox) error {
	res := tx.Create(&outbox)
	if res.Error != nil {
		return db.logError(errors.Create(res.Error))
	}
	return nil
}

// buildMergeDiscoveredAssetsDTO builds a MergeDiscoveredAssets action DTO for
// outbox.  Expected input:
//	- teamID
//  - []api.Asset
//  - groupName
//  - jobID
func (db vulcanitoStore) buildMergeDiscoveredAssetsDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 4 {
		return nil, errInvalidParams
	}
	teamID, ok := data[0].(string)
	if !ok {
		return nil, errInvalidParams
	}
	assets, ok := data[1].([]api.Asset)
	if !ok {
		return nil, errInvalidParams
	}
	groupName, ok := data[2].(string)
	if !ok {
		return nil, errInvalidParams
	}
	jobID, ok := data[3].(string)
	if !ok {
		return nil, errInvalidParams
	}

	return cdc.OpMergeDiscoveredAssetsDTO{TeamID: teamID, Assets: assets, GroupName: groupName, JobID: jobID}, nil
}
