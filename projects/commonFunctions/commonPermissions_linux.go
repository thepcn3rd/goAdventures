package common

// This file contains functions that are used by multiple programs
// Place in the src folder under a folder called commonFunctions
// In the commonFunctions folder after creating and dropping common.go in it
// Execute "go mod init commonFunctions"
// Then the files in common functions can be referenced in import as:
//    cf "commonFunctions"

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func SetupPermissionsProcess(userInput string) {
	var userInfo *user.User
	var err error
	// Modify the process the webserver runs as to the user and group specified...
	// The permissions on uploading files and accessing keys need to be set appropriately for the user and group...

	// Who is the current user logged in
	currentUserInfo, err := user.Current()
	CheckError("Unable to get current user", err, true)
	//fmt.Println("Current User: " + currentUserInfo.Username)
	//fmt.Println("Current UID: " + currentUserInfo.Uid)
	//fmt.Println("Current GID: " + currentUserInfo.Gid)

	// Change the UID and the GID if the current user is root
	if currentUserInfo.Uid == "0" {
		if IsFlagPassed("user") {
			userInfo, err = user.Lookup(userInput)
			CheckError("Unable to change user", err, true)
			if userInfo.Uid == "0" {
				fmt.Printf("DANGER: Running the server as root is not recommended!\n\n")
			}
			// When running as the root account the GID does not set
			//fmt.Println(userInfo.Gid)
			//gid, _ := strconv.Atoi(userInfo.Gid)
			//gid := 65534
			//if err := syscall.Setgid(gid); err != nil {
			//      fmt.Println("Error setting GID:", err, true)
			//      return
			//}
		} else {
			// If the flag is not set use the default of nobody to launch the server...
			userInfo, err = user.Lookup(userInput)
			CheckError("Unable to change user", err, true)
		}
		// The error setting uid will trigger if the server is not started with the a low privileged user...
		// If you are root then you can set the UID to whatever user you need... (Need to test on windows...)
		///uid, _ := strconv.Atoi(userInfo.Uid)
		///if err := syscall.Setuid(uid); err != nil {
		///     fmt.Println("Error setting UID:", err, true)
		///     os.Exit(0)
		///}
		//else {
		//      userInfo, err = user.Lookup("nobody")
		//      cf.CheckError("Unable to change user", err, true)
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
