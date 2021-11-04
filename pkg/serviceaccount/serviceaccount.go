// Package serviceaccount provides access to service account associated with eks-connector pod
package serviceaccount

import (
	"crypto/x509"
	"io/ioutil"
	"path"
)

const (
	BaseDir     = "/var/run/secrets/kubernetes.io/serviceaccount"
	FileToken   = "token"
	FileRootCAs = "ca.crt"
)

type Secret struct {
	RootCAs *x509.CertPool
	Token   string
}

type SecretProvider interface {
	Get() (*Secret, error)
}

func NewProvider() SecretProvider {
	return &mountedSecretProvider{
		baseDir: BaseDir,
	}
}

type mountedSecretProvider struct {
	baseDir string
}

func (r *mountedSecretProvider) Get() (*Secret, error) {
	caCert, err := r.rootCAs()
	if err != nil {
		return nil, err
	}

	token, err := r.token()
	if err != nil {
		return nil, err
	}

	return &Secret{
		Token:   token,
		RootCAs: caCert,
	}, nil
}

func (r *mountedSecretProvider) rootCAs() (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(path.Join(r.baseDir, FileRootCAs))
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	return pool, nil
}

func (r *mountedSecretProvider) token() (string, error) {
	token, err := ioutil.ReadFile(path.Join(r.baseDir, FileToken))
	if err != nil {
		return "", err
	}

	return string(token), nil
}
