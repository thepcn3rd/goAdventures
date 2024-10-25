package main

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
Purpose: Build a ivanti honeypot

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/ivantiHoneypot/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ivantihp.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o ivantihp.exe -ldflags "-w -s" main.go

Modifications
1. Saving the POST output to a randomly named file in postInfo
2. Output to stdout the hex and the strings of the POST when it is a post
3. Extended the response to include /dana/ /dana-cached/ and /dana-ws/
4. Added the functionality to only respond with TLS1.1, TLS1.2 and specific ciphers


Future Enhancements
1. Logging to JSON


*/

func createCertificates() {
	configFileExists := cf.FileExists("/keys/certConfig.json")
	//fmt.Println(configFileExists)
	if configFileExists == false {
		cf.CreateCertConfigFile()
		fmt.Println("Created keys/certConfig.json, modify the values to create the self-signed cert utilized")
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists("/keys/server.crt")
	keyFileExists := cf.FileExists("/keys/server.key")
	if crtFileExists == false || keyFileExists == false {
		cf.CreateCerts()
		crtFileExists := cf.FileExists("/keys/server.crt")
		keyFileExists := cf.FileExists("/keys/server.key")
		if crtFileExists == false || keyFileExists == false {
			fmt.Println("Failed to create server.crt and server.key files")
			os.Exit(0)
		}
	}
}

func loggingOutput(r *http.Request) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	timeNow := time.Now()
	stringTime := timeNow.Format(time.RFC822)
	fmt.Print(colorGreen + "\n<-- HTTP Request from Client -->\n" + colorReset)
	remoteAddrItems := strings.Split(r.RemoteAddr, ":") // The displays the <sourceIP>:<sourcePort>
	fmt.Print("Date: " + stringTime + " SIP: " + remoteAddrItems[0] + " Method: " + r.Method + " URL: " + r.URL.String() + "\n")
	fmt.Print(colorGreen + "<-- HTTP Request Headers from Client -->\n" + colorReset)
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Println(key + ": " + value)
			//rw.Header().Set(key, value)
		}
	}

}

func generateRandomString(stringType string, length int) string {
	// Define the characters allowed in the password
	var chars string
	if stringType == "hex" {
		chars = "abcdef0123456789"
	} else if stringType == "alphanumeric" {
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	} else {
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	// Generate the random password
	password := make([]byte, length)
	for i := range password {
		password[i] = chars[rand.Intn(len(chars))]
	}

	return string(password)
}

func outputPost(data []byte) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	fmt.Printf("\n%s<-- POST Body of Response -->%s\n", colorGreen, colorReset)
	fmt.Printf("%-36s| %-20s\n", "Hex", "ASCII")
	fmt.Printf("------------------------------------------------------\n")

	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}

		//fmt.Printf("%-20x ", data[i:end])
		countChar := 0
		for x := i; x < end; x++ {
			fmt.Printf("%x", data[x])
			countChar++
			if countChar%4 == 0 && countChar > 0 {
				fmt.Printf(" ")
			}
		}

		fmt.Printf("| ")
		for j := i; j < end; j++ {
			fmt.Printf("%c", printableChar(data[j]))
		}

		fmt.Println()
	}
}

func printableChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}
	return b
}

func honeyPotResponse(w http.ResponseWriter, r *http.Request) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	loggingOutput(r)

	// If POST information exists in the body, capture the data to files
	// Test command: curl -k -X POST -d "key1=value1&key2=value2" https://127.0.0.1:9000/dana-ws/testing
	if r.Method == "POST" {
		responseBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read response body from the client", err, false)
		// Specify the path where you want to save the data
		fileName := generateRandomString("alphanumeric", 12) + ".raw"
		filePath := "postInfo/" + fileName

		err = os.WriteFile(filePath, responseBodyBytes, 0644)
		cf.CheckError("Unable to save the response body to the file system", err, false)

		outputPost(responseBodyBytes)
	}

	// Extract dynamic part from URI
	//uri := r.URL.Path[len("/"):]

	// Output the URI that was accessed
	//fmt.Printf("\n%sDynamic URI:%s %s\n", colorGreen, colorReset, uri)
	//fmt.Fprint(w, message)

	/////////////////////////////////////////////////////////////
	// Set custom headers
	// Get the current date and time
	currentTime := time.Now()
	// Format the date as a string
	currentDate := currentTime.Format("Mon, 02 Jan 2006 15:04:05")
	//fmt.Printf("Current Date: %s\n", currentDate)
	w.Header().Set("Date", currentDate)
	uriLocationString := generateRandomString("alphanumberic", 16)
	uriLocation := "/dana-na/auth/url_" + uriLocationString + "/welcome.cgi?p=more%2Dcred"
	w.Header().Set("location", uriLocation)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Expires", "-1")
	w.Header().Set("X-XSS-Protection", "1")
	w.Header().Set("accept-ch", "Sec-CH-UA-Platform-Version")

	// Set custom cookie like the headers
	// Create a new cookie
	cookieString := generateRandomString("hex", 33)
	cookieString = "state_" + cookieString
	cookie := &http.Cookie{
		Name:     "id",
		Value:    cookieString,
		Path:     "/",  // Set the path for the cookie
		HttpOnly: true, // Set HttpOnly flag
		//Expires:  time.Now().Add(24 * time.Hour), // Set expiration time (optional)
	}

	// Set the cookie in the response
	http.SetCookie(w, cookie)

	// Print all headers to stdout
	fmt.Printf("\n%s<-- Headers Sent to Client (Response) -->%s\n", colorGreen, colorReset)
	for key, values := range w.Header() {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	fmt.Println("")
	indexHTMLB64 := buildIndexHTML()
	decodedBytes, err := base64.StdEncoding.DecodeString(indexHTMLB64)
	cf.CheckError("Unable to decode the base64 for the index html", err, true)
	respBodyBytes := bytes.NewReader([]byte(decodedBytes))
	io.Copy(w, respBodyBytes)

}

func defaultResponse(w http.ResponseWriter, r *http.Request) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"
	loggingOutput(r)

	/////////////////////////////////////////////////////////////
	// Set custom headers
	// Get the current date and time
	currentTime := time.Now()
	// Format the date as a string
	currentDate := currentTime.Format("Mon, 02 Jan 2006 15:04:05")
	//fmt.Printf("Current Date: %s\n", currentDate)
	w.Header().Set("Date", currentDate)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Expires", "-1")
	w.Header().Set("X-XSS-Protection", "1")
	w.Header().Set("accept-ch", "Sec-CH-UA-Platform-Version")

	// Print all headers to stdout
	fmt.Printf("\n%sHeaders sent to client:%s\n", colorGreen, colorReset)
	for key, values := range w.Header() {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	fmt.Println("")
	indexHTML := "<html><body><a href='/dana-na/auth/url_default/welcome.cgi?'>Login to VPN</a></body></html>"
	respBodyBytes := bytes.NewReader([]byte(indexHTML))
	io.Copy(w, respBodyBytes)
}

func main() {
	listeningPortPtr := flag.String("port", "9000", "Change default listening port (default: -port 9000)")
	httpSecurePtr := flag.String("https", "true", "Disable site from running HTTPS (default: -https true)")
	changeUserPtr := flag.String("user", "nobody", "Change default user (default: -user nobody)")
	flag.Parse()

	cf.CreateDirectory("/keys")
	cf.CreateDirectory("/postInfo")
	createCertificates()

	// Modify the permissions on the process running the server...
	cf.SetupPermissionsProcess(*changeUserPtr)

	// Configure TLS to only support TLS1.2 and TLS1.3 with appropriate ciphers
	// Define a list of allowed cipher suites (modify based on your requirements)
	allowedCipherSuites := []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}

	// Create a TLS configuration with the allowed cipher suites
	tlsConfig := &tls.Config{
		CipherSuites: allowedCipherSuites,
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}

	listeningPort := *listeningPortPtr
	listeningPort = ":" + listeningPort

	// Create an HTTP server with TLS configuration
	server := &http.Server{
		Addr:      listeningPort, // You can change the port if needed
		TLSConfig: tlsConfig,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defaultResponse(w, r)
	})

	http.HandleFunc("/dana/", func(w http.ResponseWriter, r *http.Request) {
		honeyPotResponse(w, r)
	})

	http.HandleFunc("/dana-na/", func(w http.ResponseWriter, r *http.Request) {
		honeyPotResponse(w, r)
	})

	http.HandleFunc("/dana-ws/", func(w http.ResponseWriter, r *http.Request) {
		honeyPotResponse(w, r)
	})

	http.HandleFunc("/dana-cached/", func(w http.ResponseWriter, r *http.Request) {
		honeyPotResponse(w, r)
	})

	if *httpSecurePtr == "true" {
		fmt.Printf("Started the webserver with TLS on port: %s\n", listeningPort)
		log.Fatal(server.ListenAndServeTLS("./keys/server.crt", "./keys/server.key"))
	} else {
		fmt.Printf("Started the webserver with no encryption on port: %s\n", listeningPort)
		log.Fatal(server.ListenAndServe())
	}

}

func buildIndexHTML() string {
	base64 := "PGh0bWw+CjxoZWFkPgoKPG1ldGEgaHR0cC1lcXVpdj0iWC1VQS1Db21wYXRpYmxlIiBjb250ZW50"
	base64 = base64 + "PSJJRT1lZGdlIj4KPG1ldGEgaHR0cC1lcXVpdj0iQ29udGVudC1MYW5ndWFnZSI+CjxtZXRhIGh0"
	base64 = base64 + "dHAtZXF1aXY9IkNvbnRlbnQtVHlwZSIgY29udGVudD0idGV4dC9odG1sIj4KPG1ldGEgbmFtZT0i"
	base64 = base64 + "cm9ib3RzIiBjb250ZW50PSJub25lIj4KCjx0aXRsZT5JdmFudGkgQ29ubmVjdCBTZWN1cmU8L3Rp"
	base64 = base64 + "dGxlPgoKPHN0eWxlPgoJdGQgICB7IGZvbnQtZmFtaWx5OiB2ZXJkYW5hLCBhcmlhbCwgaGVsdmV0"
	base64 = base64 + "aWNhLCBzYW5zLXNlcmlmIDsgfQoJYm9keSB7IGZvbnQtZmFtaWx5OiB2ZXJkYW5hLCBhcmlhbCwg"
	base64 = base64 + "aGVsdmV0aWNhLCBzYW5zLXNlcmlmOyB9Cjwvc3R5bGU+Cgo8L2hlYWQ+Cjxib2R5PgoJPHRhYmxl"
	base64 = base64 + "IGlkPSJ0YWJsZV9Mb2dpblBhZ2VfMSIgYm9yZGVyPSIwIiB3aWR0aD0iMTAwJSIgY2VsbHNwYWNp"
	base64 = base64 + "bmc9IjAiIGNlbGxwYWRkaW5nPSIzIj4KCQk8dHI+CgkJCTx0ZCBiZ2NvbG9yPSIjRkZGRkZGIj4K"
	base64 = base64 + "CQkJCTxpbWcgYm9yZGVyPTAgc3JjPSJkYXRhOmltYWdlL3BuZztiYXNlNjQsIGlWQk9SdzBLR2dv"
	base64 = base64 + "QUFBQU5TVWhFVWdBQUFKQUFBQUEzQ0FJQUFBQnNQVlM1QUFBQUNYQklXWE1BQUE3RUFBQU94QUdW"
	base64 = base64 + "S3c0YkFBQUd5VWxFUVZSNG5PMmJUMURUV0J6SG56dDcyeWE5TGcwOUxyUWMyeVlaTDBwcDYwWDUw"
	base64 = base64 + "M3JZdFhhaEhCQlgvbHgwUjBCMnhORVZuWEdkV1JGMnJUTUxGR0U5Q0ZiUnkwb1ZQVWtySFB1SFBa"
	base64 = base64 + "YWtYRW5nekI0eXhremVTd2lZdEJ2SjU1VCs4dkx5bW0vZTcvM2U3NzBjMmQzZEJSYm00YXRxTjhC"
	base64 = base64 + "aWYxaUNtUXhMTUpQeGRiVWJjT2dZSEw1VzJpaXJGTUJ4MjhTOU8wcG5MY0VxemNyS2FvbGhWUW80"
	base64 = base64 + "Q1lmS1djc2xtZ3hMTUpOaENXWXlOSTFoVGFFVzJPM09UaWNvMG1OQWs3NXdYcjk2RGh0ajhmTXJt"
	base64 = base64 + "UTlhTHJkNm1NbXdvc1JLSXczcjFTTjRKSlpnbFVZYTFxdEg4RWdzbDJneUxNRk1oaVdZeVRpaVpU"
	base64 = base64 + "MHNrMTJEalc1WEhZYlpER2lTV2VINW5YeWg4RTk2dWJNOVNqaHFETHFMSnNFc1ZHRFk4cXYwbTZY"
	base64 = base64 + "MHUzeXh5RzN4QUlEbHBVWGpCTE9peE0vQ2U5UXZpRlF4ckRIc3M2aXdXc0FTekhTWXpDVXliQmtB"
	base64 = base64 + "d0RBc2htRzFCSUZoMytoVk04L3ZjRHdIQU1BeFhNZHFkVWVUWUZNemN6eTNMVE5Hd3MzQzBNcHZi"
	base64 = base64 + "MDlOejhGWEJRT05ibGVkU3JVTHFVV0drYSs5aXRXSzhQek9mT3FaZEZRWHdlMll1NzQrMG5hU3Br"
	base64 = base64 + "akM4UzN5TGpkdS9iYXpMVy84NkkycndnSExidjZWbkgyZFhpNUpXdUlrSERUdDdiM1FEZGVaeWE1"
	base64 = base64 + "dE1Jektud0lBcE5OdmJSTEpNUXdMQlJyRm44aC92ZWV6RXRFVUplNlpyVWVPdlRUbGV6VDE1NzZx"
	base64 = base64 + "eFhGczlmMGJxU1U1OC9qM2lRZGFob3IrM25OOUY4N0JkbVFpWEFqa2tqT1ByNCtxcGZMZ09nZUdS"
	base64 = base64 + "dVpUTC9ac2pCUW40WkJtNkpIdHVYM3phcVN0V1V0dCtveGg0ZFpUc0RGZktQSzgvTlVXeVdRUksr"
	base64 = base64 + "VW5taHFsUHdlSHIxMGZ2YU54WUw5M1A5RjYrcXpLSFdXTWpTZlUxUkxxbkpwQk9JOHFvbzlnSndK"
	base64 = base64 + "KzJNaHgvS3YwRzlndXNQQjBFVGFHdzUvZXNySHh4Sk1GUkJrVmN2bmkyQjhQdFpTY1RNN2RHMDlv"
	base64 = base64 + "S1RrMi9wQmhOL2ZWREVQUlJ6Q0s5RFM0WGJCOUlmVlM2Ukk0ZStJa0hLS1BaZGxONUFPbEtkL3Nk"
	base64 = base64 + "T0xmM0llM1M0djlQZDF3Z2NtcFdWa25RNlpqcHBOL0t6Vk1Cc2Z4azhsWmpZVXJnRzVSWXJEcFdD"
	base64 = base64 + "NWZrQmtGcndnL01xUS9wRW12ZVB3K2s0VnY0U1FjNHFEb2NOVDA5WFFCc0F2ck9wOTZIdjh4S3Y2"
	base64 = base64 + "MEsrZlBhTXJYMzNOT3lMRmxzbXNEUXlQSS9VeXYwMitIQnk0S3g1RndNMFY5YXVmbG9XdHcrVjhH"
	base64 = base64 + "TDhtQ0RxVUdIQURkQkl0M1JPRm5KM2hGZURoRitzUDI5alBpY2E2d2ppcndnNWFicm1UV3BJSXBF"
	base64 = base64 + "UW8yU2hjUEtkTHo3T2xjcktNN2x5L0tTcFlZVm56dEtOSkxrWHNJRmdnY04wRnFDclBaYU1vSHh6"
	base64 = base64 + "OExxWmV3WUVoL0tJMXJhZEpyaDE1TW12VEJOOFh0bUN3cUthTEVocm55c2ROSWE3c3ljUEZzQnlM"
	base64 = base64 + "VTVIaitmNUxwMW5QaUhBd2Nnd1dEdlNMU0g4cDZUeWpRS0oyN3FHQzN5UVhUQWszNWtKMmdsa0Qz"
	base64 = base64 + "REk3amplczArMEpQd1U2SFc4WW1Ic29lSCt3VmtmNHdoSW96cFFpTEY3bENjWU1wQ3hOaFlXZUUr"
	base64 = base64 + "aTVhSlZ5dTc1QjJoNElxMm1jTFJxT25ZSmpOUnZtOFMrbGxtVjNtRldGL3FQUytmN3g4Y1NIMVV1"
	base64 = base64 + "TXVNSTNBL3RZczZKeEw3R3lQd29KSnZTTFNIMGJhVGlKcnl4ZldMdytPNUl1YXhxUkRnczdaZW9y"
	base64 = base64 + "MDRIYjV5eXVkUVNQOUlVM0pvd2tBUUNyMUl0YlpiYWtsUS85c2ZiajFGRHd0RmIwaTdBK0RnVWJZ"
	base64 = base64 + "SDdMczVzOURJOGo2Rzl6MUpPbXBKV3B3REFNQS9IcjdidVVYcGFxSS9vS2RDUGhod1FTdm1DOFVZ"
	base64 = base64 + "WDhZQ2h5SEs0bWhZbXQzZmQzdzBDWFovdkQ3NC9JdzU4dEdmOEVvMGdOUHlBU3ZtTW1zeWdvN0NR"
	base64 = base64 + "ZHFsb1lZNTV5RVl6YVpnQ2REVzl1SFNDMWcwSW96amZwSUlwdFpROFNIa3F5QnlBcWtLd0NndDZj"
	base64 = base64 + "TE9YVTlWTjBMR0NSWXZBT1JHWHFTV29UN2pUUTlMOEtncGxhMUJBRWJNMW1FdEY4MmhnZ21wS24y"
	base64 = base64 + "TENaTnowdlpRczFTY3dWNVpoa294SnhiZkpYN0hKeU4xQkdqTnVFb1RhMmtJUDBoVU1pdko1T1Ba"
	base64 = base64 + "ZW1HZkdFZHVmakxjVlVXYlBUVzNieTJmT1lCTUVxd1VOQVBUOGhrOVBZaVFrRUFnSFQ5UXFURXNL"
	base64 = base64 + "MlJhQ2E3eXJCbGhpMlBUU1JpbllqMXNBcUQvUHlreExBdGthajNxTDhwMU93OTZvL0Z6K3Q0UjZO"
	base64 = base64 + "MlRXRTJXOGgvWEdYN1EwT0RTeWtkRlFyNmtiT3JFc09lN2FpK1NGTHFYWFZLeVV4dWl6Y2lJREp3"
	base64 = base64 + "WDJJazNLSnl0aVAydmRJcHpHYnIrNmxMeXkxd096WTcvUUMyQzd2aEtrQm4rOTRMYi9waW9HRElO"
	base64 = base64 + "SldJZWxRU2I0K2Via05zN0pHQzI3RkhrdzlxVVU3SjBHRmZDa1Y2a0RzVmpNUFluYi94MkJta1hU"
	base64 = base64 + "MDlMM0RyNWtoL1Q3ZlNONG8wNVhzK1ArZDIxV0U0NHAxZzJJT3N1UnlNdnA2dTJ6ZXZLclVUdDJO"
	base64 = base64 + "T1hSZlNUUEQxU2lhN2xpc1VlRzZiTFpjZE5UVVliZ3NGbXBTMmpWYVJmR0Y5ZzJINWo1TUtETU1h"
	base64 = base64 + "M0M3ZDIya0N3U3lrV0I5RG1BeExNSk5oQ1dZeUxNRk1oaVdZeWJBRU14bi9BZkxsNStQSjhTcGVB"
	base64 = base64 + "QUFBQUVsRlRrU3VRbUNDIiBhbHQ9IkxvZ28iPgoJCQk8L3RkPgoJCQk8dGQgYmdjb2xvcj0iI0ZG"
	base64 = base64 + "RkZGRiIgYWxpZ249InJpZ2h0Ij4KCQkJCSZuYnNwOwoJCQk8L3RkPgoJCTwvdHI+Cgk8L3RhYmxl"
	base64 = base64 + "PgoJPGhyIC8+Cgk8dGFibGUgaWQ9InRhYmxlX0xvZ2luUGFnZV8zIiBib3JkZXI9IjAiIGNlbGxw"
	base64 = base64 + "YWRkaW5nPSIyIiBjZWxsc3BhY2luZz0iMCI+CgkJPHRyPgoJCQk8dGQgd2lkdGg9MTUlPiZuYnNw"
	base64 = base64 + "OzwvdGQ+CgkJCTx0ZCBub3dyYXAgY29sc3Bhbj0iMyI+PGI+V2VsY29tZSB0byB0aGU8L2I+PC90"
	base64 = base64 + "ZD4KCQk8L3RyPgoJCTx0cj4KCQkJPHRkIHdpZHRoPTE1JT4mbmJzcDs8L3RkPgoJCQk8dGQgbm93"
	base64 = base64 + "cmFwIGNvbHNwYW49IjMiPjxzcGFuIGNsYXNzPSJjc3NMYXJnZSI+PGI+SXZhbnRpIENvbm5lY3Qg"
	base64 = base64 + "U2VjdXJlPC9iPjwvc3Bhbj48L3RkPgoJCTwvdHI+CgkJPHRyPgoJCQk8dGQgd2lkdGg9MTUlPiZu"
	base64 = base64 + "YnNwOzwvdGQ+CgkJCTx0ZCBjb2xzcGFuPSIyIj4mbmJzcDs8L3RkPgoJCTwvdHI+CgkJPHRyPgoJ"
	base64 = base64 + "CQk8dGQgd2lkdGg9MTUlPiZuYnNwOzwvdGQ+CgkJCTx0ZCB2YWxpZ249InRvcCI+CgkJCQk8dGFi"
	base64 = base64 + "bGUgYm9yZGVyPSIwIiBjZWxsc3BhY2luZz0iMCIgY2VsbHBhZGRpbmc9IjIiPgoJCQkJCTx0cj4K"
	base64 = base64 + "CQkJCQkJPHRkPgoJCQkJCQkJdXNlcm5hbWUmbmJzcDsmbmJzcDsmbmJzcDsKCQkJCQkJPC90ZD4K"
	base64 = base64 + "CQkJCQkJPHRkPgoJCQkJCQkJPGlucHV0IGlkPSJ1c2VybmFtZSIgdHlwZT0idGV4dCIgbmFtZSJ1"
	base64 = base64 + "c2VybmFtZSIgc2l6ZT0iMjAiPgoJCQkJCQk8L3RkPgoJCQkJCQk8dGQ+CgkJCQkJCQkmbmJzcDsm"
	base64 = base64 + "bmJzcDsmbmJzcDsmbmJzcDsmbmJzcDsmbmJzcDsmbmJzcDsmbmJzcDtQbGVhc2Ugc2lnbiBpbiB0"
	base64 = base64 + "byBiZWdpbiB5b3VyIHNlY3VyZSBzZXNzaW9uLgoJCQkJCQk8L3RkPgoJCQkJCTwvdHI+CgkJCQkJ"
	base64 = base64 + "PHRyPgoJCQkJCQk8dGQ+CgkJCQkJCQlwYXNzd29yZAoJCQkJCQk8L3RkPgoJCQkJCQk8dGQ+CgkJ"
	base64 = base64 + "CQkJCQk8aW5wdXQgaWQ9InBhc3N3b3JkIiB0eXBlPSJ0ZXh0IiBuYW1lInBhc3N3b3JkIiBzaXpl"
	base64 = base64 + "PSIyMCI+CgkJCQkJCTwvdGQ+CgkJCQkJCTx0ZD48L3RkPgoJCQkJCTwvdHI+CgkJCQkJPHRyPgoJ"
	base64 = base64 + "CQkJCQk8dGQ+CgkJCQkJCQkmbmJzcDs8YnIgLz48YnIgLz48YnIgLz4KCQkJCQkJPC90ZD4KCQkJ"
	base64 = base64 + "CQkJPHRkPgoJCQkJCQkJPGlucHV0IGlkPWJ0blN1Ym1pdCIgdHlwZT0ic3VibWl0IiB2YWx1ZT0i"
	base64 = base64 + "U2lnbiBJbiIgbmFtZT0iYnRuU3VibWl0Ij4KCQkJCQkJPC90ZD4KCQkJCQkJPHRkPjwvdGQ+CgkJ"
	base64 = base64 + "CQkJPC90cj4KCQkJCTwvdGFibGU+CgkJCTwvdGQ+CgkJPC90cj4KCTwvdGFibGU+CjwvYm9keT4K"
	base64 = base64 + "PC9odG1sPgoK"

	return base64
}
