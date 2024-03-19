/*
Copyright 2021 Adevinta
*/

package cgcatalogue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/adevinta/vulcan-api/pkg/awscatalogue"
	"github.com/google/go-cmp/cmp"
)

type MemCGCAPI struct {
	APIKey   string
	Accounts []account
}

func (m *MemCGCAPI) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check the key.
	key := req.Header.Get("Authorization")
	if key != fmt.Sprintf("apiKey %s", m.APIKey) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
		}, nil
	}
	p := req.URL.Path
	if p == "/v1/accounts" {
		return m.GetAccounts()
	}
	if strings.HasPrefix(p, "/v1/accounts/") {
		parts := strings.Split(p, "/")
		if len(parts) != 4 {
			return nil, errors.New("invalid path")
		}
		return m.GetProviderAccounts(parts[3])
	}
	return nil, errors.New("not implemented")
}

func (m *MemCGCAPI) GetAccounts() (*http.Response, error) {
	buff := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buff)
	err := enc.Encode(m.Accounts)
	if err != nil {
		return nil, err
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(buff.String())),
	}
	return resp, nil
}

func (m *MemCGCAPI) GetProviderAccounts(provider string) (*http.Response, error) {
	var paccounts []account
	for _, a := range m.Accounts {
		if a.Provider == provider {
			paccounts = append(paccounts, a)
			continue
		}
	}

	buff := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buff)
	err := enc.Encode(paccounts)
	if err != nil {
		return nil, err
	}
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(buff),
	}
	resp.ContentLength = int64(len(buff.Bytes()))
	return resp, nil
}

var (
	testAccounts = []account{
		{
			ID:             0,
			AccountName:    "Account 1",
			Asset:          "Asset1",
			CostCenter:     "CostCenter",
			IsProduction:   true,
			Provider:       "AWS",
			Administrators: []string{},
		},
		{
			ID:             1,
			AccountName:    "Account 2",
			Asset:          "Asset1",
			CostCenter:     "CostCenter1",
			IsProduction:   true,
			Provider:       "Datadog",
			Administrators: []string{},
		},
	}

	testKey = "Key"
)

func TestClient_Accounts(t *testing.T) {
	type fields struct {
		rt      http.RoundTripper
		baseURL string
		APIKey  string
	}

	tests := []struct {
		name    string
		fields  fields
		want    []awscatalogue.Account
		wantErr bool
	}{
		{
			name: "ReturnAccountsOfProvider",
			fields: fields{
				baseURL: "http://localhost",
				rt: &MemCGCAPI{
					Accounts: testAccounts,
					APIKey:   testKey,
				},
				APIKey: testKey,
			},
			want: []awscatalogue.Account{
				{
					ID:          testAccounts[0].ProviderID,
					AccountName: testAccounts[0].AccountName,
					Status:      testAccounts[0].Status,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			c := New(tt.fields.APIKey, tt.fields.baseURL, tt.fields.rt)
			got, err := c.Accounts()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ProviderAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("Client.ProviderAccounts() != want, diff %s", diff)
			}
		})
	}
}
