package main

// Simple web request

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o webRequest.bin -ldflags "-w -s" webRequest.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o webRequest.exe -ldflags "-w -s" webRequest.go

// Working with the custom header, as of now it is not working...  I think it has to do with the split function

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type requestConf struct {
	reqMethod      string
	reqURL         string
	reqOutput      string
	reqOutputSave  string
	reqPath        string
	reqCookie      string
	reqContentType string
	reqHeaderKey   string // Add a unique header
	reqHeaderValue string
	reqUserAgent   string
	reqData        string
	reqInjection   string
	reqPayload     string
	// fmt.Print("\nset injection AND 1=1")
}

func defaultConf() *requestConf {
	d := requestConf{}
	d.reqMethod = "post"
	//d.reqURL = "http://10.13.38.20/book-trip.php"
	d.reqURL = "http://10.27.20.173/book-trip.php"
	d.reqOutput = "true" // Displayed on the screen
	d.reqOutputSave = "true"
	currentPath, _ := os.Getwd()
	d.reqPath = currentPath + "/output.txt"
	d.reqCookie = ""
	d.reqContentType = ""
	d.reqHeaderKey = ""
	d.reqHeaderValue = ""
	d.reqUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/114.0"
	d.reqData = "destination=xxx&adults=3&children=3"
	d.reqInjection = "AND 1 = 1"
	d.reqPayload = ""
	return &d
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
	fmt.Print("\nset method get (Change HTTP Method to GET)")
	fmt.Print("\nset method post (Change HTTP Method to POST, DELETE, PUT, HEAD)")
	fmt.Print("\nset url http://127.0.0.1:8000 (Change URL)")
	fmt.Print("\nset cookie PHPSESSID=blahblahblah")
	fmt.Print("\nset content-type application/json")
	fmt.Print("\nset user-agent webrequest (Default: Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/114.0)")
	fmt.Print("\nset header (Key:Value) <-- No space between the key and the value")
	fmt.Print("\nset data {'p': 'test'}  <-- Currently only supports JSON")
	fmt.Print("\nset output (Toggle the Output to the Screen)")
	fmt.Print("\nset save (Toggle if the Output is Saved to the Path")
	fmt.Print("\nset path /tmp/output.txt - (Set the path where to save the file)\n\n")
	fmt.Print("\nset injection AND 1 = 1")
}

func padString(str string, length int) string {
	for len(str) < length {
		str = str + " "
	}
	return str
}

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func main() {
	osDetected := runtime.GOOS
	requestConfig := defaultConf()
	var colorReset = "\033[0m"
	var colorRed = "\033[31m"
	var colorGreen = "\033[32m"
	var jsonData []byte
	// The below setup ignores the security of the certificate that is presented... (Self-signed and revoked)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	//client = &client{Transport: tr}
	fmt.Print("<-- Simple Browser -->\n")
	columnWidth := 30
	optionSelected := "beginLoop"
	for optionSelected != "e" && optionSelected != "E" && optionSelected != "exit" {
		// HTTP Method (GET, POST, HEAD, PUT, DELETE, ...)
		fmt.Print(padString("Method", columnWidth) + "-  " + padString(requestConfig.reqMethod, columnWidth) + "\n")
		fmt.Print(padString("URL", columnWidth) + "-  " + padString(requestConfig.reqURL, columnWidth) + "\n")
		fmt.Print(padString("Cookie", columnWidth) + "-  " + padString(requestConfig.reqCookie, columnWidth) + "\n")
		fmt.Print(padString("Content-Type", columnWidth) + "-  " + padString(requestConfig.reqContentType, columnWidth) + "\n")
		fmt.Print(padString("User-Agent", columnWidth) + "-  " + padString(requestConfig.reqUserAgent, columnWidth) + "\n")
		fmt.Print(padString("Custom Header", columnWidth) + "-  " + padString(requestConfig.reqHeaderKey+": "+requestConfig.reqHeaderValue, columnWidth) + "\n")
		fmt.Print(padString("Data", columnWidth) + "-  " + padString(requestConfig.reqData, columnWidth) + "\n")
		fmt.Print(padString("Injection to xxx: ", columnWidth) + "-  " + padString(requestConfig.reqInjection, columnWidth) + "\n")
		fmt.Print(padString("Payload: ", columnWidth) + "-  " + padString(requestConfig.reqPayload, columnWidth) + "\n")
		fmt.Print(padString("Output to Screen (Response)", columnWidth) + "-  " + padString(requestConfig.reqOutput, columnWidth) + "\n")
		fmt.Print(padString("Save to Path (Response)", columnWidth) + "-  " + padString(requestConfig.reqOutputSave, columnWidth) + "\n")
		fmt.Print(padString("Path (Output Path)", columnWidth) + "-  " + padString(requestConfig.reqPath, columnWidth) + "\n")

		fmt.Print("\nshow help\n")
		//fmt.Print("load config\n")
		//fmt.Print("save config\n")
		fmt.Print("run\n")
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
				requestConfig.reqMethod = setMethod
				continue
			} else {
				fmt.Printf(colorRed + "\nIncorrect method, choose from the following: (get, post)\n\n" + colorReset)
			}
		} else if strings.Contains(optionSelectedLower, "set url") {
			setURL := strings.TrimPrefix(optionSelectedLower, "set url ")
			requestConfig.reqURL = setURL
		} else if strings.Contains(optionSelectedLower, "set user-agent") {
			setUA := strings.TrimPrefix(optionSelectedLower, "set user-agent ")
			requestConfig.reqUserAgent = setUA
		} else if strings.Contains(optionSelectedLower, "set data") {
			setData := strings.TrimPrefix(optionSelectedLower, "set data ")
			requestConfig.reqData = setData
		} else if strings.Contains(optionSelected, "set cookie") {
			setCookie := strings.TrimPrefix(optionSelected, "set cookie ")
			requestConfig.reqCookie = setCookie
		} else if strings.Contains(optionSelected, "set output") {
			if requestConfig.reqOutput == "true" {
				requestConfig.reqOutput = "false"
			} else {
				requestConfig.reqOutput = "true"
			}
		} else if strings.Contains(optionSelected, "set save") {
			if requestConfig.reqOutputSave == "true" {
				requestConfig.reqOutputSave = "false"
			} else {
				requestConfig.reqOutputSave = "true"
			}
		} else if strings.Contains(optionSelected, "set content-type") {
			setContentType := strings.TrimPrefix(optionSelected, "set content-type ")
			requestConfig.reqContentType = setContentType
		} else if strings.Contains(optionSelected, "set header") {
			setCustomHeader := strings.TrimPrefix(optionSelected, "set header ")
			setCustomHeaderItems := strings.Split(setCustomHeader, ":")
			//fmt.Print(setCustomHeaderItems[1])
			//fmt.Print("\n")
			requestConfig.reqHeaderKey = setCustomHeaderItems[0]
			requestConfig.reqHeaderValue = setCustomHeaderItems[1]
		} else if strings.Contains(optionSelected, "set injection") {
			setInjection := strings.TrimPrefix(optionSelected, "set injection ")
			requestConfig.reqInjection = setInjection
		} else if optionSelectedLower == "run" {
			var req *http.Request
			var err error
			if requestConfig.reqData == "" {
				req, err = http.NewRequest(strings.ToUpper(requestConfig.reqMethod), requestConfig.reqURL, nil)
				checkError("Unable to build new request", err)
			} else {
				if strings.Contains(requestConfig.reqData, "{") {
					// Need to extend this in the event it is not JSON data...
					jsonData = []byte(requestConfig.reqData)
					req, err = http.NewRequest(strings.ToUpper(requestConfig.reqMethod), requestConfig.reqURL, bytes.NewBuffer(jsonData))
					// Set the content-type header to be JSON
					requestConfig.reqContentType = "application/json"
				} else {
					requestConfig.reqPayload = strings.Replace(requestConfig.reqData, "xxx", requestConfig.reqInjection, -1)

					fmt.Println(requestConfig.reqPayload)
					asciiData := []byte(requestConfig.reqPayload)
					req, err = http.NewRequest(strings.ToUpper(requestConfig.reqMethod), requestConfig.reqURL, bytes.NewBuffer(asciiData))
					// Set the content-type header to be ASCII
					requestConfig.reqContentType = "application/x-www-form-urlencoded"
					//req.Header.Set("Content-Type", "text/plain; charset=us-ascii")
				}
			}
			if requestConfig.reqCookie != "" {
				items := strings.Split(requestConfig.reqCookie, "=")

				req.AddCookie(&http.Cookie{
					Name: items[0], Value: items[1], MaxAge: 60,
				})
			}
			if requestConfig.reqContentType != "" {
				req.Header.Add("Content-Type", requestConfig.reqContentType)
			}
			if requestConfig.reqHeaderKey != "" {
				//fmt.Print(requestConfig.reqHeaderKey)
				if requestConfig.reqHeaderKey == "Host" {
					req.Host = requestConfig.reqHeaderValue
				} else {
					req.Header.Add(requestConfig.reqHeaderKey, requestConfig.reqHeaderValue)
				}
			}
			if requestConfig.reqUserAgent != "" {
				req.Header.Set("User-Agent", requestConfig.reqUserAgent)
			}
			resp, err := client.Do(req)
			checkError("Unable to request URL specified", err)
			respBody, _ := io.ReadAll(resp.Body)
			fmt.Println(resp.Header)
			// Output to console
			if requestConfig.reqOutput == "true" {
				fmt.Printf(string(respBody))
			}
			if requestConfig.reqOutputSave == "true" {
				// Output to File - Overwrites if file exists...
				f, err := os.Create(requestConfig.reqPath)
				checkError("Unable create file to save output", err)
				defer f.Close()
				fmt.Printf(colorGreen+"\nSaved the file to %s\n\n"+colorReset, requestConfig.reqPath)
				f.Write(respBody)
				f.Close()
			}
		}
	}
}
