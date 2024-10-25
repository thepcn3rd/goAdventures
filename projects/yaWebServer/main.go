package main

// Setup the following for the application
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/chapter3/yaWebServer"

Create the directories of src, bin, and pkg

Need the following package dependency
go get github.com/gomarkdown/markdown

// Create a folder under src called commonFunctions
// Execute "go mod init commonFunctions" to initialize it as a module, this creates a go.mod file in the directory
cf "commonFunctions"
Then the files inside of commonFunctions can be copied between projects and utilized
*/

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o yaWebServer.bin -ldflags "-w -s" main.go

// To work with parrot linux docker
// Manjaro uses glibc 2.32 2.34, the parrot linux docker uses 2.31 due to the difference CGO_ENABLED=0 will remove the dependency...
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o yaWebServer.bin main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o yaWebServer.exe -ldflags "-w -s" main.go

// Create the TLS keys for the https web server
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key
// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365

// Directory structure
// - simpleWebServer.go
// - code/ (Location of uncompiled go src files, intent is to compile them..., Not searchable inside of static)
// - keys/
// - - server.key
// - - server.crt
// - library/ (Location of the markdown files)
// - notesLog/ (Storage of Notes)
// - - notes.log (Read)
// - static/  (Web site contents)
// - - index.html
// - - chef/ (CyberChef location)
// - - - c.html (CyberChef html file copied to be this shortened filename)
// - - codelist (Function codeList, intent is to list the contents of code ready to compile, called by a form as a POST after password is entered)
// - - codelist.html (Lists the directory contents of the files that can be compiled)
// - - compilecode (Function compileCode, intent is to compile the file listed in the code directory, called by POST from codecompile.html)
// - - compilecode.html (Allows interaction to input password and input the file to compile and is posted to compilecode)
// - - downloads/ (Location where files are downloaded)
// - - downloads/index.html (Disable the directory indexing with this file)
// - - library (Allows you to view the markdown files that are in the library directory)
// - - notes (Allows to create notes of what is occurring)
// - - upload (Function uploadFile, called in the form as a POST request in upload.html)
// - - upload.html
// - uploads/ (Location where files are uploaded protected from download)

/*
Added a simple way to take notes; to be able to send commands or information back and forth...
Added a way to add markdown files in the library folder and then reference them (Filenames can only contain uppercase and lowercase letters with a period allowed)
When a golang program is created it links to glibc libraries, adjusted a command above that removes that linkage

Future Activities:
Add sanitization of the codefile due to it executing a command this could be dangerous
Add the AMSI bypass in Amazon Lambda
Add the Download of Github pages through the Amazon Lambda Function
Remove the deprecated dependency ioutil
Add instructions to build the TLS keys if they are missing when the prog starts
Add the ability to save the cookies to a file as they are discovered...

*/

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	// Create a folder under src called commonFunctions
	// Execute "go mod init commonFunctions" to initialize it as a module, this creates a go.mod file in the directory
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Reference:
// https://onlinetool.io/goplayground/#txO7hJ-ibeU
func mdToHTML(md []byte) []byte {
	var printAst = false
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	if printAst {
		fmt.Print("--- AST tree:\n")
		ast.Print(os.Stdout, doc)
		fmt.Print("\n")
	}

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func uploadFile(w http.ResponseWriter, r *http.Request, passSHA256 string) {
	loggingOutput(r)
	// Read HTML Formatting
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ParseMultipartForm(10 << 20)
	// Reads the variable password ...  The contained password has to match for the upload to be successful
	// Due to injection we are going to compare the hashes of the passwords...
	passInputSHA256 := cf.CalcSHA256Hash(r.FormValue("password"))
	if passInputSHA256 != passSHA256 {
		fmt.Fprintf(w, "Failed to Upload File\n")
		return
	}
	file, handler, err := r.FormFile("myFile")
	cf.CheckError("Error Loading the File", err, true)
	defer file.Close()
	fmt.Fprintf(w, "Uploaded File: %+v<br />", handler.Filename)
	fmt.Fprintf(w, "File Size: %+v<br />", handler.Size)
	fmt.Fprintf(w, "MIME Header: %+v<br /><br />", handler.Header)
	//tempFile, err := ioutil.TempFile("./uploads", handler.Filename)
	//cf.CheckError("Error creating temp file", err, true)
	//defer tempFile.Close()
	fileBytes, err := ioutil.ReadAll(file)
	cf.CheckError("Unable to Read File Selected", err, true)
	// Had the directory in static/uploads moved it to be private
	f, err := os.Create("./uploads/" + handler.Filename)
	cf.CheckError("Unable create file to save output", err, true)
	defer f.Close()
	f.Write(fileBytes)
	fmt.Fprintf(w, "Successfully Uploaded File<br />")
	//Change based on the location...
	fmt.Fprint(w, "<a href='/upload.html'>Link to Upload</a>")
}

func compileCode(w http.ResponseWriter, r *http.Request, passSHA256 string) {
	loggingOutput(r)
	// Read HTML Formatting
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ParseMultipartForm(10 << 20)
	// Reads the variable password ...  The contained password has to match for the upload to be successful
	// Due to injection we are going to compare the hashes of the passwords...
	passInputSHA256 := cf.CalcSHA256Hash(r.FormValue("password"))
	if passInputSHA256 != passSHA256 {
		return
	}
	// codeFile is not sanitized... Security issue...
	codeFile := r.FormValue("codefile")
	osType := r.FormValue("ostype")
	// Identify the current working directory
	currentDir, err := os.Getwd()
	cf.CheckError("Unable to identify current directory", err, true)
	// The directory of the current codelist (linux based)
	dirPath := currentDir + "/code/" + codeFile
	if _, err := os.Stat(dirPath); err == nil {
		fmt.Fprintf(w, "<html><body>")
		fmt.Fprintf(w, "File exists...<br />")
		fmt.Fprintf(w, "Compiling server side...<br />")
		var outputPath string
		if strings.ToLower(osType) == "linux" {
			// Remove the go extension (linux)
			outputPath = currentDir + "/static/downloads/" + codeFile[:len(codeFile)-2] + "bin"
		} else if strings.ToLower(osType) == "windows" {
			// Remove the go extension (windows)
			outputPath = currentDir + "/static/downloads/" + codeFile[:len(codeFile)-2] + "exe"
		} else {
			return
		}
		// With the ldflags it fails to execute...
		cmd := exec.Command("go", "build", "-o", outputPath, dirPath)
		outputCompile, err := cmd.Output()
		//outputCompile, err := exec.Command(a, b, c, d, outputPath, dirPath).Output()
		cf.CheckError("Unable to execute the compile command...", err, true)
		cmd.Env = os.Environ()
		if strings.ToLower(osType) == "windows" {
			cmd.Env = append(cmd.Env, "GOOS=windows")
			cmd.Env = append(cmd.Env, "GOARCH=amd64")
		} else {
			cmd.Env = append(cmd.Env, "GOOS=linux")
			cmd.Env = append(cmd.Env, "GOARCH=amd64")
		}
		fmt.Fprintf(w, "<pre>")
		//cmdOutput, err := outputCompile.Output()
		outputCompileStr := string(outputCompile[:])
		fmt.Fprintf(w, outputCompileStr)
		fmt.Fprintf(w, "</pre>")
		fmt.Fprintf(w, "</body></html>")
	} else {
		fmt.Fprintf(w, "File does not exist...<br />")
	}

}

func codeList(w http.ResponseWriter, r *http.Request, passSHA256 string) {
	loggingOutput(r)
	// Read HTML Formatting
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ParseMultipartForm(10 << 20)
	// Reads the variable password ...  The contained password has to match for the upload to be successful
	// Due to injection we are going to compare the hashes of the passwords...
	passInputSHA256 := cf.CalcSHA256Hash(r.FormValue("password"))
	if passInputSHA256 != passSHA256 {
		fmt.Fprintf(w, "Failed to list code available...\n")
		return
	}
	// Identify the current working directory
	currentDir, err := os.Getwd()
	cf.CheckError("Unable to identify current directory", err, true)
	// The directory of the current codelist (linux based)
	dirPath := currentDir + "/code"
	//fmt.Print(dirPath)
	files, err := os.ReadDir(dirPath)
	cf.CheckError("Unable to read contents of the code directory", err, true)
	fmt.Fprint(w, "<b>Files in Directory</b><br />")
	for _, file := range files {
		if !file.IsDir() {
			outputHTML := file.Name() + "<br />"
			fmt.Fprint(w, outputHTML)
		}
	}
	fmt.Fprint(w, "<br />")
	fmt.Fprint(w, "<a href='/compilecode.html'>Link to Compile Code Page</a>")

}

func headerHTML() string {
	hHTML := `<!DOCTYPE html>
			  <html lang="en">
  			  <head>
    			<meta charset="UTF-8" />
    			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    			<meta http-equiv="X-UA-Compatible" content="ie=edge" />
  			  </head>
  			  <body>`
	return hHTML
}

func tailHTML() string {
	tHTML := "</body></html>"
	return tHTML
}

func notesList(w http.ResponseWriter, r *http.Request, passSHA256 string) {
	loggingOutput(r)
	fmt.Fprint(w, headerHTML())
	cListHTML_1 := `<form enctype="multipart/form-data" action="/notes" method="post">
		Password to list the notes:&nbsp;
		<input type="password" name="password" /><br />
		<br />
		Information to Post:&nbsp;
		<input type="text" name="inputTXT" width=100 /><br />
		<br />`
	fmt.Fprint(w, cListHTML_1)
	cListHTML_2 := `<br /><br />
			<input type="submit" value="Submit" />
			</form><br /><br />`
	fmt.Fprint(w, cListHTML_2)

	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20)
		passInputSHA256 := cf.CalcSHA256Hash(r.FormValue("password"))
		if passInputSHA256 != passSHA256 {
			fmt.Fprint(w, headerHTML())
			fmt.Fprintf(w, "<br />Failed to list notes...<br />")
			fmt.Fprint(w, tailHTML())
		} else {
			inputTXT := r.FormValue("inputTXT")
			// if the length of the input is more than 3 characters then save the information to a file
			if len(inputTXT) > 3 {
				// save the notes to a file, read the file, etc.
				//fmt.Fprint(w, inputTXT)
				// Place the notes
				// Built for Linux
				f, err := os.OpenFile("notesLog/notes.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				cf.CheckError("Unable to open the notes log", err, true)
				defer f.Close()
				f.WriteString("\n" + inputTXT + "\n")
				f.Close()
			}
			fmt.Fprint(w, "<br />")
			//fmt.Fprint(w, "<code>")
			// Read and display the information on the page as <code>
			f, err := os.OpenFile("notesLog/notes.log", os.O_RDONLY, 0640)
			cf.CheckError("Unable to open the notes log", err, true)
			fileScanner := bufio.NewScanner(f)
			//fileScanner.Split(bufio.ScanLines)
			for fileScanner.Scan() {
				lineTXT := fileScanner.Text() + "<br />"
				fmt.Fprint(w, lineTXT)
			}
			fmt.Fprint(w, "</br />")
		}
	}
	fmt.Fprint(w, tailHTML())

}

func libraryList(w http.ResponseWriter, r *http.Request, passSHA256 string) {
	loggingOutput(r)
	fmt.Fprint(w, headerHTML())
	cListHTML_1 := `<form enctype="multipart/form-data" action="/library" method="post">
		Password to interact with the library:&nbsp;
		<input type="password" name="password" /><br />
		<br />
		File to View:&nbsp;
		<input type="text" name="inputTXT" width=100 /><br />
		(Leave blank to display contents of the library)
		<br />`
	fmt.Fprint(w, cListHTML_1)
	cListHTML_2 := `<br />
			<input type="submit" value="Show File" />
			</form><br /><br />`
	fmt.Fprint(w, cListHTML_2)

	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20)
		passInputSHA256 := cf.CalcSHA256Hash(r.FormValue("password"))
		if passInputSHA256 != passSHA256 {
			fmt.Fprint(w, headerHTML())
			fmt.Fprintf(w, "<br />Failed to show file...<br />")
			fmt.Fprint(w, tailHTML())
		} else {
			//Open markdown file
			mdFileTXT := r.FormValue("inputTXT")
			// Reference: https://www.golangprograms.com/how-to-remove-special-characters-from-a-string-in-golang.html
			// Remove any possibilities of injection
			// Assumes the file read only contains lowercase and upper-case letters and a period
			mdFileTXT = regexp.MustCompile(`[^a-zA-Z\.]+`).ReplaceAllString(mdFileTXT, "")
			if len(mdFileTXT) > 3 {
				filePath := "library/" + mdFileTXT
				mdFile, err := os.Open(filePath)
				if err == nil {
					defer mdFile.Close()
					stat, err := mdFile.Stat()
					cf.CheckError("Unable to read size of file", err, true)
					mdByteArray := make([]byte, stat.Size())
					_, err = bufio.NewReader(mdFile).Read(mdByteArray)
					cf.CheckError("Unable to read file into byte array", err, true)
					//md := []byte(mdByteArray)
					html := mdToHTML(mdByteArray)
					fmt.Fprint(w, string(html))
					fmt.Fprint(w, "</br />")
				} else {
					fmt.Fprint(w, "Failed to open the file specified")
				}
			} else {
				// List the directory contents of the library
				libraryFiles, err := os.ReadDir("library/")
				cf.CheckError("Unable to read the directory contents of the library", err, true)
				for _, file := range libraryFiles {
					fileInfo := "<br />" + file.Name()
					if file.IsDir() == false {
						fmt.Fprint(w, fileInfo)
					}
				}
			}
		}
	}
	fmt.Fprint(w, tailHTML())

}

func uploadFileHTML(w http.ResponseWriter, r *http.Request) {
	loggingOutput(r)
	fmt.Fprint(w, headerHTML())
	ufHTML := `<form enctype="multipart/form-data" action="/upload" method="post">
      Password for file upload:&nbsp;
      <input type="password" name="password" /><br />
      <input type="file" name="myFile" /><br />
      <input type="submit" value="Upload" />
    </form>`
	fmt.Fprint(w, ufHTML)
	fmt.Fprint(w, tailHTML())
}

func codeListHTML(w http.ResponseWriter, r *http.Request) {
	loggingOutput(r)
	fmt.Fprint(w, headerHTML())
	ufHTML := `<form enctype="multipart/form-data" action="/codelist" method="post">
	  Password to list code available to compile:&nbsp;
      <input type="password" name="password" /><br />
      <input type="submit" value="List Code" />
    </form>`
	fmt.Fprint(w, ufHTML)
	fmt.Fprint(w, tailHTML())
}

func compileCodeHTML(w http.ResponseWriter, r *http.Request) {
	loggingOutput(r)
	fmt.Fprint(w, headerHTML())
	ufHTML := `<form enctype="multipart/form-data" action="/compilecode" method="post">
	  Password for compiling code:&nbsp;
      <input type="password" name="password" /><br />
	  Name of file to compile:&nbsp;
	  <input type="text" name="codefile" /><br />
	  OS Type (ie. Windows, Linux):&nbsp;
	  <input type="text" name="ostype" /><br />
      <input type="submit" value="Compile Code" />
    </form>`
	fmt.Fprint(w, ufHTML)
	fmt.Fprint(w, tailHTML())
}

func createIndexHTML(folderDir string) {
	currentDir, _ := os.Getwd()
	newDir := currentDir + folderDir
	//cf.CheckError("Unable to get the working directory", err, true)
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		// Output to File - Overwrites if file exists...
		f, err := os.Create(newDir)
		cf.CheckError("Unable create file index.html "+currentDir, err, true)
		defer f.Close()
		f.Write([]byte(headerHTML()))
		f.Write([]byte("Swayzee Merchantile"))
		f.Write([]byte(tailHTML()))
		f.Close()
	}
}

func loggingOutput(r *http.Request) {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	timeNow := time.Now()
	stringTime := timeNow.Format(time.RFC822)
	fmt.Print(colorGreen + "<-- Request -->\n" + colorReset)
	remoteAddrItems := strings.Split(r.RemoteAddr, ":") // The displays the <sourceIP>:<sourcePort>
	fmt.Print("Date: " + stringTime + " SIP: " + remoteAddrItems[0] + " Method: " + r.Method + " URL: " + r.URL.String() + "\n")
	fmt.Print(colorGreen + "<-- Headers -->\n" + colorReset)
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Println(key + ": " + value)
			//rw.Header().Set(key, value)
		}
	}
}

func main() {
	// Setup flags for the accessPass, HTTP or HTTPS, and Listening Port
	accessPassPtr := flag.String("pwd", "t3sting", "Password needed for Web UI (default: -pwd t3sting)")
	httpSecurePtr := flag.String("https", "true", "Disable site from running HTTPS (default: -https true)")
	listeningPortPtr := flag.String("port", "9000", "Change default listening port (default: -port 9000)")
	changeUserPtr := flag.String("user", "nobody", "Change default user (default: -user nobody)")
	flag.Parse()

	// Modify the accessPassword as needed
	//accessPass := "t3sting"
	accessPass := *accessPassPtr
	accessPassSHA256 := cf.CalcSHA256Hash(accessPass)

	// Create the directory structure if it does not exist...
	cf.CreateDirectory("/code")
	cf.CreateDirectory("/keys")
	cf.CreateDirectory("/library")
	cf.CreateDirectory("/notesLog")
	cf.CreateDirectory("/static")
	createIndexHTML("/static/index.html")
	cf.CreateDirectory("/static/downloads")
	createIndexHTML("/static/downloads/index.html")
	cf.CreateDirectory("/uploads")

	// Does the certConfig.json  file exist in the keys folder
	configFileExists := cf.FileExists("/keys/certConfig.json")
	//fmt.Println(configFileExists)
	if configFileExists == false {
		cf.CreateCertConfigFile()
		fmt.Println("Created keys/certConfig.json, modify the values to create the self-signed cert utilized")
		os.Exit(0)
	}

	// Does the server.crt and server.key files exist in the keys folder
	crtFileExists := cf.FileExists("/keys/server.crt")
	keyFileExists := cf.FileExists("/keys/server.key")
	if crtFileExists == false || keyFileExists == false {
		cf.CreateCerts()
		crtFileExists := cf.FileExists("/keys/server.crt")
		keyFileExists := cf.FileExists("/keys/server.key")
		if crtFileExists == false || keyFileExists == false {
			fmt.Println("Failed to create server.crt and server.key files")
			os.Exit(0)
		}
	}

	// Modify the permissions on the process running the server...
	cf.SetupPermissionsProcess(*changeUserPtr)

	// Setup of the webserver and the handlers...
	//http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//http.Handle("/", http.FileServer(http.Dir("./static")))
		loggingOutput(r)
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	})
	// Handling the Upload Function
	http.HandleFunc("/upload.html", uploadFileHTML)
	// http.HandleFunc("/upload", uploadFile)
	// Below I created a wrapper function to then allow the passing of a variable to
	// uploadFile
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadFile(w, r, accessPassSHA256)
	})
	// Handling Code
	http.HandleFunc("/codelist.html", codeListHTML)
	http.HandleFunc("/codelist", func(w http.ResponseWriter, r *http.Request) {
		codeList(w, r, accessPassSHA256)
	})
	http.HandleFunc("/compilecode.html", compileCodeHTML)
	http.HandleFunc("/compilecode", func(w http.ResponseWriter, r *http.Request) {
		compileCode(w, r, accessPassSHA256)
	})
	// Handling the Simple Notes Feature
	http.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		notesList(w, r, accessPassSHA256)
	})

	// Handling looking at the library feature
	http.HandleFunc("/library", func(w http.ResponseWriter, r *http.Request) {
		libraryList(w, r, accessPassSHA256)
	})

	httpSecure := *httpSecurePtr
	// Set the port appropriately... (Can set here to bind to a specific IP Address)
	listeningPort := *listeningPortPtr
	listeningPort = ":" + listeningPort
	if httpSecure == "true" {
		fmt.Printf("Started the webserver with TLS on port: %s\n", listeningPort)
		log.Fatal(http.ListenAndServeTLS(listeningPort, "./keys/server.crt", "./keys/server.key", nil))
	} else {
		fmt.Printf("Started the webserver with no encryption on port: %s\n", listeningPort)
		log.Fatal(http.ListenAndServe(listeningPort, nil))
	}

}
