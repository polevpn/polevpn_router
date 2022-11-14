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
		go kcpServer.ListenTLS(wg,
			config.Get("kcp.listen").AsStr(),
			config.Get("kcp.cert_file").AsStr(),
			config.Get("kcp.key_file").AsStr(),
		)
		elog.Infof("listen kcp %v", config.Get("kcp.listen").AsStr())
	}

	if config.Get("tls.enable").AsBool() {
		httpServer := NewTLSServer(requestHandler)
		wg.Add(1)
		go httpServer.ListenTLS(wg,
			config.Get("tls.listen").AsStr(),
			config.Get("tls.cert_file").AsStr(),
			config.Get("tls.key_file").AsStr(),
		)
		elog.Infof("listen tls %v", config.Get("tls.listen").AsStr())
	}

	wg.Wait()
}
