package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"golang.org/x/exp/rand"

	"gonum.org/v1/gonum/stat/distuv"
)

type ReqConfig struct {
	Addr     string
	Size     int // Size of each object requested
	Parallel int // No. of byte ranges
	Count    int // No. of requests to generate

	Dist bool
	Std  int
}

func RequestGenerator(config ReqConfig) []*http.Request {
	var requests []*http.Request
	norm := distuv.Normal{
		Mu:    float64(config.Size),
		Sigma: float64(config.Std),
		Src:   rand.NewSource(0),
	}
	for i := 0; i < config.Count; i++ {
		filesize := config.Size
		if config.Dist {
			filesize = int(norm.Rand())
		}
		url := fmt.Sprintf("http://%s/download?size=%d", config.Addr, filesize)
		//length := GetBodyLength(url)
		//log.Debug().Int("filesize", filesize).Int("bodylen", length).Msg("assert")
		lenEach := filesize / config.Parallel
		diff := filesize % config.Parallel // Get the remaining for the last request
		for i := 0; i < config.Parallel; i++ {
			min := lenEach * i       // Min range
			max := lenEach * (i + 1) // Max range
			if i == config.Parallel-1 {
				max += diff // Add the remaining bytes in the last request
			}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Fatal().Err(err).Msg("unable to build request")
			}
			rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
			req.Header.Add("Range", rangeHeader)
			requests = append(requests, req)
		}
	}
	return requests
}

func GetBodyLength(url string) int {
	start := time.Now()
	res, err := http.Head(url)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to request head")
	}
	length, err := strconv.Atoi(res.Header.Get("Content-Length")) // Get the content length from the header request
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get content length")
	}
	log.Debug().Str("url", url).Dur("durMS", time.Since(start)).Msg("got body length")
	return length
}
