// Package main 提供一个CLI界面
package main

import (
	"fmt"
	"membership"
	"net"
	"os"
	"strconv"
)

func main() {
	p, _ := strconv.Atoi(os.Args[1])
	listener, err := net.ListenUDP(
		"udp",
		&net.UDPAddr{IP: net.ParseIP("192.168.70.30"), Port: p})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Local: <%s> \n", listener.LocalAddr().String())
	data := make([]byte, 1024)

	for index := 0; index < 100; index++ {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		fmt.Printf("<%s> %v\n", remoteAddr, membership.DecodeMessage(data[:n]))
	}
}
