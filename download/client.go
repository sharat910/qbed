package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

type Client struct {
}

type Counter struct {
	bytesDownloaded int64
	lastLog         time.Time
	lastDownloaded  int64
	lock            sync.Mutex
}

func (c *Counter) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&c.bytesDownloaded, int64(len(p)))
	now := time.Now()
	c.lock.Lock()
	if now.Sub(c.lastLog) > time.Second {
		rate := 8 * (c.bytesDownloaded - c.lastDownloaded) / now.Sub(c.lastLog).Milliseconds()
		log.Info().
			Int64("rate_kbps", rate).
			Str("bytes_downloaded", humanize.Bytes(uint64(c.bytesDownloaded))).
			Msg("progress")
		c.lastDownloaded = c.bytesDownloaded
		c.lastLog = now
	}
	c.lock.Unlock()
	return len(p), nil
}

func (c *Client) Download(addr string, filesize, parallel int) {
	log.Info().Time("start_time", time.Now()).Str("filesize", humanize.Bytes(uint64(filesize))).
		Int("num_goroutines", parallel).Msg("Starting download")
	start := time.Now()
	var wg sync.WaitGroup
	url := fmt.Sprintf("http://%s/download?size=%d", addr, filesize)
	length := c.GetBodyLength(url)
	lenEach := length / parallel
	diff := length % parallel // Get the remaining for the last request
	counter := &Counter{}
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
	log.Info().Int("parallel", parallel).
		Str("filesize", humanize.Bytes(uint64(filesize))).
		Dur("durMS", time.Since(start)).
		Int64("avg_speed", 8*int64(filesize)/time.Since(start).Milliseconds()).
		Msg("download finished")
}

func (c *Client) GetBodyLength(url string) int {
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
