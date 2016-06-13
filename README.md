# EDFparser
Parse .edf files to .json + .csv
_____
###Using:
**Command-line (for Windows):** `edfparser.exe -f "[Path to source edf file]" -r [Save only header record - 0, only data record - 1, both - 2] -l [Data record without labels - false]` <br />

**More info:** edfparser.exe -h

Or you can build EDFparser on your system:
`$ go build edfparser.go`

###Screenshots:
Header record (json):<br />
![Header record in .json file](https://raw.githubusercontent.com/SpycerLviv/EDFparser/master/screenshots/header.png)<br />
Data record (csv): <br />
![Data record in .csv file](https://raw.githubusercontent.com/SpycerLviv/EDFparser/master/screenshots/data.png)

###TODO: 
- Ability to convert multiple files
- EDF+ support


