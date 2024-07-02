package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
)

/*
go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/xorFile/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o xorFile.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o xorFile.exe -ldflags "-w -s" main.go

Summary: Created this tool to XOR a file with the following options specified. Used in the hackthebox sherlock challenge called lockpick
Usage:
  -action string
        Specify the action to take i.e. (encrypt, decrypt) (default "encrypt")
  -file string
        Specify the filename to xor
  -flag string
        Specify the key to use with the XOR function
  -output string
        Specify the output for the filename

*/

func xorInfo(info []byte, key string, state string) string {
	var err error
	if state == "decrypt" {
		info, err = base64.StdEncoding.DecodeString(string(info))
		if err != nil {
			fmt.Println("Error decoding:", err)
			os.Exit(0)
		}
	}
	xorInfo := make([]byte, len(info))
	keyLen := len(key)
	for i := 0; i < len(info); i++ {
		xorInfo[i] = info[i] ^ key[i%keyLen]
	}
	if state == "encrypt" {
		b64encoded := base64.StdEncoding.EncodeToString(xorInfo)
		return b64encoded
	}
	return string(xorInfo)
}

// CheckError checks for errors
func CheckError(reasonString string, err error, exitBool bool) {
	if err != nil && exitBool == true {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n\n", err)
		os.Exit(0)
	} else if err != nil && exitBool == false {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n", err)
		return
	}
}

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		CheckError("Unable to write to txt file", err, true)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

func main() {
	filenamePtr := flag.String("file", "", "Specify the filename to xor")
	outputnamePtr := flag.String("output", "", "Specify the output for the filename")
	actionPtr := flag.String("action", "encrypt", "Specify the action to take i.e. (encrypt, decrypt)")
	keyPtr := flag.String("key", "", "Specify the key to use with the XOR function")
	flag.Parse()

	if len(*filenamePtr) == 0 || len(*outputnamePtr) == 0 {
		fmt.Printf("\nYou must specify the file to XOR and the output file\n\n")
		flag.Usage()
		os.Exit(0)
	}

	file, err := os.Open(*filenamePtr)
	if err != nil {
		CheckError("Error opening file specified", err, true)
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		CheckError("Error gathering the file size", err, true)
	}
	fileSize := fileInfo.Size()

	// Read file content into a byte slice
	fileContent := make([]byte, fileSize)
	_, err = io.ReadFull(file, fileContent)
	if err != nil {
		CheckError("Error reading the file contents", err, true)
	}

	outputInfo := xorInfo([]byte(fileContent), *keyPtr, *actionPtr)
	SaveOutputFile(outputInfo, *outputnamePtr)
	fmt.Println("Output was saved to the following file: ", *outputnamePtr)

}
