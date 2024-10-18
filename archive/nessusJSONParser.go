package nessusJSONParser

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
References:
https://github.com/kkirsche/nessusControl/blob/master/api/structsScans.go

*/

// Made the struct public with a captial letter so it could be referenced in a function in main.go
// *************************  Scans Structure **********************************
type ScanStruct struct {
	Scans []scanIDStruct `json:"scans"`
}

type scanIDStruct struct {
	Control              bool              `json:"control"`
	CreationDate         int               `json:"creation_date"`
	Enabled              bool              `json:"enabled"`
	ID                   int               `json:"id"`
	LastModificationDate int               `json:"last_modification_date"`
	Name                 string            `json:"name"`
	Owner                string            `json:"owner"`
	Read                 bool              `json:"read"`
	Rrules               string            `json:"rrules"`
	Schedule_UUID        string            `json:"schedule_uuid"`
	Shared               bool              `json:"shared"`
	Starttime            string            `json:"starttime"`
	Status               string            `json:"status"`
	Template_UUID        string            `json:"template_uuid"`
	Timezone             string            `json:"timezone"`
	HasTriggers          bool              `json:"has_triggers"`
	Type                 string            `json:"type"`
	Permissions          int               `json:"permissions"`
	UserPermissions      int               `json:"user_permissions"`
	UUID                 string            `json:"uuid"`
	Wizard_UUID          string            `json:"wizard_uuid"`
	Progress             int               `json:"progress"`
	TotalTargets         int               `json:"total_targets"`
	StatusTimes          statusTimesStruct `json:"status_times"`
}

type statusTimesStruct struct {
	Initializing int `json:"initializing"`
	Pending      int `json:"pending"`
	Processing   int `json:"processing"`
	Publishing   int `json:"publishing"`
	Running      int `json:"running"`
}

// *************************   End Scans Structure *************************

// *************************   Scan Details Strucutre **********************

type ScanDetailsStruct struct {
	Info            infoStruct         `json:"info"`
	Hosts           []hostsStruct      `json:"hosts"`
	Vulnerabilities []vulnStruct       `json:"vulnerabilities"`
	CompHosts       []string           `json:"comphosts,omitempty"`
	Compliance      []string           `json:"compliance,omitempty"`
	Filters         []filtersStruct    `json:"filters"`
	History         []historyStruct    `json:"history"`
	Notes           []notesStruct      `json:"notes"`
	Remediations    remediationsStruct `json:"remediations"`
}

type infoStruct struct {
	Owner            string `json:"owner"`
	Name             string `json:"name"`
	No_target        bool   `json:"no_target"`
	Folder_id        int    `json:"folder_id,omitempty"`
	Control          bool   `json:"control"`
	User_permissions int    `json:"user_permissions"`
	Schedule_UUID    string `json:"schedule_uuid"`
	Edit_allowed     bool   `json:"edit_allowed"`
	Scanner_name     string `json:"scanner_name"`
	Policy           string `json:"policy"`
	Shared           bool   `json:"shared"`
	Object_id        int    `json:"object_id,omitempty"`
	Tag_targets      string `json:"tag_targets,omitempty"`
	Acls             string `json:"acls,omitempty"`
	Hostcount        int    `json:"hostcount"`
	Uuid             string `json:"uuid"`
	Status           string `json:"status"`
	Scan_type        string `json:"scan_type"`
	Targets          string `json:"targets"`
	Alt_targets_used bool   `json:"alt_targets_used"`
	Pci_can_upload   bool   `json:"pci-can-upload"`
	Scan_start       int    `json:"scan_start"`
	Timestamp        int    `json:"timestamp"`
	Is_archived      bool   `json:"is_archived"`
	Reindexing       bool   `json:"reindexing"`
	Scan_end         int    `json:"scan_end"`
	Haskb            bool   `json:"haskb"`
	Hasaudittrail    bool   `json:"hasaudittrail"`
	Scanner_start    string `json:"scanner_start,omitempty"`
	Scanner_end      string `json:"scanner_end,omitempty"`
}

type hostsStruct struct {
	Asset_id              int                 `json:"asset_id"`
	Host_id               int                 `json:"host_id"`
	UUID                  string              `json:"uuid"`
	Hostname              string              `json:"hostname"`
	Progress              string              `json:"progress"`
	Scanprogresscurrent   int                 `json:"scanprogresscurrent"`
	Scanprogresstotal     int                 `json:"scanprogresstotal"`
	Numchecksconsidered   int                 `json:"numchecksconsidered"`
	Totalchecksconsidered int                 `json:"totalchecksconsidered"`
	Severitycount         severityCountStruct `json:"severitycount"`
	Severity              int                 `json:"severity"`
	Score                 int                 `json:"score"`
	Info                  int                 `json:"info"`
	Low                   int                 `json:"low"`
	Medium                int                 `json:"medium"`
	High                  int                 `json:"high"`
	Critical              int                 `json:"critical"`
	Host_index            int                 `json:"host_index"`
}

type severityCountStruct struct {
	Item []severityCountItemStruct `json:"item"`
}

type severityCountItemStruct struct {
	Count         int `json:"count"`
	SeverityLevel int `json:"severitylevel"`
}

type vulnStruct struct {
	Count         int    `json:"count"`
	Plugin_id     int    `json:"plugin_id"`
	Plugin_name   string `json:"plugin_name"`
	Severity      int    `json:"severity"`
	Plugin_family string `json:"plugin_family"`
	Vuln_index    int    `json:"vuln_index"`
}

type filtersStruct struct {
	Name          string        `json:"name"`
	Readable_name string        `json:"readable_name"`
	Control       controlStruct `json:"control"`
	Operators     []string      `json:"operators"`
	Group_name    string        `json:"group_name"`
}

type controlStruct struct {
	Type           string              `json:"type"`
	Regex          string              `json:"regex,omitempty"`
	Readable_regex string              `json:"readable_regex,omitempty"`
	List           []listControlStruct `json:"list,omitempty"`
}

type listControlStruct struct {
	NameList string `json:name,omitempty`
	IDList   int    `json:id,omitempty`
}

type historyStruct struct {
	History_id             int    `json:"history_id"`
	Owner_id               int    `json:"owner_id"`
	Creation_date          int    `json:"creation_date"`
	Last_modification_date int    `json:"last_modification_date"`
	Uuid                   string `json:"uuid"`
	Type                   string `json:"type"`
	Status                 string `json:"status"`
	Scheduler              int    `json:"scheduler"`
	Alt_targets_used       bool   `json:"alt_targets_used"`
	Is_archived            bool   `json:"is_archived"`
}

type notesStruct struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type remediationsStruct struct {
	Num_CVEs            int                   `json:"num_cves"`
	Num_hosts           int                   `json:"numb_hosts"`
	Num_remediated_cves int                   `json:"num_remediated_cves"`
	Num_impacted_hosts  int                   `json:"num_impacted_hosts"`
	RRemediations       []rRemediationsStruct `json:"remediations"`
}

type rRemediationsStruct struct {
	Vulns          int    `json:"vulns"`
	Value          string `json:"value"`
	Hosts          int    `json:"hosts"`
	RemDescription string `json:"remediation"`
}

// *************************   End Scan Details Structure *************************

func checkError(reason string, err error) {
	if err != nil {
		fmt.Printf("%s...\n", reason)
		fmt.Printf("%s", err)
		os.Exit(0)
	}
}

// Use the struct to read the JSON from the following URL
// URL: "https://cloud.tenable.com/scans?folder_id=12&last_modification_date=1771537711"
func ScanParser(httpResponseBody []byte) *ScanStruct {
	var scanInformation ScanStruct
	//json.NewDecoder(httpResponseBody).Decode(&scanInformation)
	json.Unmarshal(httpResponseBody, &scanInformation)
	return &scanInformation
}

// func ScanDetailsParser(httpResponseBody io.Reader) *ScanDetailsStruct {
// Changed the function to read in the byte array instead of the actual httpResponse.Body // Seems to work better in the parsing...
func ScanDetailsParser(httpResponseBody []byte) *ScanDetailsStruct {
	var scanDetails ScanDetailsStruct
	json.Unmarshal(httpResponseBody, &scanDetails)
	//json.NewDecoder(result).Decode(&scanDetails)
	//json.NewDecoder(httpResponseBody).Decode(&scanDetails)
	return &scanDetails
}
