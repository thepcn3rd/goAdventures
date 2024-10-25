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
Purpose: Build a ConnectWise honeypot

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/connectwiseHoneypot/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o connectwisehp.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o connectwisehp.exe -ldflags "-w -s" main.go

Future Enhancements
1. Logging to JSON


Headers captured...
Connection: keep-alive
Cache-Control: private
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Xss-Protection: 1; mode=block
Content-Security-Policy: frame-ancestors 'self' blob: *.myconnectwise.net *.connectwisedev.com *.itboost.com; default-src 'self' 'unsafe-inline' 'unsafe-eval' blob: *.walkme.com *.connectwise *.connectwise.com az416426.vo.msecnd.net dc.services.visualstudio.com/v2/track *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.wise-pay.com *.wise-sync.com;style-src 'self' 'unsafe-inline' fonts.googleapis.com files.connectwise.com *.itsupport247.net *.myconnectwise.net; font-src 'self' 'unsafe-inline' 'unsafe-eval' data: *.walkme.com *.connectwise.com *.googleapis.com; img-src * data: snapshot:; frame-src * data: mailto:; connect-src 'self' *.walkme.com *.connectwise.com *.connectwisedev.com *.myconnectwise.net *.itsupport247.net cwview.com *.cwview.com dc.services.visualstudio.com/v2/track cheetah quotewerks://* wss://*.amazonaws.com; script-src 'self' 'unsafe-inline' 'unsafe-eval' *.connectwise.com *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.walkme.com *.cwnet.io *.itsupport247.net
Referrer-Policy: strict-origin-when-cross-origin
X-OneAgent-JS-Injection: true
X-ruxit-JS-Agent: true
Server-Timing: dtSInfo;desc="0", dtRpid;desc="-1756887520"
Set-Cookie: dtCookie=v_4_srv_3_sn_11CE2F40CED797EB65666B90F1ACCFD2_perc_100000_ol_0_mul_1_app-3Aea7c4b59f27d43eb_1_rcs-3Acss_0; Path=/


*/

func createCertificates() {
	configFileExists := cf.FileExists("/keys/certConfig.json")
	//fmt.Println(configFileExists)
	if !configFileExists {
		cf.CreateCertConfigFile()
		fmt.Println("Created keys/certConfig.json, modify the values to create the self-signed cert utilized")
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists("/keys/server.crt")
	keyFileExists := cf.FileExists("/keys/server.key")
	if !crtFileExists || !keyFileExists {
		cf.CreateCerts()
		crtFileExists := cf.FileExists("/keys/server.crt")
		keyFileExists := cf.FileExists("/keys/server.key")
		if !crtFileExists || !keyFileExists {
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
	fmt.Print(colorGreen + "<-- Request -->\n" + colorReset)
	remoteAddrItems := strings.Split(r.RemoteAddr, ":") // The displays the <sourceIP>:<sourcePort>
	fmt.Print("Date: " + stringTime + " SIP: " + remoteAddrItems[0] + " Method: " + r.Method + " URL: " + r.URL.String() + "\n")
	fmt.Print(colorGreen + "<-- Headers -->\n" + colorReset)
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

	fmt.Printf("\n%s<-- POST Body of Request -->%s\n", colorGreen, colorReset)
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

func honeyPotResponse(w http.ResponseWriter, r *http.Request, page string) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	loggingOutput(r)

	// If POST information exists in the body, capture the data to files
	// Test command: curl -k -X POST -d "key1=value1&key2=value2" https://127.0.0.1:11000/remote/login?lang=en
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

	// Fortinet Custom Headers
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "private")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubdomains")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'self' blob: *.myconnectwise.net *.connectwisedev.com *.itboost.com; default-src 'self' 'unsafe-inline' 'unsafe-eval' blob: *.walkme.com *.connectwise *.connectwise.com az416426.vo.msecnd.net dc.services.visualstudio.com/v2/track *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.wise-pay.com *.wise-sync.com;style-src 'self' 'unsafe-inline' fonts.googleapis.com files.connectwise.com *.itsupport247.net *.myconnectwise.net; font-src 'self' 'unsafe-inline' 'unsafe-eval' data: *.walkme.com *.connectwise.com *.googleapis.com; img-src * data: snapshot:; frame-src * data: mailto:; connect-src 'self' *.walkme.com *.connectwise.com *.connectwisedev.com *.myconnectwise.net *.itsupport247.net cwview.com *.cwview.com dc.services.visualstudio.com/v2/track cheetah quotewerks://* wss://*.amazonaws.com; script-src 'self' 'unsafe-inline' 'unsafe-eval' *.connectwise.com *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.walkme.com *.cwnet.io *.itsupport247.net")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-OneAgent-JS-Injection", "true")
	w.Header().Set("X-ruxit-JS-Agent", "true")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Server-Timing", "dtSInfo;desc=\"0\", dtRpid;desc=\"-1756887520\"")
	/*
		Connection: keep-alive
		Cache-Control: private
		Strict-Transport-Security: max-age=31536000; includeSubDomains
		X-Content-Type-Options: nosniff
		X-Xss-Protection: 1; mode=block
		Content-Security-Policy: frame-ancestors 'self' blob: *.myconnectwise.net *.connectwisedev.com *.itboost.com; default-src 'self' 'unsafe-inline' 'unsafe-eval' blob: *.walkme.com *.connectwise *.connectwise.com az416426.vo.msecnd.net dc.services.visualstudio.com/v2/track *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.wise-pay.com *.wise-sync.com;style-src 'self' 'unsafe-inline' fonts.googleapis.com files.connectwise.com *.itsupport247.net *.myconnectwise.net; font-src 'self' 'unsafe-inline' 'unsafe-eval' data: *.walkme.com *.connectwise.com *.googleapis.com; img-src * data: snapshot:; frame-src * data: mailto:; connect-src 'self' *.walkme.com *.connectwise.com *.connectwisedev.com *.myconnectwise.net *.itsupport247.net cwview.com *.cwview.com dc.services.visualstudio.com/v2/track cheetah quotewerks://* wss://*.amazonaws.com; script-src 'self' 'unsafe-inline' 'unsafe-eval' *.connectwise.com *.connectwisedev.com *.myconnectwise.net cwview.com *.cwview.com *.walkme.com *.cwnet.io *.itsupport247.net
		Referrer-Policy: strict-origin-when-cross-origin
		X-OneAgent-JS-Injection: true
		X-ruxit-JS-Agent: true
		Server-Timing: dtSInfo;desc="0", dtRpid;desc="-1756887520"
		Set-Cookie: dtCookie=v_4_srv_3_sn_11CE2F40CED797EB65666B90F1ACCFD2_perc_100000_ol_0_mul_1_app-3Aea7c4b59f27d43eb_1_rcs-3Acss_0; Path=/
	*/

	// Set custom cookie like the headers
	// Create a new cookie
	cookie := &http.Cookie{
		Name:  "dtCookie",
		Value: "v_4_srv_3_sn_11CE2F40CED797EB65666B90F1ACCFD2_perc_100000_ol_0_mul_1_app-3Aea7c4b59f27d43eb_1_rcs-3Acss_0",
		Path:  "/", // Set the path for the cookie
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
	indexHTMLB64 := buildLogonHTML(page)
	decodedBytes, err := base64.StdEncoding.DecodeString(indexHTMLB64)
	cf.CheckError("Unable to decode the base64 for the index html", err, true)
	respBodyBytes := bytes.NewReader([]byte(decodedBytes))
	io.Copy(w, respBodyBytes)

}

func buildLogonHTML(page string) string {
	// Open the designated file
	file, err := os.Open(page)
	cf.CheckError("Unable to read file for the logon html", err, true)
	defer file.Close()

	fileInfo, err := file.Stat()
	cf.CheckError("Unable to determine the size of the file", err, true)
	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	cf.CheckError("Unable to read the file into the byte array", err, true)

	encoded := base64.StdEncoding.EncodeToString(data)

	return encoded

}

func main() {
	listeningPortPtr := flag.String("port", "9000", "Change default listening port (default: -port 9000)")
	httpSecurePtr := flag.String("https", "true", "Disable site from running HTTPS (default: -https true)")
	changeUserPtr := flag.String("user", "nobody", "Change default user (default: -user nobody)")
	logonPagePtr := flag.String("index", "logon_connectwise.html", "Set the logon page that is displayed")
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
		honeyPotResponse(w, r, *logonPagePtr)
	})

	if *httpSecurePtr == "true" {
		fmt.Printf("Started the webserver with TLS on port: %s\n", listeningPort)
		log.Fatal(server.ListenAndServeTLS("./keys/server.crt", "./keys/server.key"))
	} else {
		fmt.Printf("Started the webserver with no encryption on port: %s\n", listeningPort)
		log.Fatal(server.ListenAndServe())
	}

}
