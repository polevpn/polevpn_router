package main

import (
	"net"

	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
	"github.com/polevpn/netstack/tcpip/header"
)

type RequestHandler struct {
	connmgr *ConnMgr
}

func NewRequestHandler(connmgr *ConnMgr) *RequestHandler {

	return &RequestHandler{
		connmgr: connmgr,
	}
}

func (r *RequestHandler) OnRequest(pkt []byte, conn Conn) {

	ppkt := PolePacket(pkt)
	switch ppkt.Cmd() {
	case CMD_ROUTE_REGISTER:
		r.handleRouteRegister(ppkt, conn)
	case CMD_C2S_IPDATA:
		r.handleC2SIPData(ppkt, conn)
	case CMD_HEART_BEAT:
		r.handleHeartBeat(ppkt, conn)
	case CMD_CLIENT_CLOSED:
		r.handleClientClose(ppkt, conn)
	default:
		elog.Error("invalid pkt cmd=", ppkt.Cmd())
	}
}

func (r *RequestHandler) OnConnection(conn Conn) {
	elog.Info("accpet new conn ", conn.String())
	r.connmgr.SetConnById(conn.String(), conn)
}

func (r *RequestHandler) handleRouteRegister(pkt PolePacket, conn Conn) {

	elog.Info("received route register request from", conn.String())

	req, err := anyvalue.NewFromJson(pkt.Payload())

	resp := anyvalue.New()

	if err != nil {
		resp.Set("error", err.Error())
	} else {

		oldConn := r.connmgr.GetConnByGateway(req.Get("gateway").AsStr())
		if oldConn != nil {
			oldConn.Close(true)
		}

		if req.Get("network").IsStr() {
			elog.Info("register route ", req.Get("network").AsStr())
			r.connmgr.AttachRouteToConn(req.Get("network").AsStr(), conn)
		} else if req.Get("network").IsArray() {
			networks := req.Get("network").AsStrArr()
			elog.Info("register route ", networks)
			for _, network := range networks {
				r.connmgr.AttachRouteToConn(network, conn)
			}
		}
		elog.Info("register gateway ", req.Get("gateway").AsStr())
		r.connmgr.AttachGatewayToConn(req.Get("gateway").AsStr(), conn)

		r.connmgr.UpdateConnActiveTime(conn)

	}

	body, _ := resp.MarshalJSON()
	buf := make([]byte, POLE_PACKET_HEADER_LEN+len(body))
	copy(buf[POLE_PACKET_HEADER_LEN:], body)
	resppkt := PolePacket(buf)
	resppkt.SetLen(uint16(len(buf)))
	resppkt.SetCmd(pkt.Cmd())
	resppkt.SetSeq(pkt.Seq())
	conn.Send(resppkt)
}

func (r *RequestHandler) handleC2SIPData(pkt PolePacket, conn Conn) {

	ipv4pkg := header.IPv4(pkt.Payload())

	dstIp := ipv4pkg.DestinationAddress().To4()
	dstIpStr := dstIp.String()

	srcIp := ipv4pkg.SourceAddress().To4()

	elog.Debug("received pkt to ", dstIpStr)

	toconn := r.connmgr.GetConnByGateway(dstIpStr)

	if toconn == nil {
		toconn = r.connmgr.FindRoute(net.IP(dstIp))
	}

	if toconn == nil {
		elog.Debugf("can't find route from %v to %v", srcIp, dstIp)
		return
	}
	pkt.SetCmd(CMD_S2C_IPDATA)
	toconn.Send(pkt)

}

func (r *RequestHandler) handleHeartBeat(pkt PolePacket, conn Conn) {
	elog.Debug("received heartbeat request", conn.String())
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
	r.connmgr.DetachGatewayFromConn(conn)
	//just process proactive close event
	if proactive {
		elog.Info(conn.String(), "proactive close")
	}

}
