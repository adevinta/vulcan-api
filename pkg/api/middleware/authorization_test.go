/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

const okMsg = "Everything is gona be ok"

var logger log.Logger

func init() {
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
}

type proxyAuthorizer struct {
	tenant func(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error)
	rol    func(ctx context.Context, tenant interface{}) (bool, error)
}

func (a proxyAuthorizer) AuthTenant(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error) {
	return a.tenant(ctx, request)
}

func (a proxyAuthorizer) AuthRol(ctx context.Context, tenant interface{}) (bool, error) {
	return a.rol(ctx, tenant)
}

type testAuthorizationMiddlewareArgs struct {
	next    endpoint.Endpoint
	ctx     context.Context
	auth    Authorizer
	method  string
	path    string
	request string
}

func TestAuthorizationMiddleware(t *testing.T) {

	tests := []struct {
		name string
		args testAuthorizationMiddlewareArgs
		want string
	}{
		{
			name: "AllowsSuperAdminAccess",
			args: testAuthorizationMiddlewareArgs{
				next: func(ctx context.Context, request interface{}) (interface{}, error) {
					return okMsg, nil
				},
				auth: proxyAuthorizer{
					tenant: func(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error) {
						return "", true, nil
					},
				},
				path: "/test",
			},
			want: okMsg,
		},
		{
			name: "ForbidUserNotAuthInTenant",
			args: testAuthorizationMiddlewareArgs{
				next: func(ctx context.Context, request interface{}) (interface{}, error) {
					return okMsg, nil
				},
				auth: proxyAuthorizer{
					tenant: func(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error) {
						return nil, false, nil
					},
				},
				path: "/test",
			},
			want: "tenant is nil",
		},
		{
			name: "AllowWhenRoleIsOKForATenant",
			args: testAuthorizationMiddlewareArgs{
				next: func(ctx context.Context, request interface{}) (interface{}, error) {
					return okMsg, nil
				},
				auth: proxyAuthorizer{
					tenant: func(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error) {
						return "team1", false, nil
					},
					rol: func(ctx context.Context, tenant interface{}) (bool, error) {
						t, ok := tenant.(string)
						if !ok {
							return false, errors.New("Unexpected tenant type")
						}
						if t == "team1" {
							return true, nil
						}
						return false, nil
					},
				},
				path: "/test/12/do",
			},
			want: okMsg,
		},
		{
			name: "ForbidWhenRoleIsNotOkForATenant",
			args: testAuthorizationMiddlewareArgs{
				next: func(ctx context.Context, request interface{}) (interface{}, error) {
					return okMsg, nil
				},
				auth: proxyAuthorizer{
					tenant: func(ctx context.Context, request interface{}) (tenant interface{}, passThrough bool, err error) {
						return "team1", false, nil
					},
					rol: func(ctx context.Context, tenant interface{}) (bool, error) {
						_, ok := tenant.(string)
						if !ok {
							return false, errors.New("Unexpected tenant type")
						}
						return false, nil
					},
				},
				path: "/test/12/do",
			},
			want: "access not granted",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := &authorizerMiddleware{
				auth:   tt.args.auth,
				logger: logger,
			}
			tt.args.next = a.Authorize(tt.args.next)
			handler := func(w http.ResponseWriter, r *http.Request) {
				response, err := tt.args.next(tt.args.ctx, tt.args.request)
				if err != nil {
					w.Write([]byte(err.Error())) // nolint
					return
				}
				str, ok := response.(string)
				if !ok {
					w.Write([]byte("error expected response to be a string")) // nolint
				}
				w.Write([]byte(str)) // nolint
			}

			mux := http.NewServeMux()
			mux.HandleFunc(tt.args.path, handler)

			req, err := http.NewRequest(tt.args.method, tt.args.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp, _ := ioutil.ReadAll(w.Body)
			got := string(resp)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
