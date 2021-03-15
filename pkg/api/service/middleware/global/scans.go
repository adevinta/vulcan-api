/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/adevinta/errors"
	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/adevinta/vulcan-api/pkg/api"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	scanengineApi "github.com/adevinta/vulcan-scan-engine/pkg/api"
	scanengineData "github.com/adevinta/vulcan-scan-engine/pkg/api/endpoint"
)

// ScanNotification holds the required fields to unmarshal
type ScanNotification struct {
	ProgramID string `json:"program_id"`
}

func (e *globalEntities) CreateScan(ctx context.Context, scan api.Scan, teamID string) (*api.Scan, error) {
	_, programID := decodeGlobalProgramRequest(scan.ProgramID)
	_, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.CreateScan(ctx, scan, teamID)
	}
	program, err := e.FindProgram(ctx, programID, teamID)
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
	externalID := encodeGlobalProgramRequestID(programID, teamID)
	team, err := e.VulcanitoService.FindTeam(ctx, teamID)
	if err != nil {
		return nil, errors.Default(err)
	}

	scanengine := scanengine.NewClient(ctx, http.DefaultClient, e.scanEngineConfig)
	req, err := scanengine.CreateScanRequest(*program, scan.ScheduledTime, externalID, scan.RequestedBy, team.Tag)
	if err != nil {
		return nil, err
	}
	scanResponse, err := scanengine.Create(req)
	if err != nil {
		return nil, errors.Default(err)
	}

	if scanResponse.ScanID == "" {
		return nil, errors.Default("Scan engine did not return scan id")
	}
	createdScan, err := e.FindScan(ctx, scanResponse.ScanID, teamID)
	if err != nil {
		return nil, errors.Default(err)
	}

	e.pushScanMetrics(team, program)

	return createdScan, nil
}

// pushScanMetrics pushes metrics related to the created scan and its checks.
func (e *globalEntities) pushScanMetrics(team *api.Team, program *api.Program) {
	componentTag := "component:api"
	scanTag := buildScanTag(team, program)
	scanStatusTag := "scanstatus:requested"

	e.metricsClient.Push(metrics.Metric{
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
		teamTagParts := strings.Split(team.Tag, ":")
		teamLabel = teamTagParts[len(teamTagParts)-1]
	}

	if programLabel = program.ID; programLabel == "" {
		programLabel = "unknown"
	}

	return fmt.Sprint("scan:", teamLabel, "-", programLabel)
}

// GenerateReport triggers the generation of a new report.
func (e *globalEntities) GenerateReport(ctx context.Context, teamID, teamName, scanID string, autosend bool) error {
	scan, err := e.findScan(ctx, scanID, teamID)
	if err != nil {
		return err
	}
	// If scan is nil that means it does not belong to a global program so let
	// the VulcanitoService deal with the request. For this we will perform a
	// second request to the scan engine to fetch again the data of the scan.
	// Given that by now majority of the times the report will be generated for
	// a scan that belongs to a global program it is acceptable. Improving this
	// in future is a matter to add a parameter to the GenerateReport func that
	// can be nil, but if not, it will already contain the data about a scan.
	if scan == nil {
		return e.VulcanitoService.GenerateReport(ctx, teamID, teamName, scanID, autosend)
	}
	programName := scan.Program.Name
	return e.VulcanitoService.RunGenerateReport(ctx, autosend, scanID, programName, teamID, teamName)
}

// ProcessScanCheckNotification looks if the message is regarding a scan
// triggered from a global program.
func (e *globalEntities) ProcessScanCheckNotification(ctx context.Context, msg []byte) error {
	var m scanengineApi.ScanNotification
	err := json.Unmarshal(msg, &m)
	if err != nil {
		return errors.Default(err)
	}
	teamID, programID := decodeGlobalProgramRequest(m.ProgramID)
	_, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.ProcessScanCheckNotification(ctx, msg)
	}
	program, err := e.FindProgram(ctx, programID, teamID)
	if err != nil {
		return err
	}
	if m.Status == "FINISHED" {
		team, err := e.VulcanitoService.FindTeam(ctx, teamID)
		if err != nil {
			return errors.NotFound(err)
		}

		var autosend bool
		if program.Autosend != nil {
			autosend = *program.Autosend
		}
		err = e.GenerateReport(ctx, team.ID, team.Name, m.ScanID, autosend)
		if err != nil {
			_ = e.logger.Log("ErrGenerateSecurityOverview", err)
			return errors.Default(err)
		}
	}
	return nil
}

// ListScans is completely handled by the middleware because the scans of a team
// can contain scans triggered by global or not global programs.
func (e *globalEntities) ListScans(ctx context.Context, teamID string, programID string) ([]*api.Scan, error) {
	// Get possible scans for each global program.
	_, ok := e.store.Programs()[programID]
	if !ok {
		return e.VulcanitoService.ListScans(ctx, teamID, programID)
	}
	id := encodeGlobalProgramRequestID(programID, teamID)
	scans := []*api.Scan{}
	scanengine := scanengine.NewClient(ctx, http.DefaultClient, e.scanEngineConfig)
	response, err := scanengine.GetScans(id)
	if err != nil {
		return nil, err
	}
	program, err := e.FindProgram(ctx, programID, teamID)
	if err != nil {
		return nil, err
	}
	for _, scan := range response.Scans {
		s := e.buildScanWithProgram(&scan, program)
		scans = append(scans, s)
	}
	return scans, nil
}

func (e *globalEntities) FindScan(ctx context.Context, scanID, teamID string) (*api.Scan, error) {
	scan, err := e.findScan(ctx, scanID, teamID)
	if err != nil {
		return nil, err
	}
	// If the scan is nil that means it does not belong to a global program
	// so let the VulcanitoService deal with the request. For this we will
	// perform a second request to the scan engine to fetch again the data of
	// the scan. Given that, by now, majority of the times the report will be
	// generated for a scan that belongs to a global program it is acceptable.
	// Improving this in future is matter to add a parameter to the
	// GenerateReport func that can be nil, but if not, it will already contain
	// the data about a scan.
	if scan == nil {
		return e.VulcanitoService.FindScan(ctx, scanID, teamID)
	}
	return scan, nil
}

func (e *globalEntities) findScan(ctx context.Context, scanID, teamID string) (*api.Scan, error) {
	scanengine := scanengine.NewClient(ctx, http.DefaultClient, e.scanEngineConfig)
	scanResponse, err := scanengine.Get(scanID)
	if err != nil {
		return nil, err
	}
	programID := scanResponse.ExternalID
	_, programID = decodeGlobalProgramRequest(programID)
	_, ok := e.store.Programs()[programID]
	if !ok {
		return nil, nil
	}
	program, err := e.FindProgram(ctx, programID, teamID)
	if err != nil {
		return nil, err
	}
	scan := e.buildScanWithProgram(scanResponse, program)
	return scan, nil
}

func (e *globalEntities) buildScanWithProgram(scanInfo *scanengineData.GetScanResponse, program *api.Program) *api.Scan {
	scan := &api.Scan{
		ID: scanInfo.ID,
	}
	scan.RequestedBy = scanInfo.Trigger
	scan.Progress = scanInfo.Progress
	scan.CheckCount = scanInfo.CheckCount
	scan.ScheduledTime = scanInfo.ScheduledTime
	scan.StartTime = scanInfo.StartTime
	scan.EndTime = scanInfo.EndTime
	scan.Status = scanInfo.Status
	scan.ProgramID = program.ID
	scan.Program = program
	return scan
}

// decodeGlobalProgramRequest looks if a program id matches the pattern
// teamID@programID. If that is the case it returns the teamID and
// programID.
func decodeGlobalProgramRequest(id string) (teamID, programID string) {
	// Check if the call comes from the scheduler the programID will be have the form:
	// team_id@global_program
	if i := strings.Index(id, "@"); (i > 0) && (i < len(id)-1) {
		teamID = id[0:i]
		programID = id[i+1:]
		return
	}
	programID = id
	return
}

func encodeGlobalProgramRequestID(id, teamID string) string {
	return fmt.Sprintf("%s@%s", teamID, id)
}
