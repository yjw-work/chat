package main

import (
	"fmt"
	"os"
)

func main() {
	svr := NewChatClient()
	if nil == svr {
		fmt.Println("new chat client fail")
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
