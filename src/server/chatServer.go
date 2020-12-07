package main

import (
	"bufio"
	"common"
	"container/list"
	"fmt"
	"io"
	"library/calgorithm"
	"library/chelper"
	"library/cnet"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
)

const (
	cstCmdPrefix  string = "/"
	cstCmdPopular string = cstCmdPrefix + "popular"
	cstCmdStats   string = cstCmdPrefix + "stats"
)

type TChatServerConfig struct {
	cnet.TTcpServerConfig
	RecentLength  int
	FiltWordsFile string
}

type TChatRecord struct {
	ConnId  string
	Content string
}

type TWordStatistic struct {
	WordsCnt map[string]int
	Time     int64
}

type MsgProc func(conn *cnet.TTcpConnect, content string)

type TChatServer struct {
	svr     *cnet.TTcpServer
	cfg     *TChatServerConfig
	clients sync.Map
	ctx     *chelper.TContext

	cmdProcs map[string]MsgProc

	trie *TTrie

	// recent chat list
	recentLock sync.RWMutex
	recentChat *calgorithm.TCircleQueue

	// cmd popular
	statisticLock      sync.Mutex
	recentStatistic    list.List
	wordStatistic      map[string]int
	wordStatisticOrder []string
}

func NewChatServer() *TChatServer {
	defLen := 100
	this := &TChatServer{
		wordStatistic:      make(map[string]int, defLen),
		wordStatisticOrder: make([]string, 0, defLen),
		trie:               NewTrie(),
	}
	this.ctx = chelper.NewContext()
	if nil == this.ctx {
		return nil
	}
	this.cmdProcs = map[string]MsgProc{
		cstCmdPopular: this.cmdPopular,
		cstCmdStats:   this.cmdStats,
	}
	return this
}

func (this *TChatServer) init(cfgPath string) error {
	if len(cfgPath) <= 0 {
		d, n, _, err := chelper.BinPathSplit()
		if nil != err {
			return err
		}
		cfgPath = path.Join(d, n+".json")
	}
	cfg := &TChatServerConfig{}
	if err := chelper.ReadJsonFile(cfgPath, cfg); nil != err {
		return err
	}
	if cfg.RecentLength <= 0 || cfg.MaxMessageSize <= 0 {
		return fmt.Errorf("wrong cfg: %v", cfg)
	}

	this.svr = cnet.NewTcpServer(this, &cfg.TTcpServerConfig)
	if nil == this.svr {
		return fmt.Errorf("new tcp server fail")
	}
	this.cfg = cfg
	this.recentChat = calgorithm.NewCircleQueue(cfg.RecentLength)
	if nil == this.recentChat {
		return fmt.Errorf("new recent chat queue fail")
	}

	f, err := os.Open(this.cfg.FiltWordsFile)
	if nil != err {
		return err
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		line = strings.TrimRight(line, "\n")
		this.trie.Insert(line)
	}

	return nil
}

func (this *TChatServer) run() {

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	this.ctx.WG.Add(1)
	go this.signalLoop(signalChan)

	this.ctx.WG.Add(1)
	go this.svr.Run(this.ctx)

	this.ctx.WG.Wait()
}

func (this *TChatServer) stop() {
	this.svr.Stop()
}

func (this *TChatServer) OnConnect(conn *cnet.TTcpConnect) {
	val, loaded := this.clients.LoadOrStore(conn.ConnId(), conn)
	if loaded {
		oldConn, ok := val.(*cnet.TTcpConnect)
		if !ok {
			fmt.Println("connect convert fail: ", conn.NetInfo())
		} else if nil != oldConn {
			oldConn.Close("repeat connect")
		}
	}

	this.recentLock.RLock()
	records := make([]*TChatRecord, 0, this.recentChat.Length())
	for i := 0; i < this.recentChat.Length(); i++ {
		data := this.recentChat.Item(i)
		if nil == data {
			continue
		}
		if r, ok := data.(*TChatRecord); ok && nil != r {
			records = append(records, r)
		}
	}
	this.recentLock.RUnlock()

	content := ""
	for _, r := range records {
		content += fmt.Sprintf("%s: %s\n", r.ConnId, r.Content)
	}
	this.send2Conn(conn, content)
}

func (this *TChatServer) OnDisConnect(conn *cnet.TTcpConnect) {
	this.clients.Delete(conn.ConnId())
}

func (this *TChatServer) OnMessage(msg cnet.IMessage) {
	data := msg.Data()
	if nil == data || len(data) <= 0 {
		return
	}

	content := string(data)
	if content[:1] == cstCmdPrefix {
		cmd := content
		cmdContent := content
		idx := strings.Index(content, " ")
		if idx >= 0 {
			cmd = content[:idx]
			cmdContent = content[idx+1:]
		}
		if proc, ok := this.cmdProcs[cmd]; ok {
			proc(msg.Connect(), cmdContent)
			return
		}
	}
	this.onChat(msg.Connect(), content)
}

func (this *TChatServer) send2Conn(conn *cnet.TTcpConnect, content string) {
	msg := this.MessagePool().Get(conn).SetData([]byte(content))
	conn.Send(msg)
}

func (this *TChatServer) onChat(conn *cnet.TTcpConnect, content string) {
	content = this.trie.Filt(content)

	this.recentLock.Lock()
	this.recentChat.Push(&TChatRecord{
		ConnId:  fmt.Sprint(conn.ConnId()),
		Content: content,
	})
	this.recentLock.Unlock()

	this.clients.Range(func(key, value interface{}) bool {
		select {
		case <-this.ctx.Done():
			return false
		default:
			if dstConn, ok := value.(*cnet.TTcpConnect); ok && nil != dstConn && dstConn.ConnId() != conn.ConnId() {
				this.send2Conn(dstConn, fmt.Sprintf("%v: %v", conn.ConnId(), content))
			}
			return true
		}
	})

	this.addPopular(content)
}

func (this *TChatServer) MessagePool() cnet.IMessagePool {
	return common.DefMessagePool
}

func (this *TChatServer) signalLoop(ch chan os.Signal) {
	defer this.ctx.WG.Done()

	for {
		select {
		case <-this.ctx.Done():
			return
		case <-ch:
			this.svr.Stop()
			this.ctx.CloseDone()
			return
		}
	}
}
