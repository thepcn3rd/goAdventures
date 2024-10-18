package main

import (
	"fmt"
	"net/http"
	"nvdJSONParser"
	"os"
)

// Setup the environment variables for go
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/nistCVE"

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o nistCVE.bin -ldflags "-w -s" main.go

// To cross compile windows (Not tested...)
// GOOS=windows GOARCH=amd64 go build -o nistCVE.exe -ldflags "-w -s" main.go

Directory setup and file placement

// nistCVE/ (Create Directory)
// - main.go (Renamed to nistCVE.go to live in this repo for the moment)
// - bin/
// - pkg/
// - src/
// - - nvdJSONParser/
// - - - nvdJSONParser.go

Future additions
1. Add msteams webhook functionality
2. Add function to search based on timeframe with keywords
3. Split the project, teams integration and stand-alone NVD query tool

*/

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func main() {
	var httpClient http.Client
	var httpRequest *http.Request
	var err error
	nvdBaseURL := "http://services.nvd.nist.gov/rest/json/cves/2.0?cveId=CVE-2019-18935"

	// Build the request for the Base URL above...
	httpRequest, err = http.NewRequest("GET", nvdBaseURL, nil)
	checkError("Unable to build http request for the NIST NVD API", err)

	// Receive response through the httpClient connection
	httpResponse, err := httpClient.Do(httpRequest)
	checkError("Unable to pull http response from NIST NVD API", err)

	// Verify we receive a 200 response and if not exit the program...
	if httpResponse.Status != "200 OK" {
		fmt.Println("Response Status: " + httpResponse.Status)
		os.Exit(0)
	}

	// Pass the httpResponse.Body to the Parser to Place it into a struct
	// Left the below lines for debugging
	//responseBody, err := io.ReadAll(httpResponse.Body)
	//fmt.Print(string(responseBody))
	nvdJSON := nvdJSONParser.NVDParser(httpResponse.Body)
	fmt.Println(nvdJSON.Vulns[0].CVE.ID)

}
