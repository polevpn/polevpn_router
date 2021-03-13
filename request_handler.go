package main

import (
	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
	"github.com/polevpn/netstack/tcpip/header"
	"github.com/polevpn/netstack/tcpip/transport/icmp"
)

type RequestHandler struct {
	connmgr *ConnMgr
}

func NewRequestHandler() *RequestHandler {

	return &RequestHandler{}
}

func (r *RequestHandler) SetConnMgr(connmgr *ConnMgr) {
	r.connmgr = connmgr
}

func (r *RequestHandler) OnRequest(pkt []byte, conn Conn) {

	ppkt := PolePacket(pkt)
	switch ppkt.Cmd() {
	case CMD_ROUTER_REGISTER:
		elog.Info("received router register request from", conn.String())
		r.handleRouterRegister(ppkt, conn)
	case CMD_C2S_IPDATA:
		r.handleC2SIPData(ppkt, conn)
	case CMD_HEART_BEAT:
		//elog.Info("received heart beat request", conn.String())
		r.handleHeartBeat(ppkt, conn)
	case CMD_CLIENT_CLOSED:
		r.handleClientClose(ppkt, conn)
	default:
		elog.Error("invalid pkt cmd=", ppkt.Cmd())
	}
}

func (r *RequestHandler) OnConnection(conn Conn) {

}

func (r *RequestHandler) handleRouterRegister(pkt PolePacket, conn Conn) {

	av, err := anyvalue.NewFromJson(pkt.Payload())

	if err != nil {
		av.Set("error", err.Error())
	} else {
		r.connmgr.AttachRouteToConn(av.Get("network").AsStr(), conn)
	}

	body, _ := av.MarshalJSON()
	buf := make([]byte, POLE_PACKET_HEADER_LEN+len(body))
	copy(buf[POLE_PACKET_HEADER_LEN:], body)
	resppkt := PolePacket(buf)
	resppkt.SetLen(uint16(len(buf)))
	resppkt.SetCmd(pkt.Cmd())
	resppkt.SetSeq(pkt.Seq())
	conn.Send(resppkt)
}

func (r *RequestHandler) handleC2SIPData(pkt PolePacket, conn Conn) {

	ipv4pkg := header.IPv4(pkt)

	if ipv4pkg.Protocol() == uint8(icmp.ProtocolNumber4) {

		conn := r.connmgr.FindRoute(ipv4pkg.DestinationAddress())

		if conn == nil {
			elog.Error("can't find route for ", ipv4pkg.DestinationAddress())
			return
		}
		conn.Send(pkt)
	}
}

func (r *RequestHandler) handleHeartBeat(pkt PolePacket, conn Conn) {
	buf := make([]byte, POLE_PACKET_HEADER_LEN)
	resppkt := PolePacket(buf)
	resppkt.SetLen(POLE_PACKET_HEADER_LEN)
	resppkt.SetCmd(CMD_HEART_BEAT)
	resppkt.SetSeq(pkt.Seq())
	conn.Send(resppkt)
	r.connmgr.UpdateConnActiveTime(conn)
}

func (r *RequestHandler) handleClientClose(pkt PolePacket, conn Conn) {
	elog.Info(conn.String(), "close")
}

func (r *RequestHandler) OnClosed(conn Conn, proactive bool) {

	elog.Info("connection closed event from", conn.String())

	r.connmgr.DetachRouteFromConn(conn)
	//just process proactive close event
	if proactive {
		elog.Info(conn.String(), "proactive close")
	}

}
