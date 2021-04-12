/*
Copyright 2021 Adevinta
*/

package reports

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/adevinta/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

const (
	scanType       = "scan"
	livereportType = "livereport"

	endpointFmt            = "%s%s"
	apiBasePath            = "/api/v1"
	getReportPathFmt       = "/reports/scan/%s"
	getReportNotifPathFmt  = "/reports/scan/%s/notification"
	sendReportNotifPathFmt = "/reports/scan/%s/send"
)

// Client represents a client to interact with
// reports generator microservice.
type Client interface {
	GetReport(ctx context.Context, scanID string) (*ScanReport, error)
	GetReportNotification(ctx context.Context, scanID string) (*Notification, error)
	SendReportNotification(ctx context.Context, scanID string, recipients []string) error
	GenerateReport(scanID, teamID, teamName, programName string, recipients []string, autoSend bool) error
	GenerateDigestReport(teamID, teamName, dateFrom, dateTo, liveReportURL string,
		recipients []string, severitiesStats map[string]int, autoSend bool) error
}

type client struct {
	cfg        Config
	httpClient *http.Client
	snsAPI     snsiface.SNSAPI
}

// NewClient builds a new reports client.
func NewClient(cfg Config) (Client, error) {
	arn, err := arn.Parse(cfg.SNSARN)
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	awsCfg := aws.NewConfig()
	if arn.Region != "" {
		awsCfg = awsCfg.WithRegion(arn.Region)
	}
	if cfg.SNSEndpoint != "" {
		awsCfg = awsCfg.WithEndpoint(cfg.SNSEndpoint)
	}
	// Build http client.
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.InsecureTLS, // nolint
			},
		},
	}

	return &client{
		cfg:        cfg,
		httpClient: httpClient,
		snsAPI:     sns.New(sess, awsCfg),
	}, nil
}

// GetReport performs an HTTP GET request to reports generator API to retrieve
// the report for the specified scan ID.
func (c *client) GetReport(ctx context.Context, scanID string) (*ScanReport, error) {
	getReportPath := fmt.Sprintf(getReportPathFmt, scanID)
	resp, err := c.performRequest(ctx, http.MethodGet, getReportPath, nil)
	if err != nil {
		return nil, err
	}

	var scanReport ScanReport
	err = json.Unmarshal(resp, &scanReport)
	return &scanReport, err
}

// GetReportNotification performs an HTTP GET request to reports generator API to retrieve
// the report notification for the specified scan ID.
func (c *client) GetReportNotification(ctx context.Context, scanID string) (*Notification, error) {
	getReportNotifPath := fmt.Sprintf(getReportNotifPathFmt, scanID)
	resp, err := c.performRequest(ctx, http.MethodGet, getReportNotifPath, nil)
	if err != nil {
		return nil, err
	}

	var notif Notification
	err = json.Unmarshal(resp, &notif)
	return &notif, err
}

// SendReportNotification performs an HTTP POST request to reports generator API to trigger
// the sending of the report notification for the specified scan ID.
func (c *client) SendReportNotification(ctx context.Context, scanID string, recipients []string) error {
	sendReportNotifPath := fmt.Sprintf(sendReportNotifPathFmt, scanID)
	jsonPayload, err := json.Marshal(struct {
		Recipients []string `json:"recipients"`
	}{recipients})
	if err != nil {
		return err
	}
	body := bytes.NewReader(jsonPayload)
	_, err = c.performRequest(ctx, http.MethodPost, sendReportNotifPath, body)
	return err
}

// GenerateReport pushes an SNS event to trigger the report generation for the specified scanID.
func (c *client) GenerateReport(scanID, teamID, teamName, programName string, recipients []string, autoSend bool) error {
	event := genReportEvent{
		Typ: scanType,
		TeamInfo: teamInfo{
			ID:         teamID,
			Name:       teamName,
			Recipients: recipients,
		},
		Data: scanData{
			ScanID:      scanID,
			ProgramName: programName,
		},
		AutoSend: autoSend,
	}

	return c.publish(event)
}

// GenerateReport pushes an SNS event to trigger the digest report generation for the specified teamID.
func (c *client) GenerateDigestReport(teamID, teamName, dateFrom, dateTo, liveReportURL string,
	recipients []string, severitiesStats map[string]int, autoSend bool) error {
	event := genDigestReportEvent{
		Typ: livereportType,
		TeamInfo: teamInfo{
			ID:         teamID,
			Name:       teamName,
			Recipients: recipients,
		},
		Data: digestReportData{
			TeamID:        teamID,
			DateFrom:      dateFrom,
			DateTo:        dateTo,
			LiveReportURL: liveReportURL,
			Info:          severitiesStats["info"],
			InfoDiff:      severitiesStats["infoDiff"],
			InfoFixed:     severitiesStats["infoFixed"],
			Low:           severitiesStats["low"],
			LowDiff:       severitiesStats["lowDiff"],
			LowFixed:      severitiesStats["lowFixed"],
			Medium:        severitiesStats["medium"],
			MediumDiff:    severitiesStats["mediumDiff"],
			MediumFixed:   severitiesStats["mediumFixed"],
			High:          severitiesStats["high"],
			HighDiff:      severitiesStats["highDiff"],
			HighFixed:     severitiesStats["highFixed"],
			Critical:      severitiesStats["critical"],
			CriticalDiff:  severitiesStats["criticalDiff"],
			CriticalFixed: severitiesStats["criticalFixed"],
		},
		AutoSend: autoSend,
	}

	return c.publish(event)
}

func (c *client) publish(event interface{}) error {
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = c.snsAPI.Publish(&sns.PublishInput{
		Message:  aws.String(string(eventPayload)),
		TopicArn: aws.String(c.cfg.SNSARN),
	})

	return err
}

func (c *client) performRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	u, err := url.Parse(c.cfg.APIBaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("%s%s", apiBasePath, path)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, parseErr(resp.StatusCode, string(content))
	}

	return ioutil.ReadAll(resp.Body)
}

func parseErr(statusCode int, mssg string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.Assertion(mssg)
	case http.StatusUnauthorized:
		return errors.Unauthorized(mssg)
	case http.StatusForbidden:
		return errors.Forbidden(mssg)
	case http.StatusNotFound:
		return errors.NotFound(mssg)
	case http.StatusMethodNotAllowed:
		return errors.MethodNotAllowed(mssg)
	case http.StatusUnprocessableEntity:
		return errors.Validation(mssg)
	default:
		return errors.Default(mssg)
	}
}
