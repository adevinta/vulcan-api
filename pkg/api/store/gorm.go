/*
Copyright 2021 Adevinta
*/

package store

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jinzhu/gorm"

	// Import postgres dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/adevinta/vulcan-api/pkg/api"
)

const (
	dialect = "postgres"
)

var forbiddenCharsRe = regexp.MustCompile(`[^a-z^A-Z^_^-]`)

// DB is a postgres driver
type vulcanitoStore struct {
	Conn     *gorm.DB
	logger   log.Logger
	defaults DefaultEntities
}

// Store provides access to the storage layer of the Vulcan API.
type Store struct {
	*vulcanitoStore
}

// DefaultEntities that may be used accross the API
// NOTE: This is a map to allow for testing without
// circular dependencies in the testutils package
type DefaultEntities map[string][]string

// NewStore returns an initialized Vulcan Store.
func NewStore(pDialect, connectionString string, logger log.Logger, logMode bool, defaults map[string][]string) (Store, error) {
	if pDialect == "" {
		pDialect = dialect
	}
	conn, err := gorm.Open(pDialect, connectionString)
	if err != nil {
		return Store{}, err
	}
	conn.LogMode(logMode)
	vs := vulcanitoStore{
		Conn:     conn,
		logger:   logger,
		defaults: defaults,
	}
	return Store{&vs}, nil
}

// TODO: Refactor to return a public struct and not a interface defined in other package.

// NewDB returns a connection to a Postgres instance
func NewDB(pDialect, connectionString string, logger log.Logger, logMode bool, defaults map[string][]string) (api.VulcanitoStore, error) {
	return NewStore(pDialect, connectionString, logger, logMode, defaults)
}

// NotFoundError is an utility method to check if a returned error is a not found error.
// This is needed because sometimes the services will need to act different depppending if an error is because a record it's not
// found or because other causes.
func (db vulcanitoStore) NotFoundError(err error) bool {
	return gorm.ErrRecordNotFound == err
}

// DuplicateError is an utility method to check if a returned error is a duplicate key error.
func (db vulcanitoStore) IsDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key")
}

// Close close vulcanitoStore db connection
func (db vulcanitoStore) Close() error {
	return db.Conn.Close()
}

func (db vulcanitoStore) logError(err error) error {
	_, file, line, _ := runtime.Caller(1)
	_ = level.Error(db.logger).Log("caller", file, "line", line, "err", err)
	return err
}

// lockTablesUnchecked locks the given tables for writing. WARNING: the tables
// paramenter must never be controlled by an external user, because the
// function can't use prepared statements.
func (db vulcanitoStore) lockTablesUnchecked(tx *gorm.DB, tables ...string) error {
	for _, t := range tables {
		t = forbiddenCharsRe.ReplaceAllString(t, "-")
		stm := fmt.Sprintf("LOCK TABLE ONLY %s IN EXCLUSIVE MODE", t)
		result := tx.Exec(stm)
		if result.Error != nil {
			err := fmt.Errorf("error locking table %s: %w", t, result.Error)
			return err
		}
	}
	return nil
}
