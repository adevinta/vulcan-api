/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/vulcan-api/pkg/api"
)

var (
	loggerPolicy log.Logger
)

func init() {
	loggerPolicy = log.NewLogfmtLogger(os.Stderr)
	loggerPolicy = log.With(loggerPolicy, "ts", log.DefaultTimestampUTC)
	loggerPolicy = log.With(loggerPolicy, "caller", log.DefaultCaller)

}

func TestVulcanitoService_CreatePolicu(t *testing.T) {
	srv := vulcanitoService{
		db:     nil,
		logger: loggerPolicy,
	}
	policy := api.Policy{}
	_, err := srv.CreatePolicy(context.Background(), policy)
	if err == nil {
		t.Error("Should return validation error if empty name")
	}
	expectedErrorMessage := "Key: 'Policy.Name' Error:Field validation for 'Name' failed on the 'required' tag"
	diff := cmp.Diff(expectedErrorMessage, err.Error())
	if diff != "" {
		t.Errorf("Wrong error message, diff: %v", diff)
	}
}
