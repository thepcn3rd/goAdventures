package main

/*

Setup the Environment

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/lnkCreator"

Create the directories of src, bin, and pkg

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o lnkCreator.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o createCerts.exe -ldflags "-w -s" main.go

References:
https://winprotocoldoc.blob.core.windows.net/productionwindowsarchives/MS-SHLLINK/%5bMS-SHLLINK%5d.pdf
https://github.com/gotopkg/mslnk/tree/master
https://github.com/libyal/liblnk/blob/main/documentation/Windows%20Shortcut%20File%20(LNK)%20format.asciidoc
https://www.trendmicro.com/en_ph/research/17/e/rising-trend-attackers-using-lnk-files-download-malware.html
https://github.com/parsiya/golnk/tree/master



In windows to copy the LNK file to linux
1. Goto registry key for HKEY_CLASSES_ROOT find LNK
2. Remove the registry entry of "Computer\HKEY_CLASSES_ROOT\lnkfile\NeverShowExt REG_SZ" - Remember to place back if you do not want the extension to show...
3. Reboot and the extension of .lnk will show
4. Rename the file to .txt or an extension of choice and then it can be copied to another computer.


*/

import (
	"bytes"
	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// 76 bytes in size
type ShellLinkHeader struct {
	HeaderSize     [4]byte
	LinkCLSID      [16]byte
	LinkFlags      [4]byte
	FileAttributes [4]byte
	CreationTime   [8]byte
	AccessTime     [8]byte
	WriteTime      [8]byte
	FileSize       [4]byte
	IconIndex      [4]byte
	ShowCommand    [4]byte
	HotKey         [2]byte
	Reserved1      [2]byte
	Reserved2      [4]byte
	Reserved3      [4]byte
}

type StructLinkFlags struct {
	HasLinkTargetIDList         bool // bit00 - ShellLinkHeader is followed by a LinkTargetIDList structure.
	HasLinkInfo                 bool // bit01 - LinkInfo in file.
	HasName                     bool // bit02 - NAME_String in file.
	HasRelativePath             bool // bit03 - RELATIVE_PATH in file.
	HasWorkingDir               bool // bit04 - WORKING_DIR in file.
	HasArguments                bool // bit05 - COMMAND_LINE_ARGUMENTS
	HasIconLocation             bool // bit06 - ICON_LOCATION
	IsUnicode                   bool // bit07 - Strings are in unicode
	ForceNoLinkInfo             bool // bit08 - LinkInfo is ignored
	HasExpString                bool // bit09 - The shell link is saved with an EnvironmentVariableDataBlock
	RunInSeparateProcess        bool // bit10 - Target runs in a 16-bit virtual machine
	Unused1                     bool // bit11 - ignore
	HasDarwinID                 bool // bit12 - The shell link is saved with a DarwinDataBlock
	RunAsUser                   bool // bit13 - The application is run as a different user when the target of the shell link is activated.
	HasExpIcon                  bool // bit14 - The shell link is saved with an IconEnvironmentDataBlock
	NoPidlAlias                 bool // bit15 - The file system location is represented in the shell namespace when the path to an item is parsed into an IDList.
	Unused2                     bool // bit16 - ignore
	RunWithShimLayer            bool // bit17 - The shell link is saved with a ShimDataBlock.
	ForceNoLinkTrack            bool // bit18 - The TrackerDataBlock is ignored.
	EnableTargetMetadata        bool // bit19 - The shell link attempts to collect target properties and store them in the PropertyStoreDataBlock (section 2.5.7) when the link target is set.
	DisableLinkPathTracking     bool // bit20 - The EnvironmentVariableDataBlock is ignored.
	DisableKnownFolderTracking  bool // bit21 - The SpecialFolderDataBlock (section 2.5.9) and the KnownFolderDataBlock (section 2.5.6) are ignored when loading the shell link. If this bit is set, these extra data blocks SHOULD NOT be saved when saving the shell link.
	DisableKnownFolderAlias     bool // bit22 - If the link has a KnownFolderDataBlock (section 2.5.6), the unaliased form of the known folder IDList SHOULD be used when translating the target IDList at the time that the link is loaded.
	AllowLinkToLink             bool // bit23 - Creating a link that references another link is enabled. Otherwise, specifying a link as the target IDList SHOULD NOT be allowed.
	UnaliasOnSave               bool // bit24 - When saving a link for which the target IDList is under a known folder, either the unaliased form of that known folder or the target IDList SHOULD be used.
	PreferEnvironmentPath       bool // bit25 - The target IDList SHOULD NOT be stored; instead, the path specified in the EnvironmentVariableDataBlock (section 2.5.4) SHOULD be used to refer to the target.
	KeepLocalIDListForUNCTarget bool // bit26 - When the target is a UNC name that refers to a location on a local machine, the local path IDList in the PropertyStoreDataBlock (section 2.5.7) SHOULD be stored, so it can be used when the link is loaded on the local machine.
}

func (s *StructLinkFlags) Init() {
	s.HasLinkTargetIDList = false
	s.HasLinkInfo = false
	s.HasName = false
	s.HasRelativePath = false
	s.HasWorkingDir = false
	s.HasArguments = false
	s.HasIconLocation = false
	s.IsUnicode = false
	s.ForceNoLinkInfo = false
	s.HasExpString = false
	s.RunInSeparateProcess = false
	s.Unused1 = false
	s.HasDarwinID = false
	s.RunAsUser = false
	s.HasExpIcon = false
	s.NoPidlAlias = false
	s.Unused2 = false
	s.RunWithShimLayer = false
	s.ForceNoLinkTrack = false
	s.EnableTargetMetadata = false
	s.DisableLinkPathTracking = false
	s.DisableKnownFolderTracking = false
	s.DisableKnownFolderAlias = false
	s.AllowLinkToLink = false
	s.UnaliasOnSave = false
	s.PreferEnvironmentPath = false
	s.KeepLocalIDListForUNCTarget = false
}

type StructFileAttributesFlags struct {
	FILE_ATTRIBUTE_READONLY            bool
	FILE_ATTRIBUTE_HIDDEN              bool
	FILE_ATTRIBUTE_SYSTEM              bool
	Reserved1                          bool
	FILE_ATTRIBUTE_DIRECTORY           bool
	FILE_ATTRIBUTE_ARCHIVE             bool
	Reserved2                          bool
	FILE_ATTRIBUTE_NORMAL              bool
	FILE_ATTRIBUTE_TEMPORARY           bool
	FILE_ATTRIBUTE_SPARSE_FILE         bool
	FILE_ATTRIBUTE_REPARSE_POINT       bool
	FILE_ATTRIBUTE_COMPRESSED          bool
	FILE_ATTRIBUTE_OFFLINE             bool
	FILE_ATTRIBUTE_NOT_CONTENT_INDEXED bool
	FILE_ATTRIBUTE_ENCRYPTED           bool
}

func (s *StructFileAttributesFlags) Init() {
	s.FILE_ATTRIBUTE_READONLY = false
	s.FILE_ATTRIBUTE_HIDDEN = false
	s.FILE_ATTRIBUTE_SYSTEM = false
	s.Reserved1 = false
	s.FILE_ATTRIBUTE_DIRECTORY = false
	s.FILE_ATTRIBUTE_ARCHIVE = false
	s.Reserved2 = false
	s.FILE_ATTRIBUTE_NORMAL = false
	s.FILE_ATTRIBUTE_TEMPORARY = false
	s.FILE_ATTRIBUTE_SPARSE_FILE = false
	s.FILE_ATTRIBUTE_REPARSE_POINT = false
	s.FILE_ATTRIBUTE_COMPRESSED = false
	s.FILE_ATTRIBUTE_OFFLINE = false
	s.FILE_ATTRIBUTE_NOT_CONTENT_INDEXED = false
	s.FILE_ATTRIBUTE_ENCRYPTED = false
}

type LinkTargetIDList struct {
	IDListSize [2]byte
	//IDListItems []byte
}

type ItemID struct {
	ItemIDSize [2]byte
	//ItemIDData []byte
}

type LinkInfo struct {
	LinkInfoSize [4]byte
	//Stuff        []byte
}

func writeShellLinkHeader(bFile *os.File) {
	var header ShellLinkHeader
	header.HeaderSize = [4]byte{0x4C, 0x00, 0x00, 0x00}
	header.LinkCLSID = [16]byte{0x01, 0x14, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}
	binary.Write(bFile, binary.LittleEndian, header)
}

func displayFileAttributesFunc(binaryString string, readFileAttributes *StructFileAttributesFlags) {
	// Convert Link Flags Hex into a Binary String
	binaryString = strings.Replace(binaryString, " ", "", -1)
	binaryString = strings.Replace(binaryString, "[", "", -1)
	binaryString = strings.Replace(binaryString, "]", "", -1)
	//fmt.Printf(binaryString)

	for bitLocation := 0; bitLocation < len(binaryString); bitLocation++ {
		// Remember that the binaryString has a line break at the end
		//fmt.Printf("%c\n", binaryString[i])
		charString := fmt.Sprintf("%c", binaryString[bitLocation])
		//fmt.Println(charString)
		switch bitLocation {
		case 0:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_READONLY (A)")
				readFileAttributes.FILE_ATTRIBUTE_READONLY = true
			}
		case 1:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_HIDDEN (B)")
				readFileAttributes.FILE_ATTRIBUTE_HIDDEN = true
			}
		case 2:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_SYSTEM (C)")
				readFileAttributes.FILE_ATTRIBUTE_SYSTEM = true
			}
		case 3:
			if charString == "1" {
				fmt.Println("\tReserved 1 - Must be Zero (D)")
				readFileAttributes.Reserved1 = true
			}
		case 4:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_DIRECTORY - (E)")
				readFileAttributes.FILE_ATTRIBUTE_DIRECTORY = true
			}
		case 5:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_ARCHIVE - (F)")
				readFileAttributes.Reserved1 = true
			}
		case 6:
			if charString == "1" {
				fmt.Println("\tReserved 2 - Must be Zero (G)")
				readFileAttributes.Reserved1 = true
			}
		case 7:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_NORMAL - (H)")
				readFileAttributes.FILE_ATTRIBUTE_NORMAL = true
			}
		case 8:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_TEMPORARY - (I)")
				readFileAttributes.FILE_ATTRIBUTE_TEMPORARY = true
			}
		case 9:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_SPARSE_FILE - (J)")
				readFileAttributes.FILE_ATTRIBUTE_SPARSE_FILE = true
			}
		case 10:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_REPARSE_POINT - (K)")
				readFileAttributes.FILE_ATTRIBUTE_REPARSE_POINT = true
			}
		case 11:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_COMPRESSED - (L)")
				readFileAttributes.FILE_ATTRIBUTE_COMPRESSED = true
			}
		case 12:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_OFFLINE - (M)")
				readFileAttributes.FILE_ATTRIBUTE_OFFLINE = true
			}
		case 13:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_NOT_CONTENT_INDEXED - (N)")
				readFileAttributes.FILE_ATTRIBUTE_NOT_CONTENT_INDEXED = true
			}
		case 14:
			if charString == "1" {
				fmt.Println("\tFILE_ATTRIBUTE_ENCRYPTED - (O)")
				readFileAttributes.FILE_ATTRIBUTE_ENCRYPTED = true
			}
		}
	}
}

func displayLinkFlagsFunc(binaryString string, readLinkFlags *StructLinkFlags) {
	// Convert Link Flags Hex into a Binary String
	binaryString = strings.Replace(binaryString, " ", "", -1)
	binaryString = strings.Replace(binaryString, "[", "", -1)
	binaryString = strings.Replace(binaryString, "]", "", -1)
	//fmt.Printf(binaryString)

	for bitLocation := 0; bitLocation < len(binaryString); bitLocation++ {
		// Remember that the binaryString has a line break at the end
		//fmt.Printf("%c\n", binaryString[i])
		charString := fmt.Sprintf("%c", binaryString[bitLocation])
		//fmt.Println(charString)
		switch bitLocation {
		case 0:
			if charString == "1" {
				fmt.Println("\tHasLinkTargetIDList (A)")
				readLinkFlags.HasLinkTargetIDList = true
			}
		case 1:
			if charString == "1" {
				fmt.Println("\tHasLinkInfo (B)")
				readLinkFlags.HasLinkInfo = true
			}
		case 2:
			if charString == "1" {
				fmt.Println("\tHasName (C)")
				readLinkFlags.HasName = true
			}
		case 3:
			if charString == "1" {
				fmt.Println("\tHasRelativePath (D)")
				readLinkFlags.HasRelativePath = true
			}
		case 4:
			if charString == "1" {
				fmt.Println("\tHasWorkingDir (E)")
				readLinkFlags.HasWorkingDir = true
			}
		case 5:
			if charString == "1" {
				fmt.Println("\tHasArguments (F)")
				readLinkFlags.HasArguments = true
			}
		case 6:
			if charString == "1" {
				fmt.Println("\tHasIconLocation (G)")
				readLinkFlags.HasIconLocation = true
			}
		case 7:
			if charString == "1" {
				fmt.Println("\tIsUnicode (H)")
				readLinkFlags.IsUnicode = true
			}
		case 8:
			if charString == "1" {
				fmt.Println("\tForceNoLinkInfo (I)")
				readLinkFlags.ForceNoLinkInfo = true
			}
		case 9:
			if charString == "1" {
				fmt.Println("\tHasExpString (J)")
				readLinkFlags.HasExpString = true
			}
		case 10:
			if charString == "1" {
				fmt.Println("\tRunInSeparateProcess (K)")
				readLinkFlags.RunInSeparateProcess = true
			}
		case 11:
			if charString == "1" {
				fmt.Println("\tUnused1 (L)")
				readLinkFlags.Unused1 = true
			}
		case 12:
			if charString == "1" {
				fmt.Println("\tHasDarwinID (M)")
				readLinkFlags.HasDarwinID = true
			}
		case 13:
			if charString == "1" {
				fmt.Println("\tRunAsUser (N)")
				readLinkFlags.RunAsUser = true
			}
		case 14:
			if charString == "1" {
				fmt.Println("\tHasExpIcon (O)")
				readLinkFlags.HasExpIcon = true
			}
		case 15:
			if charString == "1" {
				fmt.Println("\tNoPidlAlias (P)")
				readLinkFlags.NoPidlAlias = true
			}
		case 16:
			if charString == "1" {
				fmt.Println("\tUnused2 (Q)")
				readLinkFlags.Unused2 = true
			}
		case 17:
			if charString == "1" {
				fmt.Println("\tRunWithShimLayer (R)")
				readLinkFlags.RunWithShimLayer = true
			}
		case 18:
			if charString == "1" {
				fmt.Println("\tForceNoLinkTrack (S)")
				readLinkFlags.ForceNoLinkTrack = true
			}
		case 19:
			if charString == "1" {
				fmt.Println("\tEnableTargetMetadata (T)")
				readLinkFlags.EnableTargetMetadata = true
			}
		case 20:
			if charString == "1" {
				fmt.Println("\tDisableLinkPathTracking (U)")
				readLinkFlags.DisableLinkPathTracking = true
			}
		case 21:
			if charString == "1" {
				fmt.Println("\tDisableKnownFolderTracking (V)")
				readLinkFlags.DisableKnownFolderTracking = true
			}
		case 22:
			if charString == "1" {
				fmt.Println("\tDisableKnownFolderAlias (W)")
				readLinkFlags.DisableKnownFolderAlias = true
			}
		case 23:
			if charString == "1" {
				fmt.Println("\tAllowLinkToLink (X)")
				readLinkFlags.AllowLinkToLink = true
			}
		case 24:
			if charString == "1" {
				fmt.Println("\tUnaliasOnSave (Y)")
				readLinkFlags.UnaliasOnSave = true
			}
		case 25:
			if charString == "1" {
				fmt.Println("\tPreferEnvironmentPath (Z)")
				readLinkFlags.PreferEnvironmentPath = true
			}
		case 26:
			if charString == "1" {
				fmt.Println("\tKeepLocalIDListForUNCTarget (AA)")
				readLinkFlags.KeepLocalIDListForUNCTarget = true
			}
		}
	}
}

func displayTime(timeInfo [8]byte) uint64 {
	// Convert the hex from little endian to correct order for mathematical conversion to a decimal number
	hexString := fmt.Sprintf("%x", timeInfo)
	// The below 3 lines changes the hex string in little endian to the readable format...
	newHexString, _ := hex.DecodeString(hexString)
	hexStringBE := binary.LittleEndian.Uint64(newHexString)
	hexStringStr := fmt.Sprintf("%016x", hexStringBE)
	dateNanoSeconds, _ := strconv.ParseUint(hexStringStr, 16, 64)
	//fmt.Printf("Creation Time: %x - %d\n", readHeader.CreationTime, dateNanoSeconds)
	return dateNanoSeconds
}

func displaySize(sizeInfo [4]byte) uint64 {
	// Convert the hex from little endian to correct order for mathematical conversion to a decimal number
	hexString := fmt.Sprintf("%x", sizeInfo)
	// The below 3 lines changes the hex string in little endian to the readable format...
	newHexString, _ := hex.DecodeString(hexString)
	hexStringBE := binary.LittleEndian.Uint16(newHexString)
	hexStringStr := fmt.Sprintf("%08x", hexStringBE)
	dateNanoSeconds, _ := strconv.ParseUint(hexStringStr, 16, 32)
	//fmt.Printf("Creation Time: %x - %d\n", readHeader.CreationTime, dateNanoSeconds)
	return dateNanoSeconds
}

func displaySize2(sizeInfo [2]byte) uint64 {
	// Convert the hex from little endian to correct order for mathematical conversion to a decimal number
	hexString := fmt.Sprintf("%x", sizeInfo)
	// The below 3 lines changes the hex string in little endian to the readable format...
	newHexString, _ := hex.DecodeString(hexString)
	hexStringBE := binary.LittleEndian.Uint16(newHexString)
	hexStringStr := fmt.Sprintf("%08x", hexStringBE)
	dateNanoSeconds, _ := strconv.ParseUint(hexStringStr, 16, 32)
	//fmt.Printf("Creation Time: %x - %d\n", readHeader.CreationTime, dateNanoSeconds)
	return dateNanoSeconds
}

func displayHotkey(sizeInfo [2]byte) uint64 {
	// Convert the hex from little endian to correct order for mathematical conversion to a decimal number
	hexString := fmt.Sprintf("%x", sizeInfo)
	// The below 3 lines changes the hex string in little endian to the readable format...
	newHexString, _ := hex.DecodeString(hexString)
	hexStringBE := binary.LittleEndian.Uint16(newHexString)
	hexStringStr := fmt.Sprintf("%08x", hexStringBE)
	dateNanoSeconds, _ := strconv.ParseUint(hexStringStr, 16, 32)
	//fmt.Printf("Creation Time: %x - %d\n", readHeader.CreationTime, dateNanoSeconds)
	return dateNanoSeconds
}

// toTime converts an 8-byte Windows Filetime to time.Time.
func toTime(t [8]byte) time.Time {
	// Taken from https://golang.org/src/syscall/types_windows.go#L352, which is only available on Windows
	nsec := int64(binary.LittleEndian.Uint32(t[4:]))<<32 + int64(binary.LittleEndian.Uint32(t[:4]))
	// change starting time to the Epoch (00:00:00 UTC, January 1, 1970)
	nsec -= 116444736000000000
	// convert into nanoseconds
	nsec *= 100
	return time.Unix(0, nsec)
}

// formatTime converts a 8-byte Windows Filetime to time.Time and then formats
// it to string.
func formatTime(t [8]byte) string {
	return toTime(t).Format("2006-01-02 15:04:05.999999 -07:00")
}

func main() {
	// Read lnk file example.lnk
	inputLNK, err := os.Open("example.lnk")
	// The above lnk was generated from right-click on the desktop and create shortcut
	// Link Flags present in the above is A, D, E, G, H, U

	//inputLNK, err := os.Open("example_mslnk.lnk")

	// inputLNK, err := os.Open("example_with_args.lnk")
	// The above lnk was generated from a powershell script
	// Link Flags present in the above is A,B,C,D,E,F,G,H,J

	cf.CheckError("Unable to open example.lnk", err, true)
	defer inputLNK.Close()

	// Create a variable with the struct
	readHeader := ShellLinkHeader{}

	// Parse into ShellLinkHeader
	//binary.Read(inputLNK, binary.LittleEndian, &readHeader)
	binary.Read(inputLNK, binary.LittleEndian, &readHeader)

	// Output the contents of the ShellLinkHeader
	fmt.Println("ShellLinkHeader")
	fmt.Printf("Header Size: %s\n", fmt.Sprintf("%x", readHeader.HeaderSize))
	fmt.Printf("Link CLSID: %s\n", fmt.Sprintf("%x", readHeader.LinkCLSID))

	// Output the Link Flags ----------------------------------------------------------------------------------------------------
	fmt.Printf("Link Flags: %s - %s\n", fmt.Sprintf("%x", readHeader.LinkFlags), fmt.Sprintf("%08b", readHeader.LinkFlags))
	// Read the binaryString into the struct for Link Flags
	readLinkFlags := &StructLinkFlags{}
	readLinkFlags.Init()
	binaryString := fmt.Sprintf("%08b", readHeader.LinkFlags)
	displayLinkFlagsFunc(binaryString, readLinkFlags)
	// fmt.Println(readLinkFlags)

	// Output the FileAttributesFlags -------------------------------------------------------------------------------------------
	fmt.Printf("File Attributes Flags: %x - %08b\n", readHeader.FileAttributes, readHeader.FileAttributes)
	readFileAttributes := &StructFileAttributesFlags{}
	readFileAttributes.Init()
	binStrFileAttributes := fmt.Sprintf("%08b", readHeader.FileAttributes)
	displayFileAttributesFunc(binStrFileAttributes, readFileAttributes)

	// Output the Creation Time --------------------------------------------------------------------------------------------------
	fmt.Printf("Creation Time: %x - %d\n", readHeader.CreationTime, displayTime(readHeader.CreationTime))
	creationTime := formatTime(readHeader.CreationTime)
	fmt.Printf("%s\n", creationTime)
	// Output the Access Time --------------------------------------------------------------------------------------------------
	fmt.Printf("Access Time: %x - %d\n", readHeader.AccessTime, displayTime(readHeader.AccessTime))
	// Output the Write Time --------------------------------------------------------------------------------------------------
	fmt.Printf("Write Time: %x - %d\n", readHeader.WriteTime, displayTime(readHeader.WriteTime))
	// Output the File Size --------------------------------------------------------------------------------------------------
	fmt.Printf("File Size: %x - %d\n", readHeader.FileSize, displaySize(readHeader.FileSize))
	// Output the Icon Index --------------------------------------------------------------------------------------------------
	fmt.Printf("Icon Index: %x - %d\n", readHeader.IconIndex, displaySize(readHeader.IconIndex))
	// Output the Show Command --------------------------------------------------------------------------------------------------
	// Can be changed to the following:
	//"Show Normal      {01000000}
	//"Show Maximized":   {03000000}
	//"Show Minimized": {07000000}
	fmt.Printf("Show Command: %x - %d\n", readHeader.ShowCommand, displaySize(readHeader.ShowCommand))
	// Output the Display Hotkey --------------------------------------------------------------------------------------------------
	// Could place in a struct for the HotKeys based on the spec...
	fmt.Printf("Display Hotkey: %x - %08b\n", readHeader.HotKey, readHeader.HotKey)

	// Link Target ID List -------------------------------------------------------------------------------------------------------
	if readLinkFlags.HasLinkTargetIDList == true {
		fmt.Printf("\nLink Target ID List\n")
		// Create a variable with the struct
		readLinkTargetIDList := LinkTargetIDList{}
		binary.Read(inputLNK, binary.LittleEndian, &readLinkTargetIDList)
		fmt.Printf("ID List Size: %x - %d\n", readLinkTargetIDList.IDListSize, displaySize2(readLinkTargetIDList.IDListSize))
		//fmt.Printf("ID Lists: %x\n", readLinkTargetIDList.IDListItems)

		idList := make([]byte, displaySize2(readLinkTargetIDList.IDListSize))
		binary.Read(inputLNK, binary.LittleEndian, idList)
		fmt.Printf("ID Lists: %x\n", idList)
		idListReader := bytes.NewReader(idList)

		// Link Target ID List - Item 1
		//var idList []byte
		//idList = make([]byte, len(readLinkTargetIDList.IDListItems))
		// Remeber the copy is dst first from the src second...
		//copy(idList, readLinkTargetIDList.IDListItems[:])
		//fmt.Printf("\n%x\n", idList)
		//idListReader := bytes.NewReader(idList)
		readItemID := ItemID{}
		binary.Read(idListReader, binary.LittleEndian, &readItemID)
		fmt.Printf("\tItem ID Size: %x - %d\n", readItemID.ItemIDSize, displaySize2(readItemID.ItemIDSize))
		itemIDData := make([]byte, displaySize2(readItemID.ItemIDSize))
		binary.Read(idListReader, binary.LittleEndian, itemIDData)
		fmt.Printf("\tItem ID Data: %x\n", itemIDData)
	}

	// Link Info ------------------------------------------------------------------------------------------------------------------
	if readLinkFlags.HasLinkInfo == true {
		fmt.Printf("\nLink Info\n")
		readLinkInfo := LinkInfo{}
		binary.Read(inputLNK, binary.LittleEndian, &readLinkInfo)
		fmt.Printf("Link Info Size: %x - %d\n", readLinkInfo.LinkInfoSize, displaySize([4]byte(readLinkInfo.LinkInfoSize)))

	}

	//for i := 0; i <= (len(hexString) - 1); i += 2 {
	/* Manual way of getting a Big Endian representation of a number
	var newHexString string
	for i := (len(hexString) - 2); i >= 0; i -= 2 {
		fmt.Printf("%s\n", hexString[i:i+2])
		newHexString += hexString[i : i+2]
	}
	fmt.Printf("%s\n", newHexString)

	creationDateDecimal, err := strconv.ParseInt(fmt.Sprintf("%s", newHexString), 16, 64)
	fmt.Printf("%d\n", creationDateDecimal)
	*/

	//fmt.Printf("%d\n", (((creationDateDecimal - 116444736000000000) / 10000000) + 130610000))
	// Create a binary file to test with...
	/*
		outputBinary, err := os.Create("test.lnk")
		cf.CheckError("Unable to create the test.lnk file", err, true)
		writeShellLinkHeader(outputBinary)
		outputBinary.Close()
	*/
}
