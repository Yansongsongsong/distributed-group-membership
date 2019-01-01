package membership

type nodeAddr = string
type timestamp = int

// Node 是机器节点
type Node struct {
	// endpoint
	addr        string
	maintenance map[nodeAddr]timestamp
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
