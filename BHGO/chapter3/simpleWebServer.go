package main

// In the upload.html adjust to the IP Address of the upload server...

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o simpleWebServer.bin -ldflags "-w -s" simpleWebServer.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o simpleWebServer.exe -ldflags "-w -s" simpleWebServer.go

// Create the TLS keys for the https web server
// openssl genrsa -out server.key 2048
// openssl ecparam -genkey -name secp384r1 -out server.key
// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365

// Directory structure
// - simpleWebServer.go
// - keys/
// - - server.key
// - - server.crt
// - uploads/ (Location where files are uploaded protected from download)
// - static/  (Web site contents)
// - - index.html
// - - upload.html
// - - chef/ (CyberChef location)
// - - downloads/ (Location where files are downloaded)
// - - downloads/index.html (Disable the directory indexing with this file)

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Read HTML Formatting
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.ParseMultipartForm(10 << 20)
	// Reads the variable password ...  The contained password has to match for the upload to be successful
	passwordInput := r.FormValue("password")
	if passwordInput != "t3sting" {
		fmt.Fprintf(w, "Failed to Upload File\n")
		return
	}
	file, handler, err := r.FormFile("myFile")
	checkError("Error Loading the File", err)
	defer file.Close()
	fmt.Fprintf(w, "Uploaded File: %+v<br />", handler.Filename)
	fmt.Fprintf(w, "File Size: %+v<br />", handler.Size)
	fmt.Fprintf(w, "MIME Header: %+v<br /><br />", handler.Header)
	//tempFile, err := ioutil.TempFile("./uploads", handler.Filename)
	//checkError("Error creating temp file", err)
	//defer tempFile.Close()
	fileBytes, err := ioutil.ReadAll(file)
	checkError("Unable to Read File Selected", err)
	// Had the directory in static/uploads moved it to be private
	f, err := os.Create("./uploads/" + handler.Filename)
	checkError("Unable create file to save output", err)
	defer f.Close()
	f.Write(fileBytes)
	fmt.Fprintf(w, "Successfully Uploaded File<br />")
	//Change based on the location...
	//fmt.Fprint(w, "<a href='https://127.0.0.1:9000/upload.html'>Link to Upload</a>")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/upload", uploadFile)
	// Set the port appropriately...
	log.Fatal(http.ListenAndServeTLS(":9000", "./keys/server.crt", "./keys/server.key", nil))

}

// References:
// https://tutorialedge.net/golang/creating-simple-web-server-with-golang/
// https://tutorialedge.net/golang/go-file-upload-tutorial/

// Contents of upload.html
/*
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
  </head>
  <body>
    <form
      enctype="multipart/form-data"
      action="https://127.0.0.1:9000/upload"
      method="post"
    >
      Password for file upload:&nbsp;
      <input type="password" name="password" /><br />
      <input type="file" name="myFile" /><br />
      <input type="submit" value="Upload" />
    </form>
  </body>
</html>
*/
