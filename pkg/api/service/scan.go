/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	errs "errors"
	"fmt"
	"net/http"

	"gopkg.in/go-playground/validator.v9"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	metrics "github.com/adevinta/vulcan-metrics-client"
	scanengineData "github.com/adevinta/vulcan-scan-engine/pkg/api/endpoint"
)

// ListScans returns the list of scans belonging to a a team. The method can
// only return scans that belong to a program that has not been deleted because
// is the only way to know if a scan belongs to a team.
func (s vulcanitoService) ListScans(ctx context.Context, teamID string, programID string) ([]*api.Scan, error) {
	program, err := s.FindProgram(ctx, programID, teamID)
	if err != nil {
		return nil, err
	}
	scanengine := scanengine.NewClient(ctx, http.DefaultClient, s.scanEngineConfig)
	scans := []*api.Scan{}
	response, err := scanengine.GetScans(programID)
	if err != nil {
		return nil, err
	}
	for _, scanInfo := range response.Scans {
		scanInfo := scanInfo
		scan, err := s.fillScanInfo(ctx, &scanInfo, program, programID, teamID)
		if err != nil {

			return nil, err
		}
		scans = append(scans, scan)
	}

	return scans, nil
}

// CreateScan runs a program by calling the scan engine component with the
// parameters defined in the program. The function can only return scans from
// programs that are not deleted because the program is the only way to know if
// a scan belongs to a team.
func (s vulcanitoService) CreateScan(ctx context.Context, scan api.Scan, teamID string) (*api.Scan, error) {
	validationErr := validator.New().Struct(scan)
	if validationErr != nil {
		return nil, errors.Validation(validationErr)
	}
	// TODO add a new function that returns all the programs so we can return
	// also scans that belong to a program that has been deleted.
	program, err := s.FindProgram(ctx, scan.ProgramID, teamID)
	if err != nil {
		return nil, err
	}

	if program.Disabled != nil && *program.Disabled == true {
		return nil, errors.Validation(fmt.Errorf("Program %s is disabled. [Program ID: %s]", program.Name, program.ID))
	}

	err = program.ValidateGroupsPolicies()
	if err != nil {
		return nil, err
	}

	team, err := s.db.FindTeam(teamID)
	if err != nil {
		return nil, err
	}
	scanengineClient := scanengine.NewClient(ctx, http.DefaultClient, s.scanEngineConfig)
	scanRequest, err := scanengineClient.CreateScanRequest(*program, scan.ScheduledTime, program.ID, scan.RequestedBy, team.Tag)
	if err != nil {
		if errs.Is(err, scanengine.ErrProgramWithoutPolicyGroups) {
			return nil, errors.Validation(err)
		}
		if errs.Is(err, scanengine.ErrNotFound) {
			return nil, errors.NotFound(err)
		}
		return nil, errors.Default(err)
	}

	scanResponse, err := scanengineClient.Create(scanRequest)
	if err != nil {
		if errs.Is(err, scanengine.ErrUnprocessableEntity) {
			return nil, errors.Validation(err)
		}
		return nil, errors.Default(err)
	}

	if scanResponse.ScanID == "" {
		return nil, errors.Default("Scan engine did not return scan id")
	}

	createdScan, err := s.FindScan(ctx, scanResponse.ScanID, teamID)
	if err != nil {
		return nil, err
	}

	s.pushScanMetrics(team, program)

	return createdScan, nil
}

// pushScanMetrics pushes metrics related to the created scan and its checks.
func (s vulcanitoService) pushScanMetrics(team *api.Team, program *api.Program) {
	componentTag := "component:api"
	scanTag := buildScanTag(team, program)
	scanStatusTag := "scanstatus:requested"

	s.metricsClient.Push(metrics.Metric{
		Name:  "vulcan.scan.count",
		Typ:   metrics.Count,
		Value: 1,
		Tags:  []string{componentTag, scanTag, scanStatusTag},
	})
}

func buildScanTag(team *api.Team, program *api.Program) string {
	var teamLabel, programLabel string

	if team.Tag == "" {
		teamLabel = "unknown"
	} else {
		teamLabel = team.Tag
	}

	if programLabel = program.ID; programLabel == "" {
		programLabel = "unknown"
	}

	return fmt.Sprint("scan:", teamLabel, "-", programLabel)
}

func (s vulcanitoService) FindScan(ctx context.Context, scanID, teamID string) (*api.Scan, error) {
	scan := &api.Scan{
		ID: scanID,
	}
	// query scan-engine
	scanengine := scanengine.NewClient(ctx, http.DefaultClient, s.scanEngineConfig)
	scanResponse, err := scanengine.Get(scanID)
	if err != nil {
		return nil, errors.Default(err)
	}

	program, err := s.FindProgram(ctx, scanResponse.ExternalID, teamID)
	if err != nil {
		return nil, err
	}

	scan.ProgramID = program.ID
	scan.Program = program
	scan.RequestedBy = scanResponse.Trigger
	scan.Progress = scanResponse.Progress
	scan.CheckCount = scanResponse.CheckCount
	scan.ScheduledTime = scanResponse.ScheduledTime
	scan.StartTime = scanResponse.StartTime
	scan.EndTime = scanResponse.EndTime
	scan.Status = scanResponse.Status

	return scan, nil
}

func (s vulcanitoService) fillScanInfo(ctx context.Context, scanInfo *scanengineData.GetScanResponse, p *api.Program, programID, teamID string) (*api.Scan, error) {
	scan := &api.Scan{
		ID: scanInfo.ID,
	}
	scan.ProgramID = programID
	scan.Program = p
	scan.RequestedBy = scanInfo.Trigger
	scan.Progress = scanInfo.Progress
	scan.CheckCount = scanInfo.CheckCount
	scan.EndTime = scanInfo.EndTime
	scan.Status = scanInfo.Status
	return scan, nil
}

func (s vulcanitoService) AbortScan(ctx context.Context, scanID string, teamID string) (*api.Scan, error) {
	scan, err := s.FindScan(ctx, scanID, teamID)
	if err != nil {
		return nil, err
	}
	if scan.Program.TeamID != teamID {
		return nil, errors.Forbidden("Scan does not belong to given team")
	}
	scanengine := scanengine.NewClient(ctx, http.DefaultClient, s.scanEngineConfig)
	scanEngineScan, err := scanengine.Abort(scanID)
	if err != nil {
		return nil, errors.Default(err)
	}
	scan.Status = scanEngineScan.Status
	return scan, nil
}

// TODO: no endpoint exists for update/delete scan
func (s vulcanitoService) UpdateScan(ctx context.Context, scan api.Scan) (*api.Scan, error) {
	return nil, errors.Default("not implemented")
}

func (s vulcanitoService) DeleteScan(ctx context.Context, scan api.Scan) error {
	return errors.Default("not implemented")
}
