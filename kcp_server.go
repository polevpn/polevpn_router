package main

import (
	"crypto/tls"
	"net"
	"sync"

	"github.com/pion/dtls/v2"
	"github.com/polevpn/elog"
	"github.com/polevpn/kcp"
)

const (
	KCP_SHARED_KEY_LEN = 16
	KCP_MTU            = 1350
	KCP_RECV_WINDOW    = 512
	KCP_SEND_WINDOW    = 512
	KCP_READ_BUFFER    = 4194304
	KCP_WRITE_BUFFER   = 4194304
)

type KCPServer struct {
	requestHandler *RequestHandler
}

func NewKCPServer(requestHandler *RequestHandler) *KCPServer {
	return &KCPServer{
		requestHandler: requestHandler,
	}
}

func (ks *KCPServer) ListenTLS(wg *sync.WaitGroup, addr string, certFile string, keyFile string) {
	defer wg.Done()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		elog.Error("resolve address fail,", err)
		return
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		elog.Error("load cert fail,", err)
		return
	}

	// Prepare the configuration of the DTLS connection
	config := &dtls.Config{
		Certificates: []tls.Certificate{certificate},
		MTU:          1400,
	}

	listener, err := kcp.Listen(udpAddr, config)

	if err != nil {
		elog.Error("kcp listen fail,", err)
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			elog.Error("accept kcp conn fail,", err)
			continue
		}
		conn.SetMtu(KCP_MTU)
		conn.SetACKNoDelay(true)
		conn.SetStreamMode(true)
		conn.SetNoDelay(1, 10, 2, 1)
		conn.SetWindowSize(KCP_SEND_WINDOW, KCP_RECV_WINDOW)
		conn.SetReadBuffer(KCP_READ_BUFFER)
		conn.SetReadBuffer(KCP_WRITE_BUFFER)
		go ks.handleConn(conn)
	}

}

func (ks *KCPServer) handleConn(conn *kcp.UDPSession) {

	kcpconn := NewKCPConn(conn, ks.requestHandler)
	if ks.requestHandler != nil {
		ks.requestHandler.OnConnection(kcpconn)
	}
	kcpconn.StartProcess()
}
