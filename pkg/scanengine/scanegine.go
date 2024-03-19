/*
Copyright 2021 Adevinta
*/

package scanengine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/adevinta/vulcan-api/pkg/api"
	scanengineAPI "github.com/adevinta/vulcan-scan-engine/pkg/api"
	scanengine "github.com/adevinta/vulcan-scan-engine/pkg/api/endpoint"
)

var (
	ScanEngineErrorKind = "Error calling scan engine"

	ErrProgramWithoutPolicyGroups = errors.New("Program has no policy groups")
	ErrNotFound                   = errors.New("Not found")
	ErrCreatingScan               = errors.New("Error creating scan")
	ErrGettingScan                = errors.New("Error getting scan")
	ErrGettingScans               = errors.New("Error getting scans")
	ErrAbortingScans              = errors.New("Error aborting scans")
	ErrUnprocessableEntity        = errors.New("UnprocessableEntity")
)

type GenericError struct {
	Code int
	Msg  string
}

func (g GenericError) Error() string {
	return g.Msg
}

type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
	ctx        context.Context
}
type Config struct {
	Url string `mapstructure:"url"`
}

// NewClient returns a new scan engine client.
func NewClient(ctx context.Context, httpClient *http.Client, config Config) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(config.Url)
	return &Client{
		httpClient: httpClient,
		BaseURL:    baseURL,
		ctx:        ctx,
	}
}

// CreateScanRequest creates a scan by calling the scan engine using the information in the parameters.
func (c *Client) CreateScanRequest(program api.Program, scheduledTime *time.Time, externalID, requestedBy, tag string) (scanengine.ScanRequest, error) {
	scanRequest := scanengine.ScanRequest{}
	scanRequest.Trigger = requestedBy
	if len(program.ProgramsGroupsPolicies) < 1 {
		errMssg := fmt.Sprintf("no PoliciesGroups defined in the current program %v ", program)
		err := fmt.Errorf("%w: %v", ErrProgramWithoutPolicyGroups, GenericError{Msg: errMssg})
		return scanengine.ScanRequest{}, err // nolint to avoid complaining about the standard errors package usage.
	}
	targetGroups := []scanengineAPI.TargetsChecktypesGroup{}
	for _, tg := range program.ProgramsGroupsPolicies {
		scTargetGroup := scanengineAPI.TargetGroup{
			Name:    tg.Group.Name,
			Options: tg.Group.Options,
		}

		scannableAssetGroups := []*api.AssetGroup{}
		for _, ag := range tg.Group.AssetGroup {
			// append only if the asset group is scannable
			if ag.Asset != nil && *ag.Asset.Scannable {
				scannableAssetGroups = append(scannableAssetGroups, ag)
			}
		}
		tg.Group.AssetGroup = scannableAssetGroups

		// Get the targets for the group. If no assets are defined in the group
		// we just skip the the entire target group as no checks are going to be
		// generated.
		if len(tg.Group.AssetGroup) < 1 {
			continue
		}
		targets, err := targetsFromAssetGroups(tg.Group.AssetGroup)
		if err != nil {
			return scanengine.ScanRequest{}, err
		}
		scTargetGroup.Targets = targets

		// Get the checktypes from the policy.
		scCheckTypesGroup := scanengineAPI.ChecktypesGroup{Name: tg.Policy.Name}
		scCheckTypesGroup.Checktypes = checktypesFromChecktypesSettings(tg.Policy.ChecktypeSettings)

		targetGroups = append(targetGroups, scanengineAPI.TargetsChecktypesGroup{
			ChecktypesGroup: scCheckTypesGroup,
			TargetGroup:     scTargetGroup,
		})
	}
	scanRequest.ScheduledTime = scheduledTime
	scanRequest.TargetGroups = targetGroups
	scanRequest.ExternalID = externalID
	scanRequest.Tag = tag
	return scanRequest, nil
}

func checktypesFromChecktypesSettings(settings []*api.ChecktypeSetting) []scanengineAPI.Checktype {
	checktypes := []scanengineAPI.Checktype{}
	for _, checktypesetting := range settings {
		setting := *checktypesetting
		checktype := scanengineAPI.Checktype{}
		checktype.Name = setting.CheckTypeName
		checktype.Options = ptrStrToStr(setting.Options)
		checktypes = append(checktypes, checktype)
	}
	return checktypes
}

func targetsFromAssetGroups(assetGroups []*api.AssetGroup) ([]scanengineAPI.Target, error) {
	targets := []scanengineAPI.Target{}
	for _, asset := range assetGroups {
		if asset.Asset == nil {
			errMssg := fmt.Sprintf("asset.Asset is nil for asset with ID: %v", asset.AssetID)
			err := fmt.Errorf("%w: %v", ErrNotFound, GenericError{Msg: errMssg})
			return nil, err
		}
		a := *asset.Asset
		target := scanengineAPI.Target{
			Identifier: a.Identifier,
			Type:       a.AssetType.Name,
		}
		if asset.Asset.Options != nil {
			target.Options = *a.Options
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// Create creates a scan in the scan engine.
func (c *Client) Create(request scanengine.ScanRequest) (*scanengine.ScanResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.BaseURL.String()+"scans", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	v := c.ctx.Value(kithttp.ContextKeyRequestXRequestID)
	if v != nil {
		xReqID, ok := v.(string)
		if ok {
			req.Header.Set("X-Request-Id", xReqID)
		}
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusCreated {
		// Try to read the response for a possible error message
		b, errR := io.ReadAll(response.Body)
		if errR != nil {
			return nil, errR
		}

		var err error
		err = GenericError{
			Msg:  string(b),
			Code: response.StatusCode,
		}
		if response.StatusCode == http.StatusUnprocessableEntity {
			err = fmt.Errorf("%w: %v", ErrUnprocessableEntity, err)
		}

		err = fmt.Errorf("%v: %w", ErrCreatingScan, err)
		return nil, fmt.Errorf("%w", err)
	}
	scanResponse := &scanengine.ScanResponse{}
	err = json.Unmarshal(responseBytes, scanResponse)
	if err != nil {
		return nil, err
	}
	return scanResponse, nil
}

func (c *Client) Get(scanID string) (*scanengine.GetScanResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL.String()+fmt.Sprintf("scans/%v", scanID), nil)
	if err != nil {
		return nil, err
	}

	v := c.ctx.Value(kithttp.ContextKeyRequestXRequestID)
	if v != nil {
		xReqID, ok := v.(string)
		if ok {
			req.Header.Set("X-Request-Id", xReqID)
		}
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		// Try to read the response for a possible error message
		b, errR := io.ReadAll(response.Body)
		if errR != nil {
			return nil, errR
		}

		errMssg := string(b)
		errCode := response.StatusCode
		err := fmt.Errorf("%w: %v", ErrGettingScan, GenericError{Code: errCode, Msg: errMssg})
		return nil, err
	}

	getScanResponse := &scanengine.GetScanResponse{}
	err = json.Unmarshal(responseBytes, getScanResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling")
	}
	return getScanResponse, nil
}

// GetScans returns all the scans belonging to a team.
func (c *Client) GetScans(externalID string) (*scanengine.GetScansResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL.String()+fmt.Sprintf("scans?external_id=%v&all=true", externalID), nil)
	if err != nil {
		return nil, err
	}

	v := c.ctx.Value(kithttp.ContextKeyRequestXRequestID)
	if v != nil {
		xReqID, ok := v.(string)
		if ok {
			req.Header.Set("X-Request-Id", xReqID)
		}
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		// Try to read the response for a possible error message
		b, errR := io.ReadAll(response.Body)
		if errR != nil {
			return nil, errR
		}

		errMssg := string(b)
		errCode := response.StatusCode
		err := fmt.Errorf("%w: %v", ErrGettingScans, GenericError{Code: errCode, Msg: errMssg})
		return nil, err
	}
	getScansResponse := &scanengine.GetScansResponse{}
	err = json.Unmarshal(responseBytes, getScansResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling")
	}
	return getScansResponse, nil
}

// Abort send a signal to the scan engine to try to abort a scan.
func (c *Client) Abort(scanID string) (*scanengine.GetScanResponse, error) {
	req, err := http.NewRequest("PUT", c.BaseURL.String()+fmt.Sprintf("scans/%v/abort", scanID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Request-Id", c.ctx.Value(kithttp.ContextKeyRequestXRequestID).(string))
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusAccepted {
		// Try to read the response for a possible error message
		b, errR := io.ReadAll(response.Body)
		if errR != nil {
			return nil, errR
		}

		errMssg := string(b)
		errCode := response.StatusCode
		err := fmt.Errorf("%w: %v", ErrAbortingScans, GenericError{Code: errCode, Msg: errMssg})
		return nil, err
	}
	getScanResponse := &scanengine.GetScanResponse{}
	err = json.Unmarshal(responseBytes, getScanResponse)
	if err != nil {
		return nil, err
	}
	return getScanResponse, nil
}

func ptrStrToStr(input *string) string {
	if input == nil {
		return ""
	}
	return *input
}
