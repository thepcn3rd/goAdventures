# HTML Smuggling

Create files that can be placed on a web server to conduct HTML Smuggling. The prog takes a file specified and will base64 encode it, detect its mime type and then auto-download the file as someone visits the site. This is created primarily for ISO files which if downloaded may work around the Mark of the Web control that is built into Windows 10+

Using the project you can specify the file that you are going to be base64 encoded and placed in javascript to then be downloaded if the web page is visited.  The -o flag indicates what filename will be seen as it is being downloaded.

The program creates 2 files in the output folder, index.html that contains the reference to the my.js file to load the necessary javascript to initiate the automatic download.

Command Line Options
```txt
Specify a file to include in the HTML for smuggling...
Usage of prog.bin:
  -i string
    	Specify the file that you are smuggling in HTML
  -o string
    	Specify the file that is downloaded by the browser (default "info.iso")
```


![wolfWhite3.png](/images/wolfWhite3.png)
