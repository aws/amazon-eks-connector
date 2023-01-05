package server

import (
	"net"
)

func NewTcpListener(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}
