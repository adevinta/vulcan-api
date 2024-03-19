/*
Copyright 2021 Adevinta
*/

package schedule

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/robfig/cron"
)

var (
	// ErrScheduleNotFound is returned when crontinuos answers with 404 status.
	ErrScheduleNotFound = errors.New("ScheduleNotFound")
	// ErrInvalidSchedulePeriod is returned when the schedule period is less than the minimum allowed.
	ErrInvalidSchedulePeriod = errors.New("Schedule program period less than the minimun allowed")
	// ErrInvalidCronExpr is returned when the given cron expression is not valid.
	ErrInvalidCronExpr = errors.New("Invalid Cron Expression")
)

func errorCreatingSchedule(code int, msg string) error {
	return fmt.Errorf("ErrorCreatingSchedule Status: %d, msg: %s", code, msg)
}

func errorBulkCreatingSchedules(code int, msg string) error {
	return fmt.Errorf("ErrorCreatingSchedule Status: %d, msg: %s", code, msg)
}

func errorGettingSchedule(code int, msg string) error {
	return fmt.Errorf("ErrorGettingSchedule Status: %d, msg: %s", code, msg)
}

func errorDeletingSchedule(code int, msg string) error {
	return fmt.Errorf("ErrorDeletingSchedule Status: %d, msg: %s", code, msg)
}

const (
	bulkCreateScanSchedulePath = "entries"
	createScanSchedulePath     = "settings"
	getScanScheduleByIDPath    = "entries"

	bulkCreateReportSchedulePath = "report/entries"
	createReportSchedulePath     = "report/settings"
	getReportScheduleByIDPath    = "report/entries"

	jsonContentType = "application/json"
)

// Config holds the configuration needed by the schuduler client.
type Config struct {
	URL             string  `mapstructure:"url"`
	MinimumInterval float64 `mapstructure:"minimum_interval"`
}

type createScheduleRequest struct {
	Str string `json:"str"`
}

type getScanScheduleByIDResponse struct {
	ID       string `json:"program_id"`
	CronSpec string `json:"cron_spec"`
}

type getReportScheduleByIDResponse struct {
	ID       string `json:"team_id"`
	CronSpec string `json:"cron_spec"`
}

// ScanBulkSchedule defines the information needed to create
// a schedule in a bulk create operation.
type ScanBulkSchedule struct {
	Str       string `json:"str"`
	ProgramID string `json:"program_id"`
	TeamID    string `json:"team_id"`
	Overwrite bool   `json:"overwrite"`
}

// ReportBulkSchedule defines the information needed to create
// a report schedule in a bulk create operation.
type ReportBulkSchedule struct {
	Str       string `json:"str"`
	TeamID    string `json:"team_id"`
	Overwrite bool   `json:"overwrite"`
}

// Client provides functionalities for calling vulcan-scheduler.
type Client struct {
	baseURL   string
	minPeriod float64
	c         *http.Client
}

// NewClient creates a new vulcan-cron client.
func NewClient(cfg Config) *Client {
	c := Client{
		baseURL:   cfg.URL,
		c:         http.DefaultClient,
		minPeriod: cfg.MinimumInterval,
	}
	return &c
}

// DeleteScanSchedule executes a request against the scheduler component for deleting a scan schedule.
func (c *Client) DeleteScanSchedule(programID string) error {
	path := path.Join(getScanScheduleByIDPath, programID)

	status, body, err := c.performRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if status == http.StatusNotFound {
		return ErrScheduleNotFound
	}
	if status != http.StatusOK {
		return errorDeletingSchedule(status, string(body))
	}

	return nil
}

// DeleteReportSchedule executes a request against the scheduler component for deleting a report schedule.
func (c *Client) DeleteReportSchedule(teamID string) error {
	path := path.Join(getScanScheduleByIDPath, teamID)

	status, body, err := c.performRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if status == http.StatusNotFound {
		return ErrScheduleNotFound
	}
	if status != http.StatusOK {
		return errorDeletingSchedule(status, string(body))
	}

	return nil
}

// GetScanScheduleByID gets the cron string defining the scan schedule for a given id.
// If the scheduler doesn't have a schedule defined for the
// given id the func will return empty cron string.
func (c *Client) GetScanScheduleByID(programID string) (string, error) {
	path := path.Join(getScanScheduleByIDPath, programID)

	status, body, err := c.performRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	if status == http.StatusNotFound {
		return "", ErrScheduleNotFound
	}
	if status != http.StatusOK {
		return "", errorGettingSchedule(status, string(body))
	}

	r := getScanScheduleByIDResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}
	return r.CronSpec, nil
}

// GetReportScheduleByID gets the cron string defining the report schedule for a given id.
// If the scheduler doesn't have a schedule defined for the
// given id the func will return empty cron string.
func (c *Client) GetReportScheduleByID(teamID string) (string, error) {
	path := path.Join(getScanScheduleByIDPath, teamID)

	status, body, err := c.performRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	if status == http.StatusNotFound {
		return "", ErrScheduleNotFound
	}
	if status != http.StatusOK {
		return "", errorGettingSchedule(status, string(body))
	}

	r := getReportScheduleByIDResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", err
	}
	return r.CronSpec, nil
}

// BulkCreateScanSchedules creates scan schedules in bulk. It only creates a schedule for
// a program if no schedule for thar program already exist.
func (c *Client) BulkCreateScanSchedules(schedules []ScanBulkSchedule) error {
	for _, s := range schedules {
		err := scheduleAllowed(s.Str, c.minPeriod)
		if err != nil {
			return err
		}
	}

	status, body, err := c.performRequest(http.MethodPost, bulkCreateScanSchedulePath, schedules)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return errorBulkCreatingSchedules(status, string(body))
	}

	return nil
}

// BulkCreateReportSchedules creates report schedules in bulk. It only creates a schedule for
// a program if no schedule for thar program already exist.
func (c *Client) BulkCreateReportSchedules(schedules []ReportBulkSchedule) error {
	for _, s := range schedules {
		err := scheduleAllowed(s.Str, c.minPeriod)
		if err != nil {
			return err
		}
	}

	status, body, err := c.performRequest(http.MethodPost, bulkCreateReportSchedulePath, schedules)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return errorBulkCreatingSchedules(status, string(body))
	}

	return nil
}

// CreateScanSchedule creates a new scan schedule for executing a program.
func (c *Client) CreateScanSchedule(programID, teamID, cronExpr string) error {
	// Verify the cron str
	err := scheduleAllowed(cronExpr, c.minPeriod)
	if err != nil {
		return err
	}

	req := createScheduleRequest{
		Str: cronExpr,
	}
	path := path.Join(createScanSchedulePath, programID, teamID)

	status, body, err := c.performRequest(http.MethodPost, path, req)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return errorCreatingSchedule(status, string(body))
	}

	return nil
}

// CreateReportSchedule creates a new report schedule for executing a program.
func (c *Client) CreateReportSchedule(teamID, cronExpr string) error {
	// Verify the cron str
	err := scheduleAllowed(cronExpr, c.minPeriod)
	if err != nil {
		return err
	}

	req := createScheduleRequest{
		Str: cronExpr,
	}
	path := path.Join(createReportSchedulePath, teamID)

	status, body, err := c.performRequest(http.MethodPost, path, req)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return errorCreatingSchedule(status, string(body))
	}

	return nil
}

func (c *Client) performRequest(httpMethod, path string, payload interface{}) (int, []byte, error) {
	status := http.StatusInternalServerError

	content, err := json.Marshal(payload)
	if err != nil {
		return status, nil, err
	}
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(content))
	if err != nil {
		return status, nil, err
	}
	if httpMethod == http.MethodPost {
		req.Header.Set("Content-Type", jsonContentType)
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return status, nil, err
	}
	defer resp.Body.Close() //nolint

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return status, nil, err
	}
	return resp.StatusCode, body, nil
}

func scheduleAllowed(cronExpr string, min float64) error {
	s, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return ErrInvalidCronExpr
	}
	t1 := s.Next(time.Now())
	t2 := s.Next(t1)
	d := t2.Sub(t1)
	p := d.Minutes()
	if p < min {
		return ErrInvalidSchedulePeriod
	}
	return nil
}
