#!/bin/bash

export SEND_TIMEOUT=10s
export RECV_TIMEOUT=10s
export INPUT_ADDR=127.0.0.1:5000
export OUTPUT_ADDRS=127.0.0.1:5001,127.0.0.1:5002
export CHANNEL_SIZE=100
./udp-replicate
