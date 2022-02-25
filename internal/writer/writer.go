package writer

import (
	"context"
	//"fmt"
	"net"
	"sync"
	"time"

	"github.com/tjmerritt/udp-replicate/internal/packet"

	log "github.com/sirupsen/logrus"
)

type Net interface {
	DialUDP(network string, laddr, raddr *net.UDPAddr) (Conn, error)
}

type Conn interface {
	Close() error
	SetWriteDeadline(time.Time) error
	Write([]byte) (int, error)
	WriteTo([]byte, net.Addr) (int, error)
}

type UDPWriter struct {
	ctx     context.Context
	addr    *net.UDPAddr
	timeout time.Duration
	conn    Conn
	input   <-chan *packet.Packet
	wg      sync.WaitGroup
	done    context.Context
	cancel  context.CancelFunc
}

func New(ctx context.Context, netf Net, timeout time.Duration, addr *net.UDPAddr) (*UDPWriter, error) {
	conn, err := netf.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return &UDPWriter{
		ctx:     ctx,
		addr:    addr,
		timeout: timeout,
		conn:    conn,
	}, nil
}

func (uw *UDPWriter) SetInput(input <-chan *packet.Packet) {
	uw.input = input
}

func (uw *UDPWriter) Start() {
	if uw.done == nil {
		uw.done, uw.cancel = context.WithCancel(uw.ctx)
		uw.wg.Add(1)
		go uw.run()
	}
}

func (uw *UDPWriter) Stop() {
	if uw.done != nil {
		uw.cancel()
		uw.wg.Wait()
		uw.done = nil
	}
}

func (uw *UDPWriter) Close() error {
	uw.Stop()
	return uw.conn.Close()
}

func (uw *UDPWriter) run() {
	defer uw.wg.Done()
	input := uw.input
	for {
		select {
		case pkt, ok := <-input:
			if !ok {
				input = nil
				continue
			}
			//fmt.Printf("Sending %s len %d\n", pkt, len(pkt.Buf()))

			uw.conn.SetWriteDeadline(time.Now().Add(uw.timeout))
			_, err := uw.conn.Write(pkt.Buf())
			//			_, err := uw.conn.WriteTo(pkt.Buf(), uw.addr)
			if err != nil {
				log.WithError(err).Error("Error writing to UDP socket")
			}
			pkt.Put()
		case <-uw.done.Done():
			return
		}
	}
}
