# Creation of a DLL in Golang

#### Summary

For a couple of projects I had to turn my code into a dll.  This was not meant to side-load or conduct dll hijacking though it could be used for that.  This is to capture my notes on creating a dll in golang.

#### Creation of a DLL 

1. Normally in go you begin with code in the function main.  You will want to change that from main to the function you want to call with the execution of the DLL.  This could contain more than one that you would like called.  In the code snippet below you can see the main() {} function is empty.
2.  I then added the function of Executeprogram(). For this simple dll, I will use this function as if it was main.
3. Then for the function to appear in the export tables or the "import address table" you need to place the "//export Executeprogram" comment above the function.  Note the export name and the function name are the same, I am not sure if that needs to be the case.
```go
//export Executeprogram
func Executeprogram() {
	launchProg()

}

func main() {}
```

4. Add import "C" at the top underneath the package name as shown below

```go
package main
import "C"
import (
	"syscall"
	"unsafe"
)
```

5.  You do need a gcc compiler, at current I am using Manjaro and this is the command I used to install it.

```bash
sudo pacman -S mingw-w64-gcc
```

6. Then to compile it with go the following command is what I have used.  Not in most of the scripts that I write CGO_ENABLED is set to 0.  This needs to be set to 1.  I discovered this with a little pain, I understand why but could not explain it very well.

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -ldflags="-w -s -H=windowsgui" -o simple.dll
```
7. The dll now can be transfered over to windows and loaded with powershell, rundll32.exe or other methodologies also.
8. Example cmd to run it with rundll32.exe

```cmd
rundll32.exe v2fake.dll,Executeprogram
```

![Wolf Driving](/projects/simpleDLL/wolfDriving.png)