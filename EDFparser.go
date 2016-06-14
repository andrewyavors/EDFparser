package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

type HeaderRecord struct {
	PatientID    string
	RecordID     string
	StartDate    string
	StartTime    string
	NumOfBytes   int      `json:"-"`
	NumOfDR      int      `json:"Records"`    // Number of data records
	DurationOfDR int      `json:"Duration"`   // Duration of data record, in seconds
	NumOfSignals int      `json:"Signals"`    // Number of signals in data record
	NumOfSamples []int    `json:"Samples"`    // Number of samples in each data record
	Label        []string `json:"Labels"`     // Label of each signal
	TrType       []string `json:"Transducer"` // Transducer type
	PhDim        []string // Physical dimension
	PhMin        []string // Physical minimum
	PhMax        []string // Physical maximum
	DigMin       []string // Digital minimum
	DigMax       []string // Digital maximum
	Prefiltering []string
}

// Reads header record from source file and write it into the structure
func (header *HeaderRecord) ReadHeader(source *os.File) {
	if !(bytes.Equal(bytes.TrimSpace(readBytes(source, 8, 0, 0)), []byte{48})) {
		log.Fatal("File format is not valid.", readBytes(source, 8, 0, 0))
	}
	header.PatientID = string(bytes.TrimSpace(readBytes(source, 80, 8, 0)))
	header.RecordID = string(bytes.TrimSpace(readBytes(source, 80, 0, 1)))
	header.StartDate = string(bytes.TrimSpace(readBytes(source, 8, 0, 1)))
	header.StartTime = string(bytes.TrimSpace(readBytes(source, 8, 0, 1)))
	header.NumOfBytes, _ = strconv.Atoi(string(bytes.TrimSpace(readBytes(source, 8, 0, 1))))
	header.NumOfDR, _ = strconv.Atoi(string(bytes.TrimSpace(readBytes(source, 8, 44, 1))))
	header.DurationOfDR, _ = strconv.Atoi(string(bytes.TrimSpace(readBytes(source, 8, 0, 1))))
	header.NumOfSignals, _ = strconv.Atoi(string(bytes.TrimSpace(readBytes(source, 4, 0, 1))))

	for ns := 0; ns < header.NumOfSignals; ns++ {
		header.Label = append(header.Label, string(bytes.TrimSpace(readBytes(source, 16, int64(256+ns*16), 0))))
		header.TrType = append(header.TrType, string(bytes.TrimSpace(readBytes(source, 80, int64(256+header.NumOfSignals*16+ns*80), 0))))
		header.PhDim = append(header.PhDim, string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*96+ns*8), 0))))
		header.PhMin = append(header.PhMin, string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*104+ns*8), 0))))
		header.PhMax = append(header.PhMax, string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*112+ns*8), 0))))
		header.DigMin = append(header.DigMin, string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*120+ns*8), 0))))
		header.DigMax = append(header.DigMax, string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*128+ns*8), 0))))
		header.Prefiltering = append(header.Prefiltering, string(bytes.TrimSpace(readBytes(source, 80, int64(256+header.NumOfSignals*136+ns*80), 0))))
		tmp, _ := strconv.Atoi(string(bytes.TrimSpace(readBytes(source, 8, int64(256+header.NumOfSignals*216+ns*8), 0))))
		header.NumOfSamples = append(header.NumOfSamples, tmp)
	}
}

// Converts header structure to .json file
func (header *HeaderRecord) HeaderToJSON() {
	jsonHeader, _ := json.MarshalIndent(header, "", "    ")
	result, err := os.Create("header_" + header.StartDate + "_" + header.StartTime + ".json")
	if err != nil {
		log.Fatal("Creating JSON file failed: ", err)
	}
	defer result.Close()
	result.Write(jsonHeader)
}

// Converts record data from source file to .csv file
func (header *HeaderRecord) DataToCSV(source *os.File, showLabels *bool) {
	if _, err := source.Seek(int64(header.NumOfBytes), 0); err != nil {
		log.Fatal("Seeking failed: ", err)
	}

	// String array of samples. Row - signals, column - number of sample
	record := make([][]string, header.NumOfSignals)
	for row := range record {
		if *showLabels == true {
			record[row] = make([]string, header.NumOfDR*header.NumOfSamples[row]+1)
			record[row][0] = header.Label[row]
		} else {
			record[row] = make([]string, header.NumOfDR*header.NumOfSamples[row])
		}
	}

	// Filling array
	var tmp int16
	for DR := 0; DR < header.NumOfDR; DR++ {
		for row := range record {
			if *showLabels == true {
				for col := header.NumOfSamples[row]*DR + 1; col < (DR+1)*header.NumOfSamples[row]; col++ {
					if err := binary.Read(bytes.NewReader(readBytes(source, 2, 0, 1)), binary.LittleEndian, &tmp); err != nil {
						log.Fatal("Converting bytes to int16 failed: ", err)
					}
					record[row][col] = strconv.Itoa(int(tmp))
				}

			} else {
				for col := header.NumOfSamples[row] * DR; col < (DR+1)*header.NumOfSamples[row]; col++ {
					if err := binary.Read(bytes.NewReader(readBytes(source, 2, 0, 1)), binary.LittleEndian, &tmp); err != nil {
						log.Fatal("Converting bytes to int16 failed: ", err)
					}
					record[row][col] = strconv.Itoa(int(tmp))
				}
			}
		}
	}

	result, err := os.Create("data_" + header.StartDate + "_" + header.StartTime + ".csv")
	if err != nil {
		log.Fatal("Creating CSV file failed: ", err)
	}
	defer result.Close()
	w := csv.NewWriter(result)
	if err = w.WriteAll(record); err != nil {
		log.Fatal("Error writing records to csv:", err)
	}

}

//Read numToRead bytes from source file with some offset (More about Seek func: https://golang.org/pkg/os/#File.Seek)
func readBytes(source *os.File, numToRead int, seekOffset int64, seekWhence int) []byte {
	if _, err := source.Seek(seekOffset, seekWhence); err != nil {
		log.Fatal("Seeking failed: ", err)
	}

	bytesSlice := make([]byte, numToRead)
	if _, err := source.Read(bytesSlice); err != nil {
		log.Fatal("Reading from file failed: ", err)
	}
	return bytesSlice
}

func main() {
	fPath := flag.String("f", "source.edf", "Path to source edf/edf+ file")
	whatToRecord := flag.Int("r", 2, "Save only header record - 0, only data record - 1, both - 2")
	showLabels := flag.Bool("l", true, "Data record without labels - false")
	flag.Parse()

	source, err := os.Open(*fPath)
	if err != nil {
		log.Fatal("Opening file failed: ", err)
	}
	defer source.Close()

	header := HeaderRecord{}
	header.ReadHeader(source)

	if *whatToRecord == 0 || *whatToRecord == 2 {
		header.HeaderToJSON()
	}
	if *whatToRecord == 1 || *whatToRecord == 2 {
		header.DataToCSV(source, showLabels)
	}
}
