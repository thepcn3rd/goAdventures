# Overview
MITRE ATTACK
Tactic: Execution
Technique ID: T1204.T1204.002
Technique: User Execution
Sub-Technique: Malicious File

tags: #execution #powershell #lnkfiles #windows 

# Description

Create a lnk file that embeds a payload in the arguments section of the LNK file.  This is an attack vector to hide payloads.

# References:
URL: https://www.trendmicro.com/en_ph/research/17/e/rising-trend-attackers-using-lnk-files-download-malware.html

URL: https://gist.github.com/vector-sec/05733c8758005cd6c4da2e1926917dce # Create lnk files with a payload for powershell

URL: https://github.com/hexachordanu/Red-Team-Essentials/blob/master/Generate-LNK.ps1

https://winprotocoldoc.blob.core.windows.net/productionwindowsarchives/MS-SHLLINK/%5bMS-SHLLINK%5d.pdf

https://github.com/gotopkg/mslnk/tree/master

https://github.com/libyal/liblnk/blob/main/documentation/Windows%20Shortcut%20File%20(LNK)%20format.asciidoc

https://github.com/parsiya/golnk/tree/master

# Creation of LNK Files

### Powershell

```powershell
$shellObject = New-Object -ComObject WScript.Shell
# Specifically creating the extension so that it can be copied
$newLNK = $shellObject.CreateShortcut("c:\users\administrator\desktop\new.lnk") 
$newLNK.TargetPath = "C:\windows\system32\cmd.exe"
$newLNK.WindowStyle = 1 # Could be 3 or 7 refer to specs of LNK from Microsoft v5
$newLNK.IconLocation = "C:\windows\system32\cmd.exe"
$newLNK.WorkingDirectory = "c:\users\public\"

# The arguments can be upwards of 4096 characters the below
# generates a string that is 4000 characters long to embed in the arguments section
$charString = -join ((65..90) + (97..122) | Get-Random -Count 4000 | ForEach-Object {[char]$_})
$newLNK.Arguments = $charString
$newLNK.Description = "This is not the droid you are looking for..."
$newLNK.Save()
```

Copy a LNK File from Windows

```txt

1. Goto registry key for HKEY_CLASSES_ROOT find LNK

2. Remove the registry entry of "Computer\HKEY_CLASSES_ROOT\lnkfile\NeverShowExt REG_SZ" - Remember to place back if you do not want the extension to show...

3. Reboot and the extension of .lnk will show

4. Rename the file to .txt or an extension of choice and then it can be copied to another computer.
```

