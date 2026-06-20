package main

import (
	"main.go/config/app_conf"
	"main.go/udp"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	if !app_conf.TestMode {
		s, err := os.Stat("./log/")
		if err != nil {
			os.Mkdir("./log", 0755)
		} else if s.IsDir() {
			os.Mkdir("./log", 0755)
		}
	}
}

func main() {
	udp.StartServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}