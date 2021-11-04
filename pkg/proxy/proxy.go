package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"

	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/config"
	"github.com/aws/amazon-eks-connector/pkg/serviceaccount"
)

const (
	HeaderIamArn          = "x-aws-eks-identity-arn"
	HeaderUserAgent       = "User-Agent"
	HeaderAuthorization   = "Authorization"
	HeaderImpersonateUser = "Impersonate-User"

	HeaderValueUserAgent = "eks-connector/1.0"

	StatusProxyError = 502
	// MessageProxyError is the response body when there's any proxy level error occurs.
	// It is a static json so we are putting it here without needing to serializing it on every request.
	MessageProxyError = `{"status": 502, "message": "eks connector failed to proxy the request to kubernetes api. check eks connector logs for details."}`
)

type proxy struct {
	ProxyConfig    *config.ProxyConfig
	ServiceAccount serviceaccount.SecretProvider
}

func NewProxyHandler(proxyConfig *config.ProxyConfig,
	serviceAccountProvider serviceaccount.SecretProvider) http.Handler {
	return &proxy{
		ProxyConfig:    proxyConfig,
		ServiceAccount: serviceAccountProvider,
	}
}

func (p *proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	url := p.proxyUrl(req)

	reverseProxy, err := p.reverseProxy(url)
	if err != nil {
		p.proxyError(res, req, err)
	} else {
		reverseProxy.ServeHTTP(res, req)
	}
}

func (p *proxy) reverseProxy(target *url.URL) (*httputil.ReverseProxy, error) {
	secret, err := p.ServiceAccount.Get()
	if err != nil {
		return nil, err
	}
	director := func(req *http.Request) {
		// override the scheme and host.
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.RawQuery = target.RawQuery
		req.URL.Path = target.Path
		req.URL.RawPath = target.RawPath

		klog.V(2).Infof("rewritten URL to %s", req.URL)

		p.proxyHeader(req, secret)
	}
	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: secret.RootCAs,
			},
		},
		ErrorHandler: p.proxyError,
	}, nil
}

func (p *proxy) proxyError(res http.ResponseWriter, req *http.Request, err error) {
	klog.Infof("eks connector proxy encountered error: %v", err)
	res.WriteHeader(StatusProxyError)
	_, responseError := res.Write([]byte(MessageProxyError))
	if responseError != nil {
		klog.Error("eks connector proxy failed to write error response: %v", err)
	}
}

func (p *proxy) proxyHeader(req *http.Request, secret *serviceaccount.Secret) {
	// for security reasons we start with a new header map.
	originalHeader := req.Header
	req.Header = http.Header{}

	// extract iam identity from original request header
	iamIdentity := originalHeader.Get(HeaderIamArn)
	klog.V(2).Infof("requester IAM identity is %s", iamIdentity)
	req.Header.Set(HeaderImpersonateUser, iamIdentity)

	// inject ServiceAccount token to authorization
	req.Header.Set(HeaderAuthorization, "Bearer "+secret.Token)

	// common headers
	req.Header.Set(HeaderUserAgent, HeaderValueUserAgent)
}

func (p *proxy) proxyUrl(req *http.Request) *url.URL {
	url := &url.URL{
		Scheme:   p.ProxyConfig.TargetProtocol,
		Host:     p.ProxyConfig.TargetHost,
		Path:     req.URL.Path,
		RawPath:  req.URL.RawPath,
		RawQuery: req.URL.RawQuery,
	}

	klog.V(2).Infof("proxy url is %s", url)

	return url
}
