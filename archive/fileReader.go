package main

/*

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o fileReader.bin -ldflags "-w -s" fileReader.go

// To cross compile windows (Not tested...)
// GOOS=windows GOARCH=amd64 go build -o fileReader.exe -ldflags "-w -s" fileReader.go


To do list:
- Does not function currently
- Modify to detect the OS architecture and which files to search for...
- Research common files on linux to evaluate (Research a word list?)


*/

import (
	"bufio"
	"fmt"
	"os"
)

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func getLinFilenames() []string {
	return []string{
		`/etc/passwd`,
		`/etc/shadow`,
	}
}

func getWinFilenames() []string {
	return []string{
		`C:\unattend.xml`,
		`C:\Windows\Panther\Unattend.xml`,
		`C:\Windows\Panther\Unattend\Unattend.xml`,
		`C:\Windows\system32\sysprep.inf`,
		`C:\Windows\system32\sysprep\sysprep.xml`,
	}
}

func displayFiles(filePath string) {
	fileInfo, err := os.Open(filePath)
	checkError("Unable to open file", err)
	defer fileInfo.Close()
	scanner := bufio.NewScanner(fileInfo)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func main() {
	filenames := getWinFilenames()
	for _, file := range filenames {
		_, err := os.Stat(file)
		if err != nil {
			checkError("File does not exist", err)
		} else {
			displayFiles(file)
		}
	}
}
