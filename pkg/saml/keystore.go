/*
Copyright 2021 Adevinta
*/

package saml

import (
	"crypto/rsa"

	dsig "github.com/russellhaering/goxmldsig"
)

// TODO: support custom keystore.

// X509KeyStore represents an X509 keystore.
type X509KeyStore interface {
	GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error)
}

// RandomKeyStore is a X509KeyStore which generates a
// new random private key and certificate from it.
// This is acceptable for many IdPs as they
// often do not verify request signatures (e.g.: Okta)
type RandomKeyStore struct {
	privKey *rsa.PrivateKey
	cert    []byte
}

// NewRandomKeyStore builds a new RandomKeyStore.
func NewRandomKeyStore() *RandomKeyStore {
	privKey, cert, _ := dsig.RandomKeyStoreForTest().GetKeyPair() // nolint
	return &RandomKeyStore{
		privKey: privKey,
		cert:    cert,
	}
}

// GetKeyPair returns the keystore private key and certificate.
func (s *RandomKeyStore) GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error) {
	return s.privKey, s.cert, nil
}
