package main

// Portscanner built from "Black Hat Go"

// Added the input of the ip address to scan...
// Added the input of the ports to scan
// Had to figure out how to allow the connString when compiled for windows
// Seems to work fine...  Does not execute from a powershell prompt...

// To run the program execute:
// go run portScanner.go

// To build the program:
// go build portScanner.go -o output.bin

// To build the program without debugging information:
// go build -o output.bin -ldflags "-w -s" portScanner.go

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o yaPortScanner.bin -ldflags "-w -s" yaPortScanner.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o yaPortScanner.exe -ldflags "-w -s" yaPortScanner.go

// Additional training...  https://tour.golang.org

import (
	"bufio"
	"fmt"
	"math/big"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func prepareString(osDetected string, input string) string {
	var output string
	if osDetected == "windows" {
		output = strings.Replace(input, "\r\n", "", -1)
	} else {
		output = strings.Replace(input, "\n", "", -1)
	}
	return output
}

func IP4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}

func InttoIP4(ipInt int64) string {
	// Need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

func buildConnString(osDetected string) (int64, int64, int, int) {
	//var sliceConnString []string
	var startingIP string
	var lastIP string
	var startingPort string
	var lastPort string
	invalidOption := "True"
	for invalidOption == "True" {
		fmt.Print("Do you want to scan a range of IP Addresses? (y/n): ")
		inputReader := bufio.NewReader(os.Stdin)
		inputSelection, _ := inputReader.ReadString('\n')
		inputSelection = prepareString(osDetected, inputSelection)
		if inputSelection == "y" || inputSelection == "Y" {
			fmt.Print("Starting IP Address: ")
			startingIP, _ = inputReader.ReadString('\n')
			startingIP = prepareString(osDetected, startingIP)
			fmt.Print("Last IP Address: ")
			lastIP, _ = inputReader.ReadString('\n')
			lastIP = prepareString(osDetected, lastIP)
			invalidOption = "False"
		} else if inputSelection == "n" || inputSelection == "N" {
			fmt.Print("IP Address to Scan: ")
			startingIP, _ = inputReader.ReadString('\n')
			startingIP = prepareString(osDetected, startingIP)
			lastIP = startingIP
			invalidOption = "False"
		} else {
			fmt.Print("Invalid Option!!\n")
		}
	}
	// Calculate number of IP Addresses need to be scanned...
	startIPDecimal := IP4toInt(net.ParseIP(startingIP))
	lastIPDecimal := IP4toInt(net.ParseIP(lastIP))
	totalIPAddresses := (lastIPDecimal - startIPDecimal) + 1
	if totalIPAddresses > 255 {
		fmt.Printf("Warning: Total Addresses being scanned is more than 255!! (%d)\n", totalIPAddresses)
	}

	// Calculate the range of ports to be scanned
	invalidOption = "True"
	for invalidOption == "True" {
		// Could introduce a comma seperated list of IP Addresses
		fmt.Print("\nDo you want to scan a range of Ports? (y/n): ")
		inputReader := bufio.NewReader(os.Stdin)
		inputSelection, _ := inputReader.ReadString('\n')
		inputSelection = prepareString(osDetected, inputSelection)
		if inputSelection == "y" || inputSelection == "Y" {
			fmt.Print("Starting Port: ")
			startingPort, _ = inputReader.ReadString('\n')
			startingPort = prepareString(osDetected, startingPort)
			fmt.Print("Last Port: ")
			lastPort, _ = inputReader.ReadString('\n')
			lastPort = prepareString(osDetected, lastPort)
			invalidOption = "False"
		} else if inputSelection == "n" || inputSelection == "N" {
			fmt.Print("Port to Scan: ")
			startingPort, _ = inputReader.ReadString('\n')
			startingPort = prepareString(osDetected, startingPort)
			lastPort = startingPort
			invalidOption = "False"
		} else {
			fmt.Print("Invalid Option!!\n")
		}
	}
	// Calculate number of Ports to be scanned...
	intLastPort, _ := strconv.Atoi(lastPort)
	intStartingPort, _ := strconv.Atoi(startingPort)
	totalPorts := (intLastPort - intStartingPort) + 1
	if totalPorts > 65534 {
		fmt.Printf("Warning: Total Ports being scanned is more than 65534!! (%d)\n", totalPorts)
	} else if totalPorts > 1000 {
		fmt.Printf("Warning: Total Ports being scanned is more than 1000!! (%d)\n", totalPorts)
	}

	// Create the connection strings
	//var currentIPAddress string = "127.0.0.1"
	//var currentPort string = "0"
	//var currentConnString string = "127.0.0.1:0"
	//for numIP := startIPDecimal; numIP <= lastIPDecimal; numIP++ {
	//	for numPort := intStartingPort; numPort <= intLastPort; numPort++ {
	// 	Convert Decimal IP to IPv4 format
	//		currentIPAddress = InttoIP4(numIP)
	//		currentPort = strconv.Itoa(numPort)
	//		currentConnString = currentIPAddress + ":" + currentPort
	//		sliceConnString = append(sliceConnString, currentConnString)
	//fmt.Printf("%s\n", currentConnString)
	//	}
	//}
	return startIPDecimal, lastIPDecimal, intStartingPort, intLastPort
}

func worker(ports, results chan int, ip string) {

	for p := range ports {
		port := strconv.Itoa(p)
		connString := ip + ":" + port
		// Uncomment the below line to show the output of the connection strings being created
		//fmt.Println(connString)
		conn, err := net.Dial("tcp", connString)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

func main() {
	//ip := "127.0.0.1"
	osDetected := runtime.GOOS
	//fmt.Println("Operating System: ", osDetected)
	fmt.Print("<--- Beans Port Scanner --->\n")
	// Build the connection string slice that is used for the scanning...
	startIPDecimal, lastIPDecimal, intStartingPort, intLastPort := buildConnString(osDetected)
	//fmt.Println(sliceConnString)

	var portsChan chan int
	var resultsChan chan int
	var openports []string

	for intIPAddr := startIPDecimal; intIPAddr <= lastIPDecimal; intIPAddr++ {
		//for _, sConnString := range sliceConnString {

		if osDetected == "windows" && ((intLastPort-intStartingPort)+1)*((int(lastIPDecimal)-int(startIPDecimal))+1) > 100 {
			// The scanner running all 65k ports is really slow
			// Increased the channels to 1000
			// As stated in the book, efficiency may decrease with this setting...
			portsChan = make(chan int, 1000)
		} else {
			portsChan = make(chan int, 100)
		}
		resultsChan = make(chan int)

		//currentInfo := strings.Split(sConnString, ":")
		currentIPAddress := InttoIP4(intIPAddr)
		//currentIPAddress := currentInfo[0]
		fmt.Printf("Scanning %s\n\n", currentIPAddress)
		for i := 0; i < cap(portsChan); i++ {
			go worker(portsChan, resultsChan, currentIPAddress)
		}

		go func() {
			for i := intStartingPort; i <= intLastPort; i++ {
				portsChan <- i
			}
		}()

		for i := intStartingPort; i <= intLastPort; i++ {
			port := <-resultsChan
			if port != 0 {
				currentPort := strconv.Itoa(port)
				connString := currentIPAddress + ":" + currentPort
				openports = append(openports, connString)
			}
		}
		close(portsChan)
		close(resultsChan)
	}

	sort.Strings(openports)
	for _, port := range openports {
		if osDetected == "linux" {
			fmt.Printf("%s open\n", port)
		} else {
			fmt.Printf("%s open\r\n", port)
		}
	}

}
