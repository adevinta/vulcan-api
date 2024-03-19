/*
Copyright 2021 Adevinta
*/

package cgcatalogue

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
)

const (
	awsProvider = "AWS"
	apiVersion  = "v1"
	slash       = "/"
)

var (
	errAccounts = errors.New("error getting accounts data")
	errAccount  = errors.New("error getting account data")
)

// account represents the administrative information
// related with a provider account from CG API.
type account struct {
	ID             int64    `json:"id"`
	AccountName    string   `json:"account_name"`
	Administrators []string `json:"administrators"`
	Asset          string   `json:"asset"`
	CostCenter     string   `json:"cost_center"`
	IsProduction   bool     `json:"is_production"`
	Payer          string   `json:"payer"`
	Provider       string   `json:"provider"`
	ProviderID     string   `json:"provider_id"`
	Status         string   `json:"status"`
}

func (a account) toAWSAccount() awscatalogue.Account {
	return awscatalogue.Account{
		ID:          a.ProviderID,
		AccountName: a.AccountName,
		Status:      a.Status,
	}
}

type operationError struct {
	opErr    error
	innerErr error
}

func (op *operationError) Unwrap() error {
	return op.opErr
}

func (op *operationError) Error() string {
	return fmt.Sprintf("%v:%v", op.opErr, op.innerErr)
}

// unexpectedStatusError is returned when a call the the
// catalogue API does not return the expected status code.
type unexpectedStatusError struct {
	Expected int
	Received int
}

func (e *unexpectedStatusError) Error() string {
	return fmt.Sprintf("expected status %d, received status %d", e.Expected, e.Received)
}

// Client implements functions exposed by the CG Catalogue API.
type Client struct {
	rt      http.RoundTripper
	baseURL string
}

type addAuthRoundTripper struct {
	APIKey string
	Next   http.RoundTripper
}

func (a addAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	auth := fmt.Sprintf("apiKey %s", a.APIKey)
	req.Header.Add("Authorization", auth)
	return a.Next.RoundTrip(req)
}

// New returns a new Client given an API KEY, a baseURL and round tripper if the
// it is nil it will be set to the default http round tripper.
func New(APIKey string, baseURL string, rt http.RoundTripper) *Client {
	if rt == nil {
		rt = http.DefaultTransport
	}
	rt = addAuthRoundTripper{APIKey: APIKey, Next: rt}
	if !strings.HasSuffix(baseURL, slash) {
		baseURL = baseURL + slash
	}
	return &Client{rt, fmt.Sprint(baseURL, apiVersion, slash)}
}

// Accounts returns AWS accounts from CG catalogue.
func (c *Client) Accounts() ([]awscatalogue.Account, error) {
	url := fmt.Sprintf("%saccounts/%s", c.baseURL, awsProvider)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &operationError{errAccounts, err}
	}
	resp, err := c.rt.RoundTrip(req)
	if err != nil {
		return nil, &operationError{errAccounts, err}
	}
	if resp.StatusCode != http.StatusOK {
		uerr := &unexpectedStatusError{http.StatusOK, resp.StatusCode}
		return nil, &operationError{errAccounts, uerr}
	}
	accs := []account{}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &operationError{errAccounts, err}
	}
	json.Unmarshal(content, &accs)
	if err != nil {
		return nil, &operationError{errAccounts, err}
	}
	awsAccs := []awscatalogue.Account{}
	for _, a := range accs {
		awsAccs = append(awsAccs, a.toAWSAccount())
	}
	return awsAccs, nil
}

// Account returns account information for the given AWS account ID.
func (c *Client) Account(ID string) (awscatalogue.Account, error) {
	var acc account
	url := fmt.Sprintf("%saccounts/%s/%s", c.baseURL, awsProvider, ID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return awscatalogue.Account{}, &operationError{errAccount, err}
	}
	resp, err := c.rt.RoundTrip(req)
	if err != nil {
		return awscatalogue.Account{}, &operationError{errAccount, err}
	}
	if resp.StatusCode != http.StatusOK {
		uerr := &unexpectedStatusError{http.StatusOK, resp.StatusCode}
		return awscatalogue.Account{}, &operationError{uerr, err}
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&acc)
	if err != nil {
		return awscatalogue.Account{}, &operationError{errAccount, err}
	}
	return acc.toAWSAccount(), nil
}
