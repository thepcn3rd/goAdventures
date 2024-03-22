package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Initial program came from Black Hat Go
// Added flags to be able to specify the domain, dns server and port to utilize
// Added the capability to display the answer if it contains TXT records
// Extended the capability to conduct a reverse dns lookup with a given IP Address
// Added the capability if it is a range of IP Addresses that it will parse the CIDR
// Added flags for the IP address or the CIDR specification
// Added a throttle into the DNS Queries if a CIDR is provided of 2 seconds, this is so a DNS server does not blacklist the queries

/* Future Items
1.
2.

*/

// Setup the environment variables for go
/*
go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter5/dnsClient"

In the dnsClient directory create the following directorys, src, bin, pkg
*/

// Install the dns dependency
// go get github.com/miekg/dns

// Install the package for tcpdump to verify the dns queries
// sudo pacman -S tcpdump

// Verify using tcpdump the DNS queries
// sudo tcpdump -i wlp2s0 -n udp port 53

// Run the code
/*

go run main.go


DNS Message Structure

type Msg struct {
	MsgHdr
	Compress	bool	`json:"-"`
	Question	[]Question
	Answer		[]RR
	Ns			[]RR
	Extra		[]RR
}

*/

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func ipLookup(ipAddr string, dnsServer string, dnsServerPort string) {
	// Testing with ksl.com or 64.147.131.201 works...
	var msg dns.Msg

	ipArpa, err := dns.ReverseAddr(ipAddr)
	checkError("Unable to place IP Address in ARPA Format", err)
	//fmt.Println(ipArpa)

	msg.SetQuestion(ipArpa, dns.TypePTR)
	dnsServerInfo := dnsServer + ":" + dnsServerPort
	dnsStruct, err := dns.Exchange(&msg, dnsServerInfo)
	checkError("Unable to communicate to dns server", err)
	if len(dnsStruct.Answer) < 1 {
		fmt.Printf("%s : No records found for the IP Address\n", ipAddr)
		return
	}
	for _, answer := range dnsStruct.Answer {
		if ptr, ok := answer.(*dns.PTR); ok {
			fmt.Printf("%s : %s\n", ipAddr, ptr.String()) // Outputs the IP address of the response...
		}
	}
}

func domainLookup(domain string, dnsServer string, dnsServerPort string) {
	var msg dns.Msg
	fqdn := dns.Fqdn(domain)

	msg.SetQuestion(fqdn, dns.TypeA) // dns.TypeCNAME is also available
	dnsServerInfo := dnsServer + ":" + dnsServerPort
	dnsStruct, err := dns.Exchange(&msg, dnsServerInfo)
	checkError("Unable to communicate to dns server", err)
	if len(dnsStruct.Answer) < 1 {
		fmt.Printf("%s : No records found from the domain\n", domain)
		return
	}
	for _, answer := range dnsStruct.Answer {
		if a, ok := answer.(*dns.A); ok {
			fmt.Printf("%s : %s\n", domain, a.String()) // Outputs the IP address of the response...
		}
		if txt, ok := answer.(*dns.TXT); ok {
			fmt.Println(txt.String()) // Outputs the TXT record of the response...
		}
	}
}

// Reference: https://stackoverflow.com/questions/67788289/iterating-over-a-multiline-variable-to-return-all-ips
func ipGenerateList(out chan string, ipCIDR string) {
	ip, ipnet, err := net.ParseCIDR(ipCIDR)
	if err != nil {
		log.Fatal(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		out <- ip.String()
	}
	close(out)
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func main() {
	domainPtr := flag.String("d", "", "Specify FQDN to Query")
	ipPTR := flag.String("i", "", "Specify IP Address or IP in CIDR notation to conduct reverse DNS Lookup")
	dnsSrvPortPtr := flag.String("p", "53", "Specify DNS Server Port to Query")
	dnsSrvPtr := flag.String("s", "8.8.8.8", "Specify DNS Server to Query (UDP)")
	flag.Parse()

	if *domainPtr == "" && *ipPTR == "" {
		fmt.Println("\nPlease specify a domain to query or IP Address to conduct a reverse DNS Lookup")
		fmt.Println("./dnsClient.bin -d google.com")
		fmt.Println("./dnsClient.bin -i 192.168.0.5")
		fmt.Println("./dnsClient.bin -i 192.168.0.0/24")
		os.Exit(1)
	}

	if *domainPtr != "" {
		domainLookup(*domainPtr, *dnsSrvPtr, *dnsSrvPortPtr)
	} else if *ipPTR != "" {
		// What if the IP Address is a CIDR block?
		if strings.Contains(*ipPTR, "/") {
			ipAddressesChan := make(chan string)
			go ipGenerateList(ipAddressesChan, *ipPTR)
			for ip := range ipAddressesChan {
				ipLookup(ip, *dnsSrvPtr, *dnsSrvPortPtr)
				// Rate limiting the CIDRs for 2 seconds between queries
				time.Sleep(2 * time.Second)
			}
		} else {
			ipLookup(*ipPTR, *dnsSrvPtr, *dnsSrvPortPtr)
		}
	}

}
