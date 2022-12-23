package main

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/polevpn/anyvalue"
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

func (ks *TLSServer) ListenTLS(wg *sync.WaitGroup, addr string, certFile string, keyFile string) {

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

		err = ks.auth(conn)

		if err != nil {
			elog.Error("auth fail,", err)
			conn.Close()
			continue
		}

		go ks.handleConn(conn)
	}

}

func (ks *TLSServer) auth(conn net.Conn) error {

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

func (ks *TLSServer) handleConn(conn net.Conn) {

	tlsconn := NewTLSConn(conn, ks.requestHandler)
	if ks.requestHandler != nil {
		ks.requestHandler.OnConnection(tlsconn)
	}
	tlsconn.StartProcess()
}
