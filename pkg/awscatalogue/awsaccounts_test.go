/*
Copyright 2021 Adevinta
*/

package awscatalogue

import (
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
)

type memCGServices struct {
	provider string
	accounts map[string]Account
}

func (m memCGServices) Accounts() ([]Account, error) {
	var accs []Account
	for _, acc := range m.accounts {
		accs = append(accs, acc)
	}
	return accs, nil
}

func (m memCGServices) Account(ID string) (Account, error) {
	acc, ok := m.accounts[ID]
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return acc, nil
}

func TestAWSAccounts_Name(t *testing.T) {
	tests := []struct {
		name    string
		c       Client
		aNames  map[string]string
		l       log.Logger
		accID   string
		want    string
		wantErr error
	}{
		{
			name:   "AccNotInCacheButInCatalog",
			aNames: map[string]string{},
			l:      log.NewNopLogger(),
			accID:  "1",
			c: memCGServices{
				accounts: map[string]Account{
					"1": {
						AccountName: "Account1",
					},
				},
			},
			want: "Account1",
		},
		{
			name:   "AccNotInCacheNeitherInCatalogue",
			aNames: map[string]string{},
			l:      log.NewNopLogger(),
			accID:  "1",
			c: memCGServices{
				accounts: map[string]Account{},
			},
			wantErr: ErrAccountNotFound,
		},
		{
			name:   "AccInCache",
			aNames: map[string]string{"1": "Account1"},
			l:      log.NewNopLogger(),
			accID:  "1",
			c: memCGServices{
				accounts: map[string]Account{},
			},
			want: "Account1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSAccounts{
				c:      tt.c,
				aNames: tt.aNames,
				l:      tt.l,
			}
			got, err := c.Name(tt.accID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AWSAccounts.Name() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AWSAccounts.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAWSAccounts_RefreshCache(t *testing.T) {
	tests := []struct {
		name    string
		c       Client
		aNames  map[string]string
		l       log.Logger
		want    map[string]string
		wantErr bool
	}{
		{
			name:   "HappyPath",
			aNames: map[string]string{},
			l:      log.NewNopLogger(),
			c: memCGServices{
				accounts: map[string]Account{
					"1": {AccountName: "Alias1", ID: "1"},
					"2": {AccountName: "Alias2", ID: "2"},
				},
			},
			want: map[string]string{"1": "Alias1", "2": "Alias2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSAccounts{
				c:      tt.c,
				aNames: tt.aNames,
				l:      tt.l,
			}
			if err := c.RefreshCache(); (err != nil) != tt.wantErr {
				t.Errorf("AWSAccounts.RefreshCache() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := tt.aNames
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("got != want, diff %s", diff)
			}
		})
	}
}
