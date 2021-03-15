/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	jwtkit "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	"github.com/adevinta/vulcan-api/pkg/jwt"
	"github.com/adevinta/vulcan-api/pkg/api/store"
	"github.com/adevinta/vulcan-api/pkg/testutil"
)

func makeFooEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return "foo", nil
	}
}

func TestAuthentication(t *testing.T) {
	var logger = log.NewLogfmtLogger(os.Stderr)
	var jwtSignKey = "S3KR3T"

	db, err := testutil.PrepareDatabaseLocal("../../../testdata/fixtures", store.NewDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	fooEndpoint := Authentication(logger, jwtSignKey, db)(makeFooEndpoint())

	tokenGenTime := time.Now()

	sessionToken, err := jwt.NewJWTConfig(jwtSignKey).GenerateToken(map[string]interface{}{
		"first_name": "Newman",
		"last_name":  "testuser",
		"email":      "vulcan-team@vulcan.com",
		"username":   "testuser",
		"iat":        tokenGenTime.Unix(),
		"exp":        tokenGenTime.Add(6 * time.Hour).Unix(),
		"sub":        "vulcan-team@vulcan.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	apiToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2MTUzNjQ3NTcsInN1YiI6InZ1bGNhbi10ZWFtQHZ1bGNhbi5jb20iLCJ0eXBlIjoiQVBJIn0.Ym7gocLREqa7fXYq9hP1lNunckDXGaSrCYAXBEi5DlA"

	invalidAPIToken := "WRONG"

	tests := []struct {
		name         string
		token        string
		wantResponse interface{}
		wantErr      error
	}{
		{
			name:         "SessionToken",
			token:        sessionToken,
			wantResponse: "foo",
			wantErr:      nil,
		},
		{
			name:         "APIToken",
			token:        apiToken,
			wantResponse: "foo",
			wantErr:      nil,
		},
		{
			name:         "InvalidAPIToken",
			token:        invalidAPIToken,
			wantResponse: nil,
			wantErr:      errors.New(`cannot parse jwt token`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var ctx = context.WithValue(context.Background(), jwtkit.JWTTokenContextKey, tt.token)

			response, errFoo := fooEndpoint(ctx, struct{}{})
			diff := cmp.Diff(errToStr(tt.wantErr), errToStr(errFoo))
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}

			diff = cmp.Diff(tt.wantResponse, response)
			if diff != "" {
				t.Fatalf("%v\n", diff)
			}
		})
	}

}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}
