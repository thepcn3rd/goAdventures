# Project 

Evaluates a LNK File

THIS PROJECT IS INCOMPLETE - However the lessons learned as I worked through what I worked through are beneficial to understaning LNK files.

The goals of the project was to create a project to read a LNK file and modify as needed or create a similar file. I was going to use the modified LNK files for testing controls at detecting and blocking LNK files in email and on the desktop.

Here is a way to create LNK files in powershell and be able to copy them to linux if you need them.  WARNING: Verify the meta-data created by powershell is either scrubbed with a hexeditor or anonymized.

# Powershell

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


