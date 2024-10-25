# FortiManager Honeypot 

#### FortiManager very-low interaction honeypot built to capture traffic related to recent Fortinet vulnerability: CVE-2024-47575 / FG-IR-24-423

#### Created: October 2024

This is a basic very low interaction honeypot to capture requests for the unauthenticated remote code execution vulnerability identified in CVE-2024-47575. The honeypot simulates a Fortinet login page that requires no interaction, designed to capture and log all incoming GET, POST, or other requests. Any POST requests received are recorded in both HEX and ASCII formats, similar to a hexdump, in the directory postInfo.

At Saintcon 2024, I learned about and explored using Go and launching within a Docker container. I created `runDocker.sh` and a `Dockerfile` to support launching this honeypot in Docker. This was purely an experiment to apply my learning and was not intended as a permanent hosting solution. The honeypot's logging system is not yet configured to operate within Docker. To reiterate, this was a practical exercise based on what I learned and observed. 

Also, due to pulling in the common functions that I use in goAdventures, I had to create a separate repo on github called goAdvsCommonFunctions.  The primary reason was speed to deploy the docker, because how I was doing it required the goAdventures repo to be downloaded.  Enjoy!

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

If user is allowed to run with the nobody account...
```bash
./ihp -user nobody -port 541 | tee -a honeypotOutput.txt &
```

Not recommended, running with sudo...
```bash
sudo ./ihp -user root -port 541 | tee -a honeypotOutput.txt &
```

Run code inside of a docker. Modify config as necessary...
```bash
./runDocker.sh
```



![Wolf Stuck in Honey](/projects/honeypots/fortimanagerHP/wolfStuckHoney.png)