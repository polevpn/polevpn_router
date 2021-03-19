package main

import (
	"io"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/polevpn/elog"
)

const (
	CH_WEBSOCKET_WRITE_SIZE = 2000
	TRAFFIC_LIMIT_INTERVAL  = 10
)

type WebSocketConn struct {
	conn    *websocket.Conn
	wch     chan []byte
	closed  bool
	handler *RequestHandler
	wg      *sync.WaitGroup
}

func NewWebSocketConn(conn *websocket.Conn, handler *RequestHandler) *WebSocketConn {
	return &WebSocketConn{
		conn:    conn,
		closed:  false,
		wch:     make(chan []byte, CH_WEBSOCKET_WRITE_SIZE),
		handler: handler,
		wg:      &sync.WaitGroup{},
	}
}

func (wsc *WebSocketConn) Close(flag bool) error {
	if wsc.closed == false {
		wsc.closed = true
		if wsc.wch != nil {
			wsc.wch <- nil
			close(wsc.wch)
		}
		err := wsc.conn.Close()
		if flag {
			go wsc.handler.OnClosed(wsc, false)
		}
		wsc.wg.Wait()
		return err
	}
	return nil
}

func (wsc *WebSocketConn) String() string {
	return wsc.conn.RemoteAddr().String() + "->" + wsc.conn.LocalAddr().String()
}

func (wsc *WebSocketConn) IsClosed() bool {
	return wsc.closed
}

func (wsc *WebSocketConn) StartProcess() {
	wsc.wg.Add(2)
	go wsc.read()
	go wsc.write()

}

func (wsc *WebSocketConn) read() {
	defer func() {
		wsc.wg.Done()
		wsc.Close(true)
	}()

	defer PanicHandler()

	for {
		mtype, pkt, err := wsc.conn.ReadMessage()
		if err != nil {
			if err == io.ErrUnexpectedEOF || err == io.EOF {
				elog.Info(wsc.String(), "conn closed")
			} else {
				elog.Error(wsc.String(), "conn read exception:", err)
			}
			return
		}
		if mtype == websocket.BinaryMessage {
			wsc.handler.OnRequest(pkt, wsc)
		}
	}

}

func (wsc *WebSocketConn) write() {

	defer func() {
		wsc.wg.Done()
		wsc.Close(true)
	}()

	defer PanicHandler()

	for {

		pkt, ok := <-wsc.wch
		if !ok {
			elog.Error(wsc.String(), "get pkt from write channel fail,maybe channel closed")
			return
		}
		if pkt == nil {
			elog.Info(wsc.String(), "exit write process")
			return
		}

		err := wsc.conn.WriteMessage(websocket.BinaryMessage, pkt)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				elog.Info(wsc.String(), "conn closed")
			} else {
				elog.Error(wsc.String(), "conn write exception:", err)
			}
			return
		}
	}
}

func (wsc *WebSocketConn) Send(pkt []byte) {
	if wsc.closed {
		return
	}
	if wsc.wch != nil {
		wsc.wch <- pkt
	}
}
