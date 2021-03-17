package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
)

const (
	CH_TUNIO_WRITE_SIZE = 4096
)

var Config *anyvalue.AnyValue
var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./config.json", "config file path")
}

func signalHandler() {

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				elog.Fatal("receive exit signal,exit")
			default:
			}
		}
	}()
}

func main() {

	flag.Parse()
	defer elog.Flush()
	signalHandler()

	var err error

	Config, err = GetConfig(configPath)
	if err != nil {
		elog.Fatal("load config fail", err)
	}

	NewPoleVPNRouter().Start(Config)

}
