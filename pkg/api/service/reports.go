/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/adevinta/vulcan-api/pkg/api"
)

func (s vulcanitoService) FindReport(ctx context.Context, scanID string) (*api.Report, error) {
	r, err := s.reportsClient.GetReport(ctx, scanID)
	if err != nil {
		return nil, err
	}

	return &api.Report{
		ID:             r.ID,
		ScanID:         r.ScanID,
		ProgramName:    r.ProgramName,
		Report:         r.ReportURL,
		ReportJson:     r.ReportJSONURL,
		EmailBody:      r.EmailBody,
		DeliveredTo:    r.DeliveredTo,
		UpdateStatusAt: &r.UpdatedAt,
		Status:         r.Status,
		Risk:           &r.Risk,
		CreatedAt:      &r.CreatedAt,
	}, nil
}

// GenerateReport triggers the generation of a new report
func (s vulcanitoService) GenerateReport(ctx context.Context, teamID, teamName, scanID string, autosend bool) error {
	programName := ""
	// Add program name if the scan has been launched via scan engine.
	scan, err := s.FindScan(ctx, scanID, teamID)
	if err == nil && scan.Program != nil {
		// As there's no error, the scan has been launched via the scan engine.
		programName = scan.Program.Name
	}

	return s.RunGenerateReport(ctx, autosend, scanID, programName, teamID, teamName)
}

func (s vulcanitoService) RunGenerateReport(ctx context.Context, autosend bool, scanID, programName, teamID, teamName string) error {
	// Gather recipient emails.
	recipients, err := s.ListRecipients(ctx, teamID)
	if err != nil {
		_ = s.logger.Log("ErrListRecipients", err)
		return err
	}
	var emails []string
	for _, r := range recipients {
		emails = append(emails, r.Email)
	}

	return s.reportsClient.GenerateReport(scanID, teamID, teamName, programName, emails, autosend)
}

func (s vulcanitoService) SendReport(ctx context.Context, scanID, teamID string) error {
	// Gather recipient emails.
	recipients, err := s.ListRecipients(ctx, teamID)
	if err != nil {
		_ = s.logger.Log("ErrListRecipients", err)
		return err
	}
	var emails []string
	for _, r := range recipients {
		emails = append(emails, r.Email)
	}

	return s.reportsClient.SendReportNotification(ctx, scanID, emails)
}

func (s vulcanitoService) SendDigestReport(ctx context.Context, teamID string, startDate string, endDate string) error {
	// Find the team.
	team, err := s.FindTeam(ctx, teamID)
	if err != nil {
		_ = s.logger.Log("ErrFindTeam", err)
		return err
	}

	// Gather recipient emails.
	recipients, err := s.ListRecipients(ctx, teamID)
	if err != nil {
		_ = s.logger.Log("ErrListRecipients", err)
		return err
	}

	var emails []string
	for _, r := range recipients {
		emails = append(emails, r.Email)
	}

	dateFromStr := startDate
	dateToStr := endDate

	// Default value for dateFrom is 7 days ago (one week).
	// Default value for dateTo is the current date.
	if dateFromStr == "" && dateToStr == "" {
		dateTo := time.Now()
		dateFrom := dateTo.Add(-(7 * 24 * time.Hour))

		dateToStr = dateTo.Format("2006-01-02")
		dateFromStr = dateFrom.Format("2006-01-02")
	}

	params := api.StatsParams{
		Tag: team.Tag,
	}

	if dateToStr != "" {
		params.AtDate = dateToStr
	}

	currentStats, err := s.vulndbClient.StatsOpen(ctx, params)
	if err != nil {
		_ = s.logger.Log("ErrStatsOpen", err)
		return err
	}

	diffStats, err := s.vulndbClient.StatsOpen(ctx, api.StatsParams{
		Tag:     team.Tag,
		MinDate: dateFromStr,
		MaxDate: dateToStr,
	})
	if err != nil {
		_ = s.logger.Log("ErrStatsOpen", err)
		return err
	}

	fixedStats, err := s.vulndbClient.StatsFixed(ctx, api.StatsParams{
		Tag:     team.Tag,
		MinDate: dateFromStr,
		MaxDate: dateToStr,
	})
	if err != nil {
		_ = s.logger.Log("ErrStatsFixed", err)
		return err
	}

	//s.vulndbClient.StatsOpen(ctx, params)
	severitiesStats := make(map[string]int)
	severitiesStats["info"] = currentStats.OpenIssues.Informational
	severitiesStats["low"] = currentStats.OpenIssues.Low
	severitiesStats["medium"] = currentStats.OpenIssues.Medium
	severitiesStats["high"] = currentStats.OpenIssues.High
	severitiesStats["critical"] = currentStats.OpenIssues.Critical

	severitiesStats["infoDiff"] = diffStats.OpenIssues.Informational
	severitiesStats["lowDiff"] = diffStats.OpenIssues.Low
	severitiesStats["mediumDiff"] = diffStats.OpenIssues.Medium
	severitiesStats["highDiff"] = diffStats.OpenIssues.High
	severitiesStats["criticalDiff"] = diffStats.OpenIssues.Critical

	severitiesStats["infoFixed"] = fixedStats.FixedIssues.Informational
	severitiesStats["lowFixed"] = fixedStats.FixedIssues.Low
	severitiesStats["mediumFixed"] = fixedStats.FixedIssues.Medium
	severitiesStats["highFixed"] = fixedStats.FixedIssues.High
	severitiesStats["criticalFixed"] = fixedStats.FixedIssues.Critical

	liveReportURL := fmt.Sprintf("%s/report/report.html?team_id=%s&minDate=%s&maxDate=%s", s.reportsConfig.VulcanUIURL, teamID, dateFromStr, dateToStr)

	return s.reportsClient.GenerateDigestReport(teamID, team.Name, dateFromStr, dateToStr, liveReportURL, emails, severitiesStats, true)
}
