/*
Copyright 2021 Adevinta
*/

package store

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) ListAssets(teamID string, asset api.Asset) ([]*api.Asset, error) {
	findTeam := &api.Team{ID: teamID}
	res := db.Conn.Find(&findTeam)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	assets := []*api.Asset{}
	result := db.Conn.
		Preload("Team").
		Preload("AssetType").
		Preload("AssetGroups.Group").
		Preload("AssetAnnotations").
		Where("team_id = ?", teamID).
		Where(&asset).
		Find(&assets)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return assets, nil
}

func (db vulcanitoStore) CreateAssets(assets []api.Asset, groups []api.Group, annotations []*api.AssetAnnotation) ([]api.Asset, error) {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	createdAssets := []api.Asset{}

	for _, a := range assets {
		// Check if asset already exists.
		asset, err := db.findAsset(tx, a.TeamID, a.Identifier, a.AssetTypeID)
		if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
			return nil, err
		}

		// If asset does not exist, create it.
		if errors.IsKind(err, errors.ErrNotFound) {
			asset, err = db.createAsset(tx, a)
			if err != nil {
				tx.Rollback()

				assetType := ""
				if a.AssetType != nil {
					assetType = a.AssetType.Name
				}
				err = errors.Create(err.Error(), "asset", a.Identifier, assetType)
				return nil, err
			}
		}

		// Associate asset with group for each input group.
		for _, g := range groups {
			assetGroup := api.AssetGroup{AssetID: asset.ID, GroupID: g.ID}
			err := db.createAssetGroup(tx, assetGroup)
			if err != nil {
				tx.Rollback()

				if db.IsDuplicateError(err) {
					err = errors.Duplicated(err.Error())
				} else {
					err = errors.Create(err.Error(), "assetGroup", asset.ID, g.ID)
				}
				return nil, err
			}
		}

		// Associate asset with input annotations
		for _, an := range annotations {
			an.AssetID = asset.ID

			result := tx.Create(&an)
			if result.Error != nil {
				tx.Rollback()
				return nil, errors.Create(result.Error, "assetAnnotation", asset.ID, an.Key)
			}
		}

		createdAssets = append(createdAssets, *asset)
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return createdAssets, nil
}

// CreateAsset persists an asset in the database.
// It receives an asset and an array of groups.
// The asset will be associated with all groups from that array.
func (db vulcanitoStore) CreateAsset(a api.Asset, groups []api.Group) (*api.Asset, error) {
	// Creates a new transaction in the database.
	tx := db.Conn.Begin()
	if tx.Error != nil {
		// We have rceived an error when trying to obtain a new transaction.
		// No need to rollback the transaction.
		return nil, db.logError(errors.Database(tx.Error))
	}

	// We try to retrieve the asset from the database using the Team ID, Identifier and Asset Type.
	// This asset will be returned at the end of the function.
	// Abort the transaction in case of errors during the search.
	asset, err := db.findAsset(tx, a.TeamID, a.Identifier, a.AssetTypeID)
	if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
		tx.Rollback()
		return nil, err
	}

	// If the asset does not exist, then we create a new asset.
	if errors.IsKind(err, errors.ErrNotFound) {
		asset, err = db.createAsset(tx, a)
		if err != nil {
			tx.Rollback()

			assetType := ""
			if a.AssetType != nil {
				assetType = a.AssetType.Name
			}
			err = errors.Create(err.Error(), "asset", a.Identifier, assetType)
			return nil, err
		}
	}

	// Associate the asset with all groups.
	for _, g := range groups {
		// Declare an object representing the association between asset and group.
		assetGroup := api.AssetGroup{AssetID: asset.ID, GroupID: g.ID}

		// Create the association in the database.
		err := db.createAssetGroup(tx, assetGroup)
		if err != nil {
			tx.Rollback()

			// Return an specific error for the case in which the association already exists.
			if db.IsDuplicateError(err) {
				err = errors.Duplicated(err.Error())
			} else {
				err = errors.Create(err.Error(), "assetGroup", asset.ID, g.ID)
			}

			return nil, err
		}
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	// Return the asset
	return asset, nil
}

func (db vulcanitoStore) createAsset(tx *gorm.DB, asset api.Asset) (*api.Asset, error) {
	asset.AssetType = nil
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	findTeam := &api.Team{ID: asset.TeamID}
	res := db.Conn.Find(&findTeam)
	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}

	if asset.ROLFP != nil {
		now := time.Now()
		asset.ClassifiedAt = &now
	}

	res = tx.Create(&asset)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}

	res = tx.Preload("Team").Preload("AssetType").Find(&asset)
	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}

	err := db.pushToOutbox(tx, opCreateAsset, asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (db vulcanitoStore) createAssetGroup(tx *gorm.DB, assetGroup api.AssetGroup) error {
	res := tx.Create(&assetGroup)
	if res.Error != nil {
		return db.logError(errors.Create(res.Error))
	}

	return nil
}

func (db vulcanitoStore) FindAsset(teamID, assetID string) (*api.Asset, error) {
	asset := &api.Asset{ID: assetID}
	res := db.Conn.
		Preload("Team").
		Preload("AssetGroups").
		Preload("AssetGroups.Asset").
		Preload("AssetGroups.Group").
		Preload("AssetGroups.Group.AssetGroup").
		Preload("AssetAnnotations").
		Preload("AssetType").Where("team_id = ?", teamID).Find(&asset)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}
	return asset, nil
}

func (db vulcanitoStore) findAsset(tx *gorm.DB, teamID, identifier, assetTypeID string) (*api.Asset, error) {
	asset := &api.Asset{}
	res := tx.Preload("Team").
		Preload("AssetGroups").
		Preload("AssetGroups.Asset").
		Preload("AssetGroups.Group").
		Preload("AssetGroups.Group.AssetGroup").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Find(&asset, "team_id = ? and identifier = ? and asset_type_id = ?", teamID, identifier, assetTypeID)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}
	return asset, nil
}

// countTeamAssetsByIdentifier returns the number of assets for the given team
// which match with the given indentifier.
func (db vulcanitoStore) countTeamAssetsByIdentifier(teamID, identifier string) (int, error) {
	var count struct {
		Count int
	}
	res := db.Conn.Raw(`
		SELECT COUNT(*) FROM assets a
		INNER JOIN teams t ON a.team_id = t.id
		WHERE t.id = ? AND a.identifier = ?`,
		teamID, identifier).Scan(&count)

	if res.Error != nil {
		return 0, db.logError(errors.Database(res.Error))
	}

	return count.Count, nil
}

func (db vulcanitoStore) UpdateAsset(asset api.Asset) (*api.Asset, error) {
	findAsset := api.Asset{ID: asset.ID}
	if db.Conn.
		Preload("Team").
		Preload("AssetAnnotations").
		Where("team_id = ? and id = ?", asset.TeamID, asset.ID).
		First(&findAsset).
		RecordNotFound() {
		return nil, db.logError(errors.Forbidden("asset does not belong to team"))
	}

	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	result := tx.Model(&asset).Where("team_id = ?", asset.TeamID).Update(asset)
	if result.Error != nil {
		tx.Rollback()
		return nil, db.logError(errors.Update(result.Error))
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, db.logError(errors.Update("Asset was not updated"))
	}

	// If asset identifier has changed, we have to propagate the action
	// to the vulnerability DB so ownership from previous identifier is
	// removed for this team if necessary, and also the new one is created.
	if asset.Identifier != "" && asset.Identifier != findAsset.Identifier {
		err := db.pushToOutbox(tx, opUpdateAsset, findAsset, asset)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return &asset, nil
}

func (db vulcanitoStore) DeleteAsset(asset api.Asset) error {
	findAsset := api.Asset{ID: asset.ID}
	if db.Conn.
		Where("team_id = ? and id = ?", asset.TeamID, asset.ID).
		Preload("Team").
		Preload("AssetAnnotations").
		First(&findAsset).RecordNotFound() {
		return db.logError(errors.Forbidden("asset does not belong to team"))
	}

	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	result := tx.Delete(&api.AssetGroup{}, "asset_id = ?", asset.ID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	result = tx.Delete(&api.Asset{}, "id = ? and team_id = ?", asset.ID, asset.TeamID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return db.logError(errors.Delete("Asset was not deleted"))
	}

	err := db.pushToOutbox(tx, opDeleteAsset, findAsset)
	if err != nil {
		tx.Rollback()
		return err
	}

	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	return nil
}

func (db vulcanitoStore) DeleteAllAssets(teamID string) error {
	// Begin a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	// Delete all asset_group associations for this team
	result := tx.Exec("DELETE from asset_group where asset_id in (select id from assets where team_id = ?)", teamID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	// Delete all assets from this team
	result = tx.Delete(&api.Asset{}, "team_id = ?", teamID)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	// Push to outbox so distributed tx is processed
	err := db.pushToOutbox(tx, opDeleteAllAssets, teamID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	return nil
}

func (db vulcanitoStore) GetAssetType(name string) (*api.AssetType, error) {
	assetType := &api.AssetType{}
	result := db.Conn.First(&assetType, "lower(name) = lower(?)", name)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}
	return assetType, nil
}

func (db vulcanitoStore) CreateGroup(group api.Group) (*api.Group, error) {
	res := db.Conn.Preload("Team").Create(&group)
	err := res.Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, db.logError(errors.Duplicated(err))
		}
		return nil, db.logError(errors.Create(err))
	}
	db.Conn.Preload("Team").First(&group)
	return &group, nil
}

func (db vulcanitoStore) UpdateGroup(group api.Group) (*api.Group, error) {
	findGroup := api.Group{ID: group.ID}
	if db.Conn.Where("team_id = ? and id = ?", group.TeamID, group.ID).First(&findGroup).RecordNotFound() {
		return nil, db.logError(errors.Forbidden("group does not belong to team"))
	}

	result := db.Conn.Model(&group).Where("team_id = ?", group.TeamID).Update(group)
	if result.RowsAffected == 0 {
		return nil, db.logError(errors.Update("Asset group was not updated"))
	}
	if result.Error != nil {
		return nil, db.logError(errors.Update(result.Error))
	}
	return &group, nil
}

func (db vulcanitoStore) DeleteGroup(group api.Group) error {
	result := db.Conn.Model(&group).Where("team_id = ?", group.TeamID).Delete(&group)
	if result.RowsAffected == 0 {
		return db.logError(errors.Delete("Asset group was not deleted"))
	}

	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}

	assetGroup := api.AssetGroup{GroupID: group.ID}
	result = db.Conn.Delete(&assetGroup, "group_id = ?", group.ID)

	if result.Error != nil {
		return db.logError(errors.Delete(result.Error))
	}

	return nil
}

func (db vulcanitoStore) FindGroup(group api.Group) (*api.Group, error) {
	foundGroup := api.Group{}
	res := db.Conn.
		Preload("Team").
		Preload("AssetGroup").
		Preload("AssetGroup.Asset").
		Preload("AssetGroup.Asset.AssetType").
		Find(&foundGroup, group)

	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return &foundGroup, nil
}

func (db vulcanitoStore) FindGroupInfo(group api.Group) (*api.Group, error) {
	foundGroup := api.Group{}
	res := db.Conn.Find(&foundGroup, group)
	if res.Error != nil {
		if db.NotFoundError(res.Error) {
			return nil, db.logError(errors.NotFound(res.Error))
		}
		return nil, db.logError(errors.Database(res.Error))
	}

	return &foundGroup, nil
}

func (db vulcanitoStore) DisjoinAssetsInGroups(teamID, inGroupID string, notInGroupIDs []string) ([]*api.Asset, error) {
	types := []*api.AssetType{}
	if err := db.Conn.Find(&types).Error; err != nil {
		return nil, err
	}
	at := map[string]*api.AssetType{}
	for _, t := range types {
		at[t.ID] = t
	}
	assets := []*api.Asset{}
	res := db.Conn.Raw(`SELECT a.* FROM assets a
			JOIN asset_group ag ON ag.asset_id=a.id JOIN asset_types t ON t.id=a.asset_type_id
			WHERE a.scannable=true AND a.team_id=? AND ag.group_id=?
			AND NOT EXISTS(SELECT 1 FROM asset_group ag2 JOIN assets a2 ON ag2.asset_id=a2.id WHERE ag2.asset_id=a.id AND a2.team_id=a.team_id AND ag2.group_id in (?))`,
		teamID, inGroupID, notInGroupIDs).Scan(&assets)
	if res.RecordNotFound() {
		return nil, db.logError(errors.ErrNotFound)
	}
	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}

	for _, a := range assets {
		t, ok := at[a.AssetTypeID]
		if !ok {
			return nil, errors.Database("error getting assettype name for the asset")
		}
		a.AssetType = t
	}
	return assets, nil
}

func (db vulcanitoStore) CountAssetsInGroups(teamID string, groupIDs []string) (int, error) {
	var count struct {
		Count int
	}
	res := db.Conn.Raw(`SELECT COUNT(DISTINCT aa.id)
			FROM asset_group AS ag
			JOIN assets AS aa ON ag.asset_id=aa.id
			WHERE ag.group_id IN (?) AND aa.team_id=?`,
		groupIDs, teamID).Scan(&count)

	if res.RecordNotFound() {
		return 0, db.logError(errors.ErrNotFound)
	}
	if res.Error != nil {
		return 0, db.logError(errors.Database(res.Error))
	}

	return count.Count, nil
}

func (db vulcanitoStore) ListGroups(teamID, groupName string) ([]*api.Group, error) {
	findTeam := &api.Team{ID: teamID}
	res := db.Conn.Find(&findTeam)
	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}

	groups := []*api.Group{}

	var result *gorm.DB
	query := db.Conn.
		Preload("Team").
		Preload("AssetGroup").
		Preload("AssetGroup.Asset").
		Preload("AssetGroup.Asset.AssetType")

	if groupName != "" {
		groupName = "%" + groupName + "%"
		result = query.
			Where("name LIKE ?", groupName).
			Find(&groups, "team_id = ?", teamID)
	} else {
		result = query.
			Find(&groups, "team_id = ?", teamID)
	}
	if result.Error != nil {
		return nil, db.logError(errors.Database(result.Error))
	}

	return groups, nil
}

func (db vulcanitoStore) GroupAsset(assetsGroup api.AssetGroup, teamID string) (*api.AssetGroup, error) {
	asset := api.Asset{ID: assetsGroup.AssetID}
	if db.Conn.Where("team_id = ?", teamID).First(&asset).RecordNotFound() {
		return nil, db.logError(errors.Forbidden("asset does not belong to team"))
	}
	group := api.Group{ID: assetsGroup.GroupID}
	if db.Conn.Where("team_id = ?", teamID).First(&group).RecordNotFound() {
		return nil, db.logError(errors.Forbidden("group does not belong to team"))
	}
	if !db.Conn.First(&assetsGroup).RecordNotFound() {
		return nil, db.logError(errors.Duplicated("asset group relation already exists"))
	}
	res := db.Conn.Create(&assetsGroup)
	if res.Error != nil {
		return nil, db.logError(errors.Create(res.Error))
	}
	db.Conn.
		Preload("Asset").
		Preload("Asset.Team").
		Preload("Group").
		Preload("Group.Team").
		Preload("Group.AssetGroup").First(&assetsGroup)
	return &assetsGroup, nil
}

func (db vulcanitoStore) ListAssetGroup(assetGroup api.AssetGroup, teamID string) ([]*api.AssetGroup, error) {
	group := api.Group{ID: assetGroup.GroupID}
	if db.Conn.Where("team_id = ?", teamID).First(&group).RecordNotFound() {
		return nil, db.logError(errors.Forbidden("group does not belong to team"))
	}
	assetGroups := []*api.AssetGroup{}
	res := db.Conn.
		Preload("Asset").
		Preload("Group").
		Preload("Asset.AssetType").
		Find(&assetGroups, "group_id = ?", assetGroup.GroupID)
	if res.Error != nil {
		return nil, db.logError(errors.Database(res.Error))
	}
	return assetGroups, nil
}

func (db vulcanitoStore) UngroupAssets(assetGroup api.AssetGroup, teamID string) error {
	asset := api.Asset{ID: assetGroup.AssetID}
	if db.Conn.Where("team_id = ?", teamID).First(&asset).RecordNotFound() {
		return db.logError(errors.Forbidden("asset does not belong to team"))
	}
	group := api.Group{ID: assetGroup.GroupID}
	if db.Conn.Where("team_id = ?", teamID).First(&group).RecordNotFound() {
		return db.logError(errors.Forbidden("group does not belong to team"))
	}
	if db.Conn.First(&assetGroup).RecordNotFound() {
		return db.logError(errors.Duplicated("asset group relation does not exists"))
	}
	res := db.Conn.Delete(&assetGroup)
	if res.Error != nil {
		return db.logError(errors.Delete(res.Error))
	}

	return nil
}
