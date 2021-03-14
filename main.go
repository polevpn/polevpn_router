package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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
			case syscall.SIGUSR1:
			case syscall.SIGUSR2:
			default:
			}
		}
	}()
}

func main() {

	flag.Parse()
	defer elog.Flush()
	signalHandler()

	go func() {
		for range time.NewTicker(time.Minute).C {
			m := runtime.MemStats{}
			runtime.ReadMemStats(&m)
			elog.Printf("memory=%v,goroutines=%v", m.HeapAlloc, runtime.NumGoroutine())
		}
	}()

	var err error

	Config, err = GetConfig(configPath)
	if err != nil {
		elog.Fatal("load config fail", err)
	}

	NewPoleVPNRouter().Start(Config)

}
