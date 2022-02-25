package replicate

import (
	"context"
	//"fmt"
	"net"
	"sync"
	"time"

	"github.com/tjmerritt/udp-replicate/internal/packet"
	"github.com/tjmerritt/udp-replicate/internal/reader"
	"github.com/tjmerritt/udp-replicate/internal/writer"
)

type UDPReplicator struct {
	ctx         context.Context
	netr        reader.Net
	netw        writer.Net
	recvto      time.Duration
	sendto      time.Duration
	srcAddr     *net.UDPAddr
	sinkAddrs   []*net.UDPAddr
	maxLen      int
	channelSize int
	input       chan *packet.Packet
	outputs     []chan *packet.Packet
	src         *reader.UDPReader
	sinks       []*writer.UDPWriter
	wg          sync.WaitGroup
	done        context.Context
	cancel      context.CancelFunc
}

func New(ctx context.Context, netr reader.Net, netw writer.Net, recvto, sendto time.Duration, inputAddr *net.UDPAddr, outputAddrs []*net.UDPAddr, channelSize, maxLen int) *UDPReplicator {
	return &UDPReplicator{
		ctx:         ctx,
		netr:        netr,
		netw:        netw,
		recvto:      recvto,
		sendto:      sendto,
		srcAddr:     inputAddr,
		sinkAddrs:   outputAddrs,
		channelSize: channelSize,
		maxLen:      maxLen,
	}
}

func (ur *UDPReplicator) Start() error {
	if ur.done == nil {
		var err error
		done, cancel := context.WithCancel(ur.ctx)
		ur.src, err = reader.New(done, ur.netr, ur.recvto, ur.srcAddr, ur.maxLen)
		ur.sinks = make([]*writer.UDPWriter, len(ur.sinkAddrs))
		for i, a := range ur.sinkAddrs {
			var err2 error
			ur.sinks[i], err2 = writer.New(done, ur.netw, ur.sendto, a)
			if err == nil {
				err = err2
			}
		}
		if err != nil {
			ur.close()
			return err
		}
		ur.done = done
		ur.cancel = cancel
		ur.input = make(chan *packet.Packet, ur.channelSize)
		ur.src.SetOutput(ur.input)
		ur.outputs = make([]chan *packet.Packet, len(ur.sinks))
		for i, uw := range ur.sinks {
			out := make(chan *packet.Packet, 1000)
			ur.outputs[i] = out
			uw.SetInput(out)
		}
		ur.src.Start()
		for _, uw := range ur.sinks {
			uw.Start()
		}
		ur.wg.Add(1)
		go ur.run()
	}
	return nil
}

func (ur *UDPReplicator) close() {
	if ur.src != nil {
		ur.src.Close()
	}
	for _, uw := range ur.sinks {
		if uw != nil {
			uw.Close()
		}
	}
	ur.src = nil
	ur.sinks = nil
}

func (ur *UDPReplicator) Stop() {
	if ur.done != nil {
		ur.src.Stop()
		close(ur.input)
		ur.wg.Wait()
		ur.cancel()
		ur.done = nil
	}
}

func (ur *UDPReplicator) send(idx int, pkt, pkt2 *packet.Packet) (chan<- *packet.Packet, *packet.Packet, int) {
	if idx >= 0 && pkt != nil {
		for idx < len(ur.sinks) {
			sink := ur.sinks[idx]
			if sink != nil {
				if pkt2 == nil {
					pkt2 = pkt.Get()
				}
				//fmt.Printf("Attempting addr #%d\n", idx)
				return ur.outputs[idx], pkt2, idx
			}
			idx++
		}
	}
	return nil, nil, -1
}

func (ur *UDPReplicator) run() {
	defer ur.wg.Done()
	idx := -1
	input := ur.input
	var pkt *packet.Packet
	var pkt2 *packet.Packet
	for input != nil || idx >= 0 {
		in := input
		var out chan<- *packet.Packet
		out, pkt2, idx = ur.send(idx, pkt, pkt2)
		if out != nil {
			in = nil
		}
		select {
		case p, ok := <-in:
			if !ok {
				input = nil
			} else {
				//fmt.Printf("Got %+v len %d\n", p.Buf(), len(p.Buf()))
				if pkt2 != nil {
					pkt2.Put()
					pkt2 = nil
				}
				pkt = p
				idx = 0
			}
		case out <- pkt2:
			//fmt.Printf("Sent %+v len %d\n", pkt2.Buf(), len(pkt2.Buf()))
			pkt2 = nil
			idx++
		case <-ur.done.Done():
			if pkt != nil {
				pkt.Put()
			}
			if pkt2 != nil {
				pkt2.Put()
			}
			return
		}
	}
}
