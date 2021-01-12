package cmd

import (
	"bytes"
	"os"
	"fmt"
	
	"io/ioutil"
	"net/http"
	"encoding/csv"
)

// loadRecordsFromCSV reads the content of a csv file and returns headers and records.
func loadRecordsFromCSV(path string) ([]string, [][]string, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Could not read from CSV file")
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

// writeCSV combines the headers and rows to create a csv file.
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
