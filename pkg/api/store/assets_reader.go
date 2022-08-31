/*
Copyright 2022 Adevinta
*/

package store

import (
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/jinzhu/gorm"
)

// NewAssetReader creates a new [AssetsReader] with the given page size. If the
// lock param is set to true it will lock for writing the following tables:
// Assets, Teams and AssetAnnotations.
func (db vulcanitoStore) NewAssetReader(lock bool, pageSize int) (AssetsReader, error) {
	if pageSize < 1 {
		err := fmt.Errorf("invalid page size %d, it must be greater than 0", pageSize)
		return AssetsReader{}, err
	}
	tx := db.Conn.Begin()
	if tx.Error != nil {
		return AssetsReader{}, db.logError(errors.Database(tx.Error))
	}
	// Even though, theoretically, between calls to the Read method it could
	// pass enough time for the transaction to be timed out by Postgres
	// depending on the configured max amount of time for a transaction to be
	// idle, it's in practice not very likely to happen, so by now, we are not
	// implementing a `NOP loop` to ensure the transaction is not closed for
	// this reason.
	if lock {
		// Lock the teams, assets, and asset_annotations tables for writing.
		err := db.lockTablesUnchecked(tx, "teams", "assets", "asset_annotations")
		if err != nil {
			tx.Rollback()
			err := fmt.Errorf("error locking table teams: %w", err)
			return AssetsReader{}, db.logError(err)
		}
	}
	reader := AssetsReader{
		pageSize: pageSize,
		tx:       tx,
		more:     true,
		lock:     lock,
	}
	return reader, nil
}

// AssetsReader reads all the assets stored in Vulcan using pages with
// a configurable size.
type AssetsReader struct {
	next     string
	pageSize int
	tx       *gorm.DB
	more     bool
	lock     bool
	assets   []*api.Asset
	err      error
}

// Read returns the next page of the assets according to the page size of the
// [*AssetsReader]. Returns true if the read operation was successful, in that
// case the assets can be retrieved by calling [*AssetsReader.Assets].
func (a *AssetsReader) Read() bool {
	if !a.more {
		return false
	}
	// Check if this is the first call to read.
	if a.next == "" {
		return a.readFirst()
	}
	// Clean the slice.
	a.assets = a.assets[:0]
	limit := a.pageSize + 1
	next := a.next
	tx := a.tx

	res := tx.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Where("id >= ?", next).
		Order("id", true).
		Limit(limit).
		Find(&a.assets)
	if res.Error != nil {
		tx.Rollback()
		err := fmt.Errorf("error reading assets: %w", res.Error)
		a.more = false
		a.err = err
		return false
	}

	// There are more assets to read.
	if len(a.assets) == limit {
		a.next = a.assets[len(a.assets)-1].ID
		a.more = true
		a.assets = a.assets[0 : len(a.assets)-1]
		return true
	}

	// No more assets.
	a.next = ""
	a.more = false
	return len(a.assets) > 0
}

// Close closes the reader and unlocks the tables that were locked when it was
// created.
func (a *AssetsReader) Close() error {
	// Notice the tables are automatically unlocked when the transaction is
	// committed.
	return a.tx.Commit().Error
}

// Err returns the error produced by the last call to [*AssetsReader.Read],
// returns nil if the last call didn't produce any error.
func (a *AssetsReader) Err() error {
	// Notice the tables are automatically unlocked when the transaction is
	// committed.
	return a.tx.Commit().Error
}

// Assets returns the assets produced by the last call to [*AssetsReader.Read].
func (a *AssetsReader) Assets() []*api.Asset {
	return a.assets
}

func (a *AssetsReader) readFirst() bool {
	tx := a.tx
	assets := make([]*api.Asset, 0, a.pageSize)
	pageSize := a.pageSize
	limit := pageSize + 1

	res := tx.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Order("id", true).
		Limit(limit).
		Find(&assets)
	if res.Error != nil {
		tx.Rollback()
		a.err = fmt.Errorf("error reading assets: %w", res.Error)
		a.more = false
		return false
	}
	// There are more assets.
	if len(assets) == limit {
		a.next = assets[len(assets)-1].ID
		a.more = true
		assets = assets[0 : len(assets)-1]
		a.assets = assets
		return true
	}
	// No more assets.
	a.next = ""
	a.more = false
	a.assets = assets
	return len(a.assets) > 0
}
