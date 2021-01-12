package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
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
