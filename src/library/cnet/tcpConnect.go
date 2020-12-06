package cnet

import (
	"encoding/binary"
	"fmt"
	"io"
	"library/chelper"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var gConnId uint64

const (
	cstMessageLimit    = 1 << 20 //1M
	cstMessageHeadSize = 4
	cstMessageLenSize  = 4
	cstReadChanSize    = 1000
	cstSendChanSize    = 1000
	cstReadWaterMark   = 0.9
)

type IMessage interface {
	Parse(data []byte) (ok bool)
	Write2Cache(cache []byte) (writeLen int)

	SetConnect(*TTcpConnect)
	Connect() *TTcpConnect
	SetData(data []byte) IMessage
	Data() []byte
}

type IMessagePool interface {
	Get(conn *TTcpConnect) IMessage
	Put(msg IMessage)
}

type ITcpConnectCallback interface {
	OnDisConnect(conn *TTcpConnect)
	OnMessage(msg IMessage)
	MessagePool() IMessagePool
}

type TTcpConnectCfg struct {
	MaxMessageSize uint32
}

type TTcpConnect struct {
	ctx        *chelper.TContext
	conn       net.Conn
	cfg        *TTcpConnectCfg
	msgPool    IMessagePool
	sendCache  []byte
	sendChan   chan IMessage
	readChan   chan IMessage
	connId     uint64
	offLine    int32
	cb         ITcpConnectCallback
	wg         sync.WaitGroup
	chDone     chan struct{}
	OnlineTime int64
	msgLenBuf  []byte
}

func (this *TTcpConnect) ConnId() uint64 {
	return this.connId
}

func (this *TTcpConnect) NetInfo() string {
	addr := this.conn.RemoteAddr().String()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	return fmt.Sprintf("%s:%s", net.ParseIP(host).String(), port)
}

func (this *TTcpConnect) Send(msg IMessage) {
	if atomic.LoadInt32(&this.offLine) != 0 || msg == nil {
		return
	}

	this.sendChan <- msg
}

func (this *TTcpConnect) readMsg() (msg IMessage, ok bool) {
	if _, err := io.ReadFull(this.conn, this.msgLenBuf); err != nil {
		return
	}

	msgLen := binary.LittleEndian.Uint32(this.msgLenBuf)
	if msgLen > this.cfg.MaxMessageSize {
		this.Println("error receiving message len:", msgLen)
		return
	}

	data := make([]byte, msgLen)
	if n, err := io.ReadFull(this.conn, data); err != nil {
		this.Println("error receiving msg, bytes:", n, "msgLen:", msgLen, "reason:", err)
		return
	}

	msg = this.msgPool.Get(this)
	ok = msg.Parse(data)
	if !ok {
		tmp := fmt.Sprintf("receive msg msgLen=%v, len(msg.Data())=%v, msg.Data()=%#v", msgLen, len(msg.Data()), msg.Data())
		this.Println(tmp)
	}

	return
}

func (this *TTcpConnect) writePack(msg IMessage) (ok bool) {
	msgLen := msg.Write2Cache(this.sendCache[4:])

	binary.LittleEndian.PutUint32(this.sendCache, uint32(msgLen))
	n, err := this.conn.Write(this.sendCache[:uint32(msgLen)+4])
	if err != nil {
		this.Println("error writing msg, bytes:", n, "reason:", err)
		return
	}

	this.msgPool.Put(msg)

	return true
}

func (this *TTcpConnect) readLoop() {
	defer this.wg.Done()
	reason := "read done"
	defer this.Close(reason)

	for {
		msg, ok := this.readMsg()
		if !ok {
			reason = "read fail"
			return
		}
		select {
		case this.readChan <- msg:
		case <-this.ctx.Done():
			return
		case <-this.chDone:
			return
		}
	}
}

func (this *TTcpConnect) sendloop() {
	defer this.wg.Done()
	reason := "send done"
	defer this.Close(reason)

	for {
		select {
		case msg := <-this.sendChan:
			if !this.writePack(msg) {
				reason = "write fail"
				return
			}
		case <-this.ctx.Done():
			return
		case <-this.chDone:
			return
		}
	}
}

func (this *TTcpConnect) msgLoop() {
	defer this.wg.Done()
	reason := "message loop done"
	defer this.Close(reason)

	for {
		select {
		case <-this.ctx.Done():
			return
		case <-this.chDone:
			return
		case msg, ok := <-this.readChan:
			if !ok {
				return
			}
			this.cb.OnMessage(msg)
			if len(this.readChan) > cstReadChanSize*cstReadWaterMark {
				this.Println("read chan too long")
				msg.Connect().Close("msg chan too long")
			}
		}
	}
}

func (this *TTcpConnect) Close(reason string) {
	if atomic.CompareAndSwapInt32(&this.offLine, 0, 1) {
		this.Println("close socket: " + reason)
		this.cb.OnDisConnect(this)
		this.conn.Close()
		close(this.chDone)
	}
}

func (this *TTcpConnect) HandleConn() {
	defer this.ctx.WG.Done()

	this.wg.Add(3)
	go this.readLoop()
	go this.sendloop()
	go this.msgLoop()

	this.wg.Wait()
}

func (self *TTcpConnect) Println(para ...interface{}) {
	f := fmt.Sprintf("[%09d:%v---%v]", self.connId, self.conn.LocalAddr(), self.conn.RemoteAddr())
	newPara := make([]interface{}, 0, len(para)+1)
	newPara = append(newPara, f)
	newPara = append(newPara, para...)
	fmt.Println(newPara...)
}

func NewTcpConnect(ctx *chelper.TContext, conn net.Conn, cfg *TTcpConnectCfg, cb ITcpConnectCallback) *TTcpConnect {
	this := &TTcpConnect{
		ctx:       ctx,
		conn:      conn,
		cfg:       cfg,
		sendChan:  make(chan IMessage, cstSendChanSize),
		readChan:  make(chan IMessage, cstReadChanSize),
		offLine:   0,
		cb:        cb,
		chDone:    make(chan struct{}),
		connId:    atomic.AddUint64(&gConnId, 1),
		sendCache: make([]byte, cstMessageLimit+cstMessageHeadSize),
		msgPool:   cb.MessagePool(),
	}

	this.OnlineTime = time.Now().Unix()
	this.msgLenBuf = make([]byte, cstMessageLenSize)

	return this
}
