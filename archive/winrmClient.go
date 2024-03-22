package main

// Setup the following for the application
/*

go env -w GOROOT="/usr/lib/go"
go env -w GOPATH="/home/thepcn3rd/go/workspaces/winrmClient"

// Install the dns dependency
// go get github.com/masterzen/winrm

To compile the project, verify the structure is the same as below
// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o winrmClient.bin -ldflags "-w -s" main.go

// To cross compile windows (Not tested...)
// GOOS=windows GOARCH=amd64 go build -o winrmClient.exe -ldflags "-w -s" main.go


References:
Base code created by ChatGPT 8/26/2023
https://github.com/masterzen/winrm

Future tasks:
Create flags for the endpoint, username, prompt for password, and then create a loop for the commands that are sent


*/

import (
	"context"
	"os"

	"github.com/masterzen/winrm"
)

func main() {

	ipAddr := "10.10.114.194"
	port := 5985
	username := "Administrator"
	password := "Password321"
	basicAuth := false
	ntlmAuth := true

	var client *winrm.Client
	var err error
	params := winrm.DefaultParameters
	// The below works with Basic Authentication Enabled
	// winrm set winrm/config/service/Auth '@{Basic="true"}'
	// winrm set winrm/config/service '@{AllowUnencrypted="true"}'
	endpoint := winrm.NewEndpoint(ipAddr, port, false, false, nil, nil, nil, 0)
	if basicAuth == true {
		client, err = winrm.NewClient(endpoint, username, password)
		if err != nil {
			panic(err)
		}

	}

	// The below works with NTLM Authentication but allow unencrypted needs to be false
	// winrm set winrm/config/service '@{AllowUnencrypted="true"}'
	if ntlmAuth == true {

		params.TransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }

		client, err = winrm.NewClientWithParameters(endpoint, username, password, params)
		if err != nil {
			panic(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client.RunWithContext(ctx, "ipconfig /all", os.Stdout, os.Stderr)

}
