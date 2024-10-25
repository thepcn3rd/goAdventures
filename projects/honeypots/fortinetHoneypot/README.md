Due to the recent Fortinet Vulnerabilities that have been prevalent created this honeypot in golang.  Built similar to the [Ivanti Honeypot](/projects/honeypots/ivantiHoneypot/README.md), but headers are modified, the page returns a static index.html due to pictures being pulled and it reads the HTML file instead of having it embedded in HTML

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


![Bee Hive](/projects/honeypots/fortinetHoneypot/hivePicture.png)