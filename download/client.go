package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

type Client struct {
	requests chan *http.Request
	counter  *ByteCounter
	wg       sync.WaitGroup
}

func NewClient() Client {
	counter := &ByteCounter{
		stopChan: make(chan struct{}),
	}
	go counter.ComputeRate()
	return Client{
		requests: make(chan *http.Request),
		counter:  counter,
	}
}

func (c *Client) AddNewRequest(r *http.Request) {
	c.requests <- r
}

func (c *Client) SpawnThreads(n int) {
	for i := 0; i < n; i++ {
		c.wg.Add(1)
		go func(idx int) {
			defer c.wg.Done()
			log.Debug().Int("thread_index", idx).Msg("thread started")
			var httpclient http.Client
			for req := range c.requests {
				log.Debug().Int("thread_index", idx).Str("url", req.URL.String()).Msg("got req")
				start := time.Now()
				resp, err := httpclient.Do(req)
				if err != nil {
					log.Fatal().Err(err).Msg("unable to do req")
				}
				log.Debug().Msg("req sent")
				objsize, err := io.Copy(c.counter, resp.Body)
				if err != nil {
					log.Fatal().Err(err).Msg("unable to read body")
				}
				log.Debug().Msg("io copy done")
				err = resp.Body.Close()
				if err != nil {
					log.Fatal().Err(err).Msg("unable to close body")
				}
				log.Debug().Msg("resp closed")
				oct := time.Since(start)
				log.Info().Int64("size", objsize).Dur("oct", oct).
					Int64("rate_kbps", (8*objsize)/oct.Milliseconds()).
					Msg("obj downloaded")
			}
			log.Debug().Int("thread_index", idx).Msg("thread exit")
		}(i)
	}
}

func (c *Client) WaitUntilFinished() {
	close(c.requests)
	c.wg.Wait()
	c.counter.Stop()
}

func (c *Client) DownloadOneFile(addr string, filesize, parallel int) {
	log.Info().Time("start_time", time.Now()).Str("filesize", humanize.Bytes(uint64(filesize))).
		Int("num_goroutines", parallel).Msg("Starting download")
	start := time.Now()

	counter := &ByteCounter{stopChan: make(chan struct{})}
	go counter.ComputeRate()

	var wg sync.WaitGroup
	url := fmt.Sprintf("http://%s/download?size=%d", addr, filesize)
	length := GetBodyLength(url)
	lenEach := length / parallel
	diff := length % parallel // Get the remaining for the last request

	for i := 0; i < parallel; i++ {
		wg.Add(1)

		min := lenEach * i       // Min range
		max := lenEach * (i + 1) // Max range

		if i == parallel-1 {
			max += diff // Add the remaining bytes in the last request
		}

		go func(min int, max int, i int) {
			start := time.Now()
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
			rangeHeader := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
			req.Header.Add("Range", rangeHeader)
			resp, _ := client.Do(req)
			n, err := io.Copy(counter, resp.Body)
			if err != nil {
				log.Fatal().Err(err).Msg("unable to read body")
			}
			err = resp.Body.Close()
			if err != nil {
				log.Fatal().Err(err).Msg("unable to close body")
			}
			log.Debug().Int("gor_id", i).Int64("byte_count", n).Dur("durMS", time.Since(start)).
				Str("range_header", rangeHeader).Msg("goroutine done.")
			wg.Done()
		}(min, max, i)
	}
	wg.Wait()
	counter.Stop()
	log.Info().Int("parallel", parallel).
		Str("filesize", humanize.Bytes(uint64(filesize))).
		Dur("durMS", time.Since(start)).
		Int64("avg_speed", 8*int64(filesize)/time.Since(start).Milliseconds()).
		Msg("download finished")
}
