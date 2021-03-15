/*
Copyright 2021 Adevinta
*/

package saml

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockProvider struct {
	authURLFunc  func(string) (string, error)
	userDataFunc func(string) (UserData, error)
}

func (m *mockProvider) BuildAuthURL(url string) (string, error) {
	if m.authURLFunc != nil {
		return m.authURLFunc(url)
	}
	return "", nil
}
func (m *mockProvider) GetUserData(samlResp string) (UserData, error) {
	if m.userDataFunc != nil {
		return m.userDataFunc(samlResp)
	}
	return UserData{}, nil
}

func TestLoginHandler(t *testing.T) {
	testCases := []struct {
		name           string
		trustedDomains []string
		req            *http.Request
		wantHTTPStatus int
		wantHTTPBody   string
	}{
		{
			name: "happy path",
			trustedDomains: []string{
				"vulcan.com",
			},
			req: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme:   "http://",
					Host:     "vulcan.com",
					RawQuery: "redirect_to=http://vulcan.com/callback",
				},
			},
			wantHTTPStatus: http.StatusFound,
		},
		{
			name: "should return untrusted domain",
			trustedDomains: []string{
				"vulcan.com",
			},
			req: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme:   "http://",
					Host:     "example.com",
					RawQuery: "redirect_to=http://evilsite.com/pwned",
				},
			},
			wantHTTPStatus: http.StatusBadRequest,
			wantHTTPBody:   ErrUntrustedDomain.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(&mockProvider{}, tc.trustedDomains)

			rrecorder := httptest.NewRecorder()
			handler.LoginHandler()(rrecorder, tc.req)

			if rrecorder.Result().StatusCode != tc.wantHTTPStatus {
				t.Fatalf("error expected status to be %d but got %d",
					tc.wantHTTPStatus, rrecorder.Result().StatusCode)
			}
			respBody, err := ioutil.ReadAll(rrecorder.Body)
			if err != nil {
				t.Fatalf("error reading response body: %v", err)
			}
			if tc.wantHTTPBody != "" && string(respBody) != tc.wantHTTPBody {
				t.Fatalf("error expected resp body to be: %s\nbut got: %s",
					tc.wantHTTPBody, string(respBody))
			}
		})
	}
}

func TestLoginCallbackHandler(t *testing.T) {
	testCases := []struct {
		name            string
		provider        Provider
		cfg             CallbackConfig
		req             *http.Request
		wantHTTPStatus  int
		wantHTTPBody    string
		wantCookie      bool
		wantCookieName  string
		wantCookieValue string
	}{
		{
			name: "Happy path",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{"JDoe", "John", "Doe", "jd@jd.com"}, nil
				},
			},
			cfg: CallbackConfig{
				CookieName:   "vulcancookie",
				CookieDomain: "vulcan.com",
				TokenGenerator: func(data map[string]interface{}) (string, error) {
					return "vulcanToken", nil
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus:  http.StatusFound,
			wantCookie:      true,
			wantCookieName:  "vulcancookie",
			wantCookieValue: "vulcanToken",
		},
		{
			name: "Should return bad request due to ErrParsingMetadata",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{}, ErrParsingMetadata
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus: http.StatusBadRequest,
			wantHTTPBody:   ErrParsingMetadata.Error(),
			wantCookie:     false,
		},
		{
			name: "Should return bad request due to ErrMalformedSAML",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{}, ErrMalformedSAML
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus: http.StatusBadRequest,
			wantHTTPBody:   ErrMalformedSAML.Error(),
			wantCookie:     false,
		},
		{
			name: "Should return forbidden due to ErrNotInAudience",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{}, ErrNotInAudience
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus: http.StatusForbidden,
			wantHTTPBody:   ErrNotInAudience.Error(),
			wantCookie:     false,
		},
		{
			name: "Should execute user data callback",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{"JDoe", "John", "Doe", "jd@jd.com"}, nil
				},
			},
			cfg: CallbackConfig{
				CookieName:   "vulcancookie2",
				CookieDomain: "vulcan.com",
				UserDataCallback: func(ud UserData) error {
					if ud.UserName != "JDoe" || ud.FirstName != "John" ||
						ud.LastName != "Doe" || ud.Email != "jd@jd.com" {
						return errors.New("user data do not match")
					}
					return nil
				},
				TokenGenerator: func(data map[string]interface{}) (string, error) {
					return "vulcanToken2", nil
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus:  http.StatusFound,
			wantCookie:      true,
			wantCookieName:  "vulcancookie2",
			wantCookieValue: "vulcanToken2",
		},
		{
			name: "Should execute user data callback and return ErrUserDataCallback",
			provider: &mockProvider{
				userDataFunc: func(string) (UserData, error) {
					return UserData{"JDoe", "John", "Doe", "jd@jd.com"}, nil
				},
			},
			cfg: CallbackConfig{
				CookieName:   "vulcancookie2",
				CookieDomain: "vulcan.com",
				UserDataCallback: func(ud UserData) error {
					return errors.New("err")
				},
			},
			req: &http.Request{
				Method: http.MethodPost,
				URL: &url.URL{
					Scheme: "http://",
					Host:   "vulcan.com",
				},
				PostForm: url.Values{},
			},
			wantHTTPStatus: http.StatusBadRequest,
			wantHTTPBody:   ErrUserDataCallback.Error(),
			wantCookie:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(tc.provider, []string{})

			rrecorder := httptest.NewRecorder()
			handler.LoginCallbackHandler(tc.cfg)(rrecorder, tc.req)

			if rrecorder.Result().StatusCode != tc.wantHTTPStatus {
				t.Fatalf("error expected status to be %d but got %d",
					tc.wantHTTPStatus, rrecorder.Result().StatusCode)
			}
			respBody, err := ioutil.ReadAll(rrecorder.Body)
			if err != nil {
				t.Fatalf("error reading response body: %v", err)
			}
			if tc.wantHTTPBody != "" && string(respBody) != tc.wantHTTPBody {
				t.Fatalf("error expected resp body to be: %s\nbut got: %s",
					tc.wantHTTPBody, string(respBody))
			}
			if tc.wantCookie && !isCookieSet(t, rrecorder, tc.wantCookieName, tc.wantCookieValue) {
				t.Fatalf("error expected cookie '%s' to be set, but it was not",
					tc.wantCookieName)
			}
		})
	}
}

func isCookieSet(t *testing.T, w http.ResponseWriter, cookieName, cookieValue string) bool {
	t.Helper()

	cookieH := w.Header().Get("Set-Cookie")
	cookies := strings.Split(cookieH, ";")
	for _, c := range cookies {
		cookieParts := strings.Split(c, "=")
		if len(cookieParts) == 2 &&
			cookieParts[0] == cookieName && cookieParts[1] == cookieValue {
			return true
		}
	}
	return false
}
