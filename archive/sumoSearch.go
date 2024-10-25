package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sumoJSONParser"
	"time"
)

// Setup the following for the application
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/sumoSearchLogs"

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o sumoQueryAPI.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o sumoQueryAPI.exe -ldflags "-w -s" main.go

References: https://api.sumologic.com/docs/#section/Getting_Started

Build a config.json file with the following:
{
        "auth": {
                "accessID": "",
                "accessKey": ""
				"encryptedAccessID": "",
                "encryptedAccessKey": ""
        },
        "url": {
                "api": "",
				"key": "",
				"userAgent": "",
				"xAbilities": "",
				"encryptedKey": "",
				"encryptedUserAgent": "",
				"encryptedXAbilities": ""
        }
}

// After the application executes the first time the accessID and accessKey are encrypted

// Create lambda function for the pull of the secret to decrypt the accessID and accessKey
import json
import hashlib
import secrets
import string

def generate_random_string(length):
    # Define the character set for the random string
    characters = string.ascii_letters + string.digits
    # Generate a random string of the specified length
    random_string = ''.join(secrets.choice(characters) for _ in range(length))
    return random_string

def lambda_handler(event, context):

    httpData = event['requestContext']['http']
    method = httpData['method']
    ua = httpData['userAgent']
    uaSHA256 = hashlib.sha256()
    uaSHA256.update(ua.encode('utf-8'))
    uaHash = uaSHA256.hexdigest()

    headers = event['headers']
    if 'x-content-type-abilities' in headers:
        cta = headers['x-content-type-abilities']
        ctaSHA256 = hashlib.sha256()
        ctaSHA256.update(cta.encode('utf-8'))
        ctaHash = ctaSHA256.hexdigest()
    else:
        cta = "nothing"

    if method == "POST" and uaHash == "hashofvalue" and ctaHash == "hashofvalue":
        #info = json.dumps(event)
        info = "32characterpassword"
        #info = ctaHash
    else:
        info = generate_random_string(32)

    return {
        'statusCode': 200,
        'body': info
    }


*/

type Congifuration struct {
	Auth AuthStruct `json:"auth"`
	URL  URLStruct  `json:"url"`
}

type AuthStruct struct {
	AccessID           string `json:"accessID"`
	AccessKey          string `json:"accessKey"`
	EncryptedAccessID  string `json:"encryptedAccessID"`
	EncryptedAccessKey string `json:"encryptedAccessKey"`
}

type URLStruct struct {
	URL                 string `json:"api"`
	KeyURL              string `json:"key"`
	UserAgent           string `json:"userAgent"`
	XAbilities          string `json:"xAbilities"`
	EncryptedKeyURL     string `json:"encryptedKey"`
	EncryptedUserAgent  string `json:"encryptedUserAgent"`
	EncryptedXAbilities string `json:"encryptedXAbilities"`
}

type queryInformationStruct struct {
	SrcIP        string
	SrcIPList    []string
	DestIP       string
	DestIPList   []string
	SrcPort      string
	SrcPortList  []string
	DestPort     string
	DestPortList []string
	MessageLimit int
}

type searchJobStruct struct {
	Query    string `json:"query"`
	From     string `json:"from"`
	To       string `json:"to"`
	TimeZone string `json:"timeZone"`
}

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func saveDebugFile(fileBytes []byte, fileName string) {
	outFile, err := os.Create(fileName)
	checkError("Unable to create debug txt file", err)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	_, err = w.WriteString(string(fileBytes))
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func saveOutputFile(message string, fileName string) {
	outFile, err := os.Create(fileName)
	checkError("Unable to create txt file", err)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	_, err = w.WriteString(message)
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func defaultSearchConf(timeLimitString string) *searchJobStruct {
	s := searchJobStruct{}
	timeLimit, err := strconv.Atoi(timeLimitString)
	checkError("Unable to convert timeLimitString to an Integer", err)
	yesterday := time.Now().AddDate(0, 0, -1*timeLimit)
	today := time.Now()
	s.From = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", yesterday.Year(), yesterday.Month(), yesterday.Day(), yesterday.Hour(), yesterday.Minute(), yesterday.Second())
	s.To = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", today.Year(), today.Month(), today.Day(), today.Hour(), today.Minute(), today.Second())
	s.TimeZone = "MST"
	return &s
}

func connectSumo(httpVerb string, url string, postBytes []byte, debugFilename string, sumoID string, sumoKey string, createDebug bool) []byte {
	var httpClient http.Client
	sumoAccessID := sumoID
	sumoAccessKey := sumoKey
	accessInfo := sumoAccessID + ":" + sumoAccessKey
	accessInfoB64 := "Basic " + b64.StdEncoding.EncodeToString([]byte(accessInfo))
	// Create Search Job through the API
	var httpRequest *http.Request
	var err error
	//fmt.Printf("Length of Bytes: %d\n", len(postBytes))
	if len(postBytes) > 1 {
		httpRequest, err = http.NewRequest(httpVerb, url, bytes.NewBuffer(postBytes))
	} else {
		// This meets the condition for a GET Verb with nil io.Reader
		httpRequest, err = http.NewRequest(httpVerb, url, nil)
	}
	checkError("Unable to prepare http request for the SumoLogic API", err)
	httpRequest.Header.Add("Authorization", accessInfoB64)
	httpRequest.Header.Add("Accept", "application/json")
	httpRequest.Header.Add("Content-Type", "application/json")

	httpResponse, err := httpClient.Do(httpRequest)
	// Verify we receive a 200 response and if not exit the program...
	// 202 Accepted Message occurs when a job is accepted...
	if httpResponse.Status != "200 OK" && httpResponse.Status != "202 Accepted" {
		fmt.Println("Response Status: " + httpResponse.Status)
		os.Exit(0)
	}

	httpBodyBytes, err := io.ReadAll(httpResponse.Body)
	checkError("Unable to read the http response body to create debug file", err)
	if createDebug == true {
		saveDebugFile(httpBodyBytes, debugFilename)
	}

	return httpBodyBytes
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

func InttoIP4(ipInt int64) string {
	// Need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

func IP4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}

func prepareString(osDetected string, input string) string {
	var output string
	if osDetected == "windows" {
		output = strings.Replace(input, "\r\n", "", -1)
	} else {
		output = strings.Replace(input, "\n", "", -1)
	}
	return output
}

func pullKey(keyURL string, userAgentString string, xAbility string) string {
	url := keyURL
	req, err := http.NewRequest("POST", url, nil)
	checkError("Unable to pull key from URL...", err)
	req.Header.Set("User-Agent", userAgentString)
	req.Header.Set("X-Content-Type-Abilities", xAbility)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	checkError("Unable to get response from Amazon...", err)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	respBodyString := string(respBody)
	//fmt.Println(string(respBody))
	return respBodyString
}

// EncryptString encrypts a string using AES-256 encryption with a random IV.
func EncryptString(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	//fmt.Println(iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string using AES-256 encryption.
func DecryptString(key []byte, ciphertext string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("Ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return string(data), nil
}

func main() {
	var queryInfo queryInformationStruct
	var err error
	osDetected := runtime.GOOS
	ConfigPtr := flag.String("config", "config.json", "Location of the configuration file")
	// Source IP Options
	SrcIPPtr := flag.String("sip", "", "Source IP Address to Query")
	SrcIPListPtr := flag.String("slist", "", "Comma-seperated list of Source IP Addresses")
	SrcIPFilePtr := flag.String("sfile", "", "A list of Source IP Addresses to search contained in a file")
	SrcIPRangePtr := flag.String("srange", "y", "Prompts for a Source IP Range to Query")
	// Destination IP Options
	DestIPPtr := flag.String("dip", "", "Destination IP Address to Query")
	DestIPListPtr := flag.String("dlist", "", "comma-seperated list of Dest IP Addresses")
	DestIPFilePtr := flag.String("dfile", "", "A list of Dest IP Addresses to search contained in a file")
	DestIPRangePtr := flag.String("drange", "y", "Prompts for a Destination IP Range to Query")
	// Source Port Options
	SrcPortPtr := flag.String("sport", "", "Source Port to Query")
	SrcPortListPrt := flag.String("splist", "", "Source Port List")
	// Destination Port Options
	DestPortPtr := flag.String("dport", "", "Destination Port to Query")
	DestPortListPrt := flag.String("dplist", "", "Destination Port List")
	// Read custom query from file ...
	CustomQueryFile := flag.String("qfile", "", "Read a file to execute a custom query")
	LimitPtr := flag.String("limit", "10", "Limit of Messages Returned, Max is 100 at the moment")
	DurationPtr := flag.String("time", "1", "Duration of days in time to search, Max is 14 at the moment")
	OutputFilePtr := flag.String("output", "messagesOutput.txt", "Output information to the following file")
	DebugPtr := flag.String("debug", "no", "Create debug output for troubleshooting JSON messages")
	flag.Parse()

	createDebugFiles := false
	if strings.ToLower(*DebugPtr) == "yes" || strings.ToLower(*DebugPtr) == "y" {
		createDebugFiles = true
	}

	var sumoURL string
	var accessKey string
	var accessID string
	var sumoAccessKey string
	var sumoAccessID string
	var encryptedKey string
	var encryptedID string
	var pulledKey string

	fmt.Println("Config File: " + *ConfigPtr)
	configFile, err := os.Open(*ConfigPtr)
	checkError("Unable to open the configuration file", err)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Congifuration
	if err := decoder.Decode(&config); err != nil {
		checkError("Unable to decode the configuration file", err)
	}
	sumoURL = config.URL.URL
	accessKey = config.Auth.AccessKey
	accessID = config.Auth.AccessID
	encryptedKey = config.Auth.EncryptedAccessKey
	encryptedID = config.Auth.EncryptedAccessID
	keyURL := config.URL.KeyURL
	userAgent := config.URL.UserAgent
	xAbilities := config.URL.XAbilities
	encryptedKeyURL := config.URL.EncryptedKeyURL
	encryptedUserAgent := config.URL.EncryptedUserAgent
	encryptedXAbilities := config.URL.EncryptedXAbilities

	if (keyURL != "" && encryptedKeyURL == "") || (userAgent != "" && encryptedUserAgent == "") || (xAbilities != "" && encryptedXAbilities == "") {
		fmt.Println("Key URL, User Agent or X Abilities is not encrypted in the config file")
		// Had to embed the encryption key for the Key URL
		// This is a minor feature to protect the quick gathering of the Lambda URL that is used to pull the key
		if keyURL != "" {
			b64KeyURL, err := EncryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), keyURL)
			checkError("Unable to encrypt the key URL", err)
			fmt.Printf("Encrypted Key URL: %s\n", b64KeyURL)
		}
		if userAgent != "" {
			encryptedUserAgent, err := EncryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), userAgent)
			checkError("Unable to encrypt the User Agent", err)
			fmt.Printf("Encrypted User Agent: %s\n", encryptedUserAgent)
		}
		if xAbilities != "" {
			encryptedXAbilities, err := EncryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), xAbilities)
			checkError("Unable to encrypt the X Abilities", err)
			fmt.Printf("Encrypted X Abilities: %s\n", encryptedXAbilities)
		}
		os.Exit(0)
	} else if keyURL != "" && userAgent != "" && xAbilities != "" {
		fmt.Println("The unencrypted Key URL, user agent or xAbilities needs to be removed from the config file")
		os.Exit(0)
	} else {
		decryptedKeyURL, err := DecryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), encryptedKeyURL)
		checkError("Unable to decrypt the Key URL", err)
		decryptedUserAgent, err := DecryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), encryptedUserAgent)
		checkError("Unable to decrypt the User Agent", err)
		decryptedXAbilities, err := DecryptString([]byte("kJExxjzNOUldQnOwYVMlTf8wVgkao7us"), encryptedXAbilities)
		checkError("Unable to decrypt the X Abilities", err)
		//fmt.Println("Decrypted Key URL: " + decryptedKeyURL)
		//fmt.Println("Decrypted User Agent: " + decryptedUserAgent)
		//fmt.Println("Decrypted X Abilities: " + decryptedXAbilities)
		pulledKey = pullKey(decryptedKeyURL, decryptedUserAgent, decryptedXAbilities)
		//fmt.Println("Pulled Key: " + pulledKey)
	}

	if encryptedKey == "" || encryptedID == "" {
		fmt.Println("Access Key and Access ID are not encrypted in the config file")
		b64AccessKey, err := EncryptString([]byte(pulledKey), accessKey)
		checkError("Unable to encrypt the access key", err)
		fmt.Printf("Encrypted Access Key: %s\n", b64AccessKey)
		b64AccessID, err := EncryptString([]byte(pulledKey), accessID)
		checkError("Unable to encrypt the access ID", err)
		fmt.Printf("Encrypted Access ID: %s\n", b64AccessID)
		os.Exit(0)
	}

	if len(encryptedKey) > 1 && accessKey != "" {
		fmt.Println("Remove the Unencrypted Key from the config file")
		os.Exit(0)
	}

	if len(encryptedID) > 1 && accessID != "" {
		fmt.Println("Remove the Unencrypted ID from the config file")
		os.Exit(0)
	}

	sumoAccessID, err = DecryptString([]byte(pulledKey), encryptedID)
	checkError("Unable to decrypt the access ID", err)
	sumoAccessKey, err = DecryptString([]byte(pulledKey), encryptedKey)
	checkError("Unable to decrypt the access key", err)

	if !isFlagPassed("sip") && !isFlagPassed("dip") && !isFlagPassed("sport") && !isFlagPassed("dport") && !isFlagPassed("slist") && !isFlagPassed("dlist") && !isFlagPassed("splist") && !isFlagPassed("dplist") && !isFlagPassed("sfile") && !isFlagPassed("dfile") && !isFlagPassed("qfile") && !isFlagPassed("drange") && !isFlagPassed("srange") && !isFlagPassed("time") && !isFlagPassed("config") {
		fmt.Println("A source IP, source IP List, dest IP, dest IP List or a destination port needs to be specified")
		flag.Usage()
		os.Exit(0)
	}

	// A destination IP and a destination IP List can not both be passed
	if (isFlagPassed("dip") && isFlagPassed("dlist")) || (isFlagPassed("dip") && isFlagPassed("dfile")) || (isFlagPassed("dlist") && isFlagPassed("dfile")) {
		fmt.Println("Only a dest IP Address, dest IP Address file or a dest IP List can be specified")
		flag.Usage()
		os.Exit(0)
	}

	// A source IP and a source IP List can not both be passed
	if (isFlagPassed("sip") && isFlagPassed("slist")) || (isFlagPassed("sip") && isFlagPassed("sfile")) || (isFlagPassed("sfile") && isFlagPassed("slist")) {
		fmt.Println("Only a source IP Address, source IP Address file or a source IP List can be specified")
		flag.Usage()
		os.Exit(0)
	}

	queryInfo.SrcIP = *SrcIPPtr
	queryInfo.DestIP = *DestIPPtr
	queryInfo.DestPort = *DestPortPtr
	queryInfo.SrcPort = *SrcPortPtr

	checkError("Unable to convert Integer into String for the Destination Port", err)

	// Build the query based on the parameters...
	rawQuery := "_sourceCategory=\"network/firewall\""
	if isFlagPassed("sip") {
		rawQuery = rawQuery + " " + "src_ip=\"" + queryInfo.SrcIP + "\""
	}
	if isFlagPassed("slist") {
		// Using the or seems to run longer than...
		rawQuery = rawQuery + " " + "("
		SourceIPList := strings.ReplaceAll(*SrcIPListPtr, " ", "")
		SourceIPListItems := strings.Split(SourceIPList, ",")
		lenList := len(SourceIPListItems)
		for i, sip := range SourceIPListItems {
			rawQuery = rawQuery + "src_ip=\"" + sip + "\""
			if i == (lenList - 1) {
				rawQuery = rawQuery + ")"
			} else {
				rawQuery = rawQuery + " OR "
			}
		}
	}
	if isFlagPassed("sfile") {
		rawQuery = rawQuery + " " + "("
		file, err := os.Open(*SrcIPFilePtr)
		checkError("Unable to open file specified for the source IP Addresses", err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			sip := scanner.Text()
			// Clean out of string all characters not associated to an IP Address
			sip = regexp.MustCompile(`[^0-9\.]+`).ReplaceAllString(sip, "")
			rawQuery = rawQuery + "src_ip=\"" + sip + "\" OR "
		}
		// Finish the query
		rawQuery = rawQuery[:len(rawQuery)-4] + ")"
	}

	if isFlagPassed("srange") {
		fmt.Printf("Input from the user is: %s\n", *SrcIPRangePtr)
		fmt.Print("Starting IP Address: ")
		inputReader := bufio.NewReader(os.Stdin)
		startingIP, _ := inputReader.ReadString('\n')
		startingIP = prepareString(osDetected, startingIP)
		fmt.Print("Last IP Address: ")
		lastIP, _ := inputReader.ReadString('\n')
		lastIP = prepareString(osDetected, lastIP)
		startIPDecimal := IP4toInt(net.ParseIP(startingIP))
		//fmt.Println(startingIP)
		fmt.Println(startIPDecimal)
		lastIPDecimal := IP4toInt(net.ParseIP(lastIP))
		fmt.Println(lastIPDecimal)
		totalIPAddresses := (lastIPDecimal - startIPDecimal) + 1
		if totalIPAddresses > 128 {
			fmt.Printf("Error: Total Addresses being scanned is more than 128!! (%d)\n", totalIPAddresses)
			os.Exit(0)
		}
		rawQuery = rawQuery + " " + "("
		for intIPAddr := startIPDecimal; intIPAddr <= lastIPDecimal; intIPAddr++ {
			currentIPAddress := InttoIP4(intIPAddr)
			//fmt.Println(currentIPAddress)
			rawQuery = rawQuery + "src_ip=\"" + currentIPAddress + "\" OR "
		}
		// Finish the query
		rawQuery = rawQuery[:len(rawQuery)-4] + ")"
		//fmt.Println(rawQuery)
		//os.Exit(0)
	}

	if isFlagPassed("dip") {
		rawQuery = rawQuery + " " + "dest_ip=\"" + queryInfo.DestIP + "\""
	}

	if isFlagPassed("dlist") {
		// Using the or seems to run longer than...
		rawQuery = rawQuery + " " + "("
		DestIPList := strings.ReplaceAll(*DestIPListPtr, " ", "")
		DestIPListItems := strings.Split(DestIPList, ",")
		lenList := len(DestIPListItems)
		for i, dip := range DestIPListItems {
			rawQuery = rawQuery + "dest_ip=\"" + dip + "\""
			if i == (lenList - 1) {
				rawQuery = rawQuery + ")"
			} else {
				rawQuery = rawQuery + " OR "
			}
		}
	}
	if isFlagPassed("dfile") {
		rawQuery = rawQuery + " " + "("
		file, err := os.Open(*DestIPFilePtr)
		checkError("Unable to open file specified for the dest IP Addresses", err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			dip := scanner.Text()
			// Clean out of string all characters not associated to an IP Address
			dip = regexp.MustCompile(`[^0-9\.]+`).ReplaceAllString(dip, "")
			rawQuery = rawQuery + "dest_ip=\"" + dip + "\" OR "
		}
		// Finish the query
		rawQuery = rawQuery[:len(rawQuery)-4] + ")"
	}

	if isFlagPassed("drange") {
		fmt.Printf("The desination IP Subnet declared is: %s\n", *DestIPRangePtr)
		fmt.Print("Starting IP Address: ")
		inputReader := bufio.NewReader(os.Stdin)
		startingIP, _ := inputReader.ReadString('\n')
		startingIP = prepareString(osDetected, startingIP)
		fmt.Print("Last IP Address: ")
		lastIP, _ := inputReader.ReadString('\n')
		lastIP = prepareString(osDetected, lastIP)
		startIPDecimal := IP4toInt(net.ParseIP(startingIP))
		//fmt.Println(startingIP)
		//fmt.Println(startIPDecimal)
		lastIPDecimal := IP4toInt(net.ParseIP(lastIP))
		totalIPAddresses := (lastIPDecimal - startIPDecimal) + 1
		if totalIPAddresses > 128 {
			fmt.Printf("Error: Total Addresses being scanned is more than 128!! (%d)\n", totalIPAddresses)
			os.Exit(0)
		}
		rawQuery = rawQuery + " " + "("
		for intIPAddr := startIPDecimal; intIPAddr <= lastIPDecimal; intIPAddr++ {
			currentIPAddress := InttoIP4(intIPAddr)
			//fmt.Println(currentIPAddress)
			rawQuery = rawQuery + "dest_ip=\"" + currentIPAddress + "\" OR "
		}
		// Finish the query
		rawQuery = rawQuery[:len(rawQuery)-4] + ")"
		//fmt.Println(rawQuery)
		//os.Exit(0)
	}

	if isFlagPassed("dport") {
		rawQuery = rawQuery + " " + "dest_port=\"" + queryInfo.DestPort + "\""
	}

	if isFlagPassed("sport") {
		rawQuery = rawQuery + " " + "src_port=\"" + queryInfo.SrcPort + "\""
	}

	if isFlagPassed("splist") {
		// Using the or seems to run longer than...
		rawQuery = rawQuery + " " + "("
		SrcPortList := strings.ReplaceAll(*SrcPortListPrt, " ", "")
		SrcPortListItems := strings.Split(SrcPortList, ",")
		lenList := len(SrcPortListItems)
		for i, sport := range SrcPortListItems {
			rawQuery = rawQuery + "src_port=\"" + sport + "\""
			if i == (lenList - 1) {
				rawQuery = rawQuery + ")"
			} else {
				rawQuery = rawQuery + " OR "
			}
		}
	}

	if isFlagPassed("dplist") {
		// Using the or seems to run longer than...
		rawQuery = rawQuery + " " + "("
		DestPortList := strings.ReplaceAll(*DestPortListPrt, " ", "")
		DestPortListItems := strings.Split(DestPortList, ",")
		lenList := len(DestPortListItems)
		for i, dport := range DestPortListItems {
			rawQuery = rawQuery + "dest_port=\"" + dport + "\""
			if i == (lenList - 1) {
				rawQuery = rawQuery + ")"
			} else {
				rawQuery = rawQuery + " OR "
			}
		}
	}
	var limitString string
	if isFlagPassed("limit") {
		limitInt, err := strconv.Atoi(*LimitPtr)
		checkError("Unable to convert limit string to an integer", err)
		if limitInt > 100 {
			limitInt = 100
		}
		limitString = strconv.Itoa(limitInt)
	} else {
		limitString = "10"
	}

	if isFlagPassed("qfile") {
		fmt.Printf("A custom query was specified and will replace anything provided except time and limit.  The custom query needs to be on 1 line\n\n")
		file, err := os.Open(*CustomQueryFile)
		checkError("Unable to open file specified for the dest IP Addresses", err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			query := scanner.Text()
			rawQuery = query
		}
	}

	// Save the raw query prior to applying the limit variable (The limit variable may already be set)
	if isFlagPassed("qfile") && strings.Contains(rawQuery, "limit") {
		saveOutputFile(rawQuery, "lastQuery.txt")
		fmt.Printf("Saved the current query executed to the file: lastQuery.txt\n\n")
	} else {
		rawQuery = rawQuery + " " + "| limit " + limitString
		saveOutputFile(rawQuery, "lastQuery.txt")
		fmt.Printf("Saved the currentquery executed to the file: lastQuery.txt\n\n")
	}

	//var timeLimitString string
	timeLimitString := "1"
	if isFlagPassed("time") {

		timeLimitInt, err := strconv.Atoi(*DurationPtr)
		checkError("Unable to convert time duration string to an integet", err)
		if timeLimitInt > 14 {
			timeLimitInt = 14
		}
		timeLimitString = strconv.Itoa(timeLimitInt)
	} else {
		timeLimitString = "1"
	}

	//monthAgoDate = time.Now().AddDate(0, -1, 0)
	fmt.Println("Let the journey begin...")
	fmt.Printf("Query: %s\n\n", rawQuery)

	////////////// Create the Search Job in Sumo Logic
	/****************************************************************************/
	searchRequest := defaultSearchConf(timeLimitString)
	searchRequest.Query = rawQuery
	jsonData, _ := json.Marshal(searchRequest)
	fmt.Println(string(jsonData))

	sumoWebhookQueryURL := sumoURL
	httpBodyBytes := connectSumo("POST", sumoWebhookQueryURL, jsonData, "createSearchJob.debug", sumoAccessID, sumoAccessKey, createDebugFiles)
	// Parse the response into the struct for SearchJobs
	searchJobsJSON := sumoJSONParser.SearchJobs(httpBodyBytes)

	////////////////// Poll the Jobs to See if the Search Created is Complete
	sumoWebhookPollURL := searchJobsJSON.Link.HREF
	pollStateStatus := "INCOMPLETE"
	loopCount := 0
	var pollJobsJSON *sumoJSONParser.PollJobsStruct
	for pollStateStatus != "DONE" {
		httpBodyBytes = connectSumo("GET", sumoWebhookPollURL, []byte(""), "pollSearchJob.debug", sumoAccessID, sumoAccessKey, createDebugFiles)
		pollJobsJSON = sumoJSONParser.PollJobs(httpBodyBytes)
		fmt.Printf("Poll jobs state: %s\n", pollJobsJSON.State)
		pollState := strings.ToLower(pollJobsJSON.State)

		if strings.Contains(pollState, "done") {
			pollStateStatus = "DONE"
			fmt.Printf("The job to pull the logs completed...\n\n")
		} else if loopCount > 10 {
			fmt.Println("Waited for 3 minutes to gather results, failed to pull them.")
			os.Exit(1)
		} else if strings.Contains(pollState, "gathering results") {
			// Wait 15 seconds between polling to see if the search completed
			fmt.Printf("Waiting 15 seconds to continue polling...\n\n")
			time.Sleep(15 * time.Second)
		} else {
			fmt.Printf("Unknown job state provided. %s\n", pollState)
			os.Exit(1)
		}
		loopCount++
	}

	/////////////// Pull the Messages that are retrieved
	sumoWebhookPullMessagesURL := searchJobsJSON.Link.HREF + "/messages?offset=0&limit=" + limitString
	httpBodyBytes = connectSumo("GET", sumoWebhookPullMessagesURL, []byte(""), "pullMessages.debug", sumoAccessID, sumoAccessKey, createDebugFiles)
	pullJobsJSON := sumoJSONParser.PullMessages(httpBodyBytes)
	outputMessage := "\nQuery: " + rawQuery + "\n\n"
	if len(pullJobsJSON.Messages) > 0 {
		fmt.Printf("Messages Returned: %d\n\n", len(pullJobsJSON.Messages))
		// Format the messaging correctly...
		for i := 0; i < len(pullJobsJSON.Messages); i++ {
			fmt.Printf("Source IP: %s, Destination IP: %s, Destination Port: %s, Action: %s\n", pullJobsJSON.Messages[i].Map.SourceIP, pullJobsJSON.Messages[i].Map.DestIP, pullJobsJSON.Messages[i].Map.DestPort, pullJobsJSON.Messages[i].Map.Action)
			outputMessage += "Source IP: " + pullJobsJSON.Messages[i].Map.SourceIP + " "
			outputMessage += "Destination IP: " + pullJobsJSON.Messages[i].Map.DestIP + " "
			outputMessage += "Destination Port: " + pullJobsJSON.Messages[i].Map.DestPort + " "
			outputMessage += "Action: " + pullJobsJSON.Messages[i].Map.Action + "\n"
			fmt.Printf("User: %s, App: %s\n", pullJobsJSON.Messages[i].Map.SourceUser, pullJobsJSON.Messages[i].Map.App)
			outputMessage += "User: " + pullJobsJSON.Messages[i].Map.SourceUser + " "
			outputMessage += "App: " + pullJobsJSON.Messages[i].Map.App + "\n"
			fmt.Printf("Raw Log\n")
			fmt.Printf("%s\n\n", pullJobsJSON.Messages[i].Map.Raw)
			outputMessage += "Raw Log" + "\n"
			outputMessage += pullJobsJSON.Messages[i].Map.Raw
			outputMessage += "\n\n"
		}
	} else {
		fmt.Printf("No messages were returned for the following query: \n")
		fmt.Printf("%s\n\n", rawQuery)
	}
	// Save the message that was output to the screen
	saveOutputFile(outputMessage, *OutputFilePtr)
	fmt.Printf("Saved the output of the messages pulled to the file: %s\n\n", *OutputFilePtr)

}
