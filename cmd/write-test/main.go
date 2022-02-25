package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tjmerritt/udp-replicate/internal/packet"
	"github.com/tjmerritt/udp-replicate/internal/writer"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := GetConfig()
	if cfg == nil {
		os.Exit(2)
	}

	maxLen := 4096
	netw := &NetWriter{}
	writer, err := writer.New(context.Background(), netw, cfg.SendTimeout, cfg.SinkAddr)
	if err != nil {
		log.WithError(err).Error("Error starting UDP writer")
		os.Exit(1)
	}
	pktgen := NewPacketGen(cfg.ReportInterval, cfg.Rate, maxLen)
	fmt.Printf("Channel size: %d\n", cfg.ChannelSize)
	pktChan := make(chan *packet.Packet, cfg.ChannelSize)
	writer.SetInput(pktChan)
	//go func() {
	//	for range pktChan {
	//	}
	//}()
	pktgen.SetOutput(pktChan)

	writer.Start()
	pktgen.Start()

	waitForSignal()
	fmt.Printf("Shutting down\n")
	pktgen.Stop()
	writer.Stop()
	os.Exit(0)
}

func waitForSignal() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
}

type NetWriter struct {
}

func (nw *NetWriter) DialUDP(proto string, laddr, raddr *net.UDPAddr) (writer.Conn, error) {
	conn, err := net.DialUDP(proto, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return writer.Conn(conn), nil
}
