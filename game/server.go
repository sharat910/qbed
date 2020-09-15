package main

import (
	"encoding/binary"
	"github.com/rs/zerolog/log"

	"net"
	"time"
)

type Server struct {
	conn *net.UDPConn
}

func (s *Server) Listen(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to resolve udp addr")
	}

	// setup listener for incoming UDP connection
	s.conn, err = net.ListenUDP("udp", udpAddr)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Str("addr", addr).Msg("UDP server up and listening")

	for {
		// wait for UDP client to connect
		s.handleUDPConnections()
	}

}

func (s *Server) handleUDPConnections() {
	buffer := make([]byte, 1024)

	// Blocking call
	n, addr, err := s.conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read from UDP conn")
	}

	now := time.Now()
	binary.BigEndian.PutUint64(buffer[n:], uint64(now.UnixNano()))
	_, err = s.conn.WriteToUDP(buffer[:n+8], addr)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to write to UDP conn")
	}

}
