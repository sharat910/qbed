package main

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/montanaflynn/stats"
)

const MetricInterval = time.Second

type Metric struct {
	RTT                   float64
	MeanLatencyUpstream   float64
	MeanLatencyDownstream float64
	STDLatencyUpstream    float64
	STDLatencyDownstream  float64
	NPackets              int
}

func (m Metric) Log() {
	log.Info().
		Int("count", m.NPackets).
		Str("rtt", fmt.Sprintf("%.2f", m.RTT)).
		Str("mean_latency_up", fmt.Sprintf("%.2f", m.MeanLatencyUpstream)).
		Str("mean_latency_down", fmt.Sprintf("%.2f", m.MeanLatencyDownstream)).
		Str("std_latency_up", fmt.Sprintf("%.2f", m.STDLatencyUpstream)).
		Str("std_latency_down", fmt.Sprintf("%.2f", m.STDLatencyDownstream)).
		Msg("metrics (ms)")
}

type MetricGenerator struct {
	LogToCSV        bool
	packets         []PacketStat
	LastExported    time.Time
	FirstPacketSeen bool
}

func (mg *MetricGenerator) OnStat(ps PacketStat) {
	if !mg.FirstPacketSeen {
		mg.LastExported = ps.Timestamp
		mg.FirstPacketSeen = true
	}

	for ps.Timestamp.After(mg.LastExported.Add(MetricInterval)) {
		// One sec elapsed since LastExported
		mg.Generate()
		mg.packets = mg.packets[:0]
		mg.LastExported = mg.LastExported.Add(MetricInterval)
	}

	mg.packets = append(mg.packets, ps)
}

func (mg *MetricGenerator) Generate() {
	var upstreamLatencies []float64
	var downstreamLatencies []float64
	for _, p := range mg.packets {
		upstreamLatencies = append(upstreamLatencies, p.GetUpstreamLatencyMS())
		downstreamLatencies = append(upstreamLatencies, p.GetUpstreamLatencyMS())
	}
	var m Metric
	m.NPackets = len(mg.packets)
	m.MeanLatencyUpstream, _ = stats.Mean(upstreamLatencies)
	m.MeanLatencyDownstream, _ = stats.Mean(downstreamLatencies)
	m.STDLatencyUpstream, _ = stats.StandardDeviation(upstreamLatencies)
	m.STDLatencyDownstream, _ = stats.StandardDeviation(downstreamLatencies)
	m.RTT = m.MeanLatencyDownstream + m.MeanLatencyUpstream
	m.Log()
	if mg.LogToCSV {
		csvrecords <- AutoCSVRecord{m}
	}
}

func (mg *MetricGenerator) Close() {
	mg.Generate()
}
