package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/polevpn/elog"
)

const (
	TCP_WRITE_BUFFER_SIZE = 5242880
	TCP_READ_BUFFER_SIZE  = 5242880
)

type HttpServer struct {
	requestHandler *RequestHandler
	upgrader       *websocket.Upgrader
	sharedKey      string
}

func NewHttpServer(requestHandler *RequestHandler) *HttpServer {

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}

	return &HttpServer{requestHandler: requestHandler, upgrader: upgrader}
}

func (hs *HttpServer) defaultHandler(w http.ResponseWriter, r *http.Request) {
	hs.respError(http.StatusForbidden, w)
}

func (hs *HttpServer) ListenTLS(wg *sync.WaitGroup, addr string, sharedKey, certFile string, keyFile string) {

	defer wg.Done()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			hs.wsHandler(w, r)
		}
	})
	hs.sharedKey = sharedKey
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	elog.Error(server.ListenAndServeTLS(certFile, keyFile))
}

func (hs *HttpServer) respError(status int, w http.ResponseWriter) {
	if status == http.StatusBadRequest {
		w.Header().Add("Server", "nginx/1.10.3")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("<html>\n<head><title>400 Bad Request</title></head>\n<body bgcolor=\"white\">\n<center><h1>400 Bad Request</h1></center>\n<hr><center>nginx/1.10.3</center>\n</body>\n</html>"))
	} else if status == http.StatusForbidden {
		w.Header().Add("Server", "nginx/1.10.3")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("<html>\n<head><title>403 Forbidden</title></head>\n<body bgcolor=\"white\">\n<center><h1>403 Forbidden</h1></center>\n<hr><center>nginx/1.10.3</center>\n</body>\n</html>"))

	}
}

func (hs *HttpServer) wsHandler(w http.ResponseWriter, r *http.Request) {

	defer PanicHandler()

	sharedKey := r.URL.Query().Get("shared_key")

	if hs.sharedKey != sharedKey {
		elog.Errorf("shardkey:%v verify fail", sharedKey)
		hs.respError(http.StatusForbidden, w)
		return
	}

	conn, err := hs.upgrader.Upgrade(w, r, nil)
	if err != nil {
		elog.Error("upgrade http request to ws fail", err)
		return
	}

	elog.Info("accpet new ws conn", conn.RemoteAddr().String())

	if hs.requestHandler != nil {
		wsconn := NewWebSocketConn(conn, hs.requestHandler)
		hs.requestHandler.OnConnection(wsconn)
		go wsconn.Read()
		go wsconn.Write()
	} else {
		elog.Error("ws conn handler haven't set")
		conn.Close()
	}

}
