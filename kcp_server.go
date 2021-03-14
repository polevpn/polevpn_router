package main

import (
	"github.com/polevpn/kcp-go/v5"
)

const (
	KCP_MTU          = 1350
	KCP_RECV_WINDOW  = 512
	KCP_SEND_WINDOW  = 512
	KCP_READ_BUFFER  = 4194304
	KCP_WRITE_BUFFER = 4194304
)

var KCP_KEY = []byte{0x17, 0xef, 0xad, 0x3b, 0x12, 0xed, 0xfa, 0xc9, 0xd7, 0x54, 0x14, 0x5b, 0x3a, 0x4f, 0xb5, 0xf6}

type KCPServer struct {
	requestHandler *RequestHandler
}

func NewKCPServer(requestHandler *RequestHandler) *KCPServer {
	return &KCPServer{
		requestHandler: requestHandler,
	}
}

func (ks *KCPServer) Listen(addr string, sharedKey string) error {
	block, _ := kcp.NewAESBlockCrypt(KCP_KEY)
	if listener, err := kcp.ListenWithOptions(addr, block, 10, 3); err == nil {
		for {
			conn, err := listener.AcceptKCP()
			if err != nil {
				return err
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
		return err
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
