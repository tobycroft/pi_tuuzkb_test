package app_conf

import (
	"fmt"
)

var Project = "pi_tuuzkb_test"
var TestMode = true
var UdpPort = ":6666"

func init() {
	_ready()
}

func _ready() {
	fmt.Println("app_ready")
}