package membership

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// Action 表示 Node 可以采用的action
type Action int

const (
	// Join 节点加入
	Join Action = 1 << iota
	// Leave 节点退出
	Leave
	// Ping 节点更新
	Ping
	// PingBack ping 的返回值
	PingBack
)

// Message 是udp 发送的报文段内容
type Message struct {
	// From 发送者地址
	From nodeAddr
	// 操作对象地址
	TargetNode nodeAddr
	// Version 当前发送时间
	Version timestamp
	// Action 消息类型
	Action Action
}

// GetCurrentTime 获取当前时间戳
func GetCurrentTime() int {
	return int(time.Now().UTC().UnixNano())
}

// ParseOneEndpoint 将字符串 parse 成 UDPAddr
// 字符串有误会返回错误
// s format: 192.168.70.30:9980
// 只支持ipv4
func ParseOneEndpoint(s string) (addr *net.UDPAddr, err error) {
	strs := strings.Split(s, ":")
	if len(strs) != 2 {
		return nil, fmt.Errorf("format for '%s' is wrong", s)
	}

	ip := net.ParseIP(strs[0])
	port, err := strconv.Atoi(strs[1])
	if ip == nil {
		return nil, fmt.Errorf("format for '%s' is wrong", strs[0])
	}
	if err != nil {
		return nil, fmt.Errorf("format for '%s' is wrong", strs[1])
	}

	addr = &net.UDPAddr{IP: ip, Port: port}
	err = nil
	return
}

// EncodeMessage 编码信息
func EncodeMessage(msg Message) *bytes.Buffer {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(msg); err != nil {
		log.Println(err)
	}
	return &b
}

// DecodeMessage 解码信息
func DecodeMessage(bs []byte) *Message {
	var msg Message
	r := bytes.NewReader(bs)
	d := gob.NewDecoder(r)
	if err := d.Decode(&msg); err != nil {
		log.Println(err)
	}
	return &msg
}
