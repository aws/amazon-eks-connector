package proxy

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

// mockServer is an over complicated test http server
// that generates a custom certificate chain to
// simulate kube-api endpoint as much as possible.
type mockServer struct {
	handler    http.Handler
	httpServer *httptest.Server
	rootCACert *tlsCert
	leafCert   *tlsCert

	requests []*mockServerRequest
}

func (server *mockServer) Start() {
	if server.httpServer != nil {
		panic("httpServer is already started")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/pods", func(rw http.ResponseWriter, req *http.Request) {
		server.requests = append(server.requests, &mockServerRequest{
			rawRequest: req.Clone(context.TODO()),
		})
		server.handler.ServeHTTP(rw, req)
	})
	server.rootCACert = must(generateCert("Amazon Web Services Root CA", nil, nil))
	server.leafCert = must(generateCert("Kubernetes API Server leaf cert", []string{"127.0.0.1"}, server.rootCACert))

	certChain := tls.Certificate{
		Certificate:                  [][]byte{server.leafCert.raw, server.rootCACert.raw},
		PrivateKey:                   server.leafCert.privateKey,
		Leaf:                         server.leafCert.certTemplate,
		SupportedSignatureAlgorithms: []tls.SignatureScheme{tls.ECDSAWithP521AndSHA512},
	}

	server.httpServer = httptest.NewUnstartedServer(mux)
	server.httpServer.TLS = new(tls.Config)
	server.httpServer.TLS.Certificates = []tls.Certificate{certChain}
	server.httpServer.StartTLS()
}

func (server *mockServer) Stop() {
	server.httpServer.Close()
	server.httpServer = nil
}

func (server *mockServer) ProxyConfig() *config.ProxyConfig {
	targetUrl, err := url.Parse(server.httpServer.URL)
	if err != nil {
		panic(err)
	}
	return &config.ProxyConfig{
		TargetHost:     targetUrl.Host,
		TargetProtocol: targetUrl.Scheme,
	}
}

func (server *mockServer) RootCAPool() *x509.CertPool {
	pool := x509.NewCertPool()
	cert, err := x509.ParseCertificate(server.rootCACert.raw)
	if err != nil {
		panic(err)
	}
	pool.AddCert(cert)
	return pool
}

type mockServerRequest struct {
	rawRequest *http.Request
}

func (req *mockServerRequest) HeaderCount() int {
	return len(req.rawRequest.Header)
}

func (req *mockServerRequest) Header(name string) string {
	return req.rawRequest.Header.Get(name)
}

type tlsCert struct {
	certTemplate *x509.Certificate
	privateKey   crypto.PrivateKey
	publicKey    crypto.PublicKey
	raw          []byte
}

func must(cert *tlsCert, err error) *tlsCert {
	if err != nil {
		panic(err)
	}
	return cert
}

// generateCert is mostly taken from https://golang.org/src/crypto/tls/generate_cert.go
func generateCert(subject string, hosts []string, parent *tlsCert) (*tlsCert, error) {
	isCA := parent == nil
	private, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{

		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{subject},
		},
		NotBefore: time.Now().Add(-1 * time.Minute),
		NotAfter:  time.Now().Add(1 * time.Minute),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	var parentKey crypto.PrivateKey
	var parentTemplate *x509.Certificate

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		parentTemplate = template
		parentKey = private
	} else {
		parentTemplate = parent.certTemplate
		parentKey = parent.privateKey
	}

	public := private.Public()

	raw, err := x509.CreateCertificate(rand.Reader, template, parentTemplate, public, parentKey)
	if err != nil {
		return nil, err
	}

	return &tlsCert{
		certTemplate: template,
		privateKey:   private,
		publicKey:    public,
		raw:          raw,
	}, nil
}
