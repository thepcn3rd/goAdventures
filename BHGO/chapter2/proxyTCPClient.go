package main

/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter2"

// To cross compile for linux
GOOS=linux GOARCH=amd64 go build -o proxyTCP.bin -ldflags "-w -s" proxyTCPClient.go

// To work with parrot linux docker
// Manjaro uses glibc 2.32 2.34, the parrot linux docker uses 2.31 due to the difference CGO_ENABLED=0 will remove the dependency...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o proxyTCP.bin proxyTCPClient.go

// To cross compile windows
GOOS=windows GOARCH=amd64 go build -o proxyTCP.exe -ldflags "-w -s" proxyTCPClient.go


Future State:
- Create flags for the options
- Create it for UDP also from the flags, if possible...


*/

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handleConnection(src net.Conn, dIP string, dPort string) {
	// Output where connections are received from to stdout
	fmt.Printf("Received connection from: %s\n", src.RemoteAddr().String())
	// Destination of where the connection is proxied...
	dstString := dIP + ":" + dPort
	dst, err := net.Dial("tcp", dstString)
	if err != nil {
		//log.Fatalln("Unable to connect to our unreachable host")
		fmt.Printf("Unable to connect to %s\n%s\n", dstString, err)
		os.Exit(1)
	}
	defer dst.Close()

	go func() {
		// Observed with an RDP connection that if the proxied connection failed it would
		// stop the proxy with an os.Exit(1) using log.Fatal
		if _, err := io.Copy(dst, src); err != nil {
			fmt.Printf("Proxied connection for %s closed by destination %s\n", src.RemoteAddr().String(), dstString)
			src.Close()
			dst.Close()
			//log.Fatalln(err)
		}
	}()

	if _, err := io.Copy(src, dst); err != nil {
		fmt.Printf("Destination connection is not responding - %s\n", dstString)
		src.Close()
		dst.Close()
		//log.Fatalln(err)
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	// Source listening port
	listeningPortPtr := flag.String("port", "8888", "Listening Port for the Proxied TCP Connection")
	// Destination proxy IP
	dstIPPtr := flag.String("dstip", "", "Destination IP Address to Proxy")
	dstPortPtr := flag.String("dstport", "", "Destination TCP Port to Proxy")
	flag.Parse()
	if !isFlagPassed("dstip") && !isFlagPassed("dstport") {
		fmt.Println(colorGreen + "\nError: Proxied IP and TCP Port need to be specified" + colorReset)
		flag.Usage()
		fmt.Println("") // Inserted empty line for cleaner output
		os.Exit(0)
	}

	// Listening port in the format of <listeningIPAddress>:<listeningPort>
	listeningPort := ":" + *listeningPortPtr

	// Destination proxy Port
	// Listening port where clients can connect to pull through the proxy...
	/*
		Reference: https://opensource.com/article/18/5/building-concurrent-tcp-server-go
		The first parameter of the net.Listen() function defines the type of network that will be used, while the second parameter defines the server address as well as the port number the server will listen to. Valid values for the first parameter are tcp, tcp4 (IPv4-only), tcp6 (IPv6-only), udp, udp4 (IPv4- only), udp6 (IPv6-only), ip, ip4 (IPv4-only), ip6 (IPv6-only), Unix (Unix sockets), Unixgram, and Unixpacket.
	*/
	//listener, err := net.Listen("tcp", listeningPort)
	// Only listen on IPv4 IP Addresses
	listener, err := net.Listen("tcp4", listeningPort)
	if err != nil {
		fmt.Printf("Unable to bind to port %s", listeningPort)
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			// The below terminates the connection if an error occurrs
			///log.Fatalln("Unable to accept connection")
			// Modified to output the error but return to the for loop and continue to accept connections
			fmt.Printf("Unable to accept the connection from %s\n%s\n", conn.RemoteAddr().String(), err)
		}
		go handleConnection(conn, *dstIPPtr, *dstPortPtr)
	}
}
