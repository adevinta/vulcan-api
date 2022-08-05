/*
Copyright 2021 Adevinta
*/

package cdc

import (
	"context"
	"encoding/json"
	errs "errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/adevinta/errors"
	"github.com/adevinta/vulcan-api/pkg/api"
	vulndb "github.com/adevinta/vulcan-api/pkg/vulnerabilitydb"
)

const (
	// supported operations
	opDeleteTeam            = "DeleteTeam"
	opCreateAsset           = "CreateAsset"
	opDeleteAsset           = "DeleteAsset"
	opUpdateAsset           = "UpdateAsset"
	opDeleteAllAssets       = "DeleteAllAssets"
	opFindingOverwrite      = "FindingOverwrite"
	opMergeDiscoveredAssets = "MergeDiscoveredAssets"
)

var (
	errInvalidData          = errs.New("invalid data")
	errUnsupportedAction    = errs.New("unsupported action")
	errTargetNotUnique      = errs.New("target is not unique")
	errUnavailabeJobsRunner = errs.New("unavailable jobs runner")
)

// Parser defines a CDC log parser.
type Parser interface {
	// Parse should parse the log events secuentially from the beginning
	// of the slice and return the number of events that have been processed
	// correctly. So if one event processing is errored, parser should stop
	// processing and return current parsed events count.
	Parse(log []Event) (nParsed uint)
}

// AsyncTxParser implements a CDC log parser to handle distributed transactions
// for VulnDB and other API asynchronous jobs.
type AsyncTxParser struct {
	VulnDBClient vulndb.Client
	JobsRunner   *api.JobsRunner
	logger       log.Logger
}

// NewAsyncTxParser builds a new CDC log parser to handle distributed
// transactions for VulnDB and other API asynchronous jobs.
func NewAsyncTxParser(vulnDBClient vulndb.Client, jobsRunner *api.JobsRunner, logger log.Logger) *AsyncTxParser {
	return &AsyncTxParser{
		VulnDBClient: vulnDBClient,
		JobsRunner:   jobsRunner,
		logger:       logger,
	}
}

// Parse parses the log secuentially processing each event based on its action
// and returns the number of events that have been processed correctly.
// If an error happens during processing of one event, and it is not a permanent
// error, log processing is stopped.
// If a permanent error happens during processing of one event or event has reached
// max processing attempts, that event is discarded counting as if it was processed.
func (p *AsyncTxParser) Parse(log []Event) (nParsed uint) {
	var processFunc func([]byte) error

	for _, event := range log {
		switch event.Action() {
		case opDeleteTeam:
			processFunc = p.processDeleteTeam
		case opCreateAsset:
			processFunc = p.processCreateAsset
		case opDeleteAsset:
			processFunc = p.processDeleteAsset
		case opUpdateAsset:
			processFunc = p.processUpdateAsset
		case opDeleteAllAssets:
			processFunc = p.processDeleteAllAssets
		case opFindingOverwrite:
			processFunc = p.processFindingOverwrite
		case opMergeDiscoveredAssets:
			processFunc = p.processMergeDiscoveredAssets
		default:
			// If action is not supported
			// log err and stop processing
			p.logErr(event, errUnsupportedAction)
			return
		}

		// Process Event
		err := processFunc(event.Data())
		if err != nil {
			// If processing is errored
			// log err and stop processing
			p.logErr(event, err)
			return
		}

		nParsed++
	}

	return
}

func (p *AsyncTxParser) processDeleteTeam(data []byte) error {
	var dto OpDeleteTeamDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	err = p.VulnDBClient.DeleteTeam(context.Background(), dto.Team.ID, dto.Team.ID)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}

	// TODO: Get the TeamID from the Vulcan API of the teams with the same tag.

	var excludedTeams []string
	err = p.VulnDBClient.DeleteTag(context.Background(), dto.Team.ID, dto.Team.Tag, api.TargetsParams{ExceptTeams: excludedTeams})
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (p *AsyncTxParser) processCreateAsset(data []byte) error {
	var dto OpCreateAssetDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	payload := api.CreateTarget{
		Identifier: dto.Asset.Identifier,
		Tags:       []string{dto.Asset.Team.Tag},
		Teams:      []string{dto.Asset.Team.ID},
	}

	_, err = p.VulnDBClient.CreateTarget(context.Background(), payload)
	return err
}

func (p *AsyncTxParser) processDeleteAsset(data []byte) error {
	var dto OpDeleteAssetDTO
	ctx := context.Background()

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}
	// By now, we don't process the assets deleted in a "delete all assets"
	// operation.
	if dto.DeleteAllAssetsOp {
		return nil
	}
	if dto.DupAssets > 0 {
		// If there are more assets with the same
		// identifier and for the same team, do not
		// execute tag deletion from VulnDB.
		return nil
	}

	// Retrieve target from VulnDB so we
	// can use its ID in delete tag request

	params := api.TargetsParams{
		Identifier: dto.Asset.Identifier,
		// Set match to true so we look for
		// targets matching identifier completely
		IdentifierMatch: true,
	}

	ttList, err := p.VulnDBClient.Targets(ctx, params, api.Pagination{})
	if err != nil {
		return err
	}

	if len(ttList.Targets) > 1 {
		// This should never happen
		// with current VunlnDB schema
		return errTargetNotUnique
	}
	if len(ttList.Targets) == 0 {
		// If target is not present in
		// VulnDB, nothing to do
		return nil
	}

	target := ttList.Targets[0]
	teamID := dto.Asset.Team.ID
	tag := dto.Asset.Team.Tag

	err = p.VulnDBClient.DeleteTargetTeam(ctx, teamID, target.ID, teamID)
	if err != nil {
		// If target is not found or if we get a 403 HTTP response status,
		// which means that the team is no longer associated with the target,
		// there is nothing to do, so return no error.
		if errors.IsKind(err, errors.ErrNotFound) || errors.IsKind(err, errors.ErrForbidden) {
			return nil
		}
		return err
	}

	// TODO: Get the TeamID from the Vulcan API of the teams with the same tag.

	var excludedTeams []string
	err = p.VulnDBClient.DeleteTargetTag(ctx, teamID, target.ID, tag, api.TargetsParams{ExceptTeams: excludedTeams})
	if err != nil {
		// If target is not found there is nothing to do, so return no error.
		// In this case, 403 HTTP response status is only returned on failed authorization.
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}

	return nil
}

func (p *AsyncTxParser) processUpdateAsset(data []byte) error {
	// An asset update where identifier has changed can imply 2 operations in VulnDB:
	// - A delete of the asset association wih the team if team has no duplicates
	//   for the same identifier.
	// - A creation of the target with the new identifier.

	var dto OpUpdateAssetDTO
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	// Process asset deletion
	delDTO := OpDeleteAssetDTO{
		Asset:     dto.OldAsset,
		DupAssets: dto.DupAssets,
	}
	delJSON, err := json.Marshal(delDTO)
	if err != nil {
		return errInvalidData
	}
	err = p.processDeleteAsset(delJSON)
	if err != nil {
		return err
	}

	// Process asset creation
	createDTO := OpCreateAssetDTO{
		Asset: dto.NewAsset,
	}
	createJSON, err := json.Marshal(createDTO)
	if err != nil {
		return errInvalidData
	}
	return p.processCreateAsset(createJSON)
}

func (p *AsyncTxParser) processDeleteAllAssets(data []byte) error {
	var dto OpDeleteAllAssetsDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	err = p.VulnDBClient.DeleteTeam(context.Background(), dto.Team.ID, dto.Team.ID)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}

	err = p.VulnDBClient.DeleteTeamTag(context.Background(), dto.Team.ID, dto.Team.ID, dto.Team.Tag)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (p *AsyncTxParser) processFindingOverwrite(data []byte) error {
	var dto OpFindingOverwriteDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	_, err = p.VulnDBClient.UpdateFinding(
		context.Background(),
		dto.FindingOverwrite.FindingID,
		&api.UpdateFinding{
			Status: &dto.FindingOverwrite.Status,
		},
		dto.FindingOverwrite.TeamID)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// processMergeDiscoveredAssets performs the following actions:
// - Marks the Job as RUNNING
// - Calls the MergeDiscoveredAssets operation
// - Marks the Job as DONE
// In the case that the MergeDiscoveredAssets operation fails, the error is
// added to the JobResult.
// Errors are not returned from the function to avoid this operation to be
// retried. They are logged instead.
// NOTE: this operation could be time consuming, and therefore affect
// consistency introduced by latency executing the other distributed
// transactions.
func (p *AsyncTxParser) processMergeDiscoveredAssets(data []byte) error {
	var dto OpMergeDiscoveredAssetsDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		_ = level.Error(p.logger).Log(
			"component", CDCLogTag, "error", err, "action", opMergeDiscoveredAssets,
		)
		return nil
	}

	if p.JobsRunner == nil || p.JobsRunner.Client == nil {
		_ = level.Error(p.logger).Log(
			"component", CDCLogTag, "error", errUnavailabeJobsRunner, "action", opMergeDiscoveredAssets,
		)
		return nil
	}

	// Set the status of the Job to RUNNING so the user can track its progress.
	job := api.Job{
		ID:        dto.JobID,
		Status:    api.JobStatusRunning,
		Operation: opMergeDiscoveredAssets,
	}
	if err := p.updateJob(job); err != nil {
		return nil
	}

	// Execute the merge of the discovered assets.
	if err := p.JobsRunner.Client.MergeDiscoveredAssets(context.Background(), dto.TeamID, dto.Assets, dto.GroupName); err != nil {
		_ = level.Error(p.logger).Log(
			"component", CDCLogTag, "error", err, "job_id", dto.JobID, "action", opMergeDiscoveredAssets,
		)
		job.Result = &api.JobResult{
			Error: err.Error(),
		}
	}

	// Mark the job as DONE.
	job.Status = api.JobStatusDone
	_, err = p.JobsRunner.Client.UpdateJob(context.Background(), job)
	if err := p.updateJob(job); err != nil {
		return nil
	}

	return nil
}

func (p *AsyncTxParser) updateJob(job api.Job) error {
	_, err := p.JobsRunner.Client.UpdateJob(context.Background(), job)
	if err != nil {
		_ = level.Error(p.logger).Log(
			"component", CDCLogTag, "error", err, "job_id", job.ID, "action", opMergeDiscoveredAssets,
		)
	}
	return err
}

func (p *AsyncTxParser) logErr(e Event, err error) {
	_ = level.Error(p.logger).Log(
		"component", CDCLogTag, "error", err, "id", e.ID(), "action", e.Action(), "retries", e.ReadCount()+1,
	)
}
