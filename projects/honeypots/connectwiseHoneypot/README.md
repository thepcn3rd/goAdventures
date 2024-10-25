Created a honeypot to emulate the logon page of a Connectwise Server which has 2 recent vulnerabilities that would allow remote code execution and path traversal.  CVE-2024-1709 and CVE-2024-1708.

You can customize the honeypot with the following options:

```text
Usage 
  -https string
        Disable site from running HTTPS (default: -https true) (default "true")
  -port string
        Change default listening port (default: -port 9000) (default "9000")
  -user string
        Change default user (default: -user nobody) (default "nobody")
```

Current command I am using to execute the honeypot and use tee for logging.  I also combine it with screen or tmux to background the stdout.  Any POST information will be saved to a random file in the folder postInfo.

```bash
./ihp -user nobody -port 443 | tee -a honeypotOutput.txt &
```


![Inside Beehive](/projects/honeypots/connectwiseHoneypot/insideBeehive.png)