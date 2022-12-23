package main

import (
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

		pkt, err := ReadPacket(kc.conn)

		if err != nil {
			elog.Error(kc.String(), " read packet end status=", err)
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
			elog.Error(kc.String(), ",channel closed")
			return
		}
		if pkt == nil {
			elog.Info(kc.String(), ",exit write process")
			return
		}
		_, err := kc.conn.Write(pkt)
		if err != nil {
			elog.Error(kc.String(), ",conn write packet end status=", err)
			return
		}
	}
}

func (kc *TLSConn) Send(pkt []byte) {
	if kc.closed {
		return
	}

	if kc.wch != nil {

		select {
		case kc.wch <- pkt:
		default:
			elog.Error(kc.String(), " wch is full")
		}
	}
}
