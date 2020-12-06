package main

import (
	"fmt"
	"library/cnet"
	"strconv"
	"time"
)

const (
	cstMinuteSec int64 = 60
	cstHourSec         = cstMinuteSec * 60
	cstDaySec          = cstHourSec * 24
)

func (this *TChatServer) cmdStats(conn *cnet.TTcpConnect, content string) {
	id, err := strconv.Atoi(content)
	if nil != err {
		fmt.Println(err)
		return
	}
	data, ok := this.clients.Load(uint64(id))
	if !ok {
		fmt.Println("no user for id ", id)
		return
	}
	dstConn, ok := data.(*cnet.TTcpConnect)
	if !ok || nil == dstConn {
		fmt.Println("user connet convert fail: ", id)
		return
	}

	elapsed := time.Now().Unix() - dstConn.OnlineTime

	day := elapsed / cstDaySec
	elapsed %= cstDaySec

	hour := elapsed / cstHourSec
	elapsed %= cstHourSec

	minute := elapsed / cstMinuteSec
	elapsed %= cstMinuteSec

	content = fmt.Sprintf("%02dd %02dh %02dm %02ds",
		day, hour, minute, elapsed,
	)
	this.send2Conn(conn, content)
}
