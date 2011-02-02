package main

import "fmt"
type MyHandler struct{
    A int
}

func (obj *MyHandler) Handle(commandLine string) {
    fmt.Println("recv: " , commandLine)
}

func main(){
    var handler MyHandler
    obj := NewGoTelnet("zyn","0.0.0.0",2500,&handler)

    chanForQuit := make(chan bool,1)
    go obj.DoRun(chanForQuit)

    for{
    var quit string
    fmt.Scanln(&quit)
    if quit == "quit"{
        chanForQuit <- true
        chanForQuit <- true
        break
     }   
   } 
}
