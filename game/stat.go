package main

import (
	"fmt"
	"time"
)

var stats chan Stat

type Stat interface {
	GetType() string
	GetHeader() []string
	GetRow() []string
}

type PacketLatencyStat struct {
	Timestamp  time.Time
	Seq        uint32
	ClientSend int64
	ServerRcv  int64
	ClientRcv  int64
}

func (p PacketLatencyStat) GetType() string {
	return "game_packet_latency"
}

func (p PacketLatencyStat) GetHeader() []string {
	return []string{
		"Timestamp",
		"Seq",
		"ClientSend",
		"ServerRcv",
		"ClientRcv",
		"RTT_uS",
	}
}

func (p PacketLatencyStat) GetRow() []string {
	return []string{
		p.Timestamp.Format("2006-01-02 15:04:05.999999999"),
		fmt.Sprint(p.Seq),
		fmt.Sprint(p.ClientSend),
		fmt.Sprint(p.ServerRcv),
		fmt.Sprint(p.ClientRcv),
		fmt.Sprint((p.ClientRcv - p.ClientSend) / 1000),
	}
}
