package main

import (
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	InputAddr   string        `env:"INPUT_ADDR"`
	OutputAddrs []string      `env:"OUTPUT_ADDRS"`
	SendTimeout time.Duration `env:"SEND_TIMEOUT"`
	RecvTimeout time.Duration `env:"RECV_TIMEOUT"`
	SrcAddr     *net.UDPAddr
	SinkAddrs   []*net.UDPAddr
	ChannelSize int `env:"CHANNEL_SIZE"`
}

func GetConfig() *Config {
	cfg := &Config{}
	var err error
	if err = env.Parse(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	if len(cfg.OutputAddrs) < 1 {
		fmt.Printf("no target addresses\n")
		return nil
	}

	if cfg.SrcAddr, err = net.ResolveUDPAddr("udp", cfg.InputAddr); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	cfg.SinkAddrs = make([]*net.UDPAddr, 0, len(cfg.OutputAddrs))

	for _, output := range cfg.OutputAddrs {
		sink, err := net.ResolveUDPAddr("udp", output)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return nil
		}
		cfg.SinkAddrs = append(cfg.SinkAddrs, sink)
	}

	return cfg
}
