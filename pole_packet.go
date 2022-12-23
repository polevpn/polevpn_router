package main

import "encoding/binary"

const (
	CMD_AUTH       = 0x0
	CMD_REGISTER   = 0x1
	CMD_S2C_IPDATA = 0x2
	CMD_C2S_IPDATA = 0x3
	CMD_HEART_BEAT = 0x4
)

const (
	POLE_PACKET_HEADER_LEN = 8
)

type PolePacket []byte

func (p PolePacket) Len() uint16 {
	return binary.BigEndian.Uint16(p[0:2])
}

func (p PolePacket) Cmd() uint16 {
	return binary.BigEndian.Uint16(p[2:4])
}

func (p PolePacket) Seq() uint32 {
	return binary.BigEndian.Uint32(p[4:8])
}

func (p PolePacket) Payload() []byte {
	return p[8:]
}

func (p PolePacket) SetLen(len uint16) {
	binary.BigEndian.PutUint16(p[0:], len)
}

func (p PolePacket) SetCmd(cmd uint16) {
	binary.BigEndian.PutUint16(p[2:], cmd)
}
func (p PolePacket) SetSeq(seq uint32) {
	binary.BigEndian.PutUint32(p[4:], seq)
}
