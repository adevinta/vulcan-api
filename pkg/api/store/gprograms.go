/*
Copyright 2021 Adevinta
*/

package store

import (
	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
)

func (db vulcanitoStore) UpsertGlobalProgramMetadata(teamID, program string, defaultAutosend bool, defaultDisabled bool, defaultCron string, autosend *bool, disabled *bool, cron *string) error {
	var err error

	paramAutosend := defaultAutosend
	if autosend != nil {
		paramAutosend = *autosend
	}

	paramDisabled := defaultDisabled
	if disabled != nil {
		paramDisabled = *disabled
	}

	paramCron := defaultCron
	if cron != nil {
		paramCron = *cron
	}

	if autosend != nil {
		err = db.Conn.Exec(`INSERT INTO global_programs_metadata(team_id, program, autosend, disabled, cron) VALUES(?,?,?,?,?)
ON CONFLICT ON CONSTRAINT global_programs_metadata_pkey
DO
 UPDATE
	SET autosend=EXCLUDED.autosend`, teamID, program, paramAutosend, paramDisabled, paramCron).Error
		if err != nil {
			return err
		}
	}

	if disabled != nil {
		err = db.Conn.Exec(`INSERT INTO global_programs_metadata(team_id, program, autosend, disabled, cron) VALUES(?,?,?,?,?)
ON CONFLICT ON CONSTRAINT global_programs_metadata_pkey
DO
 UPDATE
	SET disabled=EXCLUDED.disabled`, teamID, program, paramAutosend, paramDisabled, paramCron).Error
		if err != nil {
			return err
		}
	}

	if cron != nil {
		err = db.Conn.Exec(`INSERT INTO global_programs_metadata(team_id, program, autosend, disabled, cron) VALUES(?,?,?,?,?)
ON CONFLICT ON CONSTRAINT global_programs_metadata_pkey
DO
 UPDATE
	SET cron=EXCLUDED.cron`, teamID, program, paramAutosend, paramDisabled, paramCron).Error
		if err != nil {
			return err
		}
	}

	return err
}

func (db vulcanitoStore) FindGlobalProgramMetadata(programID string, teamID string) (*api.GlobalProgramsMetadata, error) {
	program := &api.GlobalProgramsMetadata{
		TeamID:  teamID,
		Program: programID,
	}
	result := db.Conn.Find(&program)
	if result.Error != nil {
		if db.NotFoundError(result.Error) {
			return nil, db.logError(errors.NotFound(result.Error))
		}
		return nil, db.logError(errors.Database(result.Error))
	}
	return program, nil
}

func (db vulcanitoStore) UpdateGlobalProgramMetadata(metadata api.GlobalProgramsMetadata) error {
	return db.Conn.Model(&metadata).Updates(metadata).Error
}

func (db vulcanitoStore) DeleteProgramMetadata(program string) error {
	return db.Conn.Exec("delete from global_programs_metadata where program = ?", program).Error
}
