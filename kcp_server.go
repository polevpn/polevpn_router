package main

import (
	"sync"

	"github.com/polevpn/elog"
	"github.com/polevpn/kcp-go/v5"
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

func (ks *KCPServer) Listen(wg *sync.WaitGroup, addr string, sharedKey string) {
	defer wg.Done()
	if len(sharedKey) != KCP_SHARED_KEY_LEN {
		elog.Error("share key must be 16")
		return
	}
	block, _ := kcp.NewAESBlockCrypt([]byte(sharedKey))
	if listener, err := kcp.ListenWithOptions(addr, block, 10, 3); err == nil {
		for {
			conn, err := listener.AcceptKCP()
			if err != nil {
				elog.Error(err)
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
	} else {
		elog.Error(err)
	}

}

func (ks *KCPServer) handleConn(conn *kcp.UDPSession) {

	kcpconn := NewKCPConn(conn, ks.requestHandler)
	if ks.requestHandler != nil {
		ks.requestHandler.OnConnection(kcpconn)
	}
	go kcpconn.Read()
	go kcpconn.Write()
}
