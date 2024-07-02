#!/bin/bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o fortinethp.bin -ldflags "-w -s" main.go
./fortinethp.bin -port 11000
