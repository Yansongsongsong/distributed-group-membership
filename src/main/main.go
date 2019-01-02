// Package main 提供一个CLI界面
package main

import (
	"flag"
	"membership"
	"prompt"
)

var (
	help               = &prompt.Help
	nodeAddress        = &prompt.NodeAddress
	introducerAddr     = &prompt.IntroducerAddr
	faultsDetectTime   = &prompt.FaultsDetectTime
	pingExpireTime     = &prompt.PingExpireTime
	pingNodesMaxNumber = &prompt.PingNodesMaxNumber
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
	}

	if *nodeAddress == "" {
		flag.Usage()
	} else {
		membership.RunNode(
			*nodeAddress,
			*faultsDetectTime,
			*pingExpireTime,
			*pingNodesMaxNumber,
			*introducerAddr,
		)
	}

}
