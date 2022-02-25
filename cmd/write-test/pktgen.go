package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tjmerritt/udp-replicate/internal/packet"
)

type PacketGen struct {
	reportInterval time.Duration
	rate           float64
	maxLen         int
	pool           *packet.Pool
	output         chan<- *packet.Packet
	wg             sync.WaitGroup
	done           chan struct{}
}

func NewPacketGen(report time.Duration, rate float64, maxLen int) *PacketGen {
	return &PacketGen{
		reportInterval: report,
		rate:           rate,
		maxLen:         maxLen,
		pool:           packet.NewPool(maxLen),
	}
}

func (pg *PacketGen) SetOutput(output chan<- *packet.Packet) {
	pg.output = output
}

func (pg *PacketGen) Start() {
	if pg.done == nil {
		pg.done = make(chan struct{})
		pg.wg.Add(1)
		go pg.run()
	}

}

func (pg *PacketGen) Stop() {
	if pg.done != nil {
		pg.done <- struct{}{}
		pg.wg.Wait()
		pg.done = nil
	}
}

func (pg *PacketGen) run() {
	defer pg.wg.Done()
	num := 1
	count := 0
	lastCount := 0
	dropped := 0
	lastDropped := 0
	ticker := time.NewTicker(genInterval(pg.rate))
	defer ticker.Stop()
	output := pg.output
	var pkt *packet.Packet
	var pkt2 *packet.Packet
	reportTicker := time.NewTicker(pg.reportInterval)
	defer reportTicker.Stop()

	for {
		out := output
		out2 := output
		if pkt == nil {
			out = nil
		}
		if pkt2 == nil {
			out2 = nil
		}

		select {
		case <-reportTicker.C:
			fmt.Printf("Count %d rate %.2f, Dropped %d rate %.2f\n",
				count, rate(count-lastCount, pg.reportInterval),
				dropped, rate(dropped-lastDropped, pg.reportInterval))
			lastCount = count
			lastDropped = dropped
		case <-ticker.C:
			if pkt != nil {
				if pkt2 != nil {
					fmt.Printf("Dropped '%s'\n", pkt2)
					dropped++
				}
				pkt2 = pkt
				pkt = nil
			}
			pkt = pg.genPacket(num)
			num++
		case out <- pkt:
			//fmt.Printf("Sent pkt '%s'\n", pkt)
			pkt = nil
			count++
		case out2 <- pkt2:
			//fmt.Printf("Sent pkt2 '%s'\n", pkt2)
			pkt2 = nil
			count++
		case <-pg.done:
			ticker.Stop()
			return
		}
	}
}

func (pg *PacketGen) genPacket(num int) *packet.Packet {
	pkt := pg.pool.Get()
	s := fmt.Sprintf("Packet #%d\n", num)
	copy(pkt.Buf(), []byte(s))
	pkt.Resize(len(s))
	return pkt
}

func genInterval(rate float64) time.Duration {
	if rate == 0.0 {
		rate = 1.0
	}

	interval := 1.0 / rate
	dur := time.Duration(int(interval*1000000)) * time.Microsecond
	return dur
}

func rate(count int, interval time.Duration) float64 {
	return float64(count) / (float64(interval) / float64(time.Second))
}
