/*
Copyright 2021 Adevinta
*/

package awscatalogue

import (
	"errors"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var (
	// ErrAccountNotFound is returned when the name of an account is not found.
	ErrAccountNotFound = errors.New("Account not found")
)

// Account represents administrative information
// related to an AWS account.
type Account struct {
	ID          string `json:"id"`
	AccountName string `json:"account_name"`
	Status      string `json:"status"`
}

// Client represents the interface to interact with an
// AWS accounts catalogue service.
type Client interface {
	Accounts() ([]Account, error)
	Account(ID string) (Account, error)
}

// AWSAccounts allows to query account names given an account id.
type AWSAccounts struct {
	c      Client
	aNames map[string]string
	l      log.Logger
	sync.RWMutex
}

// NewAWSAccounts returns an initialized instance of AWSAccounts service.
func NewAWSAccounts(c Client, l log.Logger) *AWSAccounts {
	return &AWSAccounts{
		c:      c,
		l:      l,
		aNames: make(map[string]string),
	}
}

// Name returns the name of an account given its ID.
func (c *AWSAccounts) Name(accID string) (string, error) {
	c.RLock()
	n, ok := c.aNames[accID]
	c.RUnlock()
	if ok {
		return n, nil
	}
	_ = level.Info(c.l).Log("AWS account names cache miss, query catalogue API", accID)
	acc, err := c.c.Account(accID)
	if err != nil {
		return "", err
	}
	_ = level.Info(c.l).Log("AWS account name retrieved from cache", accID, n)
	n = acc.AccountName
	c.AddName(accID, n)
	return n, nil
}

// RefreshCache refreshes all the entries in the cache.
func (c *AWSAccounts) RefreshCache() error {
	accs, err := c.c.Accounts()
	if err != nil {
		return err
	}
	var accountNames = make(map[string]string)
	for _, a := range accs {
		c.aNames[a.ID] = a.AccountName
	}
	c.Lock()
	c.aNames = accountNames
	c.Unlock()
	return nil
}

// AddName adds a new account_id, name pair to the cache.
func (c *AWSAccounts) AddName(accID, name string) error {
	c.Lock()
	c.aNames[accID] = name
	c.Unlock()
	return nil
}
