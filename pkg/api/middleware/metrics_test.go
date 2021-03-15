/*
Copyright 2021 Adevinta
*/

package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/adevinta/errors"
	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/adevinta/vulcan-api/pkg/api/endpoint"
)

type mockMetricsClient struct {
	metrics.Client
	metrics         []metrics.Metric
	expectedMetrics []metrics.Metric
}

func (c *mockMetricsClient) Push(metric metrics.Metric) {
	c.metrics = append(c.metrics, metric)
}

// Verify verifies the matching between mock client
// expected metrics and the actual pushed metrics.
func (c *mockMetricsClient) Verify() error {
	nMetrics := len(c.metrics)
	nExpectedMetrics := len(c.expectedMetrics)

	if nMetrics != nExpectedMetrics {
		return fmt.Errorf(
			"Number of metrics do not match: Expected %d, but got %d",
			nExpectedMetrics, nMetrics)
	}

	for _, m := range c.metrics {
		var found bool
		for _, em := range c.expectedMetrics {
			if reflect.DeepEqual(m, em) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Metrics do not match: Expected %v, but got %v",
				c.expectedMetrics, c.metrics)
		}
	}

	return nil
}

func TestPushMetrics(t *testing.T) {
	type input struct {
		httpMethod string
		duration   int64
		failed     bool
		tags       []string
	}

	testCases := []struct {
		name            string
		input           input
		expectedMetrics []metrics.Metric
	}{
		{
			name: "Happy path",
			input: input{
				httpMethod: http.MethodGet,
				duration:   2,
				failed:     false,
				tags:       []string{"tag:mytag"},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags:  []string{"tag:mytag"},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: float64(2),
					Tags:  []string{"tag:mytag"},
				},
			},
		},
		{
			name: "Should increment failed requests due to 500",
			input: input{
				httpMethod: http.MethodPost,
				duration:   5,
				failed:     true,
				tags:       []string{"tag:sometag"},
			},
			expectedMetrics: []metrics.Metric{
				{
					Name:  metricTotal,
					Typ:   metrics.Count,
					Value: 1,
					Tags:  []string{"tag:sometag"},
				},
				{
					Name:  metricDuration,
					Typ:   metrics.Histogram,
					Value: float64(5),
					Tags:  []string{"tag:sometag"},
				},
				{
					Name:  metricFailed,
					Typ:   metrics.Count,
					Value: 1,
					Tags:  []string{"tag:sometag"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			mockClient := &mockMetricsClient{
				expectedMetrics: tc.expectedMetrics,
			}
			metricsMiddleware := &metricsMiddleware{
				metricsClient: mockClient,
			}

			// When
			metricsMiddleware.pushMetrics(tc.input.httpMethod,
				tc.input.duration, tc.input.failed, tc.input.tags)

			// Then
			if err := mockClient.Verify(); err != nil {
				t.Fatalf("Error verifying pushed metrics: %v", err)
			}
		})
	}
}

func TestParseHTTPStatus(t *testing.T) {
	type input struct {
		resp interface{}
		err  error
	}

	mockErr := fmt.Errorf("ErrMock")

	testCases := []struct {
		name           string
		input          input
		expectedStatus int
	}{
		{
			name: "Should return 200 due to Ok resp",
			input: input{
				resp: endpoint.Ok{},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Should return 201 due to Created resp",
			input: input{
				resp: endpoint.Created{},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Should return 202 due to Accepted resp",
			input: input{
				resp: endpoint.Accepted{},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Should return 500 due to ServerDown resp",
			input: input{
				resp: endpoint.ServerDown{},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 204 due to NoContent resp",
			input: input{
				resp: endpoint.NoContent{},
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Should return 207 due to MultiStatus resp",
			input: input{
				resp: endpoint.MultiStatus{},
			},
			expectedStatus: http.StatusMultiStatus,
		},
		{
			name: "Should return 200 due to default resp",
			input: input{
				resp: struct{}{},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Should return 500 due to Default err",
			input: input{
				err: errors.Default(mockErr),
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 500 due to Database err",
			input: input{
				err: errors.Database(mockErr),
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 403 due to Forbidden err",
			input: input{
				err: errors.Forbidden(mockErr),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Should return 401 due to Unauthorized err",
			input: input{
				err: errors.Unauthorized(mockErr),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Should return 404 due to NotFound err",
			input: input{
				err: errors.NotFound(mockErr),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Should return 500 due to Create err",
			input: input{
				err: errors.Create(mockErr),
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 500 due to Update err",
			input: input{
				err: errors.Update(mockErr),
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 500 due to Delete err",
			input: input{
				err: errors.Delete(mockErr),
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Should return 422 due to Validation err",
			input: input{
				err: errors.Validation(mockErr),
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Should return 409 due to Duplicated err",
			input: input{
				err: errors.Duplicated(mockErr),
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Should return 400 due to Assertion err",
			input: input{
				err: errors.Assertion(mockErr),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Should return 405 due to Assertion err",
			input: input{
				err: errors.MethodNotAllowed(mockErr),
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "Should return 500 due to default err",
			input: input{
				err: mockErr,
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpStatus := parseHTTPStatus(tc.input.resp, tc.input.err)
			if httpStatus != tc.expectedStatus {
				t.Fatalf("Error parsing HTTP status, expected %d, but got %d",
					tc.expectedStatus, httpStatus)
			}
		})
	}
}
