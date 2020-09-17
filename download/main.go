package main

import "flag"

func main() {
	addr := flag.String("addr", "localhost:8910", "Addr for server to listen or client to connect")
	server := flag.Bool("server", false, "Run in server mode")
	l := flag.String("l", "info", "Log Level")
	client := flag.Bool("client", false, "Run in client mode")
	n := flag.Int("n", 1, "number of parallel downloads")
	size := flag.Int("filesize", 100, "filesize to download in KB")
	flag.Parse()

	setupLogging(*l)
	if *server {
		Server(*addr)
	}

	if *client {
		var c Client
		c.Download(*addr, *size*1000, *n)
	}
}
