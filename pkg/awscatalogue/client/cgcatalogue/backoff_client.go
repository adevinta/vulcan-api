/*
Copyright 2021 Adevinta
*/

package cgcatalogue

import (
	"context"
	"errors"
	"time"

	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/lestrrat-go/backoff"
)

// JitterFactor defines the jitter factor used by the backoff client.
var JitterFactor = 0.5

// BackOffClient implements retries with backoff and jitter on top an CGApicatalogueClient.
type BackOffClient struct {
	*Client
	retryInterval int
	maxRetries    int
}

// NewBackOffClient returns a retries with backoff and client from an CGCatalog
// client using the given retry interval and max retries paramters.
func NewBackOffClient(c *Client, rinterval, mretries int) *BackOffClient {
	return &BackOffClient{c, rinterval, mretries}
}

// Accounts executes the Accounts operation with retries.
func (c *BackOffClient) Accounts() ([]awscatalogue.Account, error) {
	var (
		accs []awscatalogue.Account
		err  error
	)
	err = execWithBackOff(c.retryInterval, c.maxRetries, JitterFactor, func() error {
		accs, err = c.Client.Accounts()
		if err == nil {
			return nil
		}
		return err
	})
	// If the the retry function returned an error we just return error.
	if err != nil {
		return nil, err
	}
	return accs, nil
}

// Account executes the Account operation with retries.
func (c *BackOffClient) Account(ID string) (awscatalogue.Account, error) {
	var (
		acc awscatalogue.Account
		err error
	)
	err = execWithBackOff(c.retryInterval, c.maxRetries, JitterFactor, func() error {
		acc, err = c.Client.Account(ID)
		if err == nil {
			return nil
		}
		return err
	})
	return acc, err
}

func execWithBackOff(retryInterval, maxRetries int, jitter float64, run func() error) error {
	backoffPolicy := backoff.NewExponential(
		backoff.WithInterval(time.Duration(retryInterval)*time.Second),
		backoff.WithMaxRetries(maxRetries),
		backoff.WithJitterFactor(JitterFactor),
	)
	b, cancel := backoffPolicy.Start(context.Background())
	defer cancel()
	var err error
	for backoff.Continue(b) {
		err = run()
		if err == nil {
			break
		}
		// If the error is an unexpected http status we assume it's a logical
		// error, or at least not related to the stability of the catalogue api,
		// so we don't retry.
		var unerr *unexpectedStatusError
		if errors.As(err, &unerr) {
			return err
		}
	}
	return err
}
