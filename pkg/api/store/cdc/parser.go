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

// VulnDBAndJobTxParser implements a CDC log parser
// to handle distributed transactions for VulnDB.
type VulnDBAndJobTxParser struct {
	VulnDBClient vulndb.Client
	JobsRunner   api.JobsRunner
	logger       log.Logger
}

// NewVulnDBAndJobTxParser builds a new CDC log parser
// to handle distributed transactions for VulnDB.
func NewVulnDBAndJobTxParser(vulnDBClient vulndb.Client, jobsRunner api.JobsRunner, logger log.Logger) *VulnDBAndJobTxParser {
	return &VulnDBAndJobTxParser{
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
func (p *VulnDBAndJobTxParser) Parse(log []Event) (nParsed uint) {
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

func (p *VulnDBAndJobTxParser) processDeleteTeam(data []byte) error {
	var dto OpDeleteTeamDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	err = p.VulnDBClient.DeleteTag(context.Background(), dto.Team.Tag, dto.Team.Tag)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (p *VulnDBAndJobTxParser) processCreateAsset(data []byte) error {
	var dto OpCreateAssetDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	payload := api.CreateTarget{
		Identifier: dto.Asset.Identifier,
		Tags:       []string{dto.Asset.Team.Tag},
	}

	_, err = p.VulnDBClient.CreateTarget(context.Background(), payload)
	return err
}

func (p *VulnDBAndJobTxParser) processDeleteAsset(data []byte) error {
	var dto OpDeleteAssetDTO
	ctx := context.Background()

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
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
	tag := dto.Asset.Team.Tag

	err = p.VulnDBClient.DeleteTargetTag(ctx, tag, target.ID, tag)
	if err != nil {
		// If target is not found or if we get a 403 HTTP response status,
		// which means that the tag is no longer associated with the target,
		// there is nothing to do, so return no error.
		if errors.IsKind(err, errors.ErrNotFound) || errors.IsKind(err, errors.ErrForbidden) {
			return nil
		}
		return err
	}
	return nil
}

func (p *VulnDBAndJobTxParser) processUpdateAsset(data []byte) error {
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

func (p *VulnDBAndJobTxParser) processDeleteAllAssets(data []byte) error {
	var dto OpDeleteAllAssetsDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	err = p.VulnDBClient.DeleteTag(context.Background(), dto.Team.Tag, dto.Team.Tag)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (p *VulnDBAndJobTxParser) processFindingOverwrite(data []byte) error {
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
		dto.FindingOverwrite.Tag)
	if err != nil {
		if errors.IsKind(err, errors.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (p *VulnDBAndJobTxParser) processMergeDiscoveredAssets(data []byte) error {
	var dto OpMergeDiscoveredAssetsDTO

	err := json.Unmarshal(data, &dto)
	if err != nil {
		return errInvalidData
	}

	if p.JobsRunner.Client == nil {
		return errUnavailabeJobsRunner
	}

	// TODO: Who should update the Job???

	return p.JobsRunner.Client.MergeDiscoveredAssets(context.Background(), dto.TeamID, dto.Assets, dto.GroupName)
}

func (p *VulnDBAndJobTxParser) logErr(e Event, err error) {
	_ = level.Error(p.logger).Log(
		"component", CDCLogTag, "error", err, "id", e.ID(), "action", e.Action(), "retries", e.ReadCount()+1,
	)
}
