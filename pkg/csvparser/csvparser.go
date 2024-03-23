package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"

	pbar "github.com/vend/vend-cli/pkg/progressbar"
)

// writes the error data to a CSV file. It takes a struct with all fields as strings
func WriteErrorCSV(filename string, data interface{}) error {

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return fmt.Errorf("data is not a slice and empty")
	}

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(val.Len(), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return fmt.Errorf("failed to create file: %s", filename)
	}
	defer file.Close()

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
		bar.Increment()
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
	p.Wait()

	return nil
}

// readIdCSV reads a CSV that is just ids with no header
func ReadIdCSV(FilePath string) ([]string, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("Error creating progress bar:%s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file
	file, err := os.Open(FilePath)
	if err != nil {
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	// Make sure to close the file
	defer file.Close()

	// Create CSV read on our file
	reader := csv.NewReader(file)

	// Read the rest of the data from the CSV
	rows, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	var rowNumber int
	entities := []string{}

	// Loop through rows and assign them
	for _, row := range rows {
		rowNumber++
		entityIDs := row[0]
		entities = append(entities, entityIDs)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return entities, err
}
