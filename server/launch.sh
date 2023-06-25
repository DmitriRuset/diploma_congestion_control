#! /bin/bash

# Usage: ./launch.sh first/second

export GOMAXPROCS=1
./remove_logs.sh
go build $1.go
./$1
