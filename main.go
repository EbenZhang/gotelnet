package main

import "fmt"

//type CmdFunc func 
type MyHandler struct {
	A int
}

func (obj *MyHandler) Handle(commandLine string) {
	fmt.Println("recv: ", commandLine)
}

func main() {
	var handler MyHandler
	server := NewGoTelnet("zyn", "0.0.0.0", 2500, &handler)

	go server.Run()

	for {
		var quit string
		fmt.Scanln(&quit)
		if quit == "quit" {
			server.Quit()
			break
		}
	}
}
