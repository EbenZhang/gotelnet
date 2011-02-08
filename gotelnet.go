package main

import "net"
import "fmt"
import "os"
import "strconv"


type TelnetCommandHandler interface {
	Handle(commandLine string)
}

type GoTelnet struct {
	Promote           string
	Ip                string
	StartPort         uint16
	commandHandler    TelnetCommandHandler
	preConnection     *net.Conn
	chanForQuit       chan bool //用来接收退出消息
	chanForQuitOK     chan bool //用来反馈退出结果
	chanForNextAccept chan bool //用来accept下一个连接
}

func NewGoTelnet(promote string, ip string, startPort uint16, handler TelnetCommandHandler) *GoTelnet {
	var server GoTelnet
	server.Promote = promote
	if len(promote) == 0 {
		server.Promote = "telnet"
	}
	server.Ip = ip
	if len(ip) == 0 {
		server.Ip = "0.0.0.0"
	}
	server.StartPort = startPort
	if startPort <= 1024 {
		server.StartPort = 2500
	}

	server.chanForQuit = make(chan bool, 1)
	server.chanForQuitOK = make(chan bool, 1)
	server.chanForNextAccept = make(chan bool, 1)

	return &server
}

func (server *GoTelnet) getAddr() *net.TCPAddr {
	addrString := server.Ip + ":" + strconv.Uitoa(uint(server.StartPort))
	tcpAddr, _ := net.ResolveTCPAddr(addrString)
	return tcpAddr
}

func (server *GoTelnet) acceptConnection(listener *net.TCPListener) {
	conn, _ := listener.Accept()
	server.chanForNextAccept <- true
	if server.preConnection != nil {
		(*server.preConnection).Close()
	}
	server.preConnection = &conn
	conn.Write([]byte("welcome to telnet debug server"))

	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		readlen, ok := conn.Read(buf)
		if ok != nil {
			fmt.Fprintf(os.Stderr, "close connection when reading from socket: %s\n", ok.String())
			return
		}
		if readlen == 0 {
			fmt.Printf("Connection closed by remote host\n")
			return
		}
		if buf[0] == 13 {
			conn.Write([]byte(server.Promote + ">"))
		}
		if string(buf[0:3]) == "bye" {
			conn.Write(buf)
			break
		}
		server.commandHandler.Handle(string(buf))
	}
}

func (server *GoTelnet) Run() {
	var listener *net.TCPListener
	for {
		tcpAddr := server.getAddr()
		var err os.Error
		listener, err = net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			server.StartPort++
		} else {
			break
		}
	}

	server.chanForNextAccept <- true

	for {
		select {
		case <-server.chanForNextAccept:
			go server.acceptConnection(listener)
		case <-server.chanForQuit:
			fmt.Println("quiting....")
			server.chanForQuitOK <- true
			return
		}
	}
}
func (server *GoTelnet) Quit() {
	server.chanForQuit <- true
	<-server.chanForQuitOK
}
