#!/bin/bash

export RECV_TIMEOUT=1s
export INPUT_ADDR=127.0.0.1:5001
export REPORT_INTERVAL=10s
./read-test
