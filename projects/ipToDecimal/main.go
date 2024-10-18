package main

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ipToDecimal.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o ipToDecimal.exe -ldflags "-w -s" main.go

import (
	"flag"
	"fmt"
	"math/big"
	"net"
	"runtime"
	"strconv"
	"strings"
)

func prepareString(input string) string {
	osDetected := runtime.GOOS
	var output string
	if osDetected == "windows" {
		output = strings.Replace(input, "\r\n", "", -1)
	} else {
		output = strings.Replace(input, "\n", "", -1)
	}
	return output
}

func IPToDecimal(ipAddress string) *big.Int {
	// Parse the IP address string
	ipAddress = prepareString(ipAddress)
	parsedIP := net.ParseIP(ipAddress)
	if parsedIP == nil {
		return nil
	}

	// Convert the IP address bytes to a big.Int
	ipDecimal := new(big.Int)
	ipDecimal.SetBytes(parsedIP.To4())

	return ipDecimal
}

func InttoIP4(ipInt int64) string {
	// Need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

func main() {
	ipPtr := flag.String("ip", "", "Convert an IP address to to a decimal number")
	decPtr := flag.String("d", "", "Convert a Decimal Number to IP Address")
	flag.Parse()

	if len(*ipPtr) > 0 {
		ipAddress := *ipPtr
		ipDecimalValue := IPToDecimal(ipAddress)

		if ipDecimalValue == nil {
			fmt.Printf("Invalid IP address: %s\n", ipAddress)
		} else {
			fmt.Printf("IP Address: %s\n", ipAddress)
			fmt.Printf("Decimal Representation: %s\n", ipDecimalValue.String())
		}
	} else if len(*decPtr) > 0 {
		ipDecimalValue := *decPtr
		intIPDecimalValue, _ := strconv.Atoi(ipDecimalValue)

		ipAddress := InttoIP4(int64(intIPDecimalValue))
		fmt.Printf("Decimal Representation: %s\n", ipDecimalValue)
		fmt.Printf("IP Address: %s\n", ipAddress)
	} else {
		fmt.Printf("Specify the conversion of an IP Address or a Decimal Value")
		flag.Usage()
	}
}
