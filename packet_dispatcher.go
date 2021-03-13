package main

import (
	"github.com/polevpn/elog"
	"github.com/polevpn/netstack/tcpip/header"
)

const (
	IPV4_PROTOCOL = 4
	IPV6_PROTOCOL = 6
)

type PacketDispatcher struct {
	connmgr *ConnMgr
}

func NewPacketDispatcher() *PacketDispatcher {

	return &PacketDispatcher{}
}

func (p *PacketDispatcher) SetConnMgr(connmgr *ConnMgr) {
	p.connmgr = connmgr
}

func (p *PacketDispatcher) Dispatch(pkt []byte) {

	ver := pkt[0]
	ver = ver >> 4
	if ver != IPV4_PROTOCOL {
		return
	}

	ipv4pkt := header.IPv4(pkt)

	ipaddr := ipv4pkt.DestinationAddress().To4().String()
	conn := p.connmgr.GetConnByRoute(ipaddr)
	if conn == nil {
		elog.Debug("connmgr can't find wsconn for", ipaddr)
		return
	}
	buf := make([]byte, len(pkt)+POLE_PACKET_HEADER_LEN)
	copy(buf[POLE_PACKET_HEADER_LEN:], pkt)
	resppkt := PolePacket(buf)
	resppkt.SetLen(uint16(len(buf)))
	resppkt.SetCmd(CMD_S2C_IPDATA)
	conn.Send(resppkt)

}
