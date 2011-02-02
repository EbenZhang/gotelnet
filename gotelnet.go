package main

import "net"
import "fmt"
import "os"
import "strconv"


type TelnetCommandHandler interface{
    Handle(commandLine string)
}

type GoTelnet struct{
    Promote string
    Ip string
    StartPort uint16
    commandHandler TelnetCommandHandler
    preConnection *net.Conn
}

func NewGoTelnet(promote string, ip string, startPort uint16, handler TelnetCommandHandler) *GoTelnet{
    tempPromote := promote
    if len(tempPromote) == 0{
        tempPromote = "telnet"
    }
    tempIp := ip
    if len(tempIp) == 0{
        tempIp = "0.0.0.0"
    }
    tempStartPort := startPort
    if tempStartPort == 0{
        tempStartPort = 2500
    }
    
    return &GoTelnet{tempPromote,tempIp,tempStartPort,handler,nil}
}

func (obj * GoTelnet)getAddr() *net.TCPAddr{
    addrString := obj.Ip + ":" + strconv.Uitoa(uint(obj.StartPort))
    tcpAddr,_ := net.ResolveTCPAddr(addrString)
    return tcpAddr
}

func (obj * GoTelnet)handler(listener *net.TCPListener,chanForAcceptOK chan bool){
    conn,_ := listener.Accept()
    chanForAcceptOK <- true
    if obj.preConnection != nil{
       (*obj.preConnection).Close() 
    }
    obj.preConnection = &conn
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
        if buf[0] == 13{
            conn.Write([]byte(obj.Promote + ">"))
        }
        if string(buf[0:3]) == "bye"{
            conn.Write(buf)
            break
        }
        obj.commandHandler.Handle(string(buf))
    }
}

func (obj * GoTelnet)DoRun(chanForQuit chan bool){
    var listener *net.TCPListener
    for{
        tcpAddr := obj.getAddr() 
        var err os.Error
        listener,err = net.ListenTCP("tcp",tcpAddr)
        if err != nil{
            obj.StartPort++
        } else{
            break
        }
    }

    
    chanForAcceptOK := make(chan bool,1)
    chanForAcceptOK <- true
    
    for{ 
        select {
        case <-chanForAcceptOK:            
            go obj.handler(listener,chanForAcceptOK)
        case <-chanForQuit:
            fmt.Println("quiting....")
            return
        }
    }    
}
