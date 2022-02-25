package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tjmerritt/udp-replicate/internal/packet"
	"github.com/tjmerritt/udp-replicate/internal/reader"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := GetConfig()
	if cfg == nil {
		os.Exit(2)
	}

	maxLen := 4096
	netr := &NetReader{}
	reader1, err := reader.New(context.Background(), netr, cfg.RecvTimeout, cfg.SrcAddr, maxLen)
	if err != nil {
		log.WithError(err).Error("Error starting UDP reader1")
		os.Exit(1)
	}
	pktLog := NewPacketLog(cfg.ReportInterval)
	pktChan := make(chan *packet.Packet)
	reader1.SetOutput(pktChan)
	pktLog.SetInput(pktChan)

	reader1.Start()
	pktLog.Start()

	waitForSignal()
	fmt.Printf("Shutting down\n")
	reader1.Stop()
	pktLog.Stop()
	os.Exit(0)
}

func waitForSignal() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
}

type NetReader struct {
}

func (nr *NetReader) ListenUDP(proto string, addr *net.UDPAddr) (reader.Conn, error) {
	conn, err := net.ListenUDP(proto, addr)
	if err != nil {
		return nil, err
	}
	return reader.Conn(conn), nil
}
