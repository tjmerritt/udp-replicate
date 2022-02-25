package packet

import (
	//        "fmt"
	"math"
	"sync/atomic"
)

type Packet struct {
	pool *Pool
	buf  []byte
	refs int32
	size int32
}

func New(pool *Pool, buf []byte) *Packet {
	return &Packet{
		refs: 1,
		pool: pool,
		buf:  buf,
		size: int32(len(buf)),
	}
}

func (pkt *Packet) Buf() []byte {
	sz := atomic.LoadInt32(&pkt.size)
	return pkt.buf[:sz]
}

func (pkt *Packet) Get() *Packet {
	atomic.AddInt32(&pkt.refs, 1)
	return pkt
}

func (pkt *Packet) Put() {
	n := atomic.AddInt32(&pkt.refs, -1)
	switch {
	case n < 0:
		panic("Packet out of references")
	case n == 0 && pkt.pool != nil:
		pkt.pool.put(pkt)
	}
}

func (pkt *Packet) Resize(n int) {
	if n < 0 || n > math.MaxInt32 && n > cap(pkt.buf) {
		panic("Packet resize length out of range")
	}
	atomic.StoreInt32(&pkt.size, int32(n))
}

func (pkt *Packet) String() string {
	l := len(pkt.buf)
	sz := int(atomic.LoadInt32(&pkt.size))
	if l > sz {
		l = sz
	}

	// Remove trailing newline
	if l > 0 && pkt.buf[l-1] == '\n' {
		l--
	}

	return string(pkt.buf[:l])
}
