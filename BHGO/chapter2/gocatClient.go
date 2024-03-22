package main

// Tested on Linux but not on Windows...

// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o v2gocatClient.bin -ldflags "-w -s" v2gocatClient.go

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	var bytesRead int
	var output string
	var inputCommand string
	//osDetected := runtime.GOOS

	var ipAddr string
	fmt.Printf("\nConnect to IP Address: ")
	fmt.Scanln(&ipAddr)

	var dstPort string
	fmt.Printf("Connect to Port: ")
	fmt.Scanln(&dstPort)

	connString := ipAddr + ":" + dstPort
	fmt.Print("<--- gocat Client --->\n")
	fmt.Print("exit - Leave the gocat Client\n")
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		fmt.Print("Unable to connect to the host and port configured...\n")
		fmt.Println(err)
		return
	}

	defer conn.Close()

	for {
		if len(inputCommand) > 2 {
			fmt.Printf("\nPrevious Command: %s", inputCommand)
			fmt.Print("$ ")
		} else {
			fmt.Print("\n$ ")
		}

		inputReader := bufio.NewReader(os.Stdin)
		inputCommand, _ = inputReader.ReadString('\n')
		if inputCommand == "exit\n" {
			os.Exit(0)
		}
		_, err = conn.Write([]byte(inputCommand))
		if err != nil {
			fmt.Println(err)
			return
		}

		output = ""
		readNext := true
		for readNext == true {
			conn.SetReadDeadline(time.Now().Add(3 * time.Second))

			reply := make([]byte, 4096)
			bytesRead, err = conn.Read(reply)
			if err != nil {
				//fmt.Println(err)
				break
			}
			//fmt.Printf("%s", string(reply[:bytesRead]))
			output += string(reply[:bytesRead])
			//fmt.Println(bytesRead)
			if bytesRead != 47 && bytesRead != 4096 {
				readNext = false
				break
			} else if bytesRead == 47 || bytesRead == 4096 {
				readNext = true
			} else {
				readNext = false
				break
			}

		}
		fmt.Print(output)
	}
}
