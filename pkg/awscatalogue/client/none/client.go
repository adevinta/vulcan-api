/*
Copyright 2021 Adevinta
*/

// Package none provides a dummy implementation of the AWS catalogue client interface.
package none

import "github.com/adevinta/vulcan-api/pkg/awscatalogue"

type Client struct{}

func (c *Client) Accounts() ([]awscatalogue.Account, error) {
	return []awscatalogue.Account{}, nil
}
func (c *Client) Account(ID string) (awscatalogue.Account, error) {
	return awscatalogue.Account{
		ID:          ID,
		AccountName: ID,
	}, nil
}

func New() *Client {
	return &Client{}
}
