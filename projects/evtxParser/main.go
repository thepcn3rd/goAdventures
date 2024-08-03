/*
EVTX dumping utility, it can be used to carve raw data and recover EVTX events

I took the evtxdump utility and changed it for the purposes that I needed it for to parse a security.evtx file...

Allowed for a search keyword to be created
Allowed for a filter on EventIDs
Added a flag for the filename to parse

*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/0xrawsec/golang-evtx/evtx"
	"github.com/0xrawsec/golang-utils/log"
)

type Event struct {
	Event eventStruct `json:"Event"`
}

type eventStruct struct {
	EventData eventDataStruct `json:"EventData"`
	System    systemStruct    `json:"System"`
}

type eventDataStruct struct {
	AccessList            string `json:"AccessList"`
	AccessMask            string `json:"AccessMask"`
	ClientProcessID       string `json:"ClientProcessId"`
	ClientProcessStartKey string `json:"ClientProcessStartKey"`
	CommandLine           string `json:"CommandLine"`
	FQDN                  string `json:"FQDN"`
	IPAddress             string `json:"IpAddress"`
	MandatoryLabel        string `json:"MandatoryLabel"`
	NewProcessID          string `json:"NewProcessId"`
	NewProcessName        string `json:"NewProcessName"`
	ObjectName            string `json:"ObjectName"`
	ObjectServer          string `json:"ObjectServer"`
	ObjectType            string `json:"ObjectType"`
	OperationType         string `json:"OperationType"`
	ParentProcessID       string `json:"ParentProcessId"`
	ParentProcessName     string `json:"ParentProcessName"`
	ProcessID             string `json:"ProcessId"`
	ProcessName           string `json:"ProcessName"`
	RPCCallClientLocality string `json:"RpcCallClientLocality"`
	SubjectDomainName     string `json:"SubjectDomainName"`
	SubjectLogonID        string `json:"SubjectLogonId"`
	SubjectUserName       string `json:"SubjectUserName"`
	SubjectUserSid        string `json:"SubjectUserSid"`
	TargetDomainName      string `json:"TargetDomainName"`
	TargetLogonID         string `json:"TargetLogonId"`
	TargetUserName        string `json:"TargetUserName"`
	TargetUserSid         string `json:"TargetUserSid"`
	TaskContentNew        string `json:"TaskContentNew"`
	TaskName              string `json:"TaskName"`
	VirtualAccount        string `json:"VirtualAccount"`
	WorkstationName       string `json:"WorkstationName"`
}

type systemStruct struct {
	Channel       string            `json:"Channel"`
	Computer      string            `json:"Computer"`
	Correlation   correlationStruct `json:"Correlation"`
	EventID       string            `json:"EventID"`
	EventRecordID string            `json:"EventRecordID"`
	Execution     executionStruct   `json:"Execution"`
	Keywords      string            `json:"Keywords"`
	Level         string            `json:"Level"`
	Opcode        string            `json:"Opcode"`
	Properties    string            `json:"Properties"`
	Provider      providerStruct    `json:"Provider"`
	Security      securityStruct    `json:"Security"`
	Task          string            `json:"Task"`
	TimeCreated   timeCreatedStruct `json:"TimeCreated"`
	Version       string            `json:"Version"`
}

type correlationStruct struct {
	ActivityID string `json:"ActivityID"`
}

type executionStruct struct {
	ProcessID string `json:"ProcessID"`
	ThreadID  string `json:"ThreadID"`
}

type providerStruct struct {
	GUID string `json:"Guid"`
	Name string `json:"Name"`
}

type securityStruct struct {
}

type timeCreatedStruct struct {
	SystemTime string `json:"SystemTime"`
}

func printEvent(e *evtx.GoEvtxMap, keyword string, eventID string) string {
	var eventString string
	var eventInfoStruct Event
	eventString = ""
	// Output if there is an issue parsing an event ID
	if e != nil {

		event := evtx.ToJSON(e)
		err := json.Unmarshal(event, &eventInfoStruct)
		CheckError("Unable to Unmarshall the JSON", err, true)
		// Output a specific field in the Events
		//if len(eventInfoStruct.Event.EventData.WorkstationName) > 1 {
		//	fmt.Println(eventInfoStruct.Event.EventData.WorkstationName)
		//}
		//fmt.Println(eventID)
		//if eventInfoStruct.Event.System.EventID == eventID {
		//	fmt.Println("Match")
		//}
		// Conditions to search the event ID
		if len(keyword) > 0 && len(eventID) == 0 {
			//fmt.Println("Match 1")
			if strings.Contains(strings.ToLower(string(evtx.ToJSON(e))), keyword) {
				eventString = string(evtx.ToJSON(e))
			}
		} else if len(keyword) == 0 && eventInfoStruct.Event.System.EventID == eventID {
			//fmt.Println("Match 2")
			// Search by EventID
			eventString = string(evtx.ToJSON(e))
		} else if len(keyword) > 0 && eventInfoStruct.Event.System.EventID == eventID {
			//fmt.Println("Match 3")
			if strings.Contains(strings.ToLower(string(evtx.ToJSON(e))), keyword) {
				//fmt.Println("Match 4")
				eventString = string(evtx.ToJSON(e))
				//fmt.Println()
			} else {
				//fmt.Println("Match 5")
				eventString = ""
			}
		} else if len(keyword) == 0 && len(eventID) == 0 {
			eventString = string(evtx.ToJSON(e))
		} else {
			eventString = ""
		}

	} else {
		fmt.Println("ERROR parsing EVTX File")
		eventString = ""
	}
	//fmt.Println(eventString)
	return eventString
}

// CheckError checks for errors
func CheckError(reasonString string, err error, exitBool bool) {
	if err != nil && exitBool == true {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n\n", err)
		os.Exit(0)
	} else if err != nil && exitBool == false {
		fmt.Printf("%s\n", reasonString)
		//fmt.Printf("%s\n", err)
		return
	}
}

func SaveOutputFile(message string, fileName string) {
	outFile, _ := os.Create(fileName)
	//CheckError("Unable to create txt file", err, true)
	defer outFile.Close()
	w := bufio.NewWriter(outFile)
	n, err := w.WriteString(message)
	if n < 1 {
		CheckError("Unable to write to txt file", err, true)
	}
	outFile.Sync()
	w.Flush()
	outFile.Close()
}

///////////////////////////////// Main /////////////////////////////////////////

func main() {
	keywordPtr := flag.String("keyword", "", "Search for a given keyword amongst the string")
	evtxFilePtr := flag.String("file", "", "EVTX File to Search")
	eventIDPtr := flag.String("eventid", "", "Event ID to find amongst the logs")
	flag.Parse()

	var eventOutputJSON string
	eventOutputJSON = "{\"Events\":["

	// Regular EVTX file, we use OpenDirty because
	// the file might be in a dirty state
	ef, err := evtx.OpenDirty(*evtxFilePtr)

	if err != nil {
		log.Error(err)
	}

	// Loop through the events from the file
	for e := range ef.FastEvents() {
		eventString := printEvent(e, *keywordPtr, *eventIDPtr)
		//fmt.Println(eventString)
		if eventString != "" {
			eventOutputJSON += eventString + ","
		}
	}
	eventOutputJSON = eventOutputJSON[:len(eventOutputJSON)-1] + "]}"
	//prettyJSON, err := json.MarshalIndent(eventOutputJSON, "", "    ")
	//CheckError("Unable to marshall indent the JSON", err, true)
	SaveOutputFile(eventOutputJSON, "output.json")
	//SaveOutputFile(string(prettyJSON), "output.json")

}
