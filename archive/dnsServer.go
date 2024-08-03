package main

import (
    "net"
	"fmt"
	"os"
    "flag"

    "github.com/miekg/dns"
)

// Setup the environment variables for go
/*
go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter5/dnsServer"

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

# To cross compile for linux
GOOS=linux GOARCH=amd64 go build -o dnsServer.bin -ldflags "-w -s" main.go

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

/* References

https://github.com/capjamesg/dns-experiments/blob/main/dns.go

*/

/* Steps to do:
1. Create log file of DNS queries, inclusive of A, AAAA, CNAME and TXT (and others...)
2. 

*/


func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func adsFunc(w dns.ResponseWriter, r *dns.Msg) {
	resp := new(dns.Msg)
	resp.SetReply(r)

	var lines_of_text []string
    // Max length of a TXT record in DNS is 255 characters...
	lines_of_text = append(lines_of_text, "Line1")
	lines_of_text = append(lines_of_text, "Line2")

	for _, line := range lines_of_text {
		resp.Answer = append(resp.Answer, &dns.TXT{Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}, Txt: []string{line}})
	}
    fmt.Println(resp.Question)
	w.WriteMsg(resp)
}


func main() {
    var listeningPort string
    listeningPortPtr := flag.String("port", "5353", "Default Listening Port")
    flag.Parse()
    listeningPort = *listeningPortPtr

    handler := dns.NewServeMux()

    // If this specific DNS name is queried then the function of adsFunc is called and executed
    // Command that currently triggerse the below: dig @localhost ads.attacker3.com -p5353
    // Command line pull DNS TXT records with nslookup
    // nslookup <enter>
    // set type=txt <enter>
    // ads.attacker3.com
    handler.HandleFunc("ads.attacker3.com", adsFunc)

    // If none of the above is triggered then this is the default behavior
    handler.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
        var resp dns.Msg
        resp.SetReply(req)
        for _, q := range req.Question {
            a := dns.A{
                Hdr: dns.RR_Header{
                    Name:   q.Name,
                    Rrtype: dns.TypeA,
                    Class:  dns.ClassINET,
                    Ttl:    0,
                },
                A: net.ParseIP("127.0.0.1").To4(),
            }
           resp.Answer = append(resp.Answer, &a)
           
        }
        fmt.Println(resp.Question)
        w.WriteMsg(&resp)
    })

    // Place a colon in front of the port to listen on all interfaces
    // Here is where you could specify the IP Address to bind to...
    listeningPort = ":" + listeningPort
    //server := &dns.Server{Addr: ":5353", Net: "udp", Handler: handler}
    server := &dns.Server{Addr: listeningPort, Net: "udp", Handler: handler}
    server.ListenAndServe()
    defer server.Shutdown()
}



