package main

import (
	"strings"
	"sync"
	"time"

	"github.com/polevpn/netstack/tcpip"
)

const (
	CONNECTION_TIMEOUT    = 1
	CHECK_TIMEOUT_INTEVAL = 5
)

type ConnMgr struct {
	route2conns map[string]Conn
	conn2routes map[string]string

	route2actives map[string]time.Time
	mutex         *sync.Mutex
}

func NewConnMgr() *ConnMgr {
	cm := &ConnMgr{
		route2conns:   make(map[string]Conn),
		mutex:         &sync.Mutex{},
		conn2routes:   make(map[string]string),
		route2actives: make(map[string]time.Time),
	}
	go cm.CheckTimeout()
	return cm
}

func (cm *ConnMgr) CheckTimeout() {
	for range time.NewTicker(time.Second * CHECK_TIMEOUT_INTEVAL).C {
		timeNow := time.Now()
		iplist := make([]string, 0)
		cm.mutex.Lock()
		for ip, lastActive := range cm.route2actives {
			if timeNow.Sub(lastActive) > time.Minute*CONNECTION_TIMEOUT {
				iplist = append(iplist, ip)

			}
		}
		cm.mutex.Unlock()

		for _, ip := range iplist {
			conn := cm.GetConnByRoute(ip)
			if conn != nil {
				cm.DetachRouteFromConn(conn)
				conn.Close(false)
			}
		}

	}
}

func (cm *ConnMgr) UpdateConnActiveTime(conn Conn) {

	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	ip, ok := cm.conn2routes[conn.String()]
	if ok {
		cm.route2actives[ip] = time.Now()
	}
}

func (cm *ConnMgr) AttachRouteToConn(route string, conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	sconn, ok := cm.route2conns[route]
	if ok {
		delete(cm.conn2routes, sconn.String())
	}
	cm.route2conns[route] = conn
	cm.conn2routes[conn.String()] = route
}

func (cm *ConnMgr) IsDetached(route string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	_, ok := cm.route2conns[route]
	return ok
}

func (cm *ConnMgr) DetachRouteFromConn(conn Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	route, ok := cm.conn2routes[conn.String()]
	if ok {
		sconn, ok := cm.route2conns[route]
		if ok && sconn.String() == conn.String() {
			delete(cm.route2conns, route)
		}
		delete(cm.conn2routes, conn.String())
	}
}

func (cm *ConnMgr) GetConnByRoute(route string) Conn {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.route2conns[route]
}

func (cm *ConnMgr) GeRouteByConn(conn Conn) string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return cm.conn2routes[conn.String()]
}

func (cm *ConnMgr) FindRoute(ip tcpip.Address) Conn {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for route, conn := range cm.route2conns {
		address := strings.Split(route, "/")
		subnet, err := tcpip.NewSubnet(tcpip.Address(address[0]), tcpip.AddressMask(address[1]))
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return conn
		}
	}
	return nil
}
