package cnet

import (
	"fmt"
	"library/chelper"
	"net"
)

type ITcpClientCallback interface {
	ITcpConnectCallback
}

type TTcpClientConfig struct {
	Addr string
	TTcpConnectCfg
}

type TTcpClient struct {
	cfg     *TTcpClientConfig
	cb      ITcpClientCallback
	stopped int32

	conn *TTcpConnect
}

func NewTcpClient(callback ITcpClientCallback, cfg *TTcpClientConfig) *TTcpClient {
	return &TTcpClient{
		cb:  callback,
		cfg: cfg,
	}
}

func (this *TTcpClient) Run(ctx *chelper.TContext) {
	defer ctx.WG.Done()
	var err error
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", this.cfg.Addr)
	if nil != err {
		fmt.Println("resolve tcp address fail: ", err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		fmt.Println("tcp dial fail: ", err)
		return
	}
	fmt.Println("dial succ: ", tcpAddr)

	this.conn = NewTcpConnect(ctx, conn, &this.cfg.TTcpConnectCfg, this.cb)
	ctx.WG.Add(1)
	go this.conn.HandleConn()
}

func (this *TTcpClient) Connect() *TTcpConnect {
	return this.conn
}
