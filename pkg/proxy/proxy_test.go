package proxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/serviceaccount"
)

const (
	testServiceAccountToken = "rUVGEcNVnKg84iTob13n"
	testIAMIdentity         = "arn:aws:iam:123456789012::role/coder"
	testHttpResponse        = "OOMKill"
	testOriginalUserAgent   = "java/11"
	testCustomRequestHeader = "x-header-will-not-forward"
	testCustomQueryString   = "next=dUKQYLVdXnJlYjr386XA"
)

func TestProxySuite(t *testing.T) {
	suite.Run(t, new(ProxySuite))
}

type ProxySuite struct {
	suite.Suite

	secretProvider *serviceaccount.MockSecretProvider
	targetServer   *mockServer
	proxyHandler   http.Handler
}

func (suite *ProxySuite) SetupTest() {
	suite.secretProvider = &serviceaccount.MockSecretProvider{}
	suite.targetServer = &mockServer{}
	suite.targetServer.Start()
	suite.proxyHandler = NewProxyHandler(
		suite.targetServer.ProxyConfig(),
		suite.secretProvider,
	)
}

func (suite *ProxySuite) TearDownTest() {
	suite.targetServer.Stop()
}

func (suite *ProxySuite) TestServeHTTPHappyCase() {
	// prepare
	response := httptest.NewRecorder()
	requestUrl := fmt.Sprintf("http://foo-bar:12345/api/v1/pods?%s", testCustomQueryString)
	request := httptest.NewRequest("GET", requestUrl, nil)
	request.Header.Set(HeaderIamArn, testIAMIdentity)
	request.Header.Set(HeaderUserAgent, testOriginalUserAgent)
	request.Header.Set(testCustomRequestHeader, "this-header-will-not-be-forwarded")
	suite.secretProvider.On("Get").Return(&serviceaccount.Secret{
		Token:   testServiceAccountToken,
		RootCAs: suite.targetServer.RootCAPool(),
	}, nil)
	suite.targetServer.handler = newTextHandler(testHttpResponse)

	// test
	suite.proxyHandler.ServeHTTP(response, request)

	// verify
	suite.Len(suite.targetServer.requests, 1)
	proxyRequest := suite.targetServer.requests[0]
	suite.Equal(5, proxyRequest.HeaderCount())
	suite.Equal(bearer(testServiceAccountToken), proxyRequest.Header(HeaderAuthorization))
	suite.Equal(testIAMIdentity, proxyRequest.Header(HeaderImpersonateUser))
	suite.Equal(HeaderValueUserAgent, proxyRequest.Header(HeaderUserAgent))
	suite.Empty(proxyRequest.Header(testCustomRequestHeader), "custom header is not forwarded")
	suite.Equal(testCustomQueryString, proxyRequest.rawRequest.URL.RawQuery, "custom query string is forwarded")
	suite.Equal(200, response.Code)
	body, err := io.ReadAll(response.Body)
	suite.NoError(err)
	suite.Equal(testHttpResponse, string(body))
	suite.secretProvider.AssertExpectations(suite.T())
}

func (suite *ProxySuite) TestServeHTTPBadCertificate() {
	// prepare
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://foo-bar:12345/api/v1/pods", nil)
	request.Header.Set(HeaderIamArn, testIAMIdentity)
	suite.secretProvider.On("Get").Return(&serviceaccount.Secret{
		Token: testServiceAccountToken,
		// RootCAs is omitted.
	}, nil)

	// test
	suite.proxyHandler.ServeHTTP(response, request)

	// verify
	suite.Len(suite.targetServer.requests, 0, "request should not hit targetServer due to certificate error")
	suite.Equal(502, response.Code, "proxy error response should be returned")
}

func (suite *ProxySuite) TestServeHTTPSecretProviderError() {
	// prepare
	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://foo-bar:12345/api/v1/pods", nil)
	request.Header.Set(HeaderIamArn, testIAMIdentity)
	suite.secretProvider.On("Get").Return(nil, errors.New("provider error"))

	// test
	suite.proxyHandler.ServeHTTP(response, request)

	// verify
	suite.Len(suite.targetServer.requests, 0)
	suite.Equal(502, response.Code)
}

func newTextHandler(response string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
	})
}

func bearer(token string) string {
	return "Bearer " + token
}
