#!/bin/bash

export SEND_TIMEOUT=1s
export RATE=25000
export OUTPUT_ADDR=127.0.0.1:5000
export CHANNEL_SIZE=1000
export REPORT_INTERVAL=10s
./write-test
