package commonfunctions

// This file contains functions that are used by multiple programs
// Place in the src folder under a folder called commonFunctions
// In the commonFunctions folder after creating and dropping common.go in it
// Execute "go mod init commonFunctions"
// Then the files in common functions can be referenced in import as:
//    cf "commonFunctions"

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// ChatGPT helped find the below function
// Calculate a SHA256 hash of a string
func CalcSHA256Hash(message string) string {
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
}

// CheckError checks for errors
func CheckError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

// CreateDirectory creates a directory in the current working directory
func CreateDirectory(createDir string) {
	currentDir, err := os.Getwd()
	CheckError("Unable to get the working directory", err)
	newDir := currentDir + createDir
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(newDir, os.ModePerm)
		CheckError("Unable to create directory "+createDir, err)
	}
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

func SetupPermissionsProcess(userInput string) {
	var userInfo *user.User
	var err error
	// Modify the process the webserver runs as to the user and group specified...
	// The permissions on uploading files and accessing keys need to be set appropriately for the user and group...

	// Who is the current user logged in
	currentUserInfo, err := user.Current()
	CheckError("Unable to get current user", err)
	//fmt.Println("Current User: " + currentUserInfo.Username)
	//fmt.Println("Current UID: " + currentUserInfo.Uid)
	//fmt.Println("Current GID: " + currentUserInfo.Gid)

	// Change the UID and the GID if the current user is root
	if currentUserInfo.Uid == "0" {
		if IsFlagPassed("user") {
			userInfo, err = user.Lookup(userInput)
			CheckError("Unable to change user", err)
			if userInfo.Uid == "0" {
				fmt.Printf("DANGER: Running the server as root is not recommended!\n\n")
			}
			// When running as the root account the GID does not set
			//fmt.Println(userInfo.Gid)
			//gid, _ := strconv.Atoi(userInfo.Gid)
			//gid := 65534
			//if err := syscall.Setgid(gid); err != nil {
			//	fmt.Println("Error setting GID:", err)
			//	return
			//}
		} else {
			// If the flag is not set use the default of nobody to launch the server...
			userInfo, err = user.Lookup(userInput)
			CheckError("Unable to change user", err)
		}
		// The error setting uid will trigger if the server is not started with the a low privileged user...
		// If you are root then you can set the UID to whatever user you need... (Need to test on windows...)
		///uid, _ := strconv.Atoi(userInfo.Uid)
		///if err := syscall.Setuid(uid); err != nil {
		///	fmt.Println("Error setting UID:", err)
		///	os.Exit(0)
		///}
		//else {
		//	userInfo, err = user.Lookup("nobody")
		//	cf.CheckError("Unable to change user", err)
		//}

	} else {
		if IsFlagPassed("user") {
			fmt.Println("Unable to change the UID and GID due to being a non-privileged user...")
		}
		userInfo = currentUserInfo
	}

	var gid int
	if userInfo.Username == "nobody" {
		gid, _ = strconv.Atoi("65534")
	} else {
		gid, _ = strconv.Atoi(userInfo.Gid)
	}
	if err := syscall.Setgid(gid); err != nil {
		fmt.Println("Error setting GID:", err)
		os.Exit(0)
	}

	uid, _ := strconv.Atoi(userInfo.Uid)
	if err := syscall.Setuid(uid); err != nil {
		fmt.Println("Error setting UID:", err)
		os.Exit(0)
	}

	fmt.Printf("\nPermissions set for the process of the server.\nUsername: %s  UID: %s  GID: %s\n\n", userInfo.Username, userInfo.Uid, userInfo.Gid)
}
