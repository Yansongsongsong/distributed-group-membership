package membership

import (
	"net"
	"testing"
)

var (
	listener, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("192.168.70.30"), Port: 9981})
	node          = Node{
		addr:        "192.168.70.30:9981",
		maintenance: make(map[string]int),
		listener:    listener,
	}
)

func TestSendMessage(t *testing.T) {
	msg := Message{"192.168.70.30:9981", 2019, Join}
	node.sendMessage("192.168.70.30:9980", msg)
}
