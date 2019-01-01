// Package membership 基于SWIM，算法详解 @see https://zhuanlan.zhihu.com/p/39703992
// # 算法，每个 Node 都有两个组件
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
// 	- 集群关系变更的广播 组件
//		- broadcast()
//			当触发条件被触发时，执行
//			1. 对 Node 维护列表内的所有节点发送 '移除 某节点' 或者 '添加 某节点' 的 task
//		- broadcastReceiver()
//			1. 根据时间戳 判断是否执行 '移除 某节点' 或 '添加 某节点' 的操作
//	- 集群关系变更的广播触发条件
// 		- 故障检测 & 节点退出 触发
// 		- 新节点加入触发
// 思路ie
// 1. 时间周期可以用任务开始的时间戳来表示，某个人物可以在创建时就定好结束的时间戳
// 2. ack 可以包含节点名字
package membership
