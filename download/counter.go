package main

import (
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog/log"
)

type Counter struct {
	bytesDownloaded int64
	lastLog         time.Time
	lastDownloaded  int64
	lock            sync.Mutex
	wg              sync.WaitGroup
	stopChan        chan struct{}
}

func (c *Counter) Write(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.bytesDownloaded += int64(len(p))
	return len(p), nil
}

func (c *Counter) ComputeRate() {
	c.wg.Add(1)
	defer c.wg.Done()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			c.GetRate()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Counter) Stop() {
	c.stopChan <- struct{}{}
	c.wg.Wait()
}

func (c *Counter) GetRate() {
	c.lock.Lock()
	defer c.lock.Unlock()
	now := time.Now()
	rate := 8 * (c.bytesDownloaded - c.lastDownloaded) / now.Sub(c.lastLog).Milliseconds()
	log.Info().
		Int64("rate_kbps", rate).
		Str("bytes_downloaded", humanize.Bytes(uint64(c.bytesDownloaded))).
		Msg("progress")
	c.lastDownloaded = c.bytesDownloaded
	c.lastLog = now
}
