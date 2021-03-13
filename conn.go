package main

type Conn interface {
	Read()
	Write()
	Send([]byte)
	Close(bool) error
	IsClosed() bool
	String() string
}
