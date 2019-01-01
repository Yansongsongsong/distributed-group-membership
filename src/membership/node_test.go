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

func TestUpdateNodeVersion(t *testing.T) {
	node.maintenance["node1"] = 0
	node.maintenance["node2"] = 1
	node.maintenance["node3"] = 2

	node.updateNodeVersion(Message{"node4", 2019, Join})
	if node.maintenance["node4"] != 2019 {
		t.Fatal(`wrong: node.maintenance["node4"] != 2019 `)
	}

	node.updateNodeVersion(Message{"node3", 1, Join})
	if node.maintenance["node3"] != 2 {
		t.Fatal(`wrong: node.maintenance["node4"] != 2 `)
	}

	node.updateNodeVersion(Message{"node1", 1, Leave})
	if _, ok := node.maintenance["node1"]; ok {
		t.Fatal("wrong: node1 should be delete")
	}

	node.updateNodeVersion(Message{"node2", 0, Leave})
	if node.maintenance["node2"] != 1 {
		t.Fatal(`wrong: node.maintenance["node2"] != 1 `)
	}
}

func TestBroadcast(t *testing.T) {
	node.maintenance["192.168.70.30:9982"] = 1
	node.maintenance["192.168.70.30:9983"] = 1
	node.maintenance["192.168.70.30:9984"] = 1
	node.maintenance["192.168.70.30:9985"] = 1

	node.Broadcast(Message{"192.168.70.30:9981", 2019, Join})
}
