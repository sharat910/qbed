package main

import (
	"encoding/binary"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	conn *net.UDPConn
	seq  uint32
}

func (c *Client) Connect(addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	c.conn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("unable to connect to server")
	}

	log.Info().Str("addr", addr).Msg("connected to server")
	log.Debug().
		Str("server", c.conn.RemoteAddr().String()).
		Str("client", c.conn.LocalAddr().String()).
		Msg("endpoints")
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to close connection to server")
	}
}

// GetPacketData returns 12 byte data containing seq num and send timestamp
func (c *Client) GetPacketData() []byte {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[:4], c.seq)
	binary.BigEndian.PutUint64(data[4:], uint64(time.Now().UnixNano()))
	return data
}

func (c *Client) SendOnePacket() {
	data := c.GetPacketData()
	_, err := c.conn.Write(data)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to send packets")
	}
	c.seq++
}

func (c *Client) SendPacketsAtTick(tick, n int) {
	log.Info().Int("tick_rate", tick).Int("packet_count", n).Msg("sending packets")
	ticker := time.NewTicker(time.Second / time.Duration(tick))
	for i := 0; i < n; i++ {
		<-ticker.C
		c.SendOnePacket()
	}
}

func (c *Client) HandleReplies(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	buffer := make([]byte, 1024)
	numpackets := 0
	for {
		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				if numpackets == 0 {
					log.Warn().Msg("Did not receive replies from server")
				}
				return
			}
			log.Fatal().Err(err).Msg("unable to receive")
		}
		now := time.Now()
		rData := buffer[:n]
		var pls PacketLatencyStat
		pls.Timestamp = now
		pls.ClientRcv = now.UnixNano()
		pls.Seq = binary.BigEndian.Uint32(rData[:4])
		pls.ClientSend = int64(binary.BigEndian.Uint64(rData[4:12]))
		pls.ServerRcv = int64(binary.BigEndian.Uint64(rData[12:]))
		stats <- pls
		numpackets++
	}
}
