package main

import (
	"fmt"
	"sync"
	"time"
)

type PacketStat struct {
	Timestamp  time.Time
	Seq        uint32
	ClientSend int64
	ServerRcv  int64
	ClientRcv  int64
}

func (p PacketStat) GetUpstreamLatencyMS() float64 {
	return float64(p.ServerRcv-p.ClientSend) / 1e6
}

func (p PacketStat) GetDownstreamLatencyMS() float64 {
	return float64(p.ClientRcv-p.ServerRcv) / 1e6
}

func (p PacketStat) GetType() string {
	return "game_packet_latency"
}

func (p PacketStat) GetHeader() []string {
	return []string{
		"Timestamp",
		"Seq",
		"ClientSend",
		"ServerRcv",
		"ClientRcv",
		"RTTms",
	}
}

func (p PacketStat) GetRow() []string {
	rtt := (p.ClientRcv - p.ClientSend) / 1e6
	return []string{
		p.Timestamp.Format("2006-01-02 15:04:05.999999999"),
		fmt.Sprint(p.Seq),
		fmt.Sprint(p.ClientSend),
		fmt.Sprint(p.ServerRcv),
		fmt.Sprint(p.ClientRcv),
		fmt.Sprint(rtt),
	}
}

type StatHandler interface {
	OnStat(ps PacketStat)
	Close()
}

type StatManager struct {
	handlers []StatHandler
}

func (sm *StatManager) AddHandler(s StatHandler) {
	sm.handlers = append(sm.handlers, s)
}

func (sm *StatManager) Run(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for s := range packetstats {
		for _, h := range sm.handlers {
			h.OnStat(s)
		}
	}
	for _, h := range sm.handlers {
		h.Close()
	}
	close(csvrecords)
}

type PacketToCSV struct {
}

func (P PacketToCSV) OnStat(ps PacketStat) {
	csvrecords <- ps
}

func (P PacketToCSV) Close() {

}
