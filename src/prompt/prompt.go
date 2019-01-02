// Package prompt 提供一个CLI界面
package prompt

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type workerAddress []string

// new一个存放命令行参数值的slice
func newWorkerAddress(vals []string, p *[]string) *workerAddress {
	*p = vals
	return (*workerAddress)(p)
}

// 实现flag包中的Value接口，将命令行接收到的值用,分隔存到slice里
func (s *workerAddress) Set(val string) error {
	*s = workerAddress(strings.Fields(val))
	return nil
}

// 实现flag包中的Value接口，将命令行接收到的值用,分隔存到slice里
func (s *workerAddress) String() string {
	// default value
	*s = workerAddress([]string{})
	// when it is cast as string, return this value
	return ""
}

var (
	Help bool

	NodeAddress        string
	IntroducerAddr     []string
	FaultsDetectTime   int
	PingExpireTime     int
	PingNodesMaxNumber int

	// faultsDetectTime int,
	// pingExpireTime int,
	// pingNodesMaxNumber int,

)

func init() {
	flag.BoolVar(&Help, "h", false, "this help")

	// 注意 `master address`。默认是 -ma string，有了 `master address` 之后，变为 -s master address
	flag.StringVar(&NodeAddress, "na", "", "set this `node address` that we can communicate with")
	flag.Var(newWorkerAddress([]string{}, &IntroducerAddr), "ia", "set the `introducer address` that we can communicate with, spilt with space")
	flag.IntVar(&FaultsDetectTime, "fdt", 15, "set `Faults Detecting Time` for node")
	flag.IntVar(&PingExpireTime, "pet", 10, "set the `Ping Expired Time` for node")
	flag.IntVar(&PingNodesMaxNumber, "n", 2, "set the number for giving one broadcast `file`")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stdout, `node, the distributed group membership.
Usage: node [-?h] [-na nodeAddress] [-ia introducerAddress] [-fdt faultsDetectTime] [-pet pingExpireTime] [-n pingNodesMaxNumber]

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stdout, `
Example:
  1. set up node with several single introducers
	node -na localhost:9990 -ia "localhost:9991 localhost:9992 localhost:9993"
  2. to set up worker
	node -na localhost:9990 -ia "localhost:9991 localhost:9992 localhost:9993"
  3. to get help
	node -h
`)
}
