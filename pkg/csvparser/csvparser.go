package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
)

// writes the error data to a CSV file. It takes a struct with all fields as strings
func WriteErrorCSV(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", filename)
	}
	defer file.Close()

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("data is not a slice")
	}

	if val.Len() == 0 {
		return nil
	}

	// write the header. We are using the fields of the struct as the header
	elemType := val.Index(0).Type()
	headers := make([]string, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		headers[i] = elemType.Field(i).Name
	}

	writer := csv.NewWriter(file)
	err = writer.Write(headers)
	if err != nil {
		return fmt.Errorf("error writing header to file: %w", err)
	}

	// write the data
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		record := make([]string, elemType.NumField())
		for j := 0; j < elemType.NumField(); j++ {
			field := elem.Field(j)
			// all fields must be strings for this function to work
			if field.Kind() != reflect.String {
				return fmt.Errorf("field %s is not a string", headers[j])
			}
			record[j] = field.String()
		}
		err := writer.Write(record)
		if err != nil {
			return fmt.Errorf("error writing record to file: %w", err)
		}
	}
	writer.Flush()

	return nil
}
