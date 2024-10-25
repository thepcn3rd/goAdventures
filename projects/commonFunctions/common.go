package common

// This file contains functions that are used by multiple programs
// Place in the src folder under a folder called commonFunctions
// In the commonFunctions folder after creating and dropping common.go in it
// Execute "go mod init commonFunctions"
// Then the files in common functions can be referenced in import as:
//    cf "commonFunctions"

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ChatGPT helped find the below function
// Calculate a SHA256 hash of a string
func CalcSHA256Hash(message string) string {
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
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

// CreateDirectory creates a directory in the current working directory
// Example: cf.CreateDirectory("/keys") - Include the leading / on the directory
func CreateDirectory(createDir string) {
	currentDir, err := os.Getwd()
	CheckError("Unable to get the working directory", err, true)
	newDir := currentDir + createDir
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(newDir, os.ModePerm)
		CheckError("Unable to create directory "+createDir, err, true)
	}
}

// Check if the file exists, display the error message but does not exit
func FileExists(file string) bool {
	currentDir, err := os.Getwd()
	CheckError("Unable to get the working directory", err, false)
	//fmt.Println(currentDir)
	filePath := currentDir + file
	if _, err := os.Stat(filePath); err != nil {
		CheckError("Unable to find file "+filePath+"\n", err, false)
		return false
	}
	return true
}

// isFlagPassed checks if a flag is passed and parsed
func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
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

func CreateDefaultConfig(b64 string) {
	// The function requires a default config that is base64 encoded be send as a variable to this function
	// createDefaultConfig("...")
	// Check if the default config.json file exists if it does not exist create it
	if !(FileExists("/config.json")) {
		fmt.Println("Default config file does not exist...  Creating config.json.")
		b64configDefault := b64
		b64decodedBytes, err := base64.StdEncoding.DecodeString(b64configDefault)
		CheckError("Unable to decode the b64 of the config.json file", err, true)
		b64decodedString := string(b64decodedBytes)
		SaveOutputFile(b64decodedString, "config.json")
		fmt.Println("Created the config.json file, modify and rerun the app...")
		os.Exit(1)
	}
}

func CreateFileFromB64(b64 string, filename string) {
	// The function requires a default config that is base64 encoded be send as a variable to this function
	// createDefaultConfig("...")
	// Check if the default config.json file exists if it does not exist create it
	if !(FileExists("/" + filename)) {
		fmt.Println("Default config file does not exist...  Creating " + filename + ".")
		b64configDefault := b64
		b64decodedBytes, err := base64.StdEncoding.DecodeString(b64configDefault)
		CheckError("Unable to decode the b64", err, true)
		b64decodedString := string(b64decodedBytes)
		SaveOutputFile(b64decodedString, filename)
		fmt.Println("Created the " + filename + " file, modify if necessary and rerun the app...")
		os.Exit(1)
	}
}

func ZipCompression(srcFilename string, destFilename string) {
	// Create a new zip file
	zipFile, err := os.Create(destFilename)
	CheckError("Unable to create the Zip Destination Filename: "+destFilename, err, true)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Open the source file
	fileToZip, err := os.Open(srcFilename)
	CheckError("Unable to open the source filename: "+srcFilename, err, true)
	defer fileToZip.Close()

	// Get file information
	info, err := fileToZip.Stat()
	CheckError("Unable to gather the file information for the source file", err, true)

	// Create a zip header
	header, err := zip.FileInfoHeader(info)
	CheckError("Unable to create a zip header", err, true)

	// Specify the compression method
	header.Method = zip.Deflate

	// Set the name of the file inside the zip archive
	header.Name = filepath.Base(srcFilename)

	// Create a writer for the file in the zip archive
	writer, err := zipWriter.CreateHeader(header)
	CheckError("Unable to create the header", err, true)

	// Copy the file content to the zip archive
	_, err = io.Copy(writer, fileToZip)
	CheckError("Copy the contest of the zip archive", err, true)

	fmt.Println("File compressed successfully")
}
