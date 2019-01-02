package main

import (
	"flag"
	"log"
	"logger"
	"membership"
	"prompt"
	"time"
)

var (
	help = &prompt.Help
	l    = &prompt.ListenPort
	g    = &prompt.GroupMember

	heartBeat = &prompt.Heart
)

// Maintain dictionary of machines
// Each entry counts number of "heartbeats" since we heard from that machine
// If the number of heartbeats crosses a threshold, we know it is unresponsive
func main() {
	flag.Parse()
	if *help {
		flag.Usage()
	}
	listenPort := *l
	groupMember := *g

	if listenPort == "" && groupMember == "" {
		flag.Usage()
	} else {
		daemon(listenPort, groupMember)
	}

}

func daemon(listenPort string, groupMember string) {
	log.Println("Start server on port", listenPort)
	logger.Log("INFO", "Start Server on Port"+listenPort)

	// Determine the heartbeat duration time
	heartbeatInterval := 50 * time.Millisecond

	/*
	   use this to test (from command line for now):
	     echo -n "hello" >/dev/membership/localhost/4567
	*/
	daemon, err := membership.NewDaemon(listenPort)
	if err != nil {
		log.Panic("Daemon creation", err)
	}

	// Join the group by broadcasting a dummy message
	// TODO what if the packet got dropped? rebroadcast after a timeout
	firstInGroup := groupMember == ""
	if !firstInGroup {
		daemon.JoinGroup(groupMember)
		logger.Log("JOIN", "Gossiping new member to the group")
	}

	go daemon.ReceiveDatagrams(firstInGroup)

	go daemon.CheckStandardInput()

	for {
		//Get random member , increment current members
		if daemon.Active == false {
			break
		}
		daemon.HeartbeatAndGossip()
		time.Sleep(heartbeatInterval)
	}

	daemon.LeaveGroup()
}
