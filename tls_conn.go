package main

import (
	"encoding/binary"
	"io"
	"net"
	"sync"

	"github.com/polevpn/elog"
)

const (
	CH_TCP_WRITE_SIZE = 2000
)

type TLSConn struct {
	conn    net.Conn
	wch     chan []byte
	closed  bool
	handler *RequestHandler
	wg      *sync.WaitGroup
}

func NewTLSConn(conn net.Conn, handler *RequestHandler) *TLSConn {
	return &TLSConn{
		conn:    conn,
		closed:  false,
		wch:     make(chan []byte, CH_TCP_WRITE_SIZE),
		handler: handler,
		wg:      &sync.WaitGroup{},
	}
}

func (kc *TLSConn) Close(flag bool) error {
	if !kc.closed {
		kc.closed = true
		if kc.wch != nil {
			kc.wch <- nil
			close(kc.wch)
		}
		err := kc.conn.Close()
		if flag {
			go kc.handler.OnClosed(kc, false)
		}
		kc.wg.Wait()
		return err
	}
	return nil
}

func (kc *TLSConn) String() string {
	return kc.conn.RemoteAddr().String() + "->" + kc.conn.LocalAddr().String()
}

func (kc *TLSConn) IsClosed() bool {
	return kc.closed
}

func (kc *TLSConn) StartProcess() {
	kc.wg.Add(2)
	go kc.read()
	go kc.write()

}

func (kc *TLSConn) read() {
	defer func() {
		kc.wg.Done()
		kc.Close(true)
	}()

	defer PanicHandler()

	for {
		prefetch := make([]byte, 2)

		_, err := io.ReadFull(kc.conn, prefetch)

		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				elog.Info(kc.String(), ",conn closed")
			} else {
				elog.Error(kc.String(), ",conn read exception:", err)
			}
			return
		}

		len := binary.BigEndian.Uint16(prefetch)

		if len < POLE_PACKET_HEADER_LEN {
			elog.Error("invalid pkt len=", len)
			continue
		}

		pkt := make([]byte, len)
		copy(pkt, prefetch)

		_, err = io.ReadFull(kc.conn, pkt[2:])

		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				elog.Info(kc.String(), ",conn closed")
			} else {
				elog.Error(kc.String(), ",conn read exception:", err)
			}
			return
		}

		kc.handler.OnRequest(pkt, kc)
	}

}

func (kc *TLSConn) write() {
	defer func() {
		kc.wg.Done()
		kc.Close(true)
	}()

	defer PanicHandler()

	for {

		pkt, ok := <-kc.wch
		if !ok {
			elog.Error(kc.String(), ",get pkt from write channel fail,maybe channel closed")
			return
		}
		if pkt == nil {
			elog.Info(kc.String(), ",exit write process")
			return
		}
		_, err := kc.conn.Write(pkt)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				elog.Info(kc.String(), ",conn closed")
			} else {
				elog.Error(kc.String(), ",conn write exception:", err)
			}
			return
		}
	}
}

func (kc *TLSConn) Send(pkt []byte) {
	if kc.closed {
		return
	}
	if kc.wch != nil {
		kc.wch <- pkt
	}
}
