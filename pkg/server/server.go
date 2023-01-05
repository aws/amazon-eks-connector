package server

import (
	"log"
	"net"
	"net/http"
	"os"

	"k8s.io/klog/v2"

	"github.com/aws/amazon-eks-connector/pkg/config"
)

// Server for eks connector proxy
type Server struct {
	ProxyConfig  *config.ProxyConfig
	ProxyHandler http.Handler

	// private fields visible for testing
	httpServer  *http.Server
	listener    net.Listener
	serverReady chan bool
}

func (s *Server) Run() {
	proxyListener, err := NewListener(s.ProxyConfig)
	if err != nil {
		klog.Fatalf("could not start listener on %v: %v", s.ProxyConfig, err)
	}
	defer proxyListener.Close()
	s.listener = proxyListener

	klog.Infof("listening on %v", s.ProxyConfig)
	if s.serverReady != nil {
		s.serverReady <- true
		klog.Infof("notified serverReady channel for readiness")
	}

	s.httpServer = &http.Server{
		ErrorLog: log.New(os.Stdout, "[ProxyServer] ", 0),
		Handler:  s.createHandler(),
	}

	err = s.httpServer.Serve(proxyListener)
	if err != http.ErrServerClosed {
		klog.Fatalf("Proxy server exited unexpectedly: %v", err)
	} else {
		klog.Infof("Proxy server exited gracefully")
	}
}

func (s *Server) Stop() {
	s.httpServer.Close()
	s.httpServer = nil
	s.listener = nil
}

func (s *Server) createHandler() http.Handler {
	mux := &http.ServeMux{}
	mux.Handle("/", s.ProxyHandler)

	return mux
}
