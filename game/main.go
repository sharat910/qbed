package main

import (
	"flag"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var csvrecords chan CSVRecord
var packetstats chan PacketStat

func main() {
	addr := flag.String("addr", "localhost:6000", "Addr for server to listen or client to connect")
	server := flag.Bool("server", false, "Run in server mode")
	client := flag.Bool("client", false, "Run in client mode")
	tick := flag.Int("tick", 64, "tick rate to send packets at")
	n := flag.Int("n", 64, "number of packets to send")
	t := flag.Int("t", 0, "time to send packets for (supersedes n)")
	l := flag.String("l", "info", "Log Level")
	flag.Parse()

	setupLogging(*l)

	if *server {
		StartServer(addr)
	}

	if *client {
		if *t != 0 {
			*n = *t * *tick
		}
		StartClient(*addr, *tick, *n)
	}

}

func StartClient(addr string, tick int, n int) {
	var wg sync.WaitGroup
	packetstats = make(chan PacketStat, 100)
	csvrecords = make(chan CSVRecord, 10)

	go CSVDumper("data", &wg)

	var sm StatManager
	sm.AddHandler(PacketToCSV{})
	sm.AddHandler(&MetricGenerator{LogToCSV: true})
	go sm.Run(&wg)

	var c Client
	c.Connect(addr)
	go c.HandleReplies(&wg)
	c.SendPacketsAtTick(tick, n)
	time.Sleep(500 * time.Millisecond)
	c.Close()
	wg.Wait()
	log.Info().Msg("client done.")
}

func StartServer(addr *string) {
	var s Server
	s.Listen(*addr)
}
