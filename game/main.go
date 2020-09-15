package main

import (
	"flag"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	addr := flag.String("addr", "localhost:6000", "Addr for server to listen or client to connect")
	server := flag.Bool("server", false, "Run in server mode")
	client := flag.Bool("client", false, "Run in client mode")
	tick := flag.Int("tick", 64, "tick rate to send packets at")
	n := flag.Int("n", 64, "number of packets to send")
	l := flag.String("l", "info", "Log Level")
	flag.Parse()

	setupLogging(*l)

	if *server {
		var s Server
		s.Listen(*addr)
	}

	if *client {
		var wg sync.WaitGroup
		stats = make(chan Stat, 100)
		go CSVDumper("data", &wg)
		var c Client
		c.Connect(*addr)
		go c.HandleReplies(&wg)
		c.SendPacketsAtTick(*tick, *n)
		time.Sleep(500 * time.Millisecond)
		c.Close()
		close(stats)
		wg.Wait()
		log.Info().Msg("client done.")
	}

}
