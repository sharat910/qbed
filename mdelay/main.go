package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/google/gopacket/layers"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog/log"
)

type DelayData struct {
	LastMeanDelayUS int64 `json:"last_mean_delay_us"`
	LastPacketCount int64 `json:"last_packet_count"`
}

type MDelay struct {
	packets map[[16]byte]gopacket.Packet
	lock    sync.Mutex
	delays  []time.Duration
	dd      DelayData
}

func NewMDelay() *MDelay {
	return &MDelay{
		packets: make(map[[16]byte]gopacket.Packet),
	}
}

func (md *MDelay) Run(ingressIface, egressIface, bpf string, d time.Duration) {
	var wg sync.WaitGroup
	wg.Add(4)
	ctx, cancel := context.WithCancel(context.Background())
	go md.capture(ingressIface, bpf, true, ctx, &wg)
	go md.capture(egressIface, bpf, false, ctx, &wg)
	go md.DelayLoop(d, ctx, &wg)
	go md.httpserver(6789, ctx, &wg)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	wg.Wait()
}

func (md *MDelay) httpserver(port int, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		md.lock.Lock()
		err := json.NewEncoder(w).Encode(md.dd)
		md.lock.Unlock()
		if err != nil {
			log.Error().Err(err).Msg("json encoder error")
		}
	})

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: addr, Handler: nil}

	go func() {
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("server error")
		}
	}()

	log.Info().Str("addr", addr).Msg("server starting...")
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("server error")
	}

}

func (md *MDelay) capture(iface, bpf string, ingress bool, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	rcvHandle, err := pcap.OpenLive(iface, 9600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open iface")
	}

	err = rcvHandle.SetBPFFilter(bpf)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to set bpf")
	}

	//var (
	//	// Will reuse these for each packet
	//	ethLayer   layers.Ethernet
	//	ip4Layer   layers.IPv4
	//	icmp4Layer layers.ICMPv4
	//	tcpLayer   layers.TCP
	//	udpLayer   layers.UDP
	//)
	//
	//parser := gopacket.NewDecodingLayerParser(
	//	layers.LayerTypeEthernet,
	//	&ethLayer,
	//	&ip4Layer,
	//	&icmp4Layer,
	//	&tcpLayer,
	//	&udpLayer,
	//)

	packetSource := gopacket.NewPacketSource(rcvHandle, rcvHandle.LinkType())
	packetSource.DecodeOptions.Lazy = true
	packetSource.DecodeOptions.NoCopy = true

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-packetSource.Packets():

			if p.NetworkLayer() != nil {
				md.MatchPacket(ingress, p)
			}
			//var foundLayerTypes []gopacket.LayerType
			//_ = parser.DecodeLayers(p.Data(), &foundLayerTypes)
			//for _, layerType := range foundLayerTypes {
			//	switch layerType {
			//	case layers.LayerTypeICMPv4:
			//		//fmt.Println("Recv:", p.Metadata().Timestamp, p.NetworkLayer().NetworkFlow(), icmp4Layer.TypeCode, icmp4Layer.Id, icmp4Layer.Seq, icmp4Layer.Checksum)
			//		md.MatchPacket(ingress, p, icmp4Layer)
			//	case layers.LayerTypeUDP:
			//		//fmt.Printf("%s %s:%d -> %s:%d = %x\n", p.Metadata().Timestamp, ip4Layer.SrcIP, udpLayer.SrcPort, ip4Layer.DstIP, udpLayer.SrcPort, udpLayer.Checksum)
			//	}
			//}
		}
	}
}

func (md *MDelay) MatchPacket(ingress bool, p gopacket.Packet) {
	hash := md5.Sum(p.Data())
	md.lock.Lock()
	if ingress {
		futurePacket, exists := md.packets[hash]
		if !exists {
			md.packets[hash] = p
		} else {
			md.delays = append(md.delays, futurePacket.Metadata().Timestamp.Sub(p.Metadata().Timestamp))
			delete(md.packets, hash)
		}
	} else {
		lastPacket, exists := md.packets[hash]
		if !exists {
			md.packets[hash] = p
		} else {
			md.delays = append(md.delays, p.Metadata().Timestamp.Sub(lastPacket.Metadata().Timestamp))
			delete(md.packets, hash)
		}
	}
	md.lock.Unlock()
}

func (md *MDelay) DelayLoop(d time.Duration, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			md.lock.Lock()
			var sumDelay time.Duration
			var count, meanDelayUS int64
			for _, delay := range md.delays {
				sumDelay += delay
				count++
			}
			if count != 0 {
				meanDelayUS = sumDelay.Microseconds() / count
			}
			fmt.Println(meanDelayUS, count)
			md.dd.LastMeanDelayUS = meanDelayUS
			md.dd.LastPacketCount = count
			md.delays = md.delays[:0]
			md.lock.Unlock()
		}
	}
}

func (md *MDelay) MatchPacketICMP(ingress bool, p gopacket.Packet, icmp layers.ICMPv4) {
	hash := md5.Sum(p.Data())
	md.lock.Lock()
	if ingress {
		futurePacket, exists := md.packets[hash]
		if !exists {
			//fmt.Println("Adding:", p.NetworkLayer().NetworkFlow(), icmp.TypeCode, icmp.Id, icmp.Seq, icmp.Checksum)
			md.packets[hash] = p
		} else {
			t1, t2 := p.Metadata().Timestamp, futurePacket.Metadata().Timestamp
			fmt.Println(icmp.Seq, t2.After(t1), t2.Sub(t1))
			//fmt.Println("Removing:", p.NetworkLayer().NetworkFlow(), icmp.TypeCode, icmp.Id, icmp.Seq, icmp.Checksum)
			delete(md.packets, hash)
		}

	} else {
		lastPacket, exists := md.packets[hash]
		if exists {
			t1, t2 := p.Metadata().Timestamp, lastPacket.Metadata().Timestamp
			fmt.Println(icmp.Seq, t1.After(t2), t1.Sub(t2))
			//fmt.Printf("%s\r", p.Metadata().Timestamp.Sub(lastPacket.Metadata().Timestamp))
			//fmt.Println("Removing:", p.NetworkLayer().NetworkFlow(), icmp.TypeCode, icmp.Id, icmp.Seq, icmp.Checksum)
			delete(md.packets, hash)
		} else {
			//fmt.Println("Doesn't exist:", p.NetworkLayer().NetworkFlow(), icmp.TypeCode, icmp.Id, icmp.Seq, icmp.Checksum)
			//fmt.Println("Adding:", p.NetworkLayer().NetworkFlow(), icmp.TypeCode, icmp.Id, icmp.Seq, icmp.Checksum)
			md.packets[hash] = p
		}
	}
	md.lock.Unlock()
}

func main() {
	in := flag.String("in", "swport2", "Interface 1")
	eg := flag.String("eg", "swport1", "Interface 2")
	bpf := flag.String("bpf", "dst 10.0.0.1", "Filter packets")
	d := flag.Int("i", 100, "Print interval ms")
	flag.Parse()
	md := NewMDelay()
	md.Run(*in, *eg, *bpf, time.Duration(*d)*time.Millisecond)
}
