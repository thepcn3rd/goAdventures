
# APT38 

North Korean state-sponsored group that specializes in financial cyber operations.

Crowdstrike "Global Threat Report" for 2023, "North Korean adversaries maintained a consistently high tempo throughout 2023. Their activity continued to focus on financial gain via cryptocurrency theft and intelligence collection from South Korean and Western organizations, specifically in the academic, aerospace, defense, government, manufacturing, media and technology sectors."

CISA Cybersecurity Advisory AA22-108A "The U.S. government has also previously published advisories about North Korean state-sponsored cyber actors stealing money from banks using custom malware"

CISA Cybersecurity Advisory AA20-239A "North Korea's intelligence apparatus controls a hacking team dedicated to robbing banks through remote internet access. To differentiate methods from other North Korean malicious cyber activity, the U.S. Government refers to this team as BeagleBoyz, who represent a subset of HIDDEN COBRA activity. The BeagleBoyz overlap to varying degrees with groups tracked by the cybersecurity industry as Lazarus, Advanced Persistent Threat 38 (APT38), Bluenoroff, and Stardust Chollima and are responsible for the FASTCash ATM cash outs reported in October 2018, fraudulent abuse of compromised bank-operated SWIFT system endpoints since at least 2015, and lucrative cryptocurrency thefts."

UN Security Council S/2023/171 "BlueNoroff, known as a subgroup of Lazarus, was observed by a cybersecurity company renewing attacks that use new malware and updated delivery techniques, including new file types and a method of bypassing Microsoft’s Mark-of-the-Web (MotW) protections. BlueNoroff distributed optical disk image (.iso) and virtual hard disk (.vhd) files containing decoy Microsoft Office documents. This allowed them to
avoid the MotW warning that Windows typically displays when a user attempts to open a document downloaded from the Internet. The company assessed that, through phishing, BlueNoroff attempted to infect target organizations in order to intercept cryptocurrency transfers and drain accounts. In addition, as part of the campaign, the hacking group registered fake domains mimicking well-known banks and venture capital firms"

Recorded Future, "The ZIP files contained an encrypted PDF document alongside a double extension file called “Password.txt.lnk” used to trick the victim into clicking it in order to get the password for the encrypted PDF file, but it instead launches either “pcalua.exe” or “mshta.exe”, performing an indirect command execution technique"
## Tactics and Techniques

Spearphishing Messages with msi, vhd, iso and compiled HTML (chm)

Malware Applications referred to as "TraderTraitor" written in cross-platform JavaScript code with Node.js runtime environment using the Electron framework

Masquerading: Double File Extension - T1036.007

Indirect Command Execution - T1202

Clear Windows Event Logs - T1070.001





## Threat Group Names (aka)
Nickel Gladstone - Secureworks
BeagleBoyz - US Government Reference
Bluenoroff - Kaspersky
Stardust Chollima (APT38) - Crowdstrike
Labyrinth Chollima - Crowdstrike 
Chollima - Crowdstrike
Sapphire Sleet - Microsoft
Copernicium - Microsoft
Lazarus Group
TAG-71 - Overlaps with APT38 - Recorded Future
APT38 - Mandiant
CTG-6459 - Secureworks
Temp.Hermit - Fireeye
T-APT-15 - Tencent
ATK 117 - Thales
Black Alicanto - PWC
TA444 - Proofpoint
Group 77 - Talos
Hidden Cobra - TrendMicro
SectorA01 - Threat Recon




## References

https://attack.mitre.org/groups/G0082/ - MITRE Attack Information about the group

https://www.cisa.gov/news-events/cybersecurity-advisories/aa22-108a - CISA Cybersecurity Advisory - April 20, 2022

https://documents.un.org/doc/undoc/gen/n23/037/94/pdf/n2303794.pdf - UN Security Council Document - March 2023

https://documents.un.org/doc/undoc/gen/n24/032/68/pdf/n2403268.pdf - UN Security Council Document - March 2024


https://securelist.com/bluenoroff-methods-bypass-motw/108383 - Technical Document of how Bluenoroff was conducting the attacks

https://hub.elliptic.co/analysis/has-a-sanctioned-bitcoin-mixer-been-resurrected-to-aid-north-korea-s-lazarus-group 

http://www.recordedfuture.com/north-korea-aligned-tag-71-spoofs-financial-institutions

https://go.recordedfuture.com/hubfs/reports/cta-2023-0606.pdf

https://blog.qualys.com/vulnerabilities-threat-research/2022/02/08/lolzarus-lazarus-group-incorporating-lolbins-into-campaigns

https://apt.etda.or.th/cgi-bin/showcard.cgi?g=Subgroup%3A%20Bluenoroff%2C%20APT%2038%2C%20Stardust%20Chollima&n=1

https://blog.polyswarm.io/2023-recap-threat-actor-activity-highlights-north-korea

https://www.proofpoint.com/us/blog/threat-insight/ta444-apt-startup-aimed-at-your-funds

https://www.mandiant.com/sites/default/files/2021-09/rpt-apt38-2018-web_v5-1.pdf

