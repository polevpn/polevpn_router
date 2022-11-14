package main

import (
	"crypto/tls"
	"net"
	"sync"

	"github.com/polevpn/elog"
)

const (
	TCP_WRITE_BUFFER_SIZE = 5242880
	TCP_READ_BUFFER_SIZE  = 5242880
)

type TLSServer struct {
	requestHandler *RequestHandler
}

func NewTLSServer(requestHandler *RequestHandler) *TLSServer {

	return &TLSServer{requestHandler: requestHandler}
}

func (hs *TLSServer) ListenTLS(wg *sync.WaitGroup, addr string, certFile string, keyFile string) {

	defer wg.Done()

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		elog.Error("load cert fail,", err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		elog.Error("listen tls fail,", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			elog.Error("accept tls conn fail,", err)
			continue
		}
		go hs.handleConn(conn)
	}

}

func (ks *TLSServer) handleConn(conn net.Conn) {

	tlsconn := NewTLSConn(conn, ks.requestHandler)
	if ks.requestHandler != nil {
		ks.requestHandler.OnConnection(tlsconn)
	}
	tlsconn.StartProcess()
}
