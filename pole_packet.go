package main

import "encoding/binary"

const (
	CMD_ROUTER_REGISTER = 0x1
	CMD_S2C_IPDATA      = 0x2
	CMD_C2S_IPDATA      = 0x3
	CMD_HEART_BEAT      = 0x4
	CMD_CLIENT_CLOSED   = 0x5
	CMD_KICK_OUT        = 0x6
)

const (
	POLE_PACKET_HEADER_LEN = 20
)

type PolePacket []byte

func (p PolePacket) Len() uint16 {
	return binary.BigEndian.Uint16(p[0:2])
}

func (p PolePacket) Version() uint16 {
	return binary.BigEndian.Uint16(p[2:4])
}

func (p PolePacket) Cmd() uint16 {
	return binary.BigEndian.Uint16(p[4:6])
}

func (p PolePacket) Seq() uint32 {
	return binary.BigEndian.Uint32(p[6:10])
}

func (p PolePacket) Reserve() []byte {
	return p[10:20]
}

func (p PolePacket) Payload() []byte {
	return p[20:]
}

func (p PolePacket) SetLen(len uint16) {
	binary.BigEndian.PutUint16(p[0:], len)
}

func (p PolePacket) SetVersion(version uint16) {
	binary.BigEndian.PutUint16(p[2:], version)
}

func (p PolePacket) SetCmd(cmd uint16) {
	binary.BigEndian.PutUint16(p[4:], cmd)
}
func (p PolePacket) SetSeq(seq uint32) {
	binary.BigEndian.PutUint32(p[6:], seq)
}

func (p PolePacket) SetReserve(reserve []byte) {
	copy(p[10:20], reserve)
}
