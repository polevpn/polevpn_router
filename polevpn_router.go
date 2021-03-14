package main

import (
	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
)

type PoleVPNRouter struct {
}

func NewPoleVPNRouter() *PoleVPNRouter {
	return &PoleVPNRouter{}
}

func (ps *PoleVPNRouter) Start(config *anyvalue.AnyValue) error {

	kcpServer := NewKCPServer(NewRequestHandler(NewConnMgr()))

	elog.Infof("listen kcp %v", config.Get("listen").AsStr())

	err := kcpServer.Listen(config.Get("listen").AsStr(), config.Get("shared_key").AsStr())

	return err
}
