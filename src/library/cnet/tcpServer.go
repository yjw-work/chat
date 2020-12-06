package cnet

import (
	"fmt"
	"library/chelper"
	"net"
	"sync/atomic"
)

type ITcpServerCallback interface {
	ITcpConnectCallback
	OnConnect(conn *TTcpConnect)
}

type TTcpServerConfig struct {
	Addr string
	TTcpConnectCfg
}

type TTcpServer struct {
	cfg      *TTcpServerConfig
	listener *net.TCPListener
	cb       ITcpServerCallback
	stopped  int32
}

func NewTcpServer(callback ITcpServerCallback, cfg *TTcpServerConfig) *TTcpServer {
	return &TTcpServer{
		cb:  callback,
		cfg: cfg,
	}
}

func (this *TTcpServer) Run(ctx *chelper.TContext) {
	defer ctx.WG.Done()
	var err error
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", this.cfg.Addr)
	if nil != err {
		fmt.Println("resolve tcp address fail: ", err)
		return
	}
	this.listener, err = net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		fmt.Println("tcp server listen fail: ", err)
		return
	}
	fmt.Println("start on address: ", tcpAddr)
	for {
		conn, err := this.listener.AcceptTCP()
		if nil != err {
			fmt.Println("tcp server accept fail: ", err)
			return
		}
		tcpConn := NewTcpConnect(ctx, conn, &this.cfg.TTcpConnectCfg, this.cb)
		this.cb.OnConnect(tcpConn)
		ctx.WG.Add(1)
		go tcpConn.HandleConn()
	}
}

func (this *TTcpServer) Stop() {
	if atomic.CompareAndSwapInt32(&this.stopped, 0, 1) {
		if err := this.listener.Close(); nil != err {
			fmt.Println("tcp server stop fail: ", err)
		}
	}
}
