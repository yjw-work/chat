package main

import (
	"bufio"
	"common"
	"fmt"
	"library/chelper"
	"library/cnet"
	"os"
	"path"
	"strings"
)

type TChatClientConfig struct {
	cnet.TTcpClientConfig
}

type TChatClient struct {
	clt *cnet.TTcpClient
	cfg *TChatClientConfig
	ctx *chelper.TContext
}

func NewChatClient() *TChatClient {
	this := &TChatClient{}
	this.ctx = chelper.NewContext()
	if nil == this.ctx {
		return nil
	}
	return this
}

func (this *TChatClient) init(cfgPath string) error {
	if len(cfgPath) <= 0 {
		d, n, _, err := chelper.BinPathSplit()
		if nil != err {
			return err
		}
		cfgPath = path.Join(d, n+".json")
	}
	cfg := &TChatClientConfig{}
	if err := chelper.ReadJsonFile(cfgPath, cfg); nil != err {
		return err
	}
	this.clt = cnet.NewTcpClient(this, &cfg.TTcpClientConfig)
	if nil == this.clt {
		return fmt.Errorf("new tcp client fail")
	}
	this.cfg = cfg
	return nil
}

func (this *TChatClient) run() {
	this.ctx.WG.Add(1)
	go this.clt.Run(this.ctx)

	this.readCmd()
	this.ctx.WG.Wait()
}

func (this *TChatClient) OnDisConnect(conn *cnet.TTcpConnect) {
}

func (this *TChatClient) OnMessage(msg cnet.IMessage) {
	data := msg.Data()
	if nil == data || len(data) <= 0 {
		return
	}

	content := string(data)
	fmt.Println(content)
}

func (this *TChatClient) sendMessage(content string) {
	conn := this.clt.Connect()
	if nil != conn {
		msg := this.MessagePool().Get(conn).SetData([]byte(content))
		conn.Send(msg)
	}
}

func (this *TChatClient) MessagePool() cnet.IMessagePool {
	return common.DefMessagePool
}

func (this *TChatClient) readCmd() {
	reader := bufio.NewReader(os.Stdin)
	for {
		txt, _ := reader.ReadString('\n')
		txt = strings.Trim(txt, "\n")
		this.sendMessage(txt)
	}
}
