package packet

import (
	//        "fmt"
	"sync"
)

type Pool struct {
	pool   sync.Pool
	maxLen int
}

func NewPool(maxLen int) *Pool {
	pool := &Pool{
		maxLen: maxLen,
		pool:   sync.Pool{},
	}
	pool.pool.New = func() interface{} { return pool.newPacket() }
	return pool
}

func (p *Pool) newPacket() *Packet {
	return New(p, make([]byte, p.maxLen))
}

func (p *Pool) Get() *Packet {
	pkt := p.pool.Get().(*Packet)
	pkt.size = int32(len(pkt.buf))
	return pkt.Get()

}

func (p *Pool) put(pkt *Packet) {
	p.pool.Put(pkt)
}
