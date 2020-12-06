package main

import (
	"fmt"
	"os"
)

func main() {
	svr := NewChatServer()
	if nil == svr {
		fmt.Println("new chat server fail")
		return
	}
	var err error
	defer func() {
		if nil != err {
			fmt.Println(err)
		}
	}()

	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	if err = svr.init(cfgPath); nil != err {
		return
	}
	svr.run()
}
