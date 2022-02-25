package main

import (
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	OutputAddr     string        `env:"OUTPUT_ADDR"`
	SendTimeout    time.Duration `env:"SEND_TIMEOUT"`
	ChannelSize    int           `env:"CHANNEL_SIZE"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"1m"`
	Rate           float64       `env:"RATE"`
	SinkAddr       *net.UDPAddr
}

func GetConfig() *Config {
	cfg := &Config{}
	var err error
	if err = env.Parse(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	if cfg.SinkAddr, err = net.ResolveUDPAddr("udp", cfg.OutputAddr); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	return cfg
}
