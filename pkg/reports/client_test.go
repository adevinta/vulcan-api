/*
Copyright 2021 Adevinta
*/

package reports

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/adevinta/errors"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

var (
	// mockReport to return from
	// test HTTP server.
	mockReport = ScanReport{
		ID:            "mockID",
		ReportURL:     "reportURL",
		ReportJSONURL: "reportJSONURL",
		ScanID:        "scanID",
		ProgramName:   "progName",
		Status:        "FINISHED",
		Risk:          2,
		DeliveredTo:   "tom@vulcan.example.com",
	}

	// mockNotification to return from
	// test HTTP server.
	mockNotification = Notification{
		Subject: "Subject",
		Body:    "Body",
		Format:  "HTML",
	}

	mockSNSPublishErr = errs.New("ErrSNSPublish")
)

type snsMockPublishFunc func(s *sns.PublishInput) (*sns.PublishOutput, error)
type snsAPIMock struct {
	snsiface.SNSAPI
	mockFunc snsMockPublishFunc
}

func (m *snsAPIMock) Publish(s *sns.PublishInput) (*sns.PublishOutput, error) {
	return m.mockFunc(s)
}

// srvErr represents an error
// response fromr remote server
type srvErr struct {
	statusCode int
	mssg       string
}

func TestGetReport(t *testing.T) {
	testCases := []struct {
		name           string
		scanID         string
		srvReport      ScanReport // report to return from mock HTTP server
		srvErr         *srvErr    // error to return from mock HTTP server
		expectedPath   string
		expectedReport ScanReport
		expectedErr    error
	}{
		{
			name:           "Happy path",
			scanID:         "11",
			srvReport:      mockReport,
			expectedPath:   "/api/v1/reports/scan/11",
			expectedReport: mockReport,
		},
		{
			name:   "Should return 400 error from reportsgenerator",
			scanID: "12",
			srvErr: &srvErr{
				statusCode: http.StatusBadRequest,
				mssg:       "400 Err",
			},
			expectedPath: "/api/v1/reports/scan/12",
			expectedErr:  errors.Assertion("400 Err"),
		},
		{
			name:   "Should return 401 error from reportsgenerator",
			scanID: "13",
			srvErr: &srvErr{
				statusCode: http.StatusUnauthorized,
				mssg:       "401 Err",
			},
			expectedPath: "/api/v1/reports/scan/13",
			expectedErr:  errors.Unauthorized("401 Err"),
		},
		{
			name:   "Should return 403 error from reportsgenerator",
			scanID: "14",
			srvErr: &srvErr{
				statusCode: http.StatusForbidden,
				mssg:       "403 Err",
			},
			expectedPath: "/api/v1/reports/scan/14",
			expectedErr:  errors.Forbidden("403 Err"),
		},
		{
			name:   "Should return 404 error from reportsgenerator",
			scanID: "15",
			srvErr: &srvErr{
				statusCode: http.StatusNotFound,
				mssg:       "404 Err",
			},
			expectedPath: "/api/v1/reports/scan/15",
			expectedErr:  errors.NotFound("404 Err"),
		},
		{
			name:   "Should return 405 error from reportsgenerator",
			scanID: "16",
			srvErr: &srvErr{
				statusCode: http.StatusMethodNotAllowed,
				mssg:       "405 Err",
			},
			expectedPath: "/api/v1/reports/scan/16",
			expectedErr:  errors.MethodNotAllowed("405 Err"),
		},
		{
			name:   "Should return 422 error from reportsgenerator",
			scanID: "17",
			srvErr: &srvErr{
				statusCode: http.StatusUnprocessableEntity,
				mssg:       "422 Err",
			},
			expectedPath: "/api/v1/reports/scan/17",
			expectedErr:  errors.Validation("422 Err"),
		},
		{
			name:   "Should return 500 error from reportsgenerator",
			scanID: "18",
			srvErr: &srvErr{
				statusCode: http.StatusInternalServerError,
				mssg:       "500 Err",
			},
			expectedPath: "/api/v1/reports/scan/18",
			expectedErr:  errors.Default("500 Err"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.expectedPath {
					t.Fatalf("Expected req path to be: %s\nBut got: %s", tc.expectedPath, r.URL.Path)
				}

				var httpStatus int
				var resp []byte

				if tc.srvErr != nil {
					httpStatus = tc.srvErr.statusCode
					resp = []byte(tc.srvErr.mssg)
				} else {
					httpStatus = http.StatusOK
					resp, _ = json.Marshal(tc.srvReport)
				}

				w.WriteHeader(httpStatus)
				w.Write(resp) // nolint
			}))

			client := &client{
				cfg:        Config{APIBaseURL: server.URL},
				httpClient: http.DefaultClient,
			}

			report, err := client.GetReport(context.Background(), tc.scanID)
			if !reflect.DeepEqual(err, tc.expectedErr) { // nolint
				if tc.expectedErr == nil {
					t.Fatalf("No error expected, but got '%v'", err)
				} else {
					t.Fatalf("Expecting error '%v', but got '%v'", tc.expectedErr, err)
				}
			}
			if err == nil && !reflect.DeepEqual(report, &tc.expectedReport) {
				t.Fatalf("Expected report: %v\nBut got: %v", tc.expectedReport, report)
			}
		})
	}
}

func TestGetNotification(t *testing.T) {
	testCases := []struct {
		name                 string
		scanID               string
		srvNotification      Notification // report to return from mock HTTP server
		srvErr               *srvErr      // error to return from mock HTTP server
		expectedPath         string
		expectedNotification Notification
		expectedErr          error
	}{
		{
			name:                 "Happy path",
			scanID:               "11",
			srvNotification:      mockNotification,
			expectedPath:         "/api/v1/reports/scan/11/notification",
			expectedNotification: mockNotification,
		},
		{
			name:   "Should return 400 error from reportsgenerator",
			scanID: "12",
			srvErr: &srvErr{
				statusCode: http.StatusBadRequest,
				mssg:       "400 Err",
			},
			expectedPath: "/api/v1/reports/scan/12/notification",
			expectedErr:  errors.Assertion("400 Err"),
		},
		{
			name:   "Should return 401 error from reportsgenerator",
			scanID: "13",
			srvErr: &srvErr{
				statusCode: http.StatusUnauthorized,
				mssg:       "401 Err",
			},
			expectedPath: "/api/v1/reports/scan/13/notification",
			expectedErr:  errors.Unauthorized("401 Err"),
		},
		{
			name:   "Should return 403 error from reportsgenerator",
			scanID: "14",
			srvErr: &srvErr{
				statusCode: http.StatusForbidden,
				mssg:       "403 Err",
			},
			expectedPath: "/api/v1/reports/scan/14/notification",
			expectedErr:  errors.Forbidden("403 Err"),
		},
		{
			name:   "Should return 404 error from reportsgenerator",
			scanID: "15",
			srvErr: &srvErr{
				statusCode: http.StatusNotFound,
				mssg:       "404 Err",
			},
			expectedPath: "/api/v1/reports/scan/15/notification",
			expectedErr:  errors.NotFound("404 Err"),
		},
		{
			name:   "Should return 405 error from reportsgenerator",
			scanID: "16",
			srvErr: &srvErr{
				statusCode: http.StatusMethodNotAllowed,
				mssg:       "405 Err",
			},
			expectedPath: "/api/v1/reports/scan/16/notification",
			expectedErr:  errors.MethodNotAllowed("405 Err"),
		},
		{
			name:   "Should return 422 error from reportsgenerator",
			scanID: "17",
			srvErr: &srvErr{
				statusCode: http.StatusUnprocessableEntity,
				mssg:       "422 Err",
			},
			expectedPath: "/api/v1/reports/scan/17/notification",
			expectedErr:  errors.Validation("422 Err"),
		},
		{
			name:   "Should return 500 error from reportsgenerator",
			scanID: "18",
			srvErr: &srvErr{
				statusCode: http.StatusInternalServerError,
				mssg:       "500 Err",
			},
			expectedPath: "/api/v1/reports/scan/18/notification",
			expectedErr:  errors.Default("500 Err"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.expectedPath {
					t.Fatalf("Expected req path to be: %s\nBut got: %s", tc.expectedPath, r.URL.Path)
				}

				var httpStatus int
				var resp []byte

				if tc.srvErr != nil {
					httpStatus = tc.srvErr.statusCode
					resp = []byte(tc.srvErr.mssg)
				} else {
					httpStatus = http.StatusOK
					resp, _ = json.Marshal(tc.srvNotification)
				}

				w.WriteHeader(httpStatus)
				w.Write(resp) // nolint
			}))

			client := &client{
				cfg:        Config{APIBaseURL: server.URL},
				httpClient: http.DefaultClient,
			}

			notif, err := client.GetReportNotification(context.Background(), tc.scanID)
			if !reflect.DeepEqual(err, tc.expectedErr) { // nolint
				if tc.expectedErr == nil {
					t.Fatalf("No error expected, but got '%v'", err)
				} else {
					t.Fatalf("Expecting error '%v', but got '%v'", tc.expectedErr, err)
				}
			}
			if err == nil && !reflect.DeepEqual(notif, &tc.expectedNotification) {
				t.Fatalf("Expected notification: %v\nBut got: %v", &tc.expectedNotification, notif)
			}
		})
	}
}

func TestSendReportNotification(t *testing.T) {
	testCases := []struct {
		name         string
		scanID       string
		recipients   []string
		srvErr       *srvErr
		expectedPath string
		expectedErr  error
	}{
		{
			name:         "Happy path",
			scanID:       "11",
			recipients:   []string{"tom@vulcan.example.com"},
			expectedPath: "/api/v1/reports/scan/11/send",
		},
		{
			name:   "Should return 400 error from reportsgenerator",
			scanID: "12",
			srvErr: &srvErr{
				statusCode: http.StatusBadRequest,
				mssg:       "400 Err",
			},
			expectedPath: "/api/v1/reports/scan/12/send",
			expectedErr:  errors.Assertion("400 Err"),
		},
		{
			name:   "Should return 401 error from reportsgenerator",
			scanID: "13",
			srvErr: &srvErr{
				statusCode: http.StatusUnauthorized,
				mssg:       "401 Err",
			},
			expectedPath: "/api/v1/reports/scan/13/send",
			expectedErr:  errors.Unauthorized("401 Err"),
		},
		{
			name:   "Should return 403 error from reportsgenerator",
			scanID: "14",
			srvErr: &srvErr{
				statusCode: http.StatusForbidden,
				mssg:       "403 Err",
			},
			expectedPath: "/api/v1/reports/scan/14/send",
			expectedErr:  errors.Forbidden("403 Err"),
		},
		{
			name:   "Should return 404 error from reportsgenerator",
			scanID: "15",
			srvErr: &srvErr{
				statusCode: http.StatusNotFound,
				mssg:       "404 Err",
			},
			expectedPath: "/api/v1/reports/scan/15/send",
			expectedErr:  errors.NotFound("404 Err"),
		},
		{
			name:   "Should return 405 error from reportsgenerator",
			scanID: "16",
			srvErr: &srvErr{
				statusCode: http.StatusMethodNotAllowed,
				mssg:       "405 Err",
			},
			expectedPath: "/api/v1/reports/scan/16/send",
			expectedErr:  errors.MethodNotAllowed("405 Err"),
		},
		{
			name:   "Should return 422 error from reportsgenerator",
			scanID: "17",
			srvErr: &srvErr{
				statusCode: http.StatusUnprocessableEntity,
				mssg:       "422 Err",
			},
			expectedPath: "/api/v1/reports/scan/17/send",
			expectedErr:  errors.Validation("422 Err"),
		},
		{
			name:   "Should return 500 error from reportsgenerator",
			scanID: "18",
			srvErr: &srvErr{
				statusCode: http.StatusInternalServerError,
				mssg:       "500 Err",
			},
			expectedPath: "/api/v1/reports/scan/18/send",
			expectedErr:  errors.Default("500 Err"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify path.
				if r.URL.Path != tc.expectedPath {
					t.Fatalf("Expected req path to be: %s\nBut got: %s", tc.expectedPath, r.URL.Path)
				}

				// Parse and verify body.
				reqBody, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Error parsing request body: %v", err)
				}
				recipientsPayload := struct {
					Recipients []string
				}{}
				err = json.Unmarshal(reqBody, &recipientsPayload)
				if err != nil {
					t.Fatalf("Request body does not comply with recipients payload format")
				}
				if !reflect.DeepEqual(recipientsPayload.Recipients, tc.recipients) {
					t.Fatalf("Recipients payload does not match request input ones")
				}

				// Build response
				// based on test data.
				var httpStatus int
				var resp []byte

				if tc.srvErr != nil {
					httpStatus = tc.srvErr.statusCode
					resp = []byte(tc.srvErr.mssg)
				} else {
					httpStatus = http.StatusOK
					resp = []byte("OK")
				}

				w.WriteHeader(httpStatus)
				w.Write(resp) // nolint
			}))

			client := &client{
				cfg:        Config{APIBaseURL: server.URL},
				httpClient: http.DefaultClient,
			}

			err := client.SendReportNotification(context.Background(), tc.scanID, tc.recipients)
			if !reflect.DeepEqual(err, tc.expectedErr) { // nolint
				if tc.expectedErr == nil {
					t.Fatalf("No error expected, but got '%v'", err)
				} else {
					t.Fatalf("Expecting error '%v', but got '%v'", tc.expectedErr, err)
				}
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	type input struct {
		scanID      string
		teamID      string
		teamName    string
		programName string
		recipients  []string
		autoSend    bool
	}

	testCases := []struct {
		name        string
		input       input
		snsAPI      snsiface.SNSAPI
		expectedErr error
	}{
		{
			name: "Happy path",
			input: input{
				scanID:      "11",
				teamID:      "21",
				teamName:    "teamName",
				programName: "progName",
				recipients:  []string{"tom@vulcan.example.com"},
				autoSend:    true,
			},
			snsAPI: &snsAPIMock{
				mockFunc: func(s *sns.PublishInput) (*sns.PublishOutput, error) {
					expectedGenReq := genReportEvent{
						Typ: scanType,
						TeamInfo: teamInfo{
							ID:         "21",
							Name:       "teamName",
							Recipients: []string{"tom@vulcan.example.com"},
						},
						Data: scanData{
							ScanID:      "11",
							ProgramName: "progName",
						},
						AutoSend: true,
					}

					var genReq genReportEvent
					err := json.Unmarshal([]byte(*s.Message), &genReq)
					if err != nil {
						return nil, errs.New("Error parsing genRequest from SNS API mock")
					}

					if !reflect.DeepEqual(genReq, expectedGenReq) {
						return nil, fmt.Errorf("Expected genReq to be: %v\nBut got: %v",
							expectedGenReq, genReq)
					}

					return nil, nil
				},
			},
		},
		{
			name: "Should return SNS publish error",
			input: input{
				scanID:      "12",
				teamID:      "22",
				teamName:    "teamName",
				programName: "progName",
				recipients:  []string{"tom@vulcan.example.com"},
				autoSend:    true,
			},
			snsAPI: &snsAPIMock{
				mockFunc: func(s *sns.PublishInput) (*sns.PublishOutput, error) {
					return nil, mockSNSPublishErr
				},
			},
			expectedErr: mockSNSPublishErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &client{
				cfg:        Config{},
				httpClient: nil,
				snsAPI:     tc.snsAPI,
			}

			err := client.GenerateReport(tc.input.scanID, tc.input.teamID, tc.input.teamName,
				tc.input.programName, tc.input.recipients, tc.input.autoSend)
			if !errs.Is(err, tc.expectedErr) {
				if tc.expectedErr == nil {
					t.Fatalf("No error expected, but got '%v'", err)
				} else {
					t.Fatalf("Expecting error '%v', but got '%v'", tc.expectedErr, err)
				}
			}
		})
	}
}
