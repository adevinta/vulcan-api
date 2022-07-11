/*
Copyright 2022 Adevinta
*/

package store

import (
	goerrors "errors"
	"fmt"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/jinzhu/gorm"
)

// ErrReadAssetsFinished is returned by the Read operation of an AssetReader
// when there are no more assets to read.
var ErrReadAssetsFinished = goerrors.New("no more assets")

// ReadAssetsToken defines the opaque structure that keeps the required state
// between calls to the methods StartReadingAssets, ReadAssets and
// FinishReadingAssets.
type ReadAssetsToken struct {
	next     string
	pageSize int
	tx       *gorm.DB
}

// ReadAssetsResult contains the data returned by the ReadNextAssetsPage method
// of the vulcanitoStore.
type ReadAssetsResult struct {
	Assets []*api.Asset
	Token  ReadAssetsToken
	More   bool
}

// StartReadingAssetsResult contains the data returned by the
// StartReadingAssetsResult of the vulcanitoStore.
type StartReadingAssetsResult struct {
	Assets     []*api.Asset
	Token      ReadAssetsToken
	More       bool
	TotalPages int
}

// NewAssetReader creates a new AssetReader with the given page size.
// If the lock param is set to true it will lock for writing the tables:
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
	// pass enough time for the transaction to be timed out by postgres
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
	var total int
	var count = struct {
		Total int
	}{}
	res := tx.Raw("SELECT count(*) as Total from assets").Scan(&count)
	if res.Error != nil {
		tx.Rollback()
		err := fmt.Errorf("error counting assets: %w", res.Error)
		return AssetsReader{}, err
	}
	reader := AssetsReader{
		pageSize: pageSize,
		total:    total,
		tx:       tx,
		more:     true,
		lock:     lock,
	}
	return reader, nil
}

// AssetsReader allows to read all the assets stored in Vulcan using pages with
// user defined the size.
type AssetsReader struct {
	// The total number of assets that the reader will read.
	Total    int
	next     string
	pageSize int
	tx       *gorm.DB
	more     bool
	total    int
	lock     bool
}

// Read returns the next page of the assets, according to the the page size the
// reader was created with and the total number of assets to read.
func (a *AssetsReader) Read() ([]*api.Asset, error) {
	if !a.more {
		return nil, ErrReadAssetsFinished
	}
	// Check if this is the first call to read.
	if a.next == "" {
		return a.readFirst()
	}
	// stm := `SELECT * FROM assets a JOIN teams t ON a.team_id = t.id JOIN asset_annotations an ON
	// an.asset_id = a.id where a.id >= ? ORDER BY a.id LIMIT ?`
	assets := make([]*api.Asset, 0)
	limit := a.pageSize + 1
	next := a.next
	tx := a.tx
	//res := tx.Raw(stm, next, limit).Scan(&assets)
	res := tx.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Where("id >= ?", next).
		Order("id", true).
		Limit(limit).
		Find(&assets)
	if res.Error != nil {
		tx.Rollback()
		err := fmt.Errorf("error reading assets: %w", res.Error)
		return nil, err
	}
	last := ""
	more := false
	if len(assets) == limit {
		last = assets[len(assets)-1].ID
		more = true
		assets = assets[0 : len(assets)-1]
	}

	a.next = last
	a.more = more
	if !more {
		return assets, ErrReadAssetsFinished
	}
	return assets, nil
}

// Close closes the reader, if the reader was created with the lock parameter
// set to true, it unlocks the tables that were locked when it was created.
func (a *AssetsReader) Close() error {
	return a.tx.Commit().Error
}

func (a *AssetsReader) readFirst() ([]*api.Asset, error) {
	tx := a.tx
	// stm := `SELECT * FROM assets JOIN teams ON assets.team_id = teams.id JOIN asset_annotations ON
	// asset_annotations.asset_id = assets.id ORDER BY assets.id LIMIT ?`
	assets := make([]*api.Asset, 0)
	pageSize := a.pageSize
	limit := pageSize + 1
	//res := tx.Raw(stm, limit).Scan(&assets)
	res := tx.Preload("Team").
		Preload("AssetType").
		Preload("AssetAnnotations").
		Order("id", true).
		Limit(limit).
		Find(&assets)
	if res.Error != nil {
		tx.Rollback()
		err := fmt.Errorf("error reading assets: %w", res.Error)
		return nil, err
	}
	last := ""
	more := false
	if len(assets) == limit {
		last = assets[len(assets)-1].ID
		more = true
		assets = assets[0 : len(assets)-1]
	}
	a.next = last
	a.more = more
	if !more {
		return assets, ErrReadAssetsFinished
	}
	return assets, nil
}
