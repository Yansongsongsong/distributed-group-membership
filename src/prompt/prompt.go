// Package prompt 提供一个CLI界面
package prompt

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	Help bool

	ListenPort  string
	GroupMember string
	Heart       int

	// faultsDetectTime int,
	// pingExpireTime int,
	// pingNodesMaxNumber int,

)

func init() {
	flag.BoolVar(&Help, "h", false, "this help")

	// 注意 `master address`。默认是 -ma string，有了 `master address` 之后，变为 -s master address
	flag.StringVar(&ListenPort, "l", "", "set this node `listen` port. e.g. 2345")
	flag.StringVar(&GroupMember, "g", "", "set this `group member` e.g. 127.0.0.1:2345")
	flag.IntVar(&Heart, "hb", 50, "set `heart beat interval`. Millisecond is considered")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = usage
	log.SetFlags(log.Ldate | log.Ltime)
}

func usage() {
	fmt.Fprintf(os.Stdout, `node, the distributed group membership.
Usage: node [-?h] [-l listen port] [-g group member] [-hb heart beat interval]

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stdout, `
Example:
  1. set up daemon
	node -l 3334
  2. to join
	node -g 100.68.74.230:4567 -l 3334
  3. to get help
	node -h
`)
}
