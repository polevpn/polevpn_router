package main

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/polevpn/elog"
)

const (
	CONNECTION_TIMEOUT    = 1
	CHECK_TIMEOUT_INTEVAL = 5
)

type ConnMgr struct {
	route2conns   map[string]Conn
	gateway2conns map[string]Conn
	conns         map[string]Conn
	connActives   map[string]time.Time
	mutex         *sync.RWMutex
}

func NewConnMgr() *ConnMgr {
	cm := &ConnMgr{
		route2conns:   make(map[string]Conn),
		gateway2conns: make(map[string]Conn),
		mutex:         &sync.RWMutex{},
		conns:         make(map[string]Conn),
		connActives:   make(map[string]time.Time),
	}
	go cm.CheckTimeout()
	return cm
}

func (cm *ConnMgr) CheckTimeout() {
	for range time.NewTicker(time.Second * CHECK_TIMEOUT_INTEVAL).C {
		timeNow := time.Now()
		idlist := make([]string, 0)
		cm.mutex.RLock()
		for connid, lastActive := range cm.connActives {
			if timeNow.Sub(lastActive) > time.Minute*CONNECTION_TIMEOUT {
				idlist = append(idlist, connid)

			}
		}
		cm.mutex.RUnlock()

		for _, connid := range idlist {
			conn := cm.GetConnById(connid)
			if conn != nil {
				conn.Close(true)
				cm.RemoveConnById(connid)
				elog.Info("conn " + conn.String() + " have't received heartbeat for more than " + strconv.Itoa(CONNECTION_TIMEOUT) + " min")
			}
		}

	}
}

func (cm *ConnMgr) UpdateConnActiveTime(conn Conn) {

	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	conn, ok := cm.conns[conn.String()]
	if ok {
		cm.connActives[conn.String()] = time.Now()
	}
}

func (cm *ConnMgr) AttachRouteToConn(route string, conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.route2conns[route] = conn
}

func (cm *ConnMgr) AttachGatewayToConn(gateway string, conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.gateway2conns[gateway] = conn
}

func (cm *ConnMgr) IsDetached(route string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	_, ok := cm.route2conns[route]
	return ok
}

func (cm *ConnMgr) DetachRouteFromConn(conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	routeList := make([]string, 0)
	for route, item := range cm.route2conns {
		if item.String() == conn.String() {
			routeList = append(routeList, route)
		}
	}

	for _, route := range routeList {
		delete(cm.route2conns, route)
	}
}

func (cm *ConnMgr) DetachGatewayFromConn(conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	gatewayList := make([]string, 0)
	for gateway, item := range cm.gateway2conns {
		if item.String() == conn.String() {
			gatewayList = append(gatewayList, gateway)
		}
	}

	for _, gateway := range gatewayList {
		delete(cm.gateway2conns, gateway)
	}
}

func (cm *ConnMgr) GetConnByRoute(route string) Conn {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.route2conns[route]
}

func (cm *ConnMgr) SetConnById(id string, conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.conns[id] = conn
	cm.connActives[id] = time.Now()
}

func (cm *ConnMgr) RemoveConnById(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.conns, id)
	delete(cm.connActives, id)
}

func (cm *ConnMgr) GetConnById(id string) Conn {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.conns[id]
}

func (cm *ConnMgr) GetConnByGateway(gateway string) Conn {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.gateway2conns[gateway]
}

func (cm *ConnMgr) FindRoute(ip net.IP) Conn {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var defaultRouteConn Conn
	for route, conn := range cm.route2conns {

		_, subnet, err := net.ParseCIDR(route)

		if err != nil {
			continue
		}
		find := subnet.Contains(ip)
		if route == "0.0.0.0/0" {
			defaultRouteConn = conn
		} else if find && (route != "0.0.0.0/0") {
			return conn
		}
	}
	return defaultRouteConn
}
