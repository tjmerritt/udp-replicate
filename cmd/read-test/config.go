package main

import (
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	InputAddr      string        `env:"INPUT_ADDR"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"1m"`
	RecvTimeout    time.Duration `env:"RECV_TIMEOUT"`
	SrcAddr        *net.UDPAddr
}

func GetConfig() *Config {
	cfg := &Config{}
	var err error
	if err = env.Parse(cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	if cfg.SrcAddr, err = net.ResolveUDPAddr("udp", cfg.InputAddr); err != nil {
		fmt.Printf("%+v\n", err)
		return nil
	}

	return cfg
}
