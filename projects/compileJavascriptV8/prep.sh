#!/bin/bash
projectName="v8"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

go env -w GOPATH="/home/thepcn3rd/go/workspaces"
go env -w GO111MODULE='auto'

# Install dependencies
go get github.com/thepcn3rd/goAdventures/projects/commonFunctions
#go get golang.org/x/crypto/ssh
#go get github.com/rogchap/v8go
go get rogchap.com/v8go

GOOS=linux GOARCH=amd64 go build -o $bin -ldflags "-w -s" main.go
#GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go

