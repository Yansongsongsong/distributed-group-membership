package membership

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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
