package main

/*
//Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/simpleDLL"

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags="-w -s -H=windowsgui" -o simpleDLL.dll

// Create a dll from go... (CURRENT CONFIG)
1. Change the main function to Executeprogram
2. Place "//export Executeprogram"
3. Add import "C" at the top
4. Add a func main - To recognize the main function but it is not used...
5. Install if not... sudo pacman -S mingw-w64-gcc
6. GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags="-w -s -H=windowsgui" -o v2fake.dll
7. Transfer over to windows...
8. rundll32.exe v2fake.dll,Executeprogram


*/

import "C"
import (
	"syscall"
	"unsafe"
)

func LaunchProg() {
	shellExecute := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")

	shellExecute.Call(0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("notepad.exe"))), 0, 0, 1)
}

//export Executeprogram
func Executeprogram() {
	LaunchProg()
}

func main() {}
