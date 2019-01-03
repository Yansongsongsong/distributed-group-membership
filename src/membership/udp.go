package membership

/*
  Handles listening for incoming gossip requests, group joins, and quits
  Adapted from
    https://github.com/bootic/bootic_data_collector/blob/master/udp/udp.go
*/

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"logger"
)

const (
	heartbeatThreshold = 25
)

type Daemon struct {
	Conn         *net.UDPConn
	Port         string
	MemberList   map[string]*GroupMember
	InactiveList map[string]int
	Active       bool
}

func NewDaemon(port string) (daemon *Daemon, err error) {

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}
	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return
	}

	hostPort := net.JoinHostPort(addrs[0], port)
	log.Printf("Creating daemon at %s\n", hostPort)
	logger.Log("INFO", "Creating daemon at"+hostPort)
	conn, err := createUDPListener(hostPort)

	if err != nil {
		return
	}

	daemon = &Daemon{
		Conn:         conn,
		Port:         port,
		MemberList:   make(map[string]*GroupMember),
		InactiveList: make(map[string]int),
		Active:       true,
	}

	log.Println("Daemon created!")
	logger.Log("INFO", "Daemon Created")
	return
}

func (self *Daemon) ReceiveDatagrams(joinGroupOnConnection bool) {
	if self.Active == false {
		return
	}
	for {
		buffer := make([]byte, 1024)
		c, addr, err := self.Conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("%d byte datagram from %s with error %s\n", c, addr.String(), err.Error())
			logger.Log("ERROR", addr.String()+"byte datagram from %s with error "+err.Error())
			return
		}

		//log.Printf("Bytes received: %d\n", c)

		portmsg := strings.SplitN(string(buffer[:c]), "<PORT>", 2)
		port, msg := portmsg[0], portmsg[1]
		senderAddr := net.JoinHostPort(addr.IP.String(), port)

		//log.Printf("Data received from %s: %s", senderAddr, msg)
		logger.Log("INFO", "Data received from "+senderAddr+" : "+msg)
		self.handleMessage(msg, senderAddr, &joinGroupOnConnection)
	}
}

func (self *Daemon) handleMessage(msg, sender string, joinSenderGroup *bool) {
	fields := strings.SplitN(msg, "|%|", 2)
	switch fields[0] {
	case "JOIN":
		logger.Log("JOIN", "Member just joined"+sender+msg)
		self.addNewMember(sender)
		if *joinSenderGroup {
			*joinSenderGroup = false
			self.JoinGroup(sender)
		}
	case "GOSSIP":
		logger.Log("GOSSIP", "Gossiping "+sender+fields[1])
		self.handleGossip(sender, fields[1])
	case "QUIT":
		logger.Log("QUIT", "Member just quit: "+sender)
		self.removeMember(sender)
	}
}

func (self *Daemon) handleGossip(senderAddr, subject string) {
	// Reset the counter for the sender
	// TODO add sender if it doesn't exist yet
	for id, member := range self.MemberList { // todo
		if strings.Contains(id, senderAddr) {
			//log.Printf("Reset %s, %s\n", id, senderAddr)
			member.SetHeartBeat(0)
		}
		logger.Log("GOSSIP", "Reset counter for "+senderAddr)
	}

	// Update the counter for the subject
	subjectMember := Unmarshal(subject)
	if subjectMember == nil {
		return
	}

	curr := self.MemberList[subjectMember.Id]
	if curr == nil {
		if self.InactiveList[subjectMember.Id] == 0 { // todo
			self.MemberList[subjectMember.Id] = subjectMember // todo

		}
	} else {
		if curr.Heartbeat > subjectMember.Heartbeat {
			curr.SetHeartBeat(subjectMember.Heartbeat)
		}
	}

}

func (self *Daemon) addNewMember(address string) (newMember *GroupMember) {
	now := time.Now().UTC()
	machineId := address + "###" + now.String()
	newMember = NewGroupMember(machineId, address, 0)
	log.Printf("Created new member with IP: %s", address)
	logger.Log("INFO", "Created new member with IP "+address)
	self.MemberList[machineId] = newMember // todo

	return
}

func (self *Daemon) removeMember(address string) {
	var keyToDelete string
	for key, _ := range self.MemberList {
		if strings.Contains(key, address) {
			keyToDelete = key
			break
		}
	}
	self.InactiveList[keyToDelete] = 1
	delete(self.MemberList, keyToDelete)
	log.Printf("Deleted member with IP: %s", address)
	logger.Log("DELETE", "Deleted member with IP "+address)
}

func (self *Daemon) JoinGroup(address string) (err error) {
	return self.sendMessageWithPort("JOIN", address)
}

func (self *Daemon) LeaveGroup() {
	for _, member := range self.MemberList {
		self.sendMessageWithPort("QUIT", member.Address)
	}
}

func (self *Daemon) Gossip(subject, receiver *GroupMember) (err error) {
	// The message we are sending over UDP, subject can be nil
	msg := "GOSSIP|%|" + Marshal(subject)
	return self.sendMessageWithPort(msg, receiver.Address) // todo
}

func (self *Daemon) sendMessageWithPort(msg, address string) (err error) {
	msg = self.Port + "<PORT>" + msg
	return sendMessage(msg, address)
}

// Increment all heartbeat values, and gossip a random machine A to another random B
func (self *Daemon) HeartbeatAndGossip() {
	if self.Active == false {
		return
	}
	// Nobody in the list yet
	if len(self.MemberList) < 1 { // todo
		return
	}
	receiverIndex := rand.Int() % len(self.MemberList)
	var receiver *GroupMember
	i := 0
	j := 0
	deleteMembers := make([]string, len(self.MemberList))
	var BeforeList string

	for key, currMember := range self.MemberList { // todo
		if receiverIndex == i {
			receiver = currMember
		}
		currMember.IncrementHeartBeat()
		if currMember.Heartbeat > heartbeatThreshold {
			log.Println("MACHINE DEAD!", currMember.Id)
			logger.Log("EXIT", "Machine left the group "+currMember.Id)
			deleteMembers[j] = key
			j++
		}
		BeforeList += key + " , "
		i++
	}

	logger.Log("LISTSTAT", BeforeList)
	for _, subject := range self.MemberList {
		if subject.Id == receiver.Id { // todo
			self.Gossip(nil, receiver)
		}
		self.Gossip(subject, receiver)
	}
	for _, member := range deleteMembers {
		self.InactiveList[member] = 1 // todo
		delete(self.MemberList, member)
	}
}

func (self *Daemon) CheckStandardInput() {
	var input string
	for {
		fmt.Scanln(&input)
		if input == "leave" {
			self.Active = false
		}
		if input == "show" {
			log.Println("SHOW", fmt.Sprintf("%v", self.MemberList))
			logger.Log("SHOW", fmt.Sprintf("%v", self.MemberList))
		}
	}

}

// Get two random numbers that aren't the same
// TODO there's probably a better (quicker) way to do this
func distinctRandoms(max int) (a, b int) {
	a = rand.Int() % max
	for b = a; b == a; {
		b = rand.Int() % max
	}
	return
}

func sendMessage(message, address string) (err error) {
	var raddr *net.UDPAddr
	if raddr, err = net.ResolveUDPAddr("udp", address); err != nil {
		log.Panic(err)
	}

	var con *net.UDPConn
	con, err = net.DialUDP("udp", nil, raddr)
	//log.Printf("Sending '%s' to %s..", message, raddr)
	logger.Log("INFO", "Sending "+message)
	if _, err = con.Write([]byte(message)); err != nil {
		log.Panic("Writing to UDP:", err)
		logger.Log("ERROR", "Writing to UDP")
	}

	return
}

func createUDPListener(hostPort string) (conn *net.UDPConn, err error) {

	var udpaddr *net.UDPAddr
	if udpaddr, err = net.ResolveUDPAddr("udp", hostPort); err != nil {
		return
	}

	conn, err = net.ListenUDP("udp", udpaddr)

	return
}
