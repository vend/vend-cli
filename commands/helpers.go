package cmd

import (
	"bytes"
	"fmt"
	"os"

	"encoding/csv"
	"io/ioutil"
	"net/http"
)

// loadRecordsFromCSV reads the content of a csv file and returns headers and records.
func loadRecordsFromCSV(path string) ([]string, [][]string, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return nil, nil, err
	}
	return readRecords(raw)
}

// readRecords converts a byte array to string slices representing headers and records
func readRecords(csvBytes []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	return header, records, nil
}

// writeCSV combines headers and rows to create a csv file.
func writeCSV(fileName string, headers []string, rows [][]string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	strWrite := [][]string{headers}
	strWrite = append(strWrite, rows...)

	_ = csvWriter.WriteAll(strWrite)
	csvWriter.Flush()
	return nil
}

// makeRequest a custom request call that returns status code and message
func makeRequest(method, url string, body interface{}) (int, string, error) {
	req, err := vendClient.NewRequest(method, url, body)
	if err != nil {
		return 0, "", err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error making request")
		return 0, "", err
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("\nError while reading response body: %s\n", err)
		return 0, "", err
	}
	return resp.StatusCode, string(responseBody), nil
}
