package main

// Test on windows due to changes...

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o gocat.bin -ldflags "-w -s" gocat.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o gocat.exe -ldflags "-w -s" gocat.go

import (
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"
)

func handle(conn net.Conn, osDetected string) {
	var cmdProg string
	var cmdArg string
	//fmt.Println(osDetected)
	if osDetected == "linux" {
		cmdProg = "/bin/sh"
		cmdArg = "-i"
	} else {
		cmdProg = "cmd.exe"
		cmdArg = "/c"
	}
	//fmt.Println(cmdProg + cmdArg)
	cmd := exec.Command(cmdProg, cmdArg)
	//if osDetected == "linux" {
	//cmd := exec.Command("/bin/bash", "-i")
	//}
	rp, wp := io.Pipe()
	// Set stdin to our connection
	cmd.Stdin = conn
	cmd.Stdout = wp

	go io.Copy(conn, rp)
	cmd.Run()
	conn.Close()
}

func main() {
	osDetected := runtime.GOOS
	var ipAddr string
	fmt.Printf("\nListen on IP Address: ")
	fmt.Scanln(&ipAddr)

	var dstPort string
	fmt.Printf("Listen on Port: ")
	fmt.Scanln(&dstPort)

	connString := ipAddr + ":" + dstPort
	listener, err := net.Listen("tcp", connString)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handle(conn, osDetected)
	}
}
