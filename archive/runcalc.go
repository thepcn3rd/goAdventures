package main

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -buildmode=pie -o g.exe -ldflags "-w -s" main.go
// Used with go-donut to generate shellcode from the g.exe file
// ./donut.bin -i g.exe

import(
    "fmt"
    "os/exec"
)

func main(){    
    c := exec.Command("calc.exe")

    if err := c.Run(); err != nil { 
        fmt.Println("Error: ", err)
    }   
}
