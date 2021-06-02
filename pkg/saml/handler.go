/*
Copyright 2021 Adevinta
*/

package saml

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-kit/kit/log"
)

const (
	tokenExpiresAt     = 6 * time.Hour
	redirectQueryParam = "redirect_to"
)

var (
	// ErrSAMLRequest indicates there is an error on SAML callback request.
	ErrSAMLRequest = errors.New("malformed SAML callback request")
	// ErrUserDataCallback indicates there was an error executing user data callback.
	ErrUserDataCallback = errors.New("error on user data callback")
	// ErrGeneratingToken indicates there was an error genereting JWT token.
	ErrGeneratingToken = errors.New("error generating token")
	// ErrRelayStateInvalid indicates the provided "redirect_to" URL is not valid.
	ErrRelayStateInvalid = errors.New("invalid RelayState URL")
	// ErrUntrustedDomain indicates the redirect domain is not trusted.
	ErrUntrustedDomain = errors.New("redirect to an untrusted domain was requested")
)

// UserDataCallback represents the callback to
// execute when user data is obtained from SAML response.
type UserDataCallback func(UserData) error

// TokenGenerator defines the method to generate a new session token.
// Note that is designed thinking in a Bearer token, like OAuth / JWT
type TokenGenerator func(data map[string]interface{}) (string, error)

// CallbackConfig specifies config options
// for the login callback function.
type CallbackConfig struct {
	CookieName       string
	CookieDomain     string
	CookieSecure     bool
	UserDataCallback UserDataCallback
	TokenGenerator   TokenGenerator
}

// Handler represents a SAML
// authentication handler.
type Handler interface {
	LoginHandler() http.HandlerFunc
	LoginCallbackHandler(CallbackConfig) http.HandlerFunc
}

type handler struct {
	p              Provider
	trustedDomains []string
}

// NewHandler builds a new SAML handler from a SAML provider
// and a list of trusted domains.
func NewHandler(provider Provider, trustedDomains []string) Handler {
	return &handler{
		p:              provider,
		trustedDomains: trustedDomains,
	}
}

// LoginHandler returns the function to handle login
// requests through a SAML federated identity provider.
// The 'redirect_to' req query param indicates where should
// the user be redirected once the authentication process
// through de IdP is completed.
func (h *handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectPath := r.FormValue(redirectQueryParam)

		// Validate URL
		redirectURL, err := url.Parse(redirectPath)
		if err != nil {
			writeResp(w, http.StatusBadRequest, ErrRelayStateInvalid)
			return
		}
		if redirectURL.IsAbs() {
			trusted := false
			for _, domain := range h.trustedDomains {
				if redirectURL.Hostname() == domain {
					trusted = true
					break
				}
			}
			if !trusted {
				writeResp(w, http.StatusBadRequest, ErrUntrustedDomain)
				return
			}
		}

		// Build redirect URL
		relayState, _ := h.p.BuildAuthURL(url.QueryEscape(redirectPath))
		http.Redirect(w, r, relayState, http.StatusFound)
	}
}

// LoginCallbackHandler returns the function to handle the SAML callback response
// after authentication has been performed through the identity provider.
func (h *handler) LoginCallbackHandler(cfg CallbackConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			writeResp(w, http.StatusBadRequest, ErrSAMLRequest)
			return
		}

		userData, err := h.p.GetUserData(r.FormValue("SAMLResponse"))
		if err != nil {
			respStatus := http.StatusBadRequest
			if errors.Is(err, ErrNotInAudience) {
				respStatus = http.StatusForbidden
			}
			writeResp(w, respStatus, err)
			return
		}

		if cfg.UserDataCallback != nil {
			if err = cfg.UserDataCallback(userData); err != nil {
				writeResp(w, http.StatusBadRequest, ErrUserDataCallback)
				return
			}
		}
		logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		tokenGenTime := time.Now()
		logger.Log("email", userData.Email, "username", userData.UserName, "FirstName", userData.FirstName)
		claims := map[string]interface{}{
			"first_name": userData.FirstName,
			"last_name":  userData.LastName,
			"email":      userData.UserName,
			"username":   userData.UserName,
			"iat":        tokenGenTime.Unix(),
			"exp":        tokenGenTime.Add(tokenExpiresAt).Unix(),
			"sub":        userData.UserName,
		}
		token, err := cfg.TokenGenerator(claims)
		if err != nil {
			writeResp(w, http.StatusBadRequest, ErrGeneratingToken)
			return
		}

		cookie := &http.Cookie{
			Path:    "/",
			Name:    cfg.CookieName,
			Value:   token,
			Expires: tokenGenTime.Add(tokenExpiresAt),
			Domain:  cfg.CookieDomain,
			Secure:  cfg.CookieSecure,
		}
		http.SetCookie(w, cookie)

		relayState, _ := url.QueryUnescape(r.FormValue("RelayState"))

		http.Redirect(w, r, relayState, http.StatusFound)
	}
}

func writeResp(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}
