package main

import (
	"encoding/binary"
	"io"

	"github.com/polevpn/elog"
	"github.com/xtaci/kcp-go/v5"
)

const (
	CH_KCP_WRITE_SIZE = 2000
)

type KCPConn struct {
	conn    *kcp.UDPSession
	wch     chan []byte
	closed  bool
	handler *RequestHandler
}

func NewKCPConn(conn *kcp.UDPSession, handler *RequestHandler) *KCPConn {
	return &KCPConn{
		conn:    conn,
		closed:  false,
		wch:     make(chan []byte, CH_KCP_WRITE_SIZE),
		handler: handler,
	}
}

func (kc *KCPConn) Close(flag bool) error {
	if kc.closed == false {
		kc.closed = true
		if kc.wch != nil {
			kc.wch <- nil
			close(kc.wch)
		}
		err := kc.conn.Close()
		if flag {
			go kc.handler.OnClosed(kc, false)
		}
		return err
	}
	return nil
}

func (kc *KCPConn) String() string {
	return kc.conn.RemoteAddr().String() + "->" + kc.conn.LocalAddr().String()
}

func (kc *KCPConn) IsClosed() bool {
	return kc.closed
}

func (kc *KCPConn) Read() {
	defer func() {
		kc.Close(true)
	}()

	defer PanicHandler()

	for {
		var preOffset = 0
		prefetch := make([]byte, 2)

		for {
			n, err := kc.conn.Read(prefetch[preOffset:])
			if err != nil {
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					elog.Info(kc.String(), "conn closed")
				} else {
					elog.Error(kc.String(), "conn read exception:", err)
				}
				return
			}
			preOffset += n
			if preOffset >= 2 {
				break
			}
		}

		len := binary.BigEndian.Uint16(prefetch)

		if len < POLE_PACKET_HEADER_LEN {
			elog.Error("invalid pkt len=", len)
			continue
		}

		pkt := make([]byte, len)
		copy(pkt, prefetch)
		var offset uint16 = 2
		for {
			n, err := kc.conn.Read(pkt[offset:])
			if err != nil {
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					elog.Info(kc.String(), "conn closed")
				} else {
					elog.Error(kc.String(), "conn read exception:", err)
				}
				return
			}
			offset += uint16(n)
			if offset >= len {
				break
			}
		}
		kc.handler.OnRequest(pkt, kc)
	}

}

func (kc *KCPConn) Write() {

	defer PanicHandler()

	for {

		pkt, ok := <-kc.wch
		if !ok {
			elog.Error(kc.String(), "get pkt from write channel fail,maybe channel closed")
			return
		}
		if pkt == nil {
			elog.Info(kc.String(), "exit write process")
			return
		}
		_, err := kc.conn.Write(pkt)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				elog.Info(kc.String(), "conn closed")
			} else {
				elog.Error(kc.String(), "conn write exception:", err)
			}
			return
		}
	}
}

func (kc *KCPConn) Send(pkt []byte) {
	if kc.closed == true {
		return
	}
	if kc.wch != nil {
		kc.wch <- pkt
	}
}
