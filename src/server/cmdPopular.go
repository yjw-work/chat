package main

import (
	"library/cnet"
	"sort"
	"strings"
	"time"
)

const (
	cstStatisticSeconds = 5
)

func (this *TChatServer) cmdPopular(conn *cnet.TTcpConnect, content string) {
	this.statisticLock.Lock()

	this.removeOldStat()
	if len(this.wordStatisticOrder) > 0 {
		content = this.wordStatisticOrder[0]
	} else {
		content = ""
	}
	this.statisticLock.Unlock()

	if len(content) > 0 {
		this.send2Conn(conn, content)
	}
}

func (this *TChatServer) addPopular(content string) {
	content = strings.Replace(content, "*", "", -1)
	words := strings.Split(content, " \t\n\r")
	stat := &TWordStatistic{
		WordsCnt: make(map[string]int, len(words)),
		Time:     time.Now().Unix(),
	}

	this.statisticLock.Lock()
	defer this.statisticLock.Unlock()

	for _, w := range words {
		stat.WordsCnt[w] = stat.WordsCnt[w] + 1
	}

	this.removeOldStat()

	for w, cnt := range stat.WordsCnt {
		this.addWordStatistic(w, cnt)
	}

	this.recentStatistic.PushBack(stat)
}

func (this *TChatServer) removeOldStat() {
	now := time.Now().Unix()
	e := this.recentStatistic.Front()
	for nil != e {
		var statistic *TWordStatistic
		if nil != e.Value {
			var ok bool
			if statistic, ok = e.Value.(*TWordStatistic); ok && nil != statistic {
				if now < statistic.Time+cstStatisticSeconds {
					break
				}
			}
		}
		del := e
		e = e.Next()
		this.recentStatistic.Remove(del)
		for w, cnt := range statistic.WordsCnt {
			this.removeWordStatistic(w, cnt)
		}
	}
}

func (this *TChatServer) addWordStatistic(word string, cnt int) {
	oldCnt, ok := this.wordStatistic[word]
	this.wordStatistic[word] = oldCnt + cnt
	if ok {
		this.sortWordStatistic()
	} else {
		this.wordStatisticOrder = append(this.wordStatisticOrder, word)
	}
}

func (this *TChatServer) removeWordStatistic(word string, cnt int) {
	oldCnt, ok := this.wordStatistic[word]
	if !ok {
		return
	}
	cnt = oldCnt - cnt
	if cnt <= 0 {
		delete(this.wordStatistic, word)
		for i, w := range this.wordStatisticOrder {
			if w == word {
				this.wordStatisticOrder = append(this.wordStatisticOrder[:i], this.wordStatisticOrder[i+1:]...)
				break
			}
		}
	} else {
		this.wordStatistic[word] = cnt
		this.sortWordStatistic()
	}
}

func (this *TChatServer) sortWordStatistic() {
	sort.SliceStable(this.wordStatisticOrder, func(i, j int) bool {
		return this.wordStatistic[this.wordStatisticOrder[i]] > this.wordStatistic[this.wordStatisticOrder[j]]
	})
}
