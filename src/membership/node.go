package membership

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
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
	// PingReqNodesNumber PingReq时 选择节点的数目
	PingReqNodesNumber int
	// WaitingPing 是当前node等待pingback的队列 大小为1
	WaitingPing chan Message
	// Ack 是收到的Ping 的确认信息
	ACK chan Message
}

var (
	mutex sync.Mutex
)

func (n *Node) sendMessage(to string, msg Message) {
	udpAddr, err := ParseOneEndpoint(to)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", to, err)
		return
	}
	data := EncodeMessage(msg)

	_, err = n.listener.WriteToUDP(data.Bytes(), udpAddr)
	if err != nil {
		log.Printf("When send to '%s', error is: %s", udpAddr.String(), err)
	}
}

func (n *Node) updateNodeVersion(msg Message) {
	mutex.Lock()
	switch msg.Action {
	case Join:
		v, ok := n.maintenance[msg.TargetNode]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.TargetNode] = msg.Version
			n.BroadcastJoin(msg.TargetNode)
		}
	case Leave:
		v, ok := n.maintenance[msg.TargetNode]
		// 在列表里 且 版本号更大 才删除
		if ok && v < msg.Version {
			delete(n.maintenance, msg.TargetNode)
			n.BroadcastDelete(msg.TargetNode)
		}
	case Ping:
		v, ok := n.maintenance[msg.TargetNode]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		n.PingRespond(msg)
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.TargetNode] = msg.Version
		}
	case PingBack:
		v, ok := n.maintenance[msg.TargetNode]
		// 1. 不在列表里的可以加入
		// 2. 在列表里 版本号小的进行更新
		n.PingBackRespond(msg)
		if !ok || (v < msg.Version && ok) {
			n.maintenance[msg.TargetNode] = msg.Version
		}
	default:
	}
	mutex.Unlock()
}

// PingRespond 接收到Ping的信息
func (n *Node) PingRespond(msg Message) {
	// A send Ping to current:
	//		Ping.from = A, Ping.to = current, type = ping
	// A PingReq current:
	//		PingReq.from = A, PingReq.to != current, type = ping
	// A Ping_PingReq current:
	//		Ping_PingReq.from != A, Ping_PingReq.to = current, type = ping

	if msg.From == msg.RealFromAddr && msg.TargetNode == n.addr {
		// A send Ping to current:
		//		Ping.from = A, Ping.to = current, type = ping
		// 当前节点接收到的是来自 msg.From 的 Ping
		m := Message{
			RealFromAddr: n.addr,
			// From 发送者地址
			From: n.addr,
			// 操作对象地址
			TargetNode: msg.From,
			// Version 当前发送时间
			Version: GetCurrentTime(),
			// Action 消息类型
			Action: PingBack,
		}
		n.sendMessage(msg.From, m)
	}

	if msg.From == msg.RealFromAddr && msg.TargetNode != n.addr {
		// A PingReq current:
		//		PingReq.from = A, PingReq.to != current, type = ping
		// 当前节点接收到的是来自 msg.From 的 PingReq
		m := Message{
			RealFromAddr: n.addr,
			// From 发送者地址
			From: msg.From,
			// 操作对象地址
			TargetNode: msg.TargetNode,
			// Version 当前发送时间
			Version: GetCurrentTime(),
			// Action 消息类型
			Action: Ping,
		}
		n.sendMessage(msg.TargetNode, m)
	}

	if msg.From != msg.RealFromAddr && msg.TargetNode == n.addr {
		// A Ping_PingReq current:
		//		Ping_PingReq.from != A, Ping_PingReq.to = current, type = ping
		m := Message{
			RealFromAddr: n.addr,
			// From 发送者地址
			From: msg.TargetNode,
			// 操作对象地址
			TargetNode: msg.From,
			// Version 当前发送时间
			Version: GetCurrentTime(),
			// Action 消息类型
			Action: PingBack,
		}
		n.sendMessage(msg.TargetNode, m)
	}

}

// PingBackRespond 接受到PingBack的信息
func (n *Node) PingBackRespond(msg Message) {
	// @https://prakhar.me/images/swim.png 依次
	// B send PingBack to current:
	// 		PingBack.from = B, PingBack.to = current, type = pingback
	//		1. 不需要发任何报文
	//		2. 相当于在Ping的这一步获得了ack
	// B PingOfPingReq_PingBack current
	//		PingReq_PingBack.from = B, PingReq_PingBack.to != current, type = pingback
	// B transfer_PingOfPingReq_PingBack current
	// 		transfer.from != B, PingReq_PingBack.to == current, type = pingback
	if msg.From == msg.RealFromAddr && msg.TargetNode == n.addr {
		n.ACK <- msg
	}

	if msg.From == msg.RealFromAddr && msg.TargetNode != n.addr {
		m := Message{
			RealFromAddr: n.addr,
			// From 发送者地址
			From: msg.From,
			// 操作对象地址
			TargetNode: msg.TargetNode,
			// Version 当前发送时间
			Version: GetCurrentTime(),
			// Action 消息类型
			Action: PingBack,
		}
		n.sendMessage(msg.TargetNode, m)
	}

	if msg.From != msg.RealFromAddr && msg.TargetNode == n.addr {
		n.ACK <- msg
	}
}

func (n *Node) selectOneNode() nodeAddr {
	list := []nodeAddr{}
	mutex.Lock()
	for k := range n.maintenance {
		list = append(list, k)
	}
	mutex.Unlock()
	rand.Seed(int64(time.Now().UnixNano()))
	index := rand.Intn(len(list))

	return list[index]
}

func (n *Node) selectServalNodes(except nodeAddr) []nodeAddr {
	list := []nodeAddr{}
	mutex.Lock()
	for k := range n.maintenance {
		if k == except {
			continue
		}
		list = append(list, k)
	}
	mutex.Unlock()

	if len(list) <= n.PingReqNodesNumber {
		return list
	}

	rand.Seed(int64(time.Now().UnixNano()))
	indexes := rand.Perm(len(list))[:n.PingReqNodesNumber]
	res := []nodeAddr{}

	for _, i := range indexes {
		res = append(res, list[i])
	}
	return res
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
				log.Println("list: []", n.maintenance)
				oneDetectHasFinised = n.faultsDetect()
				log.Println("执行完faultsDetect() oneDetectHasFinised: ", oneDetectHasFinised)
			}
		}
	}()
}

func (n *Node) faultsDetect() bool {
	// 定时 周期 T
	if n.FaultsDetectTimer != nil {
		log.Println("重设FaultsDetectTimer")
		n.FaultsDetectTimer.Reset(time.Duration(n.T) * time.Second)
	} else {
		log.Println("新建FaultsDetectTimer")
		n.FaultsDetectTimer = time.NewTimer(time.Duration(n.T) * time.Second)
	}

	if len(n.maintenance) == 0 {
		// 当本机没有维护的机器时 不ping
		log.Println("当本机没有维护的机器时 不ping")
		return true
	}
	tar := n.selectOneNode()
	log.Println("selectOneNode: ", tar)

	log.Println("Ping: ", tar)

	go n.Ping(tar)

	for {
		select {
		case <-n.FaultsDetectTimer.C:
			log.Println("FaultsDetectTimer过期")
			n.BroadcastDelete(tar)
			return true
		case msg := <-n.ACK:
			log.Println("收到ack")
			n.FaultsDetectTimer.Stop()
			n.PingTimer.Stop()
			n.updateNodeVersion(msg)
			return true
		}
	}

}

// Ping 发送一个结构体
// 包含了本机的addr 和 时间戳
func (n *Node) Ping(to nodeAddr) {
	log.Println("Ping开始")
	// 定时
	if n.PingTimer != nil {
		log.Println("PingTimer reset")
		n.PingTimer.Reset(time.Duration(n.PingT) * time.Second)
	} else {
		log.Println("PingTimer init")
		n.PingTimer = time.NewTimer(time.Duration(n.PingT) * time.Second)
	}

	msg := Message{
		RealFromAddr: n.addr,
		From:         n.addr,
		TargetNode:   to,
		Version:      GetCurrentTime(),
		Action:       Ping,
	}
	log.Println("发ping， msg: ", msg)
	n.sendMessage(to, msg)
	n.WaitingPing <- msg

	for {
		select {
		case <-n.PingTimer.C:
			// 过期
			n.PingReq(msg)
			log.Println("PingReq msg 结束: ", msg)
			return
		}
	}

}

// PingReq 发送一个结构体
func (n *Node) PingReq(msg Message) {
	log.Println("PingReq msg: 开始", msg)
	nodes := n.selectServalNodes(msg.TargetNode)
	for _, node := range nodes {
		n.sendMessage(node, msg)
	}
}

// BroadcastDelete 对 Node 维护列表内的所有节点
// // 发送 '移除 某节点' 的 task
func (n *Node) BroadcastDelete(addr nodeAddr) {
	mutex.Lock()
	log.Println("before BroadcastDelete: ", n.maintenance)
	delete(n.maintenance, addr)
	log.Println("after BroadcastDelete: ", n.maintenance)
	mutex.Unlock()
	msg := Message{
		RealFromAddr: n.addr,
		From:         n.addr,
		TargetNode:   addr,
		Version:      GetCurrentTime(),
		Action:       Leave,
	}
	log.Println("BroadcastDelete: msg: ", msg)
	n.Broadcast(msg)
}

// BroadcastJoin 对 Node 维护列表内的所有节点
// // 发送 '移除 某节点' 的 task
func (n *Node) BroadcastJoin(addr nodeAddr) {
	v := GetCurrentTime()
	mutex.Lock()
	log.Println("before BroadcastJoin: ", n.maintenance)
	n.maintenance[addr] = v
	log.Println("after BroadcastJoin: ", n.maintenance)
	mutex.Unlock()
	msg := Message{
		RealFromAddr: n.addr,
		From:         n.addr,
		TargetNode:   addr,
		Version:      v,
		Action:       Join,
	}
	log.Println("BroadcastJoin: msg: ", msg)
	n.Broadcast(msg)
}

// Broadcast 对 Node 维护列表内的所有节点
// 发送 '移除 某节点' 或者 '添加 某节点' 的 task
func (n *Node) Broadcast(msg Message) {
	list := []nodeAddr{}
	mutex.Lock()
	for k := range n.maintenance {
		list = append(list, k)
	}
	mutex.Unlock()

	for _, end := range list {
		n.sendMessage(end, msg)
	}
}

// Receiver 根据时间戳
// 判断是否执行 '移除 某节点' 或 '添加 某节点' 的操作
func (n *Node) Receiver() {
	go func() {
		for {
			data := make([]byte, 1024)
			count, remoteAddr, err := n.listener.ReadFromUDP(data)
			if err != nil {
				fmt.Printf("error during read: %s", err)
			}
			msg := DecodeMessage(data[:count])
			fmt.Printf("Receiver from <%s> %v\n", remoteAddr, *msg)
			n.updateNodeVersion(*msg)
		}
	}()
}

// RunNode 定义了运行机制
func RunNode(
	nodeAddress string,
	faultsDetectTime int,
	pingExpireTime int,
	pingNodesMaxNumber int,
	introducerAddr []string) {
	_, err := ParseOneEndpoint(nodeAddress)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for _, in := range introducerAddr {
		_, e := ParseOneEndpoint(in)
		if e != nil {
			log.Println(e)
			os.Exit(1)
		}
	}

	node := NewNode(nodeAddress, faultsDetectTime, pingExpireTime, pingNodesMaxNumber)
	node.Receiver()
	for _, in := range introducerAddr {
		node.BroadcastJoin(in)
	}

	node.FaultsDetect()

}

// NewNode 工厂方法
func NewNode(
	nodeAddress string,
	faultsDetectTime int,
	pingExpireTime int,
	pingNodesMaxNumber int) *Node {
	node := new(Node)
	node.addr = nodeAddress
	node.T = faultsDetectTime
	node.PingT = pingExpireTime
	node.PingReqNodesNumber = pingNodesMaxNumber

	udpAdd, _ := ParseOneEndpoint(node.addr)

	listener, err := net.ListenUDP("udp", udpAdd)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	node.FaultsDetectTimer = nil
	node.PingTimer = nil
	node.listener = listener
	node.maintenance = make(map[nodeAddr]timestamp)
	node.WaitingPing = make(chan Message, 1)
	node.ACK = make(chan Message, 1)

	return node
}
