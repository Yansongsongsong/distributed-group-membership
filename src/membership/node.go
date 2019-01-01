package membership

import (
	"log"
	"net"
)

type nodeAddr = string
type timestamp = int

// Node 是机器节点
type Node struct {
	// endpoint
	addr        string
	maintenance map[nodeAddr]timestamp
	listener    *net.UDPConn
}

// Action 表示 Node 可以采用的action
type Action int

// eg
// func (this State) String() string {
// 	switch this {
// 	case Running:
// 			return "Running"
// 	case Stopped:
// 			return "Stopped"
// 	default:
// 			return "Unknow"
// 	}
// }

const (
	// Join 节点加入
	Join Action = 1 << iota
	// Leave 节点退出
	Leave
)

// Message 是udp 发送的报文段内容
type Message struct {
	from    nodeAddr
	version timestamp
	action  Action
}

func (n *Node) sendMessage(addr string, msg Message) {
	udpAddr, err := ParseOneEndpoint(addr)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", addr, err)
		return
	}

	// todo 消息编码
	_, err = n.listener.WriteToUDP([]byte("hello"), udpAddr)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", udpAddr.String(), err)
	}
}

// Ping 发送一个结构体
// 包含了本机的addr 和 时间戳
func (n *Node) Ping() {

}

func (n *Node) PingReq() {

}

// Broadcast 对 Node 维护列表内的所有节点
// 发送 '移除 某节点' 或者 '添加 某节点' 的 task
func (n *Node) Broadcast() {

}

// BroadcastReceiver 根据时间戳
// 判断是否执行 '移除 某节点' 或 '添加 某节点' 的操作
func (n *Node) BroadcastReceiver() {

}

func (n *Node) responsePing() {

}
