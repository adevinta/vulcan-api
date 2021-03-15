/*
Copyright 2021 Adevinta
*/

package global

import (
	"context"
	"fmt"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
	global "github.com/adevinta/vulcan-api/pkg/api/store/global"
	"github.com/adevinta/vulcan-api/pkg/scanengine"
	"github.com/adevinta/vulcan-api/pkg/schedule"
)

const (
	errEntityNotModifiable = "global entities can not be modified"
	errSchedulerNotFound   = "ScheduleNotFound"
)

// Middleware defines the shape of the functions that return a vulcanito service middleware
type Middleware func(api.VulcanitoService) api.VulcanitoService

// GlobalStore defines the functionality the GlobalEntitiesMiddleware
// needs to be provided by a GlobalEntities store.
type GlobalStore interface {
	Groups() map[string]global.Group
	Policies() map[string]global.Policy
	Programs() map[string]global.Program
	Reports() map[string]global.Report
}

// MetadataStore defines the functionality needed by the GlobalEntitiesMiddleware
// to store a retrive metadata about global entities.
type MetadataStore interface {
	FindGlobalProgramMetadata(programID string, teamID string) (*api.GlobalProgramsMetadata, error)
	UpsertGlobalProgramMetadata(teamID, program string, defaultAutosend bool, defaultCron string, autosend *bool, cron *string) error
	DeleteProgramMetadata(program string) error
}

type globalEntities struct {
	api.VulcanitoService
	store            GlobalStore
	metadata         MetadataStore
	logger           log.Logger
	scheduler        *globalScheduler
	scanEngineConfig scanengine.Config
	metricsClient    metrics.Client
}

// NewEntities returns a middleware to inject global entities functionality
// in the vulcanito service.
func NewEntities(l log.Logger, store GlobalStore, metadataStore MetadataStore,
	scanScheduler schedule.ScanScheduler, reportScheduler schedule.ReportScheduler,
	sconfig scanengine.Config, metricsClient metrics.Client) Middleware {

	return func(next api.VulcanitoService) api.VulcanitoService {
		gscheduler := &globalScheduler{
			scanScheduler,
			reportScheduler,
		}
		g := &globalEntities{
			store:            store,
			scheduler:        gscheduler,
			metadata:         metadataStore,
			logger:           l,
			VulcanitoService: next,
			scanEngineConfig: sconfig,
			metricsClient:    metricsClient,
		}

		// We ensure that all the teams have schedules for all the global
		// programs and default report in a goroutine because we don't want to
		// block initializing the api and don't want to make the api to be down
		// if the scheduler is down.
		go g.scheduleGlobalProgramDefaults()
		go g.scheduleGlobalReportDefaults()

		return g
	}
}

// globalScheduler is used to introduce specific logic to deal with the
// scheduler for global entities.
type globalScheduler struct {
	schedule.ScanScheduler
	schedule.ReportScheduler
}

func (e *globalEntities) scheduleGlobalProgramDefaults() {
	globalPrograms := e.store.Programs()
	for name, p := range globalPrograms {
		if p.DefaultMetadata.Cron == "" {
			continue
		}
		teams, err := e.VulcanitoService.ListTeams(context.Background())
		if err != nil {
			_ = e.logger.Log("error getting teams for ensuring global programs default schedules", err)
			return
		}

		schedulesPerTeam := make(map[string]string)
		for _, team := range teams {
			program, errProgram := e.FindProgram(context.Background(), name, team.ID)
			if errProgram != nil {
				continue
			}
			schedulesPerTeam[team.ID] = program.Cron
		}

		err = e.scheduler.EnsureGlobalProgramSchedule(name, schedulesPerTeam)
		if err != nil {
			_ = e.logger.Log("EnsureDefaultProgramsSchedulesError", err.Error())
		}
	}
}

func (e *globalEntities) scheduleGlobalReportDefaults() {
	globalReports := e.store.Reports()
	for _, r := range globalReports {
		if r.DefaultSchedule == "" {
			continue
		}
		teams, err := e.VulcanitoService.ListTeams(context.Background())
		if err != nil {
			_ = e.logger.Log("error getting teams for ensuring global report default schedules", err)
			return
		}
		teamIDs := []string{}
		for _, team := range teams {
			teamIDs = append(teamIDs, team.ID)
		}
		err = e.scheduler.EnsureGlobalReportSchedule(r.DefaultSchedule, teamIDs)
		if err != nil {
			_ = e.logger.Log("EnsureDefaultReportsSchedulesError", err.Error())
		}
	}
}

func (e *globalEntities) CreateTeam(ctx context.Context, team api.Team, ownerEmail string) (*api.Team, error) {
	teamCreated, err := e.VulcanitoService.CreateTeam(ctx, team, ownerEmail)
	if err != nil {
		return nil, err
	}
	globalPrograms := e.store.Programs()
	for name, p := range globalPrograms {
		if p.DefaultMetadata.Cron == "" {
			continue
		}
		// We do no return an error in this case because the team is already created properly
		// if for whatever reason this call does not work in any case the vulcan-api will ensure
		// this team will have the default schedule for all the global programs when it restarts.
		err = e.scheduler.EnsureGlobalProgramSchedule(name, map[string]string{teamCreated.ID: p.DefaultMetadata.Cron})
		if err != nil {
			_ = e.logger.Log("CreateTeamScheduleGlobalProgram", err.Error())
		}
	}
	globalReports := e.store.Reports()
	for _, r := range globalReports {
		if r.DefaultSchedule == "" {
			continue
		}
		// We do no return an error in this case because the team is already created properly
		// if for whatever reason this call does not work in any case the vulcan-api will ensure
		// this team will have the default schedule for all the global reports when it restarts.
		err = e.scheduler.EnsureGlobalReportSchedule(r.DefaultSchedule, []string{teamCreated.ID})
		if err != nil {
			_ = e.logger.Log("CreateTeamScheduleGlobalReport", err.Error())
		}
	}
	return teamCreated, nil
}

func (gc *globalScheduler) CreateSchedule(programID, teamID, cronExpr string) error {
	// We contatenate the program id with the team id to make id of the global
	// program unique for each team as needed by the scheduler.
	id := fmt.Sprintf("%s@%s", teamID, programID)
	return gc.ScanScheduler.CreateScanSchedule(id, teamID, cronExpr)
}

func (gc *globalScheduler) GetScheduleByID(teamID, programID string) (string, error) {
	// We contatenate the program id with the team id to make id of the global
	// program unique for each team as needed by the scheduler.
	id := fmt.Sprintf("%s@%s", teamID, programID)
	return gc.ScanScheduler.GetScanScheduleByID(id)
}

func (gc *globalScheduler) DeleteSchedule(teamID, programID string) error {
	// We contatenate the program id with the team id to make id of the global
	// program unique for each team as needed by the scheduler.
	id := fmt.Sprintf("%s@%s", teamID, programID)
	return gc.ScanScheduler.DeleteScanSchedule(id)
}

func (gc *globalScheduler) EnsureGlobalProgramSchedule(programID string, schedulesPerTeam map[string]string) error {
	schedules := []schedule.ScanBulkSchedule{}
	for teamID, cronStr := range schedulesPerTeam {
		id := fmt.Sprintf("%s@%s", teamID, programID)
		s := schedule.ScanBulkSchedule{
			Str:       cronStr,
			ProgramID: id,
			TeamID:    teamID,
			Overwrite: true,
		}
		schedules = append(schedules, s)
	}
	return gc.ScanScheduler.BulkCreateScanSchedules(schedules)
}

func (gc *globalScheduler) EnsureGlobalReportSchedule(cronStr string, teams []string) error {
	schedules := []schedule.ReportBulkSchedule{}
	for _, t := range teams {
		s := schedule.ReportBulkSchedule{
			Str:       cronStr,
			TeamID:    t,
			Overwrite: true,
		}
		schedules = append(schedules, s)
	}
	return gc.ReportScheduler.BulkCreateReportSchedules(schedules)
}
