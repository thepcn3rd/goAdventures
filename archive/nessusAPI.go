package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"msTeamsSendMessage"
	"nessusJSONParser"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cf "commonFunctions"
)

// Setup the following for the application
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/nessusAPI"

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nessusAPI.bin -ldflags "-w -s" main.go

// To cross compile windows (Not tested...)
// GOOS=windows GOARCH=amd64 go build -o nessusAPI.exe -ldflags "-w -s" main.go

Directory setup and file placement

// nessusAPI/ (Create Directory)
// - main.go (Renamed to nessusAPI.go to live in this repo for the moment)
// - bin/
// - pkg/
// - src/
// - - nessusJSONParser/
// - - - nessusParser.go
// - - msTeamsSendMessage/
// - - - sendMessage.go

// Tasks
// - Build a Nessus API connection
// - Build sqlite database to decrease the number of API calls that are needed (Hopefully...)
// - Build the ability to detect if a scan is going to run in the next 4 hours and post in teams
// - Build the ability to detect if a scan has stopped in the last 4 hours and post in teams
// - If a scan has stopped in the last 4 hours summarize the results of the scan (Critical, High, Mod)
// - (Metrics - Seperate, Go back 3 months and summarize the Critical, High, Mod for the last 3 months; Jan 1 to Jan 31, Feb 1 to Feb 28, Mar 1 to Current Date..., etc.)
// - - Conduct the above based on environment - MB CAT1, MB CAT2, CWS Ext, MB Ext, and Overall...

// Move API Information into another txt file that is read by the scanner...
// Interesting, could encrypt the text file, then decrypt when needed

// Move the teams messages into the commonFunctions


References:
https://developer.tenable.com/reference/scans-list


*/

type Configuration struct {
	Auth AuthStruct `json:"auth"`
	URL  URLStruct  `json:"url"`
}

type AuthStruct struct {
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	EncryptedAccessKey string `json:"encryptedAccessKey"`
	EncryptedSecretKey string `json:"encryptedSecretKey"`
}

type URLStruct struct {
	TeamsURL            string `json:"teamsURL"`
	EncryptedTeamsURL   string `json:"encryptedTeamsURL"`
	SecretsURL          string `json:"secretsURL"`
	UserAgent           string `json:"userAgent"`
	XAbilities          string `json:"xAbilities"`
	EncryptedSecretsURL string `json:"encryptedSecretsURL"`
	EncryptedUserAgent  string `json:"encryptedUserAgent"`
	EncryptedXAbilities string `json:"encryptedXAbilities"`
}

func todaysScans(scanJSON *nessusJSONParser.ScanStruct, webhookURL string, totalResults int) {
	currentDate := time.Now()
	currentDateFormatted := fmt.Sprintf("%d%02d%02d", currentDate.Year(), currentDate.Month(), currentDate.Day())
	//fmt.Println("Gathering information for the upcoming scans for the following date: " + currentDateFormatted)
	for i := 0; i < totalResults; i++ {
		if strings.Contains(scanJSON.Scans[i].Starttime, currentDateFormatted) {
			//fmt.Println(scanListJSON.Scans[i].Name)
			nameScan := scanJSON.Scans[i].Name
			scheduledTime := scanJSON.Scans[i].Starttime
			//timezone := scanJSON.Scans[i].Timezone

			teamsMessage := "\n**" + nameScan + "**\n\n"

			// Format the scheduledTime Variable
			//var timeOfDay string
			//timeOfDay = "am"
			scheduledYear := scheduledTime[0:4]
			scheduledMonth := scheduledTime[4:6]
			scheduledDay := scheduledTime[6:8]
			scheduledHour := scheduledTime[9:11]
			scheduledMinute := scheduledTime[11:13]
			//fmt.Println(scheduledYear)
			//fmt.Println(scheduledMonth)
			//fmt.Println(scheduledDay)
			//fmt.Println(scheduledHour)
			//fmt.Println(scheduledMinute)
			//intScheduledHour, _ := strconv.Atoi(scheduledHour)
			//var easternScheduledHour int
			//var mountainScheduledHour int
			//if intScheduledHour > 12 {
			//	timeOfDay = "pm"
			//	easternScheduledHour, _ = strconv.Atoi(scheduledHour)
			//	easternScheduledHour = easternScheduledHour - 10
			//	mountainScheduledHour, _ = strconv.Atoi(scheduledHour)
			//	mountainScheduledHour = mountainScheduledHour - 12
			//}
			stringDateTime := scheduledMonth + "/" + scheduledDay + "/" + scheduledYear + " " + scheduledHour + ":" + scheduledMinute
			//easternDateTime := scheduledMonth + "/" + scheduledDay + "/" + scheduledYear + " " + strconv.Itoa(easternScheduledHour) + ":" + scheduledMinute + timeOfDay
			//fmt.Printf(scheduledTime[11:13])
			//time.Parse("2006-01-02", dateString)
			///dateStringConverted, _ := time.Parse(time.RFC822Z, scheduledTime)
			teamsMessage += "Scheduled Time: " + stringDateTime + "\n\n"
			//teamsMessage += "Scheduled Time (Mountain): " + mountainDateTime + "\n\n"
			//teamsMessage += "Scheduled Time (Eastern): " + easternDateTime + "\n\n"
			//teamsMessage += "Timezone: " + timezone + "\n\n"
			if !strings.Contains(strings.ToLower(nameScan), "adhoc") {
				//fmt.Println(teamsMessage)
				msTeamsSendMessage.SendMessage(teamsMessage, webhookURL)
				// Wait 3 seconds between teams messages
				time.Sleep(3 * time.Second)
			}
		}
	}
}

/*
func saveJSONDebugFile(scanJSON *nessusJSONParser.ScanStruct, fileName string) {
	outFile, err := os.Create(fileName)
	checkError("Unable to create JSON debug file", err)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	jsonBuffer, err := json.MarshalIndent(scanJSON, "", "")
	_, err = w.WriteString(string(jsonBuffer))
	outFile.Sync()
	w.Flush()
	outFile.Close()
}
*/

func saveDebugFile(fileBytes []byte, fileName string) {
	cf.CreateDirectory("/debug")
	fullPath := "debug/" + fileName
	outFile, err := os.Create(fullPath)
	cf.CheckError("Unable to create debug txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	_, err = w.WriteString(string(fileBytes))
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func yesterdayResults(scanJSON *nessusJSONParser.ScanStruct, webhookURL string, totalResults int, apiHeader string) {
	// Yesterdays date and formatted date
	//yDate := time.Now().AddDate(0, 0, -1)
	//yDateFormatted := fmt.Sprintf("%d%02d%02d", yDate.Year(), yDate.Month(), yDate.Day())
	//epochDate := time.Now().Unix()
	// Evaluate the scans that have completed in the last 24 hours and display in Teams Channel
	/////// Changing the monitoring to be every hour and then will post to the channel hourly
	//epochLastHour := time.Now().Add(time.Duration(-2 * time.Hour)).Unix()
	//epochYesterday := time.Now().AddDate(0, 0, -1).Unix()
	//fmt.Println(epochDate)
	//fmt.Println(epochYesterday)
	//fmt.Println("Gathering scan results for the date of: " + yDateFormatted)

	//fmt.Println(len(scanJSON.Scans))
	for i := 0; i < len(scanJSON.Scans); i++ {
		//if strings.Contains(scanJSON.Scans[i].Starttime, yDateFormatted) {
		//fmt.Println(scanListJSON.Scans[i].Name)
		nameScan := scanJSON.Scans[i].Name

		// Format the scheduledTime Variable
		//scheduledTime := scanJSON.Scans[i].Starttime
		// If progress equals 100
		progress := scanJSON.Scans[i].Progress
		//progressStr := fmt.Sprintf("%d", progress)
		// If status equals completed
		status := scanJSON.Scans[i].Status
		scanID := scanJSON.Scans[i].ID
		scanIDStr := fmt.Sprintf("%d", scanID)

		if progress == 100 && status == "completed" {
			teamsMessage := "\n**" + nameScan + "**\n\n"
			//teamsMessage += "Progress: " + progressStr + "\n\n"
			teamsMessage += "Status: Scans Completed in Last 2 Hours\n\n"
			//teamsMessage += "Start Time: " + scheduledTime + "\n\n"
			teamsMessage += "Scan ID: " + scanIDStr + "\n\n"
			scanMessage := pullScanDetails(scanIDStr, apiHeader)
			teamsMessage += scanMessage
			//if strings.Contains(endTime, yDateFormatted) {
			if !strings.Contains(strings.ToLower(nameScan), "adhoc") {
				//fmt.Println("End time: " + strconv.Itoa(endTime))
				//**fmt.Println(teamsMessage)
				msTeamsSendMessage.SendMessage(teamsMessage, webhookURL)
				// Wait 3 seconds between teams messages
				time.Sleep(3 * time.Second)
			}
		}
		//}
	}
}

func pullScanDetails(scanID string, apiHeader string) string {
	nessusScanDetailsURL := "https://cloud.tenable.com/scans/" + scanID
	var httpClient http.Client
	var httpRequest *http.Request
	var err error

	httpRequest, err = http.NewRequest("GET", nessusScanDetailsURL, nil)
	cf.CheckError("Unable to build http request for the Nessus API", err, true)
	httpRequest.Header.Add("accept", "application/json")
	// Add the API keys as a header
	httpRequest.Header.Add("x-apikeys", apiHeader)

	// Receive response through the httpClient connection
	httpResponse, err := httpClient.Do(httpRequest)
	cf.CheckError("Unable to pull http response from the Nessus API", err, true)

	// Verify we receive a 200 response and if not exit the program...
	if httpResponse.Status != "200 OK" {
		fmt.Println("Response Status: " + httpResponse.Status)
		os.Exit(0)
	}

	httpBodyBytes, _ := io.ReadAll(httpResponse.Body)
	epochDate := time.Now().Unix()
	epochDateStr := fmt.Sprintf("%d", epochDate)
	saveDebugFilename := "ScanDetailsJSON.debug." + scanID + epochDateStr
	saveDebugFile(httpBodyBytes, saveDebugFilename)
	//fmt.Println("\n" + string(httpBodyBytes) + "\n")

	// Send the httpResponse.Body to the struct to have the JSON parsed
	//scanDetailsJSON := nessusJSONParser.ScanDetailsParser(httpResponse.Body)
	// Modified to take the byte stream and conduct the match to the struct...
	scanDetailsJSON := nessusJSONParser.ScanDetailsParser(httpBodyBytes)
	//fmt.Printf("Length of Hosts Slice: %d", len(scanDetailsJSON.Hosts)) // Always comes out to be the length of the subnet
	var informationalFindings int = 0
	var lowSevFindings int
	var medSevFindings int
	var highSevFindings int
	var criticalSevFindings int
	var hostCount int
	for h := 0; h < len(scanDetailsJSON.Hosts); h++ {
		if scanDetailsJSON.Hosts[h].Severitycount.Item[0].Count > 2 { // Observed that each IP Address has 2 informational findings
			//fmt.Printf("%d\n", scanDetailsJSON.Hosts[h].Severitycount.Item[0].Count)
			informationalFindings = informationalFindings + scanDetailsJSON.Hosts[h].Severitycount.Item[0].Count
			lowSevFindings = lowSevFindings + scanDetailsJSON.Hosts[h].Severitycount.Item[1].Count
			medSevFindings = medSevFindings + scanDetailsJSON.Hosts[h].Severitycount.Item[2].Count
			highSevFindings = highSevFindings + scanDetailsJSON.Hosts[h].Severitycount.Item[3].Count
			criticalSevFindings = criticalSevFindings + scanDetailsJSON.Hosts[h].Severitycount.Item[4].Count
			hostCount++ // Count the number of hosts
		}
	}
	//fmt.Printf("\nTotal Informational Findings: %d\n", informationalFindings)
	//fmt.Printf("Total Low Severity Findings: %d\n", lowSevFindings)
	//fmt.Printf("Total Medium Severity Findings: %d\n", medSevFindings)
	//fmt.Printf("Total High Sev Findings: %d\n", highSevFindings)
	//fmt.Printf("Total Critical Findings: %d\n", criticalSevFindings)
	//fmt.Printf("Host Count: %d\n", hostCount)

	scanMessage := "**Total Findings**\n\n"
	scanMessage += "Critical Findings: " + strconv.Itoa(criticalSevFindings) + "\n\n"
	scanMessage += "High Sev Findings: " + strconv.Itoa(highSevFindings) + "\n\n"
	scanMessage += "Medium Sev Findings: " + strconv.Itoa(medSevFindings) + "\n\n"
	scanMessage += "Low Sev Findings: " + strconv.Itoa(lowSevFindings) + "\n\n"
	scanMessage += "Informational Items: " + strconv.Itoa(criticalSevFindings) + "\n\n"
	scanMessage += "**IP Ranges Scanned**\n\n"
	scanMessage += "Targets: " + scanDetailsJSON.Info.Targets + "\n\n\n"

	//var endTime int
	//endTime = scanDetailsJSON.Info.Scan_end

	//teamsMessage := "\n**" + nameScan + "**\n\n"
	//teamsMessage += "Progress: " + progressStr + "\n\n"
	//teamsMessage += "Status: " + status + "\n\n"
	//teamsMessage += "Start Time: " + scheduledTime + "\n\n"
	//teamsMessage += "Scan ID: " + scanIDStr + "\n\n\n"
	//fmt.Println(scanMessage)
	//msTeamsSendMessage.SendMessage(scanMessage, teamsWH)
	return scanMessage
	// Save the JSON to a file for debugging
	//saveJSONDebugFile(scanDetailsJSON, "scandetails.debug.json")

}

func createConfigFile() {
	configFile := `{
        "auth": {
                "accessKey": "",
                "secretKey": "",
                "encryptedAccessKey": "",
                "encryptedSecretKey": ""
        },
        "url": {
                "teamsURL": "",
                "encryptedTeamsURL": "",
                "secretsURL": "",
                "userAgent": "",
                "xAbilities": "",
                "encryptedSecretsURL": "",
                "encryptedUserAgent": "",
                "encryptedXAbilities": ""
        }
	}`
	cf.SaveOutputFile(configFile, "config.json")
	fmt.Println("\nNew config.json file was created due to it missing...")
	fmt.Println("Edit the configuration file and run the script again to use the config file")
	os.Exit(0)
}

func main() {
	var httpClient http.Client
	var httpRequest *http.Request
	var err error

	// Creating flags to determine if the client is being used for scheduled scans or for results of scans that have finished
	ListScheduledPtr := flag.String("list", "No", "List scheduled scans for the day and send results to Teams Channel")
	MonitorFinishedPtr := flag.String("monitor", "No", "Monitor completed scans and send results to Teams Channel")
	ConfigPtr := flag.String("config", "config.json", "Location of the configuration file")
	flag.Parse()
	if !cf.IsFlagPassed("list") && !cf.IsFlagPassed("monitor") {
		fmt.Println("You need to select whether this script lists or monitors the scans")
		flag.Usage()
		os.Exit(0)
	}

	// Create the config file if it does not exist to contain the connectors
	configFileExists := cf.FileExists("/config.json")
	if configFileExists == false {
		createConfigFile()
		os.Exit(0)
	}

	fmt.Printf("Reading the config file: %s\n", *ConfigPtr)
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Verify the config file contains the encryptedSecrestsURL and others to then pull the secret for the others
	secretsURL := config.URL.SecretsURL
	userAgent := config.URL.UserAgent
	xAbilities := config.URL.XAbilities
	encryptedSecretsURL := config.URL.EncryptedSecretsURL
	encryptedUserAgent := config.URL.EncryptedUserAgent
	encryptedXAbilities := config.URL.EncryptedXAbilities
	var decryptedSecretsURL string
	var decryptedUserAgent string
	var decryptedXAbilities string
	//var pulledKey string
	functionKey := "dbGtQcnSRomEc1cW4fXoCgxkjeyISPxG"
	if (secretsURL != "" && encryptedSecretsURL == "") || (userAgent != "" && encryptedUserAgent == "") || (xAbilities != "" && encryptedXAbilities == "") {
		fmt.Println("Key URL, User Agent or X Abilities is not encrypted in the config file")
		// Had to embed the encryption key for the Key URL
		// This is a minor feature to protect the quick gathering of the Lambda URL that is used to pull the key
		if secretsURL != "" {
			b64SecretsURL, err := cf.EncryptString([]byte(functionKey), secretsURL)
			cf.CheckError("Unable to encrypt the key URL", err, true)
			fmt.Printf("Encrypted Secrets URL: %s\n", b64SecretsURL)
		}
		if userAgent != "" {
			encryptedUserAgent, err := cf.EncryptString([]byte(functionKey), userAgent)
			cf.CheckError("Unable to encrypt the User Agent", err, true)
			fmt.Printf("Encrypted User Agent: %s\n", encryptedUserAgent)
		}
		if xAbilities != "" {
			encryptedXAbilities, err := cf.EncryptString([]byte(functionKey), xAbilities)
			cf.CheckError("Unable to encrypt the X Abilities", err, true)
			fmt.Printf("Encrypted X Abilities: %s\n", encryptedXAbilities)
		}
		os.Exit(0)
	} else if secretsURL != "" && userAgent != "" && xAbilities != "" {
		fmt.Println("The unencrypted Key URL, user agent or xAbilities needs to be removed from the config file")
		os.Exit(0)
	} else {
		decryptedSecretsURL, err = cf.DecryptString([]byte(functionKey), encryptedSecretsURL)
		cf.CheckError("Unable to decrypt the Secrets URL", err, true)
		decryptedUserAgent, err = cf.DecryptString([]byte(functionKey), encryptedUserAgent)
		cf.CheckError("Unable to decrypt the User Agent", err, true)
		decryptedXAbilities, err = cf.DecryptString([]byte(functionKey), encryptedXAbilities)
		cf.CheckError("Unable to decrypt the X Abilities", err, true)
		//fmt.Println("Decrypted Secrets URL: " + decryptedSecretsURL)
		//fmt.Println("Decrypted User Agent: " + decryptedUserAgent)
		//fmt.Println("Decrypted X Abilities: " + decryptedXAbilities)

	}

	pulledKey := cf.PullKey(decryptedSecretsURL, decryptedUserAgent, decryptedXAbilities)
	//fmt.Println("Pulled Key: " + pulledKey)

	accessKey := config.Auth.AccessKey
	secretKey := config.Auth.SecretKey
	teamsURL := config.URL.TeamsURL
	encryptedAccessKey := config.Auth.EncryptedAccessKey
	encryptedSecretKey := config.Auth.EncryptedSecretKey
	encryptedTeamsURL := config.URL.EncryptedTeamsURL
	if encryptedAccessKey == "" || encryptedSecretKey == "" || encryptedTeamsURL == "" {
		fmt.Println("Access Key, Secret Key or Teams URL are not encrypted in the config file")
		b64AccessKey, err := cf.EncryptString([]byte(pulledKey), accessKey)
		cf.CheckError("Unable to encrypt the access key", err, true)
		fmt.Printf("Encrypted Access Key: %s\n", b64AccessKey)
		b64SecretKey, err := cf.EncryptString([]byte(pulledKey), secretKey)
		cf.CheckError("Unable to encrypt the access ID", err, true)
		fmt.Printf("Encrypted Access ID: %s\n", b64SecretKey)
		b64TeamsURL, err := cf.EncryptString([]byte(pulledKey), teamsURL)
		cf.CheckError("Unable to encrypt the access ID", err, true)
		fmt.Printf("Encrypted Teams URL: %s\n", b64TeamsURL)
		os.Exit(0)
	}

	if len(encryptedAccessKey) > 1 && accessKey != "" {
		fmt.Println("Remove the Unencrypted Key from the config file")
		os.Exit(0)
	}

	if len(encryptedSecretKey) > 1 && secretKey != "" {
		fmt.Println("Remove the Unencrypted ID from the config file")
		os.Exit(0)
	}

	if len(encryptedTeamsURL) > 1 && teamsURL != "" {
		fmt.Println("Remove the Unencrypted Teams URLfrom the config file")
		os.Exit(0)
	}

	// See if the values in the config file exist for encryption...
	if (config.Auth.AccessKey != "" && config.Auth.EncryptedAccessKey != "") || (config.Auth.SecretKey != "" && config.Auth.EncryptedSecretKey != "") {
		fmt.Println("The configuration file is missing the access key or secret key")
		fmt.Println("Edit the configuration file and them to the script")
		os.Exit(0)
	}

	// Decrypt the Teams URL, Access Key and the Secret Key for access
	var decryptedTeamsURL string
	var decryptedAccessKey string
	var decryptedSecretKey string
	if len(encryptedAccessKey) > 1 && len(encryptedSecretKey) > 1 && len(encryptedTeamsURL) > 1 {
		decryptedTeamsURL, err = cf.DecryptString([]byte(pulledKey), encryptedTeamsURL)
		cf.CheckError("Unable to decrypt the Teams URL", err, true)
		decryptedAccessKey, err = cf.DecryptString([]byte(pulledKey), encryptedAccessKey)
		cf.CheckError("Unable to decrypt the access key", err, true)
		decryptedSecretKey, err = cf.DecryptString([]byte(pulledKey), encryptedSecretKey)
		cf.CheckError("Unable to decrypt the secret key", err, true)
		//fmt.Println("Decrypted Teams URL: " + decryptedTeamsURL)
		//fmt.Println("Decrypted Access Key: " + decryptedAccessKey)
		//fmt.Println("Decrypted Secret Key: " + decryptedSecretKey)

	}

	// Microsoft Teams Webhook
	teamsWebhookURL := decryptedTeamsURL
	//fmt.Println(teamsWebhookURL)

	// Returns the scans with last_modification_date this limits the results that have run since the specified time
	epochLastHour := time.Now().Add(time.Duration(-2 * time.Hour)).Unix()
	epochDateStr := fmt.Sprintf("%d", epochLastHour)
	nessusBaseURL := "https://cloud.tenable.com/scans?last_modification_date=" + epochDateStr

	// API Information for Tenable API Connection
	nessusAccessKeyTXT := decryptedAccessKey
	nessusSecretKeyTXT := decryptedSecretKey
	nessusAPIHeader := "accessKey=" + nessusAccessKeyTXT + "; secretKey=" + nessusSecretKeyTXT
	//fmt.Println(nessusAPIHeader)

	httpRequest, err = http.NewRequest("GET", nessusBaseURL, nil)
	cf.CheckError("Unable to build http request for the Nessus API", err, true)
	httpRequest.Header.Add("accept", "application/json")
	// Add the API keys as a header
	httpRequest.Header.Add("x-apikeys", nessusAPIHeader)

	// Receive response through the httpClient connection
	httpResponse, err := httpClient.Do(httpRequest)
	cf.CheckError("Unable to pull http response from the Nessus API", err, true)

	// Verify we receive a 200 response and if not exit the program...
	if httpResponse.Status != "200 OK" {
		fmt.Println("Response Status: " + httpResponse.Status)
		os.Exit(0)
	}

	httpBodyBytes, err := io.ReadAll(httpResponse.Body)
	cf.CheckError("Unable to read the http response body to create debug file", err, true)
	saveDebugFile(httpBodyBytes, "ScanListJSON.debug")

	// Send the httpResponse.Body to the struct to have the JSON parsed
	//scanListJSON := nessusJSONParser.ScanDetailsParser(httpResponse.Body)
	scanListJSON := nessusJSONParser.ScanParser(httpBodyBytes)

	// Multiple scans will be returned based on the JSON
	// Total CVEs Returned Based on the Criteria
	totalResults := len(scanListJSON.Scans)
	//fmt.Printf("Total Results Returned for Scans: %d\n", totalResults)

	if strings.ToLower(*MonitorFinishedPtr) == "yes" || strings.ToLower(*MonitorFinishedPtr) == "y" {
		// Send a teams message for yesterdays scans that were scheduled
		yesterdayResults(scanListJSON, teamsWebhookURL, totalResults, nessusAPIHeader)
	} else if strings.ToLower(*ListScheduledPtr) == "yes" || strings.ToLower(*ListScheduledPtr) == "y" {
		// Send a teams message for upcoming scans
		todaysScans(scanListJSON, teamsWebhookURL, totalResults)
	} else {
		fmt.Println("You need to select whether this script lists or monitors the scans")
		flag.Usage()
		os.Exit(0)
	}

}
