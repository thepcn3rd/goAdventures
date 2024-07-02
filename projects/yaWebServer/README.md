# Yet Another Web Server

Created this webserver to overcome a couple of issues I noticed with CTFs, vulnerable machines or competitions.  The issues I designed this to overcome are the following:
* Simple HTTP Server for Golang with TLS options
* Ability to upload files through a web browser from anywhere
* Ability to view markdown files
* Ability to on-the-request compile golang code and make it available
* Create notes that can be exchanged
* Password protect the functions that are available
* Specify a user account to runas
* Location to run offline cyber chef and revshells.com js pages

HTML Pages Available
* codelist.html - Lists the directory contents of the code directory, the directory would contain the original src files to be compiled
* compilecode.html - Will compile the specified code, places the compiled code into the downloads folder (http(s)://homepage/downloads/file.ext)
* library - View the md files that are available in the library directory
* notes - Allows you to create or view notes
* upload.html - Upload any file from a client or a web browser
* upload - Displays the successful upload of a file

![steamPunkWebServer.png](/projects/yaWebServer/steamPunkWebServer.png)