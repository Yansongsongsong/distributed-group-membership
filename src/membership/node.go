package membership

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type nodeAddr = string
type timestamp = int

// Node 是机器节点
type Node struct {
	// endpoint
	addr        string
	maintenance map[nodeAddr]timestamp
	listener    *net.UDPConn
	// FaultsDetectTime 故障检测周期 秒
	FaultsDetectTimer *time.Timer
	// PingTimer ping方法等待时间 秒
	PingTimer *time.Timer
	// T 故障检测周期 秒
	T int
	// PingT ping方法等待时间 秒
	PingT int
}

var (
	mutex sync.Mutex
)

func (n *Node) sendMessage(addr string, msg Message) {
	udpAddr, err := ParseOneEndpoint(addr)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", addr, err)
		return
	}
	data := EncodeMessage(msg)

	_, err = n.listener.WriteToUDP(data.Bytes(), udpAddr)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", udpAddr.String(), err)
	}
}

// FaultsDetect 故障检测
// 	- 故障检测 组件
// 		1. 故障检测周期 T 内，每个 NodeA 从自己保持节点列表里随机选择 1 个节点(NodeB)发送 ping 指令
//			- ping 携带 NodeA 的 addr 与 时间戳
//			- 接受到 ping 的 NodeB
//				1. 如果这个 ping 来自新节点，则增加至自己的维护列表，并触发广播组件
//				2. 如果这个 ping 来自新节点，则更新自己的维护列表
// 		2. 在时间 t 内
//			2.1 没有收到响应 ack 时
//				2.1.1 Node 从自己保持节点列表里随机选择 k 个节点发送 ping-req(NodeB)
//				2.1.2 k 个节点收到 ping-req(NodeB) 后，向 NodeB 发送 ping 指令
//					2.1.2.1 当有节点收到 NodeB 返回的 ack 时，该节点将 ack 返回给 NodeA
//					2.1.2.2 当 T 到期后，NodeA 还未收到来自 NodeB 的 ack，
//						- 在它本地的节点列表里把 NodeB 标记为故障
//						- 然后把故障信息转交给传播组件处理
//			2.2 收到了响应 ack 时
//				- 更新节点信息，我觉得可以把每个node保持的节点列表做成 <addr, timestamp> 表示最近活跃的时间
func (n *Node) FaultsDetect() {
	go func() {
		// 重制检测
		oneDetectHasFinised := true

		for {
			if oneDetectHasFinised {
				oneDetectHasFinised = false
				log.Println("重制检测")
				oneDetectHasFinised = n.faultsDetect()
			}
		}
	}()
}

func (n *Node) faultsDetect() bool {
	// 定时 周期 T
	if n.FaultsDetectTimer != nil {
		n.FaultsDetectTimer.Reset(time.Duration(n.T) * time.Second)
	} else {
		n.FaultsDetectTimer = time.NewTimer(time.Duration(n.T) * time.Second)
	}

	tar := n.selectOneNode()
	ack := make(chan Message, 1)
	defer close(ack)
	n.Ping(tar, ack)

	for {
		select {
		case <-n.FaultsDetectTimer.C:
			// todo FaultsDetectTimer过期 删除某节点
			log.Println("aultsDetectTimer过期")
			n.Broadcast()
			return true
		case msg := <-ack:
			log.Println("收到ack")
			n.updateNodeVersion(msg)
			return true
		}
	}

}

func (n *Node) updateNodeVersion(msg Message) {
	mutex.Lock()
	switch msg.Action {
	case Join:
		v, ok := n.maintenance[msg.From]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.From] = msg.Version
		}
	case Leave:
		v, ok := n.maintenance[msg.From]
		// 在列表里 且 版本号更大 才删除
		if ok && v < msg.Version {
			delete(n.maintenance, msg.From)
		}
	case Ping:
		// todo
		v, ok := n.maintenance[msg.From]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.From] = msg.Version
		}
	case PingBack:
		v, ok := n.maintenance[msg.From]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.From] = msg.Version
		}
	default:

	}
	mutex.Unlock()
}

// todo
func (n *Node) selectOneNode() nodeAddr {
	log.Println("selectOneNode")
	mutex.Lock()

	mutex.Unlock()
	return ""
}

// todo
func (n *Node) selectServalNodes(except nodeAddr) []nodeAddr {
	log.Println("selectServalNodesExcept")
	return []nodeAddr{}
}

// Ping 发送一个结构体
// 包含了本机的addr 和 时间戳
func (n *Node) Ping(to nodeAddr, ack chan Message) {
	// 定时
	// if n.PingTimer != nil {
	// 	n.PingTimer.Reset(time.Duration(n.PingT) * time.Second)
	// } else {
	// 	n.PingTimer = time.NewTimer(time.Duration(n.PingT) * time.Second)
	// }

	// msg := Message{
	// 	From:    n.addr,
	// 	Version: GetCurrentTime(),
	// 	Action:  Ping,
	// }
	// n.sendMessage(to, msg)

	// select {
	// case <-n.PingTimer.C:

	// }
	log.Println("Ping")
	rand.Seed(int64(time.Now().UnixNano()))
	if rand.Intn(2) == 0 {
		msg := Message{"add", rand.Intn(100), PingBack}
		ack <- msg
		log.Println("msg: ", msg)
	}
	return
}

// PingReq 发送一个结构体
func (n *Node) PingReq() {

}

// Broadcast 对 Node 维护列表内的所有节点
// 发送 '移除 某节点' 或者 '添加 某节点' 的 task
func (n *Node) Broadcast() {
	log.Println("Broadcast")
}

// BroadcastReceiver 根据时间戳
// 判断是否执行 '移除 某节点' 或 '添加 某节点' 的操作
func (n *Node) BroadcastReceiver() {

}

// ResponseMessage 收到信息时进行反应
func (n *Node) ResponseMessage() {

}
