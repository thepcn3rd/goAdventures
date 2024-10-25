package main

// go run yaReverseHTTPProxy.go -https="Yes" -port="8070" -url="https://127.0.0.1:9000"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o yaReverseHTTPProxy.bin -ldflags "-w -s" yaReverseHTTPProxy.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o yaReverseHTTPProxy.exe -ldflags "-w -s" yaReverseHTTPProxy.go

// Create the TLS keys for the https web server
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key
// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365

// Directory structure
// - simpleReverseHTTPProxy.bin
// - keys/
// - - server.key
// - - server.crt

// References:
// https://www.youtube.com/watch?v=tWSmUsYLiE4
// https://dev.to/b0r/implement-reverse-proxy-in-gogolang-2cp4

/*
Works with proxying an evil-winrm connection through 5985 (Catch: You need to listen on port 5985 on the localhost...)


Future Steps:
1. Modify the user agent so that it is not proxied (Currently it shows the evil-winrm user-agent)

*/

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
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
	// Without the below line the reverse proxy does not handle TLS correctly
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// Introduce command line flags to dynamically update the following:
	// - Destination URL
	// - Listening Port
	dstURLPtr := flag.String("url", "http://127.0.0.1:9000", "URL to Proxy")
	var listeningPort string
	listeningPortPtr := flag.String("port", "8080", "Port to Listen for HTTP Requests")
	HTTPSPtr := flag.String("https", "Yes", "Setup HTTPS Proxy")
	userAgentPtr := flag.String("useragent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.79 Safari/537.36", "Modify User Agent")
	flag.Parse()
	if !isFlagPassed("url") && !isFlagPassed("port") {
		flag.Usage()
		os.Exit(0)
	}

	userAgentString := *userAgentPtr
	listeningPort = *listeningPortPtr

	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	// Define the destination server that is being connected to...
	dstURL := *dstURLPtr
	dstServerURL, err := url.Parse(dstURL)
	checkError("Unable to connect to destination URL", err)

	reverseProxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		req.Host = dstServerURL.Host
		req.URL.Host = dstServerURL.Host
		req.URL.Scheme = dstServerURL.Scheme

		// Filter based on the method also
		fmt.Print(colorGreen + "\n<-- New Request -->\n" + colorReset)
		fmt.Printf("Received request at: %s\n", time.Now())
		fmt.Printf("Method: %s\n", req.Method)
		fmt.Printf("URI: %s\n", req.RequestURI)
		// Can filter based on the user-agent also...
		fmt.Printf("User-Agent: %s\n", req.UserAgent())

		var dstServerResponse *http.Response
		var dstServerReq *http.Request
		var err error
		// Sweet we can filter based on the requestURI
		if req.RequestURI == "/chef/" {

			dstServerReq, err = http.NewRequest("GET", dstURL, nil)
			dstServerResponse, err = http.DefaultClient.Do(dstServerReq)
			checkError("Unable to send request to destination", err)

			// Return the response to the client
			rw.WriteHeader(http.StatusOK)
			io.Copy(rw, dstServerResponse.Body)
		} else {
			// The URI based on the documentation needs to be removed in a proxy situation
			req.RequestURI = ""

			// You do not want to pass the x-forwarded for unless you want to see that client side...
			///s, _, _ := net.SplitHostPort(req.RemoteAddr)
			///req.Header.Set("X-Forwarded-For", s)
			// This does manipulate the user-agent as it passes through the proxy
			// Modify the user agent string if it uses the Evil WinRM Useragent string to the common Microsoft WinRM Client String
			if strings.Contains(req.UserAgent(), "WinRM") {
				req.Header.Set("User-Agent", "Microsoft WinRM Client")
				fmt.Printf(colorGreen + "<-- User-Agent Modified -->\n" + colorReset)
				fmt.Printf("Modified User-Agent: " + "Microsoft WinRM Client" + "\n")
			} else {
				req.Header.Set("User-Agent", userAgentString)
				fmt.Printf(colorGreen + "<-- User-Agent Modified -->\n" + colorReset)
				fmt.Printf("Modified User-Agent: " + userAgentString + "\n")
			}

			// Send the request to the destination
			dstServerResponse, err = http.DefaultClient.Do(req)

			// Print the headers on the reverse proxy side
			// Copy the response headers to the client
			for key, values := range dstServerResponse.Header {
				for _, value := range values {
					fmt.Println(key + ": " + value)
					rw.Header().Set(key, value)
				}
			}

			//fmt.Print("\n")
			// Could filter based on status code also...
			fmt.Printf(colorGreen + "<-- Server Response -->\n" + colorReset)
			fmt.Printf("Status Code: " + strconv.Itoa(dstServerResponse.StatusCode) + "\n")
			checkError("Unable to send request to destination", err)

			//rw.WriteHeader(http.StatusOK)
			// Return the response to the client
			io.Copy(rw, dstServerResponse.Body)
		}

	})

	httpsServer := *HTTPSPtr
	//fmt.Print(httpsServer)
	//fmt.Print(listeningPort)
	if httpsServer == "No" || httpsServer == "no" || httpsServer == "N" || httpsServer == "n" {
		log.Fatal(http.ListenAndServe(":"+listeningPort, reverseProxy))
	} else {
		log.Fatal(http.ListenAndServeTLS(":"+listeningPort, "./keys/server.crt", "./keys/server.key", reverseProxy))
	}

}
