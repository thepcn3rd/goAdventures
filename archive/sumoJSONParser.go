package sumoJSONParser

// Belongs to the sumoSearch.go file

import (
	"encoding/json"
	"time"
)

// /////////////////////////  Search Jobs Struct
type SearchJobsStruct struct {
	ID   string         `json:"id"`
	Link searchJobsLink `json:"link"`
}

type searchJobsLink struct {
	Relation string `json:"rel"`
	HREF     string `json:"href"`
}

// //////////////////////// User Information Struct
type UserInfoStruct struct {
	Data []dataStruct `json:"data"`
}

type dataStruct struct {
	FirstName          string    `json:"firstName"`
	LastName           string    `json:"lastName"`
	Email              string    `json:"email"`
	RoleIDS            []string  `json:"roleIds"`
	CreatedAt          string    `json:"createdAt"`
	CreatedBy          string    `json:"createdBy"`
	ModifiedAt         string    `json:"modifiedAt"`
	ModifiedBy         string    `json:"modifiedBy"`
	ID                 string    `json:"id"`
	IsActive           bool      `json:"isActive"`
	IsLocked           bool      `json:"isLocked"`
	IsMFAEnabled       bool      `json:"isMfaEnabled"`
	LastLoginTimestamp time.Time `json:"lastLoginTimestamp"`
}

// ////////////////// Poll Jobs Struct
type PollJobsStruct struct {
	State            string             `json:"state"`
	HistogramBuckets []string           `json:"histogramBuckets"`
	MessageCount     int                `json:"messageCount"`
	RecordCount      int                `json:"recordCount"`
	PendingWarnings  []string           `json:"pendingWarnings"`
	PendingErrors    []string           `json:"pendingErrors"`
	UsageDetails     usageDetailsStruct `json:"usageDetails"`
}

type usageDetailsStruct struct {
	DataScannedInBytes int `json:"dataScannedInBytes"`
}

// //////////////////////// Message Struct
type MessageStruct struct {
	Fields   []fieldsInfoStruct   `json:"fields"`
	Messages []messagesInfoStruct `json:"messages"`
}

type fieldsInfoStruct struct {
	Name      string `json:"name"`
	FieldType string `json:"fieldType"`
	KeyField  bool   `json:"keyField"`
}

type messagesInfoStruct struct {
	Map mapInfoStruct `json:"map"`
}

type mapInfoStruct struct {
	Action         string `json:"action"`
	App            string `json:"app"`
	BlockID        string `json:"_blockid"`
	Collector      string `json:"_collector"`
	CollectorID    string `json:"_collectorid"`
	DestIP         string `json:"dest_ip"`
	DestPort       string `json:"dest_port"`
	DestUser       string `json:"dest_user"`
	DestZone       string `json:"dest_zone"`
	Format         string `json:"_format"`
	LogLevel       string `json:"_loglevel"`
	MessageCount   string `json:"_messagecount"`
	MessageID      string `json:"_messageid"`
	MessageTime    string `json:"_messagetime"`
	Packets        string `json:"packets"`
	Parser         string `json:"_parser"`
	Raw            string `json:"_raw"`
	ReceiptTime    string `json:"_receipttime"`
	Rulename       string `json:"rulename"`
	SIEMForward    string `json:"_siemforward"`
	Size           string `json:"_size"`
	Source         string `json:"_source"`
	SourceCategory string `json:"_sourcecategory"`
	SourceHost     string `json:"_sourcehost"`
	SourceID       string `json:"_sourceid"`
	SourceIP       string `json:"src_ip"`
	SourceName     string `json:"_sourcename"`
	SourcePort     string `json:"src_port"`
	SourceUser     string `json:"src_user"`
	SourceZone     string `json:"src_zone"`
	Starttime      string `json:"starttime"`
	Type           string `json:"type"`
	URL_IDX        string `json:"url_idx"`
	View           string `json:"_view"`
}

func ScanUserList(httpResponseBody []byte) *UserInfoStruct {
	var userInfo UserInfoStruct
	json.Unmarshal(httpResponseBody, &userInfo)
	return &userInfo
}

func SearchJobs(httpResponseBody []byte) *SearchJobsStruct {
	var jobsInfo SearchJobsStruct
	json.Unmarshal(httpResponseBody, &jobsInfo)
	return &jobsInfo
}

func PollJobs(httpResponseBody []byte) *PollJobsStruct {
	var jobsInfo PollJobsStruct
	json.Unmarshal(httpResponseBody, &jobsInfo)
	return &jobsInfo
}

func PullMessages(httpResponseBody []byte) *MessageStruct {
	var messageInfo MessageStruct
	json.Unmarshal(httpResponseBody, &messageInfo)
	return &messageInfo
}
