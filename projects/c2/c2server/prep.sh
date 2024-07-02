#!/bin/bash
bin="c2server.bin"
exe="c2server.exe"
# If the src directory does not exist create and copy the files, create the go.mod file
echo "Create commonFunctions directories"
if [ ! -d "src/commonFunctions" ]; then
	mkdir -p "src/commonFunctions"
	cp -r ../../commonFunctions/*.go src/commonFunctions/.
fi
go mod init commonFunctions
mv -f go.mod src/commonFunctions/.

go env -w GOPATH="`pwd`" 
go env -w GO111MODULE='auto'

# Install dependencies for c2server
go get github.com/mattn/go-sqlite3

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $bin -ldflags "-w -s" main.go
GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go
