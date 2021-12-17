/*
Copyright 2021 Adevinta
*/

package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/auth/jwt"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"
	uuid "github.com/satori/go.uuid"

	vulcanendpoint "github.com/adevinta/vulcan-api/pkg/api/endpoint"
)

// CustomCtxKey represents a custom
// key type for API requests context.
type CustomCtxKey int

const (
	// ContextKeyEndpoint is the context key
	// for the requested API endpoint.
	ContextKeyEndpoint CustomCtxKey = iota
)

func options(logger kitlog.Logger, endpoint string) []kithttp.ServerOption {
	beforeFuncs := []kithttp.RequestFunc{
		HTTPGenerateXRequestID(),
		kithttp.PopulateRequestContext,
		jwt.HTTPToContext(),
		HTTPRequestEndpoint(endpoint),
	}
	// Avoid logging on healthchecks
	if endpoint != vulcanendpoint.Healthcheck {
		beforeFuncs = append(beforeFuncs, HTTPRequestLogger(logger))
	}

	opts := []kithttp.ServerOption{
		kithttp.ServerBefore(beforeFuncs...),
		kithttp.ServerAfter(HTTPReturnXRequestID()),
		kithttp.ServerErrorEncoder(
			func(ctx context.Context, err error, w http.ResponseWriter) {
				w.Header().Set("X-Request-ID", ctx.Value(kithttp.ContextKeyRequestXRequestID).(string))
				kithttp.DefaultErrorEncoder(ctx, err, w)
			},
		),
	}
	// Avoid logging on healthchecks
	if endpoint != vulcanendpoint.Healthcheck {
		opts = append(opts, kithttp.ServerFinalizer(HTTPServerFinalizerFunc(logger)))
	}

	return opts
}

func newServer(e kitendpoint.Endpoint, request interface{}, logger kitlog.Logger, endpoint string) http.Handler {
	return kithttp.NewServer(
		e,
		makeDecodeRequestFunc(request),
		kithttp.EncodeJSONResponse,
		options(logger, endpoint)...,
	)
}

func HTTPGenerateXRequestID() kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		XRequestID, _ := uuid.NewV4()
		r.Header.Set("X-Request-ID", XRequestID.String())
		return ctx
	}
}

func HTTPRequestLogger(logger kitlog.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		_ = level.Debug(logger).Log(
			"X-Request-ID", ctx.Value(kithttp.ContextKeyRequestXRequestID).(string),
			"transport", ctx.Value(kithttp.ContextKeyRequestPath).(string),
			"Method", ctx.Value(kithttp.ContextKeyRequestMethod).(string),
			"RequestURI", ctx.Value(kithttp.ContextKeyRequestURI).(string))
		return ctx
	}
}

// HTTPRequestEndpoint includes a new request ctx entry
// indicating which endpoint was requested.
func HTTPRequestEndpoint(endpoint string) kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, ContextKeyEndpoint, endpoint)
	}
}

func HTTPReturnXRequestID() kithttp.ServerResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter) context.Context {
		w.Header().Set("X-Request-ID", ctx.Value(kithttp.ContextKeyRequestXRequestID).(string))
		return ctx
	}
}

func HTTPServerFinalizerFunc(logger kitlog.Logger) kithttp.ServerFinalizerFunc {
	return func(ctx context.Context, code int, r *http.Request) {
		_ = level.Debug(logger).Log(
			"X-Request-ID", ctx.Value(kithttp.ContextKeyRequestXRequestID).(string),
			"transport", ctx.Value(kithttp.ContextKeyRequestPath).(string),
			"Method", ctx.Value(kithttp.ContextKeyRequestMethod).(string),
			"RequestURI", ctx.Value(kithttp.ContextKeyRequestURI).(string),
			"ResponseHeaders", fmt.Sprintf("%+v", ctx.Value(kithttp.ContextKeyResponseHeaders)),
			"ResponseSize", fmt.Sprintf("%+v", ctx.Value(kithttp.ContextKeyResponseSize)),
			"HTTP-Response-Code", code,
		)
	}
}
