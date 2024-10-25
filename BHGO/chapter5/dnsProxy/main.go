package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/miekg/dns"
)

// Setup the environment variables for go
/*
go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter5/dnsProxy"

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
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dnsProxy.bin -ldflags "-w -s" main.go

DNS Message Structure

type Msg struct {
	MsgHdr
	Compress	bool	`json:"-"`
	Question	[]Question
	Answer		[]RR
	Ns			[]RR
	Extra		[]RR
}

Example config.json file:
{
        "server": {
                "port": "53",
                "primaryDNS": "8.8.8.8",  # Any request not handled will go to this dns server
                "primaryDNSPort": "53"
        },
        "proxy": {
                "domains": [
                        "attacker1.com.",
                        "attacker2.com.",
                        "ads.attacker3.com."
                ],
                "proxyServer": "127.0.0.1",  # This DNS server will need to handle the above domains
                "proxyPort": "5353"
        }
}

*/

// References
// https://freshman.tech/snippets/go/check-if-slice-contains-element/

/* Steps to do:
1. Add HandleFunc for each domain instead of reading the question for each domain (See dnsServer)
2. What if a subdomain of the domain does not exist, treat the parent domain as being the proxied domain
3. Add capability to add domains that are being proxied to a seperate file to be read when the server starts


*/

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

// If a string slice contains...
func contains(s []string, str string) bool {
	for _, v := range s {
		//if strings.Contains(v, str) {
		if v == str {
			return true
		}
	}
	return false
}

type Congifuration struct {
	Server serverStruct `json:"server"`
	Proxy  proxyStruct  `json:"proxy"`
}

type serverStruct struct {
	Port           string `json:"port"`
	PrimaryDNS     string `json:"primaryDNS"`
	PrimaryDNSPort string `json:"primaryDNSPort"`
}

type proxyStruct struct {
	Domains     []string `json:"domains"`
	ProxyServer string   `json:"proxyServer"`
	ProxyPort   string   `json:"proxyPort"`
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

func setupPermissionsProcess(userInput string) {
	var userInfo *user.User
	var err error
	// Modify the process the webserver runs as to the user and group specified...
	// The permissions on uploading files and accessing keys need to be set appropriately for the user and group...

	// Who is the current user logged in
	currentUserInfo, err := user.Current()
	checkError("Unable to get current user", err)
	//fmt.Println("Current User: " + currentUserInfo.Username)
	//fmt.Println("Current UID: " + currentUserInfo.Uid)
	//fmt.Println("Current GID: " + currentUserInfo.Gid)

	// Change the UID and the GID if the current user is root
	if currentUserInfo.Uid == "0" {
		if isFlagPassed("user") {
			userInfo, err = user.Lookup(userInput)
			checkError("Unable to change user", err)
			if userInfo.Uid == "0" {
				fmt.Printf("DANGER: Running the server as root is not recommended!\n\n")
			}
		} else {
			// If the flag is not set use the default of nobody to launch the server...
			userInfo, err = user.Lookup(userInput)
			checkError("Unable to change user", err)
		}
	} else {
		if isFlagPassed("user") {
			fmt.Println("Unable to change the UID and GID due to being a non-privileged user...")
		}
		userInfo = currentUserInfo
	}

	var gid int
	if userInfo.Username == "nobody" {
		gid, _ = strconv.Atoi("65534")
	} else {
		gid, _ = strconv.Atoi(userInfo.Gid)
	}
	if err := syscall.Setgid(gid); err != nil {
		fmt.Println("Error setting GID:", err)
		os.Exit(0)
	}

	uid, _ := strconv.Atoi(userInfo.Uid)
	if err := syscall.Setuid(uid); err != nil {
		fmt.Println("Error setting UID:", err)
		os.Exit(0)
	}

	fmt.Printf("\nPermissions set for the process of the server.\nUsername: %s  UID: %s  GID: %s\n\n", userInfo.Username, userInfo.Uid, userInfo.Gid)
}

func main() {
	var config Congifuration
	changeUserPtr := flag.String("user", "nobody", "Change default user (default: -user nobody)")
	ConfigPtr := flag.String("config", "config.json", "Location of the configuration file")
	flag.Parse()

	// Modify the permissions on the process running the server...
	setupPermissionsProcess(*changeUserPtr)

	// Read the configuration file for the DNS Proxy
	fmt.Println("Config File: " + *ConfigPtr)
	configFile, err := os.Open(*ConfigPtr)
	checkError("Unable to open the configuration file", err)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		checkError("Unable to decode the configuration file", err)
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		if len(req.Question) < 1 {
			dns.HandleFailed(w, req)
			return
		}
		// Parse out the domain name in the question
		fqdn := req.Question[0].Name
		// Below line is for troubleshooting the FQDN in the Question
		//fmt.Println("FQDN: " + fqdn)

		var resp *dns.Msg
		var err error
		proxied := false
		// Domains that are being proxied...
		//domainsProxied := []string{"attacker1.com.", "attacker2.com.", "ads.attacker3.com."}
		domainsProxied := config.Proxy.Domains
		if contains(domainsProxied, fqdn) {
			// Explicit DNS Server that it will proxy with...
			dnsProxyString := config.Proxy.ProxyServer + ":" + config.Proxy.ProxyPort
			resp, err = dns.Exchange(req, dnsProxyString)
			if err != nil {
				dns.HandleFailed(w, req)
				return
			}
			proxied = true
		} else {
			// All other requests are proxied through google...
			dnsPrimaryString := config.Server.PrimaryDNS + ":" + config.Server.PrimaryDNSPort
			//resp, err = dns.Exchange(req, "8.8.8.8:53")
			resp, err = dns.Exchange(req, dnsPrimaryString)
			if err != nil {
				dns.HandleFailed(w, req)
				return
			}
		}

		// Output the response of the A record to the DNS Request of the proxied DNS Server
		for _, answer := range resp.Answer {
			if a, ok := answer.(*dns.A); ok {
				if proxied {
					fmt.Println("Proxied: " + fqdn + " - " + a.A.String()) // Outputs the IP address of the response...
				} else {
					fmt.Println(fqdn + " - " + a.A.String()) // Outputs the IP address of the response...
				}
			}
		}

		if err := w.WriteMsg(resp); err != nil {
			dns.HandleFailed(w, req)
			return
		}

	})
	dnsString := ":" + config.Server.Port
	log.Fatal(dns.ListenAndServe(dnsString, "udp", nil))
}
