# Fake Data Generator

The fakeDataGenerator was built to test data exfiltration controls.  The fake data generator generates the following:
1. Visa CC Numbers - The numbers do pass the Luhn algorithm
2. Expiration Date for a CC
3. CVV for a CC
4. ATM Pin
5. SSN
6. Date of Birth - You can specify in the configuration the age range
7. Random Full Name, First Name, Last Name and a Middle Initial
8. Phone Numbers
9.  Username
10. Emails
11. Mothers Maiden Name
12. Random Password

The updated fakeDataGenerator.go file has an accompanying config.json.  The json allows for customization of the output.  The customization includes the following:
* Specify output quantity in rows of data to generate
* Specify the bin number used in the generation
* Specify the expiration dates of the card (min, max)
* Specify the date of birth min and max
* Length of the Random Password
* Specify the format of the username, the following formats are allowed in the config.json
	* First and Last Name separated by a delimiter
	* First Initial and Last Name separated by a delimiter
	* First Name and Last Initial separated by a delimiter 
* Display the Field in the Output - Customize Output
* Customize the Order of Columns for the Output


![wolfCavern.png](/images/wolfCavern.png)

```txt
Usage of ./fakeDataGenerator.bin:
  -config string
        Configuration file to load for the proxy (default "config.json")
  -q string
        Default is to read quantity from the config.json (default "0")
```


Batch Script to Create 3.5GB and over 250 files

```cmd
mkdir information
mkdir information\archive2010
mkdir information\archive2000
mkdir information\archive

fakeDataGenerator.exe -q 10000 > information\list.csv

FOR /L %%X IN (0,1,3) DO fakeDataGenerator.exe -q 100000 > information\list_202%%X.csv

FOR /L %%X IN (0,1,9) DO fakeDataGenerator.exe -q 100000 > information\archive2010\list_201%%X.csv

FOR /L %%X IN (0,1,9) DO fakeDataGenerator.exe -q 100000 > information\archive2000\list_200%%X.csv

FOR /L %%X IN (0,1,250) DO fakeDataGenerator.exe -q 100000 > information\archive\list-%%X.csv
```

The [fakeDataGenerator_donut.go](/projects/fakeDataCreator/fakeDataGenerator_donut.go) has the batch script built into the go program.  The paths are hardcoded due to using it as shellcode used with a powershell script in this directory, [memoryInject.ps1](/projects/fakeDataCreator/memoryInject.ps1).

I also built a version that could be compiled into a DLL and used with rundll32.dll, [fakeDataGenerator_dll.go](/projects/fakeDataCreator/fakeDataGenerator_dll.go)

1/20/2024 - Added the Expiration Date for a CC, CVV and ATM Pin to fill a database of a scam that targeted my family...

1/23/2024 - Added the new config file, a random password, if the field is displayed in the output and if the column is output

3/17/2024 - Added the username field and being able to specify the format and delimiter