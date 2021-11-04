package server

import (
	"errors"
	"net"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

func NewListener(proxyConfig *config.ProxyConfig) (net.Listener, error) {
	switch proxyConfig.SocketType {
	case config.TCP:
		return NewTcpListener(proxyConfig.SocketAddress)
	case config.Unix:
		return NewUnixListener(proxyConfig.SocketAddress)
	default:
		return nil, errors.New("unrecognized socket type: " + string(proxyConfig.SocketType))
	}
}
