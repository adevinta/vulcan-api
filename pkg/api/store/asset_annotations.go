/*
Copyright 2021 Adevinta
*/

package store

import (
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// ListAssetAnnotations returns all annotations of a given asset id
func (db vulcanitoStore) ListAssetAnnotations(teamID string, assetID string) ([]*api.AssetAnnotation, error) {
	// Find asset
	a := api.Asset{}
	result := db.Conn.Where("team_id = ?", teamID).Where("id = ?", assetID).Find(&a)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(result.Error)
	}

	// List annotations
	annotations := []*api.AssetAnnotation{}
	result = db.Conn.
		Preload("Asset").
		Joins("left join assets on assets.id = asset_annotations.asset_id").
		Where("asset_id = ?", assetID).
		Where("team_id = ?", teamID).
		Find(&annotations)

	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return annotations, nil
}

// GetAssetAnnotation retrives a single Annotation by its key
func (db vulcanitoStore) GetAssetAnnotation(teamID string, assetID string, key string) (*api.AssetAnnotation, error) {
	// Create new annotations
	annotation := api.AssetAnnotation{}
	result := db.Conn.
		Preload("Asset").
		Preload("Asset.Team").
		Joins("left join assets on assets.id = asset_annotations.asset_id").
		Joins("left join teams on teams.id = assets.team_id").
		Where("teams.id = ?", teamID).
		Where("asset_id = ?", assetID).
		Where("key = ?", key).
		First(&annotation)

	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return &annotation, nil
}

// CreateAssetAnnotations assign new annotations of a given asset id
func (db vulcanitoStore) CreateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	// Find asset
	a := api.Asset{}
	result := db.Conn.Where("team_id = ?", teamID).Where("id = ?", assetID).Find(&a)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(result.Error)
	}

	// Start a new transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	// Retrieve annotations for the asset
	createdAnnotations := []*api.AssetAnnotation{}
	for _, annotation := range annotations {
		// Ensure consistent Asset ID
		annotation.AssetID = assetID

		// Check if annotation already exists. If yes, reject the whole
		// request
		_, err := db.GetAssetAnnotation(teamID, assetID, annotation.Key)
		if err == nil {
			tx.Rollback()
			return nil, db.logError(errors.Create(fmt.Errorf("annotation '%v' already present for asset id '%v'", annotation.Key, annotation.AssetID)))
		}
		if err != nil && !errors.IsKind(err, errors.ErrNotFound) {
			tx.Rollback()
			return nil, err
		}

		// Create Annotation
		result := tx.Create(&annotation)
		if result.Error != nil {
			tx.Rollback()
			return nil, db.logError(errors.Create(result.Error))
		}
		createdAnnotations = append(createdAnnotations, annotation)
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return createdAnnotations, nil
}

// UpdateAssetAnnotations updates the value of existing annotations of a given
// asset
func (db vulcanitoStore) UpdateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	// Find asset
	a := api.Asset{}
	result := db.Conn.Where("team_id = ?", teamID).Where("id = ?", assetID).Find(&a)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(result.Error)
	}

	// Start transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	updatedAnnotations := []*api.AssetAnnotation{}
	for _, annotation := range annotations {
		// Ensure consistent Asset ID
		annotation.AssetID = assetID

		// Ensure annotation already exists. If not, error out
		an, err := db.GetAssetAnnotation(teamID, assetID, annotation.Key)
		if err != nil {
			tx.Rollback()
			if errors.IsKind(err, errors.ErrNotFound) {
				return nil, db.logError(errors.NotFound(fmt.Errorf("annotation '%v' not found for asset id '%v'", annotation.Key, annotation.AssetID)))
			}
			return nil, err
		}
		// Update Annotation
		result := tx.Model(&an).Update(annotation)
		if result.Error != nil {
			tx.Rollback()
			return nil, db.logError(errors.Update(result.Error))
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, db.logError(errors.Update("Asset Annotation was not updated"))
		}

		updatedAnnotations = append(updatedAnnotations, an)
	}

	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}

	return updatedAnnotations, nil
}

// DeleteAssetAnnotations deletes annotations "keys" from a given asset
func (db vulcanitoStore) DeleteAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) error {
	// Find asset
	a := api.Asset{}
	result := db.Conn.Where("team_id = ?", teamID).Where("id = ?", assetID).Find(&a)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return db.logError(errors.NotFound(result.Error))
		}
		return db.logError(result.Error)
	}

	// Start transaction
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	for _, annotation := range annotations {
		// Ensure annotation already exists. If not, error out
		an, err := db.GetAssetAnnotation(teamID, assetID, annotation.Key)
		if err != nil {
			tx.Rollback()
			if errors.IsKind(err, errors.ErrNotFound) {
				return db.logError(errors.NotFound(fmt.Errorf("annotation '%v' not found for asset id '%v'", annotation.Key, annotation.AssetID)))
			}
			return err
		}
		// Delete Annotation
		result := tx.Model(&an).Delete(&an)
		if result.Error != nil {
			tx.Rollback()
			return db.logError(errors.Update(result.Error))
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return db.logError(errors.Update("Asset Annotation was not delete"))
		}
	}

	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	return nil
}
