/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) CreateFindingOverride(findingOverride api.FindingOverride) error {
	// Begin transaction.
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return db.logError(errors.Database(tx.Error))
	}

	// create entry in finding_override
	result := tx.Create(&findingOverride)
	if result.Error != nil {
		tx.Rollback()
		return db.logError(errors.Delete(result.Error))
	}

	err := db.pushToOutbox(tx, opFindingOverride, findingOverride)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction.
	if tx.Commit().Error != nil {
		return db.logError(errors.Database(tx.Error))
	}
	return nil

}

func (db vulcanitoStore) ListFindingOverrides(findingID string) ([]*api.FindingOverride, error) {
	findingOverrides := []*api.FindingOverride{}
	result := db.Conn.Preload("User").Find(&findingOverrides, "finding_id = ?", findingID)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}

	return findingOverrides, nil
}
