package main

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/compileJavascriptV8"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o v8.bin -ldflags "-w -s" main.go
// GOOS=linux GOARCH=amd64 go build -o create.bin -ldflags "-w -s" backupMain.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o v8.exe -ldflags "-w -s" main.go

/*
References:

https://research.checkpoint.com/2024/exploring-compiled-v8-javascript-usage-in-malware/

https://pkg.go.dev/rogchap.com/v8go

https://github.com/rogchap/v8go

ChatGPT 7/31 results on using go and v8

*/

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"rogchap.com/v8go"
)

func serveJavaScript(w http.ResponseWriter, r *http.Request, byteData *v8go.CompilerCachedData) {

	// Create a new V8 Isolate
	iso2 := v8go.NewIsolate()
	defer iso2.Dispose()

	jsscript := "function foo() { return 'bar'; }; foo()"

	// Compile and run the cached script
	cachedData := v8go.CompileOptions{CachedData: byteData}
	script, _ := iso2.CompileUnboundScript(jsscript, "main.js", cachedData)
	//if err != nil {
	//	http.Error(w, "Failed to compile .jsc file", http.StatusInternalServerError)
	//	return
	//}
	if cachedData.CachedData.Rejected {
		fmt.Printf("expected cached data to be rejected")
	}

	// Create a new V8 Context
	ctx2 := v8go.NewContext(iso2)
	defer ctx2.Close()

	result, err := script.Run(ctx2)
	if err != nil {
		http.Error(w, "Failed to execute .jsc file", http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Hello, World! JSC Example</title>
			</head>
			<body>
				<h1>Result from JSC:</h1>
				<p>%s</p>
			</body>
			</html>
		`, result.String())

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)

	// Serve the result as JavaScript
	//w.Header().Set("Content-Type", "application/javascript")
	//fmt.Fprintf(w, "/* Decompiled JavaScript */\n%s", result)
}

func main() {

	jscData, err := os.ReadFile("script.jsc")
	if err != nil {
		log.Fatalf("Failed to read JSC file: %v", err)
	}

	cachedData := &v8go.CompilerCachedData{Bytes: jscData}

	// Start the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveJavaScript(w, r, cachedData)
	})
	log.Println("Listening on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
