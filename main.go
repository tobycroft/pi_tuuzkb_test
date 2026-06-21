package main

import (
	"os"
	"os/signal"
	"syscall"

	"pi_tuuzkb_test/config/app_conf"
	"pi_tuuzkb_test/udp"
)

func init() {
	if !app_conf.TestMode {
		s, err := os.Stat("./log/")
		if err != nil {
			os.Mkdir("./log", 0755)
		} else if !s.IsDir() {
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