/*
Copyright 2021 Adevinta
*/

package cdc

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	// OutboxVersion specifies the schema version
	// used to store data in outbox table.
	OutboxVersion    = 1
	defOutboxDBTable = "outbox"
)

// DB represents a database handle to
// perform CDC related operations synchronized
// across different instances.
type DB interface {
	GetLog() ([]Event, error)
	FailedEvent(event Event) error
	CleanEvent(event Event) error
	CleanLog(nEntries uint) error
	TryGetLock(id uint32) (*Lock, error)
	ReleaseLock(l *Lock) error
}

// Event represents an event retrieved from CDC log.
//   - ID returns the event identifier.
//   - Action returns the action related with a CBC event.
//   - Version returns the schema version for data.
//   - Data returns the data associated with the event.
//   - ReadCount returns the number of times event has been read.
type Event interface {
	ID() string
	Action() string
	Version() int
	Data() []byte
	ReadCount() int
}

// Lock represents an advisory lock
type Lock struct {
	Acquired bool
	Tx       *sql.Tx
}

// PQDB represents the PostgreSQL implementation
// of DB handle to retrieve data from an outbox table.
// Outbox pattern: https://microservices.io/patterns/data/transactional-outbox.html
type PQDB struct {
	db      *sql.DB
	dbTable string
}

// Outbox represents an entry in the
// outbox table.
type Outbox struct {
	Identifier string `gorm:"column:id"`
	Operation  string
	SchemaVer  int    `gorm:"column:version"`
	DTO        []byte `gorm:"column:data"`
	Retries    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (o Outbox) ID() string {
	return o.Identifier
}
func (o Outbox) Action() string {
	return o.Operation
}
func (o Outbox) Version() int {
	return o.SchemaVer
}
func (o Outbox) Data() []byte {
	return o.DTO
}
func (o Outbox) ReadCount() int {
	return o.Retries
}
func (o Outbox) TableName() string {
	return defOutboxDBTable
}

// NewPQDB creates a new PostgreSQL DB handle for
// CDC related operations.
func NewPQDB(conStr, dbTable string) (*PQDB, error) {
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}
	if dbTable == "" {
		dbTable = defOutboxDBTable
	}
	return &PQDB{
		db:      db,
		dbTable: dbTable,
	}, nil
}

// GetLog retrieves the log entries from the outbox table ordered by creation time.
func (p *PQDB) GetLog() ([]Event, error) {
	query := fmt.Sprintf(`SELECT id, operation, version, data, retries
		FROM %s ORDER BY created_at`, p.dbTable)
	res, err := p.db.Query(query)
	if err != nil {
		return []Event{}, err
	}
	defer res.Close()

	log := []Event{}

	for res.Next() {
		var box Outbox
		err = res.Scan(
			&box.Identifier, &box.Operation, &box.SchemaVer, &box.DTO, &box.Retries,
		)
		if err != nil {
			return []Event{}, err
		}

		log = append(log, box)
	}

	return log, nil
}

// FailedEvent increments the given event retries in DB.
func (p *PQDB) FailedEvent(event Event) error {
	query := fmt.Sprintf(`UPDATE %s 
		SET retries = retries+1, updated_at = $1
		WHERE id = $2`, p.dbTable,
	)
	_, err := p.db.Exec(query, time.Now(), event.ID())
	return err
}

// CleanEvent deletes the given event from outbox table.
func (p *PQDB) CleanEvent(event Event) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", p.dbTable)
	_, err := p.db.Exec(query, event.ID())
	return err
}

// CleanLog deletes the oldest nEntries from the outbox table.
func (p *PQDB) CleanLog(nEntries uint) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id IN (
			SELECT id FROM %s ORDER BY created_at LIMIT $1
		)`, p.dbTable, p.dbTable,
	)
	_, err := p.db.Exec(query, nEntries)
	return err
}

// TryGetLock tries to acquire the CDC advisory lock from DB.
// If no error is returned, lock should be released by calling
// ReleaseLock method, even if it was not acquired.
func (p *PQDB) TryGetLock(id uint32) (*Lock, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := p.db.Query("SELECT pg_try_advisory_xact_lock($1)", id)
	if err != nil {
		return nil, err
	}
	defer res.Close() // nolint

	var acquired bool
	res.Next()
	err = res.Scan(&acquired)
	return &Lock{Acquired: acquired, Tx: tx}, err
}

// ReleaseLock releases the input lock.
func (p *PQDB) ReleaseLock(l *Lock) error {
	if l == nil {
		return nil
	}
	return l.Tx.Commit()
}
