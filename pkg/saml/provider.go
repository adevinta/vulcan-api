/*
Copyright 2021 Adevinta
*/

package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
)

var (
	// ErrParsingMetadata indicates there was an error obtaining or parsing metadata.
	ErrParsingMetadata = errors.New("error parsing metadata")
	// ErrMalformedSAML indicates there is a format error on SAML callback request.
	ErrMalformedSAML = errors.New("malformed SAML request content")
	// ErrNotInAudience indicates SAML validation contains an audience related warning.
	ErrNotInAudience = errors.New("not in audience")
)

// UserData contains the basic auth data associated
// with a user obtained from SAML response.
type UserData struct {
	UserName  string `db:"username"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}

// Provider represents a component that is able to
// interact and communicate with a SAML IdP.
type Provider interface {
	BuildAuthURL(url string) (string, error)
	GetUserData(samlResp string) (UserData, error)
}

type provider struct {
	sp *saml2.SAMLServiceProvider
}

// NewProvider builds a new SAML provider.
// keyStore is the X509 keystore to use for request signing.
func NewProvider(metadataURL, issuerURL, callbackURL string, keyStore X509KeyStore) (Provider, error) {
	metadata, err := getMetadata(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsingMetadata, err)
	}

	if metadata == nil || metadata.IDPSSODescriptor == nil {
		return nil, ErrParsingMetadata
	}

	certStore, err := getCertStore(metadata)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParsingMetadata, err)
	}

	return &provider{
		sp: &saml2.SAMLServiceProvider{
			IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
			IdentityProviderIssuer:      metadata.EntityID,
			ServiceProviderIssuer:       issuerURL,
			AssertionConsumerServiceURL: callbackURL,
			SignAuthnRequests:           true,
			AudienceURI:                 callbackURL,
			IDPCertificateStore:         &certStore,
			SPKeyStore:                  keyStore,
		},
	}, nil
}

// BuildAuthURL builds an auth URL with the given redirect URL.
func (p *provider) BuildAuthURL(redirectURL string) (string, error) {
	return p.sp.BuildAuthURL(redirectURL)
}

// GetUserData returns UserData extracted from SAML response.
// ErrMalformedSAML is returned if an error happens when parsing assertion.
// ErrNotInAudience is returned if assertion's entity ID does not match the SP.
func (p *provider) GetUserData(samlResp string) (UserData, error) {
	assertionInfo, err := p.sp.RetrieveAssertionInfo(samlResp)
	if err != nil {
		return UserData{}, ErrMalformedSAML
	}

	if assertionInfo.WarningInfo.NotInAudience {
		return UserData{}, ErrNotInAudience
	}

	return UserData{
		UserName:  assertionInfo.Values.Get("Username"),
		FirstName: assertionInfo.Values.Get("FirstName"),
		LastName:  assertionInfo.Values.Get("LastName"),
		Email:     assertionInfo.Values.Get("Email"),
	}, nil

}

func getMetadata(metadataURL string) (*types.EntityDescriptor, error) {
	res, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	rawMetadata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	metadata := &types.EntityDescriptor{}
	err = xml.Unmarshal(rawMetadata, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func getCertStore(metadata *types.EntityDescriptor) (dsig.MemoryX509CertificateStore, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		for idx, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				err := fmt.Errorf("metadata certificate(%d) must not be empty", idx)
				return dsig.MemoryX509CertificateStore{}, err
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return dsig.MemoryX509CertificateStore{}, err
			}
			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return dsig.MemoryX509CertificateStore{}, err
			}
			certStore.Roots = append(certStore.Roots, idpCert)
		}
	}

	return certStore, nil
}
