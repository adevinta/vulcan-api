/*
Copyright 2021 Adevinta
*/

package reports

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

const (
	livereportType = "livereport"
)

type Client struct {
	cfg        Config
	httpClient *http.Client
	snsAPI     snsiface.SNSAPI
}

// NewClient builds a new reports client.
func NewClient(cfg Config) (*Client, error) {
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

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
		snsAPI:     sns.New(sess, awsCfg),
	}, nil
}

// GenerateDigestReport pushes an SNS event to trigger the digest report
// generation for the specified teamID.
func (c *Client) GenerateDigestReport(teamID, teamName, dateFrom, dateTo, liveReportURL string,
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

func (c *Client) publish(event interface{}) error {
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
