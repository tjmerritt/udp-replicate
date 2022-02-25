package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tjmerritt/udp-replicate/internal/reader"
	"github.com/tjmerritt/udp-replicate/internal/replicate"
	"github.com/tjmerritt/udp-replicate/internal/writer"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := GetConfig()
	if cfg == nil {
		os.Exit(2)
	}

	maxLen := 4096
	netr := &NetReader{}
	netw := &NetWriter{}
	replicator := replicate.New(context.Background(), netr, netw, cfg.SendTimeout, cfg.RecvTimeout, cfg.SrcAddr, cfg.SinkAddrs, cfg.ChannelSize, maxLen)
	err := replicator.Start()
	if err != nil {
		log.WithError(err).Error("Error starting UDP replicator")
		os.Exit(1)
	}

	waitForSignal()
	fmt.Printf("Shutting down\n")
	replicator.Stop()
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

type NetWriter struct {
}

func (nw *NetWriter) DialUDP(proto string, laddr, raddr *net.UDPAddr) (writer.Conn, error) {
	conn, err := net.DialUDP(proto, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return writer.Conn(conn), nil
}
