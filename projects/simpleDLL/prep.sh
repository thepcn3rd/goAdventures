#!/bin/bash
projectName="simpleDLL"
dll="$projectName.dll"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

#go env -w GOPATH="/home/thepcn3rd/go/workspaces"
#go env -w GO111MODULE='auto'

# Install dependencies
#go get github.com/thepcn3rd/goAdventures/projects/commonFunctions

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags="-w -s -H=windowsgui" -o $dll

