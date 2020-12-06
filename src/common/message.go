package common

import (
	"library/cnet"
	"sync"
)

var (
	DefMessagePool = &TMessagePool{
		packPool: sync.Pool{
			New: func() interface{} {
				return &TMessage{}
			},
		},
	}
)

type TMessagePool struct {
	packPool sync.Pool
}

func (this *TMessagePool) Get(conn *cnet.TTcpConnect) cnet.IMessage {
	newMsg := this.packPool.Get().(*TMessage)
	newMsg.conn = conn
	return newMsg
}

func (this *TMessagePool) Put(msg cnet.IMessage) {
	if m, ok := msg.(*TMessage); ok {
		m.Clear()
		this.packPool.Put(msg)
	}
}

type TMessage struct {
	data []byte
	conn *cnet.TTcpConnect
}

func (this *TMessage) Parse(data []byte) (ok bool) {
	this.data = data
	return true
}

func (this *TMessage) SetConnect(conn *cnet.TTcpConnect) {
	this.conn = conn
}

func (this *TMessage) Connect() *cnet.TTcpConnect {
	return this.conn
}

func (this *TMessage) Data() []byte {
	return this.data
}

func (this *TMessage) SetData(data []byte) cnet.IMessage {
	this.data = data
	return this
}

func (this *TMessage) Write2Cache(cache []byte) (writeLen int) {
	copy(cache, this.data)
	return len(this.data)
}

func (this *TMessage) Clear() {
	this.data = nil
}
