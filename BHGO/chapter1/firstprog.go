package main

// To run the program execute: go run main.go

// To build the program: go build main.go -o output.bin

// To build the program without debugging information:
// go build -o output.bin -ldflags "-w -s" main.go

// To cross compile for linux
// GOOS=linux GOARCH=amd64 go build -o output.linux -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o output.exe -ldflags "-w -s" main.go

// Additional training...  https://tour.golang.org

import "fmt"

func main() {
	fmt.Println("Hello")
}
