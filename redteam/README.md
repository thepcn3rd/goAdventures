# Operation "Build the House"

Operation "Build the House" is a nickname for building the architecture, initial stage preparation, exploitation, exfiltration of fake data and then using the fake data to test a simulation of Cl0p ransomware in a fictitious red team engagement.  

![mechawolf.png](/images/mechawolf.png)

Picture generated by Adobe Firefly

[Building the Environment](Sliver_BuildingEnv.md) - In this section I placed my notes on how I used AWS Lightsail to build the environment and ansible to configure it. 

[Sliver Stage Implant, AES Encryption and Initial Exploitation](Sliver_Payloads.md) - Created a go project to automate the creation of a stage implant using AES Encryption and then initial exploitation.  This also references the sources of where this was gathered

[Create ISO and Deliver Payload](/redteam/Deliver_Payload.md) - Recognizing that threat actors are using ISO files to bypass the Mark of the Web protection, created a go project to facilitate the creation of an ISO.  Then utilizing an LNK file embedding the powershell execution of the staged payload created above to establish a session with Sliver

[Generate Fake Data](/projects/fakeDataGenerator/README.md) - Testing data leak prevention (DLP) or similar controls, created this project to generate fake data that contains CC, SSN, Date of Birth, Names, and other fields.  Creates CSV files at the moment.

Used Sliver to test the exfiltration of the fake data.  

[Fake Ransomware](/projects/fakeRansomware/README.md) - Created a fake ransomware variant that provides the extension on the files for the C.l.0.p threat group with a respective ransom note left in the directory.  BE CAREFUL! No decryption is available!  The objective is to encrypt the fake data that was generated.

[Emulated Sysjoker C2](/projects/c2/README.md) - Created a c2 server and a payload that could be compiled for windows or linux.  The server executes commands as bash or cmd.exe or downloads files from the https server.


