The sliver payload generator outputs a script to create the following through a sliver client connected to the server:
- Staged Listener with a deliver mechanism for the shellcode
- Uses the server certificates created/specified
- Create the Listener with an AES Key and a random AES IV
- Creates a Payload for the Creation of a Session (Could be modified to a beacon)
- Generates a powershell file called pwsh.ps1 to be executed on the remote computer

Command Line Flags allow you to specify a config.json file
```txt
Usage of ./generator.bin:
  -config string
    	Configuration file to load for the proxy (default "config.json")
```

Below is a sample config file that would need to be created
- csharpFile - Is the shellcode that is requested through the stageListenerPort.  After the shellcode executes it will call back to the listener port to complete the establishment of the session. 
- The below makes an assumption that the sliver server will allow connection to an HTTP/S server.

```json
{
	"sliverServer": "sliver.server.local",
	"listenerPort" : "443",
	"stageListenerPort" : "80",
	"serverCRT": "keys/my.crt",
	"serverKey": "keys/my.key",
	"csharpFile": "mystuff.jpg",
	"csharpTargetBinary": "notepad.exe"
}
```

Future Enhancements
- Add a flag to create a beacon or a session with the payload