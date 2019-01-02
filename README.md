### Usage:
For macOS, 
`./src/main/dgm -h`

### how to build?
```shell
# open your terminal for Unix, just copy that
cd distributed-group-membership
export GOPATH=$PWD:$GOPATH
# get the info of your system
go env | grep GOOS
# output: GOOS="darwin"
go env | grep GOARCH
# output: GOARCH="amd64"
# with result below to build
GOOS=darwin GOARCH=amd64 go build -o dgm ./src/main/main.go
```

then you can find it in `/distributed-group-membership/dgm`

### 技术
这个项目是基于SWIM进行改进，实现了gossip协议的分布式信息传递系统。当然这里的信息传递指的是集群内所有的成员列表。
#### 算法
集群中，每个 Node 都有两个组件
	- 故障检测 组件
		1. 故障检测周期 T 内，每个 NodeA 从自己保持节点列表里随机选择 1 个节点(NodeB)发送 ping 指令
			- ping 携带 NodeA 的 addr 与 时间戳
			- 接受到 ping 的 NodeB
				1. 如果这个 ping 来自新节点，则增加至自己的维护列表，并触发广播组件
				2. 如果这个 ping 来自新节点，则更新自己的维护列表
		2. 在时间 t 内
			2.1 没有收到响应 ack 时
				2.1.1 Node 从自己保持节点列表里随机选择 k 个节点发送 ping-req(NodeB)
				2.1.2 k 个节点收到 ping-req(NodeB) 后，向 NodeB 发送 ping 指令
					2.1.2.1 当有节点收到 NodeB 返回的 ack 时，该节点将 ack 返回给 NodeA
					2.1.2.2 当 T 到期后，NodeA 还未收到来自 NodeB 的 ack，
						- 在它本地的节点列表里把 NodeB 标记为故障
						- 然后把故障信息转交给传播组件处理
			2.2 收到了响应 ack 时
				- 更新节点信息，我觉得可以把每个node保持的节点列表做成 <addr, timestamp> 表示最近活跃的时间
	- 集群关系变更的广播 组件
		- broadcast()
			当触发条件被触发时，执行
			1. 对 Node 维护列表内的所有节点发送 '移除 某节点' 或者 '添加 某节点' 的 task
		- broadcastReceiver()
			1. 根据时间戳 判断是否执行 '移除 某节点' 或 '添加 某节点' 的操作
	- 集群关系变更的广播触发条件
		- 故障检测 & 节点退出 触发
		- 新节点加入触发

其中集群关系变更采用的算法有参考Gossip的Push & Pull 模型，并通过时间戳进行版本控制。
#### 详情
在传输过程中，节点之间通信的网络协议，我们选择的是成本更低的UDP。数据传输与转换使用的核心技术是和go语言完美兼容的序列化反序列化协议：gob。
gob 定义了一整套内存数据持久化的标准，可以方便地将go对象进行持久化与序列化。gob的传化效率较protocol buffer 更高。
#### go的无锁编程与通道
在多个goroutine里，通过channel发送和接受共享的数据，达到数据同步的目的。在Go中，一个有缓冲的channel可以用来起到同步作用