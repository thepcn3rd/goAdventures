package main

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"crypto/tls"
	"encoding/base64"
	"errors"
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
go env -w GOPATH="/home/thepcn3rd/go/workspaces/fortinetHoneypot/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o fortinethp.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o fortinethp.exe -ldflags "-w -s" main.go

Modifications
1. Saving the POST output to a randomly named file in postInfo
2. Output to stdout the hex and the strings of the POST when it is a post
3. Extended the response of the logon page to include /dana/ /dana-cached/ and /dana-ws/
4. Added the functionality to only respond with TLS1.1, TLS1.2 and specific ciphers

Future Enhancements
1. Logging to JSON

HTML Example Logon Page
hxxps[:]//200[.]37.217.53:10443/remote/login?lang=en

Headers captured...
<-- New Request -->
Method: GET
URI: /remote/login
User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0
<-- User-Agent Modified -->
Modified User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.79 Safari/537.36
Content-Security-Policy: frame-ancestors 'self'; object-src 'self'; script-src 'self' https:   'unsafe-eval' 'unsafe-inline' blob:;
X-Xss-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000
Date: Tue, 20 Feb 2024 03:01:15 GMT
Server: xxxxxxxx-xxxxx
Set-Cookie: SVPNCOOKIE=; path=/; expires=Sun, 11 Mar 1984 12:00:00 GMT; secure; httponly; SameSite=Strict;
Set-Cookie: SVPNNETWORKCOOKIE=; path=/remote/network; expires=Sun, 11 Mar 1984 12:00:00 GMT; secure; httponly; SameSite=Strict
Content-Type: text/html; charset=utf-8
X-Frame-Options: SAMEORIGIN
X-Ua-Compatible: requiresActiveX=true
X-Content-Type-Options: nosniff


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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")
	w.Header().Set("X-Ua-Compatible", "requiresActiveX=true")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Server", "xxxxxxxx-xxxxx")
	/*
	   Method: GET
	   URI: /remote/login
	   User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0
	   <-- User-Agent Modified -->
	   Content-Security-Policy: frame-ancestors 'self'; object-src 'self'; script-src 'self' https:   'unsafe-eval' 'unsafe-inline' blob:;

	   Set-Cookie: SVPNCOOKIE=; path=/; expires=Sun, 11 Mar 1984 12:00:00 GMT; secure; httponly; SameSite=Strict;
	   Set-Cookie: SVPNNETWORKCOOKIE=; path=/remote/network; expires=Sun, 11 Mar 1984 12:00:00 GMT; secure; httponly; SameSite=Strict
	*/

	// Set custom cookie like the headers
	// Create a new cookie
	cookie := &http.Cookie{
		Name:     "SVPNCOOKIE",
		Value:    "",
		Path:     "/",  // Set the path for the cookie
		HttpOnly: true, // Set HttpOnly flag
		Secure:   true, // Set the Secure option of the cookie
		SameSite: http.SameSiteStrictMode,
		//Expires:  time.Now().Add(24 * time.Hour), // Set expiration time (optional)
	}

	cookie2 := &http.Cookie{
		Name:     "SVPNNETWORKCOOKIE",
		Value:    "",
		Path:     "/remote/network", // Set the path for the cookie
		HttpOnly: true,              // Set HttpOnly flag
		Secure:   true,              // Set the Secure option of the cookie
		SameSite: http.SameSiteStrictMode,
		//Expires:  time.Now().Add(24 * time.Hour), // Set expiration time (optional)
	}

	// Set the cookie in the response
	http.SetCookie(w, cookie)
	http.SetCookie(w, cookie2)

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

func createIndexHTML(folderDir string) {
	currentDir, _ := os.Getwd()
	newDir := currentDir + folderDir
	//cf.CheckError("Unable to get the working directory", err, true)
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		// Output to File - Overwrites if file exists...
		f, err := os.Create(newDir)
		cf.CheckError("Unable create file index.html "+currentDir, err, true)
		defer f.Close()
		indexHTML := "<html><body><a href='/remote/login?lang=en'>Login</a></body></html>"
		f.Write([]byte(indexHTML))
		f.Close()
	}
}

func main() {
	listeningPortPtr := flag.String("port", "9000", "Change default listening port (default: -port 9000)")
	httpSecurePtr := flag.String("https", "true", "Disable site from running HTTPS (default: -https true)")
	changeUserPtr := flag.String("user", "nobody", "Change default user (default: -user nobody)")
	logonPagePtr := flag.String("index", "fortinetLogonPage.html", "Set the logon page that is displayed")
	flag.Parse()

	cf.CreateDirectory("/keys")
	cf.CreateDirectory("/postInfo")
	cf.CreateDirectory("/static")
	cf.CreateDirectory("/static/favicon")
	createIndexHTML("/static/index.html")
	createIndexHTML("/static/favicon/index.html")
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

	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loggingOutput(r)
		//fmt.Println("Reading in this function")
		/////////////////////////////////////////////////////////////
		// Set custom headers
		// Get the current date and time
		currentTime := time.Now()
		// Format the date as a string
		currentDate := currentTime.Format("Mon, 02 Jan 2006 15:04:05")
		//fmt.Printf("Current Date: %s\n", currentDate)
		w.Header().Set("Date", currentDate)

		// Fortinet Custom Headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("X-Ua-Compatible", "requiresActiveX=true")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Server", "xxxxxxxx-xxxxx")

		// Set custom cookie like the headers
		// Create a new cookie
		cookie := &http.Cookie{
			Name:     "SVPNCOOKIE",
			Value:    "",
			Path:     "/",  // Set the path for the cookie
			HttpOnly: true, // Set HttpOnly flag
			Secure:   true, // Set the Secure option of the cookie
			SameSite: http.SameSiteStrictMode,
			//Expires:  time.Now().Add(24 * time.Hour), // Set expiration time (optional)
		}

		cookie2 := &http.Cookie{
			Name:     "SVPNNETWORKCOOKIE",
			Value:    "",
			Path:     "/remote/network", // Set the path for the cookie
			HttpOnly: true,              // Set HttpOnly flag
			Secure:   true,              // Set the Secure option of the cookie
			SameSite: http.SameSiteStrictMode,
			//Expires:  time.Now().Add(24 * time.Hour), // Set expiration time (optional)
		}

		// Set the cookie in the response
		http.SetCookie(w, cookie)
		http.SetCookie(w, cookie2)

		// Print all headers to stdout
		fmt.Printf("\n%s<-- Headers Sent to Client (Response) -->%s\n", colorGreen, colorReset)
		for key, values := range w.Header() {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	})

	http.HandleFunc("/remote/", func(w http.ResponseWriter, r *http.Request) {
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
