/*
Copyright 2021 Adevinta
*/

package store

import (
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

// DB is a postgres driver
type vulcanitoStore struct {
	Conn     *gorm.DB
	logger   log.Logger
	defaults DefaultEntities
}

// Default entities that may be used accross the API
// NOTE: This is a map to allow for testing without
// circular dependencies in the testutils package
type DefaultEntities map[string][]string

// NewDB returns a connection to a Postgres instance
func NewDB(pDialect, connectionString string, logger log.Logger, logMode bool, defaults map[string][]string) (api.VulcanitoStore, error) {
	if pDialect == "" {
		pDialect = dialect
	}
	conn, err := gorm.Open(pDialect, connectionString)
	if err != nil {
		return nil, err
	}
	conn.LogMode(logMode)
	return vulcanitoStore{
		Conn:     conn,
		logger:   logger,
		defaults: defaults,
	}, nil
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
