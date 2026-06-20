package udp

import (
	"fmt"
	"net"
	"sync"

	"github.com/bytedance/sonic"
	"main.go/config/app_conf"
)

type UdpData struct {
	Addr    *net.UDPAddr
	Message []byte
}

var (
	ServerConn   *net.UDPConn
	ReadChannel  chan UdpData
	WriteChannel chan UdpData
	AddrMap      sync.Map
)

func StartServer() {
	addr, err := net.ResolveUDPAddr("udp", app_conf.UdpPort)
	if err != nil {
		fmt.Println("ResolveUDPAddr error:", err)
		return
	}

	ServerConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("ListenUDP error:", err)
		return
	}

	ReadChannel = make(chan UdpData, 1024)
	WriteChannel = make(chan UdpData, 1024)

	fmt.Println("UDP Server started on", app_conf.UdpPort)

	go readLoop()
	go writeLoop()
	go routeLoop()
}

func readLoop() {
	buf := make([]byte, 65535)
	for {
		n, remoteAddr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ReadFromUDP error:", err)
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		ReadChannel <- UdpData{
			Addr:    remoteAddr,
			Message: data,
		}
	}
}

func writeLoop() {
	for {
		select {
		case data := <-WriteChannel:
			_, err := ServerConn.WriteToUDP(data.Message, data.Addr)
			if err != nil {
				fmt.Println("WriteToUDP error:", err)
			}
		}
	}
}

func routeLoop() {
	for {
		select {
		case data := <-ReadChannel:
			addrStr := data.Addr.String()
			if app_conf.TestMode {
				fmt.Printf("[UDP] %s: %s\n", addrStr, string(data.Message))
			}

			nd, err := sonic.Get(data.Message, "route")
			if err != nil {
				WriteChannel <- UdpData{
					Addr:    data.Addr,
					Message: []byte(`{"route":"error","code":400,"data":null,"msg":"invalid json format"}`),
				}
				continue
			}
			r, err := nd.String()
			if err != nil {
				WriteChannel <- UdpData{
					Addr:    data.Addr,
					Message: []byte(`{"route":"error","code":400,"data":null,"msg":"missing route field"}`),
				}
				continue
			}

			switch r {
			case "ping":
				AddrMap.Store(addrStr, data.Addr)
				WriteChannel <- UdpData{
					Addr:    data.Addr,
					Message: []byte(`{"route":"pong","code":0,"data":null,"msg":"pong"}`),
				}
			case "echo":
				WriteChannel <- UdpData{
					Addr:    data.Addr,
					Message: data.Message,
				}
			case "broadcast":
				msg := data.Message
				AddrMap.Range(func(key, value interface{}) bool {
					addr, ok := value.(*net.UDPAddr)
					if ok {
						WriteChannel <- UdpData{
							Addr:    addr,
							Message: msg,
						}
					}
					return true
				})
			default:
				WriteChannel <- UdpData{
					Addr:    data.Addr,
					Message: []byte(`{"route":"error","code":404,"data":null,"msg":"route not found"}`),
				}
			}
		}
	}
}

func Send(addr *net.UDPAddr, message []byte) {
	WriteChannel <- UdpData{
		Addr:    addr,
		Message: message,
	}
}

func Broadcast(message []byte) {
	AddrMap.Range(func(key, value interface{}) bool {
		addr, ok := value.(*net.UDPAddr)
		if ok {
			WriteChannel <- UdpData{
				Addr:    addr,
				Message: message,
			}
		}
		return true
	})
}