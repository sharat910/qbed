package main

import (
	"flag"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	addr := flag.String("addr", "localhost:8910", "Addr for server to listen or client to connect")
	server := flag.Bool("server", false, "Run in server mode")
	l := flag.String("l", "info", "Log Level")
	client := flag.Bool("client", false, "Run in client mode")
	p := flag.Int("p", 1, "number of parallel threads")
	c := flag.Int("c", 1, "number of files to fetch")
	irt := flag.Int("irt", 0, "inter request time")
	size := flag.Int("filesize", 100, "filesize to download in KB")
	std := flag.Int("std", 0, "std of filesize (use with dist)")
	dist := flag.Bool("norm", false, "use normal distribution")
	flag.Parse()

	setupLogging(*l)
	if *server {
		Server(*addr)
	}

	if *client {
		client := NewClient()
		client.SpawnThreads(*p)
		requests := RequestGenerator(ReqConfig{
			Addr:     *addr,
			Size:     *size,
			Parallel: *p,
			Count:    *c,
			Dist:     *dist,
			Std:      *std,
		})
		start := time.Now()
		for _, req := range requests {
			t := time.Now()
			client.AddNewRequest(req)
			elapsed := time.Since(t)
			interReqDur := time.Duration(*irt) * time.Second
			if interReqDur > elapsed {
				time.Sleep(interReqDur - elapsed)
			}
		}
		client.WaitUntilFinished()
		log.Info().Str("duration", time.Since(start).String()).Int("num_reqs", len(requests)).Msg("download complete")
	}
}
