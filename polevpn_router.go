package main

import (
	"sync"

	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
)

type PoleVPNRouter struct {
}

func NewPoleVPNRouter() *PoleVPNRouter {
	return &PoleVPNRouter{}
}

func (ps *PoleVPNRouter) Start(config *anyvalue.AnyValue) {

	wg := &sync.WaitGroup{}

	requestHandler := NewRequestHandler(NewConnMgr())

	if config.Get("kcp.enable").AsBool() {
		kcpServer := NewKCPServer(requestHandler)
		wg.Add(1)
		go kcpServer.Listen(wg, config.Get("kcp.listen").AsStr(), config.Get("shared_key").AsStr())
		elog.Infof("listen kcp %v", config.Get("kcp.listen").AsStr())
	}

	if config.Get("wss.enable").AsBool() {
		httpServer := NewHttpServer(requestHandler)
		wg.Add(1)
		go httpServer.ListenTLS(wg,
			config.Get("wss.listen").AsStr(),
			config.Get("shared_key").AsStr(),
			config.Get("wss.cert_file").AsStr(),
			config.Get("wss.key_file").AsStr(),
		)
		elog.Infof("listen wss %v", config.Get("wss.listen").AsStr())
	}

	wg.Wait()
}
