# Operation "Leia o Livro"

A few of the programs that I created from reading, studying and learning go from "Blackhat Go".

![wolfHoodie1.png](/images/wolfHoodie1.png)

[First Program](/BHGO/chapter1/firstprog.go) - Holding on to this as a starting point with go

[Console Based Web Browser](/BHGO/chapter3/webRequest.go) - Needed a tool to quickly search a website with custom parameters that could be quickly modified

[Console Coloring for Linux](/BHGO/chapter3/color.go) - A simple project to allow coloring of text when run with linux.  This does not work when the go program is compiled with windows...

[DNS Client](/BHGO/dnsClient/main.go) - Simple proxy to query domains and reverse IP lookups.  You can also specify the DNS server and port to use.

[DNS Proxy](/BHGO/chapter5/main.go) - Simple Proxy to allow the sending of domains used in a penetration test to a custom DNS server for responses.

[DNS Server](/BHGO/chapter5/dnsServer/main.go) - Simple DNS Server, extended what the book had and added TXT records.  This could allow for sending information through TXT records.

[Find Daemon Permissions](/BHGO/findDaemonPermissions.go) - This tool runs on linux to identify the permissions that a process is running as, very simple, not really necessary but created it

[gocat Client](/BHGO/chapter2/gocatClient.go) - A featureless gocat client that is similar to netcat connecting to a TCP port

[gocat Server](/BHGO/chapter2/gocat.go) - A featureless gocat server that listens on a given TCP port

[Proxy TCP Connection](/BHGO/chapter2/proxyTCPClient.go) - Simple TCP relay from 1 connection to another

[Simple Port Scanner](/BHGO/chapter1/portScanner.go) - First port scanner to determine if a TCP port is accessible through a 3-way handshake

[Simple Web Server](/BHGO/chapter3/simpleWebServer.go) - A simple web server that runs on port 9000 and is setup to run with TLS.  Primary function is to allow the upload of files.

[Site Crawler](/BHGO/chapter3/siteCrawler.go) - Built to crawl a web page and extract links from HTML HREF, IMG and other tags

[Windows or Linux File Finder](/BHGO/fileReader.go) - A simple tool that can be used to identify if a file exists on a file system.  Use to search for files that have potential information in them helpful for a penetration test.

[Yet Another Port Scanner](/BHGO/chapter2/yaPortScanner.go) - Another port scanner with some added features to scan a range of IP Addresses and/or a range of ports.  Optimized for a range of IP Addresses

[Yet Another Reverse HTTP Proxy](/BHGO/chapter3/yaReverseHTTPProxy.go) - Simple reverse proxy that can be placed in-front of a TLS Server.  Other versions exist that I have improved.





