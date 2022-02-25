package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tjmerritt/udp-replicate/internal/packet"
)

type PacketLog struct {
	reportInterval time.Duration
	input          <-chan *packet.Packet
	wg             sync.WaitGroup
	done           chan struct{}
}

func NewPacketLog(report time.Duration) *PacketLog {
	fmt.Printf("Report interval: %v\n", report)
	return &PacketLog{
		reportInterval: report,
	}
}

func (pg *PacketLog) SetInput(input <-chan *packet.Packet) {
	pg.input = input
}

func (pg *PacketLog) Start() {
	if pg.done == nil {
		pg.done = make(chan struct{}, 1)
		pg.wg.Add(1)
		go pg.run()
	}
}

func (pg *PacketLog) Stop() {
	if pg.done != nil {
		pg.done <- struct{}{}
		pg.wg.Wait()
		pg.done = nil
	}
}

func (pg *PacketLog) run() {
	defer pg.wg.Done()
	count := 0
	lastCount := 0
	input := pg.input
	ticker := time.NewTicker(pg.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("Total %d, rate %.2f\n", count, rate(count-lastCount, pg.reportInterval))
			lastCount = count
		case _, ok := <-input:
			if !ok {
				return
			}
			count++
		case <-pg.done:
			return
		}
	}
}

func rate(count int, interval time.Duration) float64 {
	return float64(count) / (float64(interval) / float64(time.Second))
}
