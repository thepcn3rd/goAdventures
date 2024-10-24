# Fake Ransomware

The fakeRansomware was built to test the controls of EDR and if it detected the variant.  With the EDRs and AVs tested, they were successful in detecting the activities.  This program requires 1 parameter and that is the path to encrypt:

```txt
-path string
    	Encrypt the files in the provided path (default "/home/thepcn3rd/go/workspaces/encryptor/testFiles/")
```

To encrypt a windows path you do need to escape the slashes
```txt
.\fakeRansomware.exe -path "c:\\temp\\encryptor\\testfiles\\"
```

The objective of the ransomware is to encrypt the fake data that is generated. 

![wolfWhite.png](/images/wolfWhite.png)