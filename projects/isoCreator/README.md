Creates ISO Files

You can specify the files to include in the ISO by using multiple "-f" flags.  The files do need to exist in a directory called isofiles.  You can specify the output filename, volume name and will compress to a zip file

Command Line Options
```txt
Usage of ./isoCreator.bin:
  -f value
    	Specify the files to place in the iso, more than one can be specified
  -output string
    	Name of the ISO Output File (default "output.iso")
  -volume string
    	Name of the Volume on the ISO (default "default")
  -z string
    	Compress the file as a zip with the provided filename
```

WARNING: At this time if you create the ISO on linux, the files you place in a directory are not accessible in windows.  Troubleshooting why this is occurring.

Future Enhancements
- Specify a directory and anything in the directory and recursively forward will be included in the ISO


![createISO.png](/projects/isoCreator/createISO.png)