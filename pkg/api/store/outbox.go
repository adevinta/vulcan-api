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
	opDeleteTeam      = "DeleteTeam"
	opDeleteAsset     = "DeleteAsset"
	opDeleteAllAssets = "DeleteAllAssets"
	opFindingOverride = "FindingOverride"
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
	case opDeleteAsset:
		buildFunc = db.buildDeleteAssetDTO
	case opDeleteAllAssets:
		buildFunc = db.buildDeleteAllAssetsDTO
	case opFindingOverride:
		buildFunc = db.buildFindingOverrideDTO
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

// buildFindingOverrideDTO builds a FindingOverride action DTO for outbox.
// Expected input:
//	- api.FindingOverride
func (db vulcanitoStore) buildFindingOverrideDTO(tx *gorm.DB, data ...interface{}) (interface{}, error) {
	if len(data) != 1 {
		return nil, errInvalidParams
	}
	findingOverride, ok := data[0].(api.FindingOverride)
	if !ok {
		return nil, errInvalidParams
	}

	return cdc.OpFindingOverrideDTO{FindingOverride: findingOverride}, nil
}

func (db vulcanitoStore) insertIntoOutbox(tx *gorm.DB, outbox cdc.Outbox) error {
	res := tx.Create(&outbox)
	if res.Error != nil {
		return db.logError(errors.Create(res.Error))
	}
	return nil
}
