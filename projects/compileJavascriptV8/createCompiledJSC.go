package main

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/compileJavascriptV8"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o v8.bin -ldflags "-w -s" main.go

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
	"log"
	"os"

	"rogchap.com/v8go"
)

func main() {
	iso := v8go.NewIsolate()
	defer iso.Dispose()

	jsscript := "function foo() { return 'bar'; }; foo()"

	compiledScript, err := iso.CompileUnboundScript(jsscript, "main.js", v8go.CompileOptions{})
	if err != nil {
		log.Fatal(err)
	}

	byteData := compiledScript.CreateCodeCache()

	err = os.WriteFile("script.jsc", byteData.Bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Serialized compiled script to script.jsc")

}
