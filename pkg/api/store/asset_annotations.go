/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

// ListAssetAnnotations returns all annotations of a given asset id.
func (db vulcanitoStore) ListAssetAnnotations(teamID string, assetID string) ([]*api.AssetAnnotation, error) {
	// List annotations.
	asset := api.Asset{}
	result := db.Conn.
		Preload("AssetAnnotations").
		Where("team_id = ? and id = ?", teamID, assetID).
		First(&asset)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}
	return asset.AssetAnnotations, nil
}

// CreateAssetAnnotations assign new annotations to a given asset belonging to
// a given team.
func (db vulcanitoStore) CreateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	asset := api.Asset{
		ID:               assetID,
		TeamID:           teamID,
		AssetAnnotations: annotations,
	}
	_, err := db.updateAssetTX(tx, asset, annotationsCreateBehavior)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	return annotations, nil
}

// UpdateAssetAnnotations updates the value of the existing annotations of a given
// asset.
func (db vulcanitoStore) UpdateAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	asset := api.Asset{
		ID:               assetID,
		TeamID:           teamID,
		AssetAnnotations: annotations,
	}
	_, err := db.updateAssetTX(tx, asset, annotationsUpdateBehavior)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	// TODO: Depending on the execution of other concurrent transactions we
	// could be returning a set of annotations with information that were never
	// present "as is" in the DB at any point in time.
	return annotations, nil
}

// PutAssetAnnotations overrides all annotations of a given asset with new content.
// Previous annotations will not ne preserved.
func (db vulcanitoStore) PutAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) ([]*api.AssetAnnotation, error) {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	asset := api.Asset{
		ID:               assetID,
		TeamID:           teamID,
		AssetAnnotations: annotations,
	}
	_, err := db.updateAssetTX(tx, asset, annotationsReplaceBehavior)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if tx.Commit().Error != nil {
		return nil, db.logError(errors.Database(tx.Error))
	}
	// Notice that depending on the execution of other concurrent transactions
	// we could be returning a set of annotations with information that were
	// never present "as is" in the DB at any point time.
	return annotations, nil
}

// DeleteAssetAnnotations deletes the given annotations belonging to the given
// asset and team.
func (db vulcanitoStore) DeleteAssetAnnotations(teamID string, assetID string, annotations []*api.AssetAnnotation) error {
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	asset := api.Asset{
		ID:               assetID,
		TeamID:           teamID,
		AssetAnnotations: annotations,
	}
	_, err := db.updateAssetTX(tx, asset, annotationsDeleteBehavior)
	if err != nil {
		tx.Rollback()
		return err
	}
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	return nil
}
