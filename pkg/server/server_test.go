package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

const (
	testResponseBodyOK = "OK returned from server"
)

func TestTCPServerSuite(t *testing.T) {
	suite.Run(t, new(TCPServerSuite))
}

type TCPServerSuite struct {
	suite.Suite
	proxyConfig *config.ProxyConfig
	server      *Server
}

func (suite *TCPServerSuite) TestServerTraffic() {
	client := http.DefaultClient
	res, err := client.Get(suite.Endpoint())

	suite.NoError(err)
	body, err := io.ReadAll(res.Body)
	suite.NoError(err)
	suite.Equal(testResponseBodyOK, string(body))
}

func (suite *TCPServerSuite) Endpoint() string {
	return "http://" + suite.server.listener.Addr().String()
}

func (suite *TCPServerSuite) SetupTest() {
	suite.proxyConfig = &config.ProxyConfig{
		SocketType: config.TCP,
		// random port
		SocketAddress: "127.0.0.1:0",
	}

	serverReady := make(chan bool)
	suite.server = &Server{
		ProxyConfig:  suite.proxyConfig,
		ProxyHandler: http.HandlerFunc(ok),
		serverReady:  serverReady,
	}
	go suite.server.Run()
	<-serverReady
}

func (suite *TCPServerSuite) TearDownTest() {
	suite.server.Stop()
}

func TestUnixServerSuite(t *testing.T) {
	suite.Run(t, new(UnixServerSuite))
}

type UnixServerSuite struct {
	suite.Suite
	proxyConfig *config.ProxyConfig
	server      *Server
}

func (suite *UnixServerSuite) TestServerTraffic() {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", suite.server.listener.Addr().String())
			},
		},
	}
	res, err := client.Get("http://foo.bar")
	suite.NoError(err)
	stat, err := os.Stat(suite.proxyConfig.SocketAddress)
	suite.NoError(err)

	body, err := io.ReadAll(res.Body)
	suite.NoError(err)
	suite.Equal(testResponseBodyOK, string(body))
	suite.Equal(os.FileMode(0700), stat.Mode()&os.ModePerm)
}

func (suite *UnixServerSuite) SetupTest() {
	file, err := os.CreateTemp("", "eks_connector_sock")
	suite.NoError(err)

	suite.proxyConfig = &config.ProxyConfig{
		SocketType: config.Unix,
		// random port
		SocketAddress: file.Name(),
	}

	serverReady := make(chan bool)
	suite.server = &Server{
		ProxyConfig:  suite.proxyConfig,
		ProxyHandler: http.HandlerFunc(ok),
		serverReady:  serverReady,
	}
	go suite.server.Run()
	<-serverReady
}

func (suite *UnixServerSuite) TearDownTest() {
	suite.server.Stop()
}

func ok(res http.ResponseWriter, req *http.Request) {
	_, _ = res.Write([]byte(testResponseBodyOK))
}
