package reader

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/tjmerritt/udp-replicate/internal/packet"

	log "github.com/sirupsen/logrus"
)

type Net interface {
	ListenUDP(proto string, addr *net.UDPAddr) (Conn, error)
}

type Conn interface {
	SetReadBuffer(lenghth int) error
	Close() error
	SetReadDeadline(time.Time) error
	ReadFrom([]byte) (int, net.Addr, error)
}

type UDPReader struct {
	ctx     context.Context
	pool    *packet.Pool
	timeout time.Duration
	conn    Conn
	output  chan<- *packet.Packet
	wg      sync.WaitGroup
	done    context.Context
	cancel  context.CancelFunc
}

func New(ctx context.Context, netf Net, timeout time.Duration, addr *net.UDPAddr, maxLen int) (*UDPReader, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	conn, err := netf.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	if err := conn.SetReadBuffer(maxLen); err != nil {
		return nil, err
	}
	return &UDPReader{
		ctx:     ctx,
		pool:    packet.NewPool(maxLen),
		timeout: timeout,
		conn:    conn,
	}, nil
}

func (ur *UDPReader) SetOutput(output chan<- *packet.Packet) {
	ur.output = output
}

func (ur *UDPReader) Start() {
	if ur.done == nil {
		ur.done, ur.cancel = context.WithCancel(ur.ctx)
		ur.wg.Add(1)
		go ur.run()
	}
}

func (ur *UDPReader) Stop() {
	if ur.done != nil {
		ur.cancel()
		ur.wg.Wait()
		ur.done = nil
	}
}

func (ur *UDPReader) Close() error {
	ur.Stop()
	return ur.conn.Close()
}

func (ur *UDPReader) run() {
	defer ur.wg.Done()
	for {
		output := ur.output
		pkt := ur.pool.Get()
		ur.conn.SetReadDeadline(time.Now().Add(ur.timeout))
		n, _, err := ur.conn.ReadFrom(pkt.Buf())
		if err != nil {
			quiet := false
			if operr, ok := err.(*net.OpError); ok && operr.Timeout() {
				//fmt.Printf("Read timeout\n")
				quiet = true
			}
			if !quiet {
				log.WithError(err).Error("Error reading from UDP socket")
			}
			output = nil
		}
		pkt.Resize(n)
		//fmt.Printf("Read packet %+v len %d\n", pkt.Buf(), n)

		// Send the packet if we can now, so that it isn't a coin flip whether it gets sent before done happens.
		if output != nil {
			select {
			case output <- pkt:
				output = nil
			default:
			}
		}

		// Send the packet and check for a done condition
		if output != nil {
			// Block waiting to send the packet or the done condition to occur.
			select {
			case output <- pkt:
			case <-ur.done.Done():
				return
			}
		} else {
			// Nothing to send so just check done condition
			select {
			case <-ur.done.Done():
				return
			default:
			}
		}
	}
}
