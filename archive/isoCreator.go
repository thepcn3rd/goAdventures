package main

/*
Summary:
This program allows you to create an ISO specifying files that are in the isofiles directory.  Then will output to the current working directory the ISO created.
Built to be able to emulate how miscreants create ISO files and provide them in phishing attacks.

MITRE ATTACK: T1566.001 and T1566.002 - Use of ISO files as attachments and in links to be downloaded...
Vulnerablity: Automounting of an ISO in modern operating systems


Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/isoCreator"

Create the directories of src, bin, and pkg

Need the following package dependency
go get github.com/kdomanski/iso9660

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o createCerts.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o createCerts.exe -ldflags "-w -s" main.go

*/

import (
	"archive/zip"
	cf "commonFunctions"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kdomanski/iso9660"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "Puff the magic dragon!"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var flags arrayFlags

func main() {
	writer, err := iso9660.NewWriter()
	cf.CheckError("Failed to create an ISO Writer", err, true)
	defer writer.Cleanup()

	flag.Var(&flags, "f", "Specify the files to place in the iso, more than one can be specified")
	outputFilename := flag.String("output", "output.iso", "Name of the ISO Output File")
	volumeName := flag.String("volume", "default", "Name of the Volume on the ISO")
	zipOption := flag.String("z", "", "Compress the file as a zip with the provided filename")
	flag.Parse()

	if cf.IsFlagPassed("f") == false {
		cf.CreateDirectory("/isofiles")
		fmt.Println("\nWelcome to the isoCreator!")
		fmt.Println("This allows you to add files to an iso by specifying the files at the command line")
		fmt.Println("The files need to exist in the directory called isofiles with the structure you expect")
		fmt.Printf("Example command: isoCreator -f test.txt -f test2.txt\n\n")
		flag.Usage()
		os.Exit(0)
	}

	for fileNumb := range flags {
		//List the files in the list for debugging
		//fmt.Println(flags[fileNumb])
		filePath := "isofiles/" + flags[fileNumb]
		f, err := os.Open(filePath)
		cf.CheckError("Unable to open the file specified: "+filePath, err, true)
		defer f.Close()
		err = writer.AddFile(f, flags[fileNumb])
		cf.CheckError("Unable to add the file to the ISO writer:"+filePath, err, true)
		f.Close()
	}

	// Create the ISO
	outputISO, err := os.OpenFile(*outputFilename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	cf.CheckError("Failed to Create the ISO File: "+*outputFilename, err, true)

	err = writer.WriteTo(outputISO, *volumeName)
	cf.CheckError("Failes to Write to the ISO File: "+*outputFilename, err, true)

	err = outputISO.Close()
	cf.CheckError("Failed to Close the ISO File", err, true)

	// Add the ISO to a zip file if the z flag is set
	if cf.IsFlagPassed("z") == true {
		archive, err := os.Create(*zipOption)
		cf.CheckError("Unable to create the zip file for the ISO", err, true)
		defer archive.Close()
		zipWriter := zip.NewWriter(archive)
		isoFile, err := os.Open(*outputFilename)
		cf.CheckError("Unable to read the new ISO file to compress", err, true)
		defer isoFile.Close()
		w, err := zipWriter.Create(*outputFilename)
		cf.CheckError("Unable to write the new ISO file into the zip file", err, true)
		if _, err := io.Copy(w, isoFile); err != nil {
			fmt.Println("Unable to create the zip file")
			os.Exit(0)
		}
		isoFile.Close()
		zipWriter.Close()
	}
}
