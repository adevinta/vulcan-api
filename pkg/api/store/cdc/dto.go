/*
Copyright 2021 Adevinta
*/

package cdc

import "github.com/adevinta/vulcan-api/pkg/api"

// OpDeleteTeamDTO represents the data to store
// as part of CDC log for a DeleteTeam operation.
type OpDeleteTeamDTO struct {
	Team api.Team `json:"team"`
}

// OpDeleteAssetDTO represents the data to store
// as part of CDC log for a DeleteAsset operation.
type OpDeleteAssetDTO struct {
	Asset api.Asset `json:"asset"`
	// DupAssets is the number of assets
	// which have the same identifier in
	// the same team as Asset
	DupAssets int `json:"duplicates"`
}

// OpDeleteAllAssetsDTO represents the data to store
// as part of CDC log for a DeleteAllAssets operation.
type OpDeleteAllAssetsDTO struct {
	Team api.Team `json:"team"`
}

// OpFindingOverwriteDTO represents the data to store
// as part of CDC log for a FindingOverwrite operation.
type OpFindingOverwriteDTO struct {
	FindingOverwrite api.FindingOverwrite `json:"finding_overwrite"`
}
