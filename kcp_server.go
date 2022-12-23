package main

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/pion/dtls/v2"
	"github.com/polevpn/anyvalue"
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

		err = ks.auth(conn)

		if err != nil {
			elog.Error("auth fail,", err)
			conn.Close()
			continue
		}

		go ks.handleConn(conn)
	}

}

func (ks *KCPServer) auth(conn *kcp.UDPSession) error {

	pkt, err := ReadPacket(conn)

	if err != nil {
		return err
	}

	ppkt := PolePacket(pkt)

	if ppkt.Cmd() != CMD_AUTH {
		return errors.New("invalid cmd")
	}

	av, err := anyvalue.NewFromJson(ppkt.Payload())

	if err != nil {
		return err
	}

	resp := anyvalue.New()

	if av.Get("key").AsStr() != Config.Get("key").AsStr() {
		resp.Set("error", "invalid key")
	}

	body, _ := resp.EncodeJson()

	buf := make([]byte, POLE_PACKET_HEADER_LEN+len(body))
	copy(buf[POLE_PACKET_HEADER_LEN:], body)
	resppkt := PolePacket(buf)
	resppkt.SetLen(uint16(len(buf)))
	resppkt.SetCmd(CMD_AUTH)
	resppkt.SetSeq(ppkt.Seq())
	_, err = conn.Write(resppkt)
	if err != nil {
		elog.Error("write error", err)
	}

	if resp.Get("error").AsStr() != "" {
		return errors.New(resp.Get("error").AsStr())
	}

	return nil
}

func (ks *KCPServer) handleConn(conn *kcp.UDPSession) {

	kcpconn := NewKCPConn(conn, ks.requestHandler)
	if ks.requestHandler != nil {
		ks.requestHandler.OnConnection(kcpconn)
	}
	kcpconn.StartProcess()
}
