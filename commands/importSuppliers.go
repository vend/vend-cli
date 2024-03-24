package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedSupplierImportRequest struct {
	Name   string
	Reason string
}

// Command config
var importSuppliersCmd = &cobra.Command{
	Use:   "import-suppliers",
	Short: "Import Suppliers",
	Long: fmt.Sprintf(`
This tool requires the Supplier CSV template, you can download it here: http://bit.ly/vendclitemplates

Example:
%s`, color.GreenString("vendcli import-suppliers -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		importSuppliers()
	},
}

var failedSupplierImportRequests []FailedSupplierImportRequest

func init() {
	// Flags
	importSuppliersCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	importSuppliersCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(importSuppliersCmd)
}

func importSuppliers() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read Suppliers from CSV file
	fmt.Println("\nReading Supplier CSV...")
	suppliers, err := readSupplierCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("couldnt read Supplier CSV file, %s", err)
		messenger.ExitWithError(err)
	}

	// Post Suppliers to Vend
	fmt.Println("\nPosting Suppliers to Vend...")
	count, err := postSuppliers(suppliers)
	if err != nil {
		err = fmt.Errorf("failed to post Suppliers, %s", err)
		messenger.ExitWithError(err)
	}

	if len(failedSupplierImportRequests) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_import_suppliers_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedSupplierImportRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}
	fmt.Println(color.GreenString("\nFinished!🎉\nImported %d out of %d suppliers\n", count, len(suppliers)))
}

// Read passed CSV, returns a slice of suppliers
func readSupplierCSV(filePath string) ([]vend.SupplierBase, error) {

	headers := []string{"name", "description", "first_name", "last_name", "company_name",
		"phone", "mobile", "fax", "email", "twitter", "website", "physical_address1",
		"physical_address2", "physical_suburb", "physical_city",
		"physical_postcode", "physical_state", "physical_country_id",
		"postal_address1", "postal_address2", "postal_suburb", "postal_city",
		"postal_postcode", "postal_state", "postal_country_id"}

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file.
	file, err := os.Open(filePath)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return nil, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			bar.AbortBar()
			p.Wait()
			err = fmt.Errorf("found error in header rows. No header match for: %s Instead got: %s",
				string(headers[i]), string(headerRow[i]))
			return nil, err
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	var suppliers []vend.SupplierBase

	// Loop through rows and assign them to supplier type.
	for idx, row := range rawData {
		// Check if supplier name is empty
		if row[0] == "" {
			reason := "supplier name is empty"
			failedSupplierImportRequests = append(failedSupplierImportRequests, FailedSupplierImportRequest{
				Name:   fmt.Sprintf("Row %d", idx+1), // csvs are 1-indexed
				Reason: reason,
			})
			continue
		}

		supplier := vend.SupplierBase{
			Name:        &row[0],
			Description: &row[1],
			Contact: &vend.Contact{
				FirstName:         &row[2],
				LastName:          &row[3],
				CompanyName:       &row[4],
				Phone:             &row[5],
				Mobile:            &row[6],
				Fax:               &row[7],
				Email:             &row[8],
				Twitter:           &row[9],
				Website:           &row[10],
				PhysicalAddress1:  &row[11],
				PhysicalAddress2:  &row[12],
				PhysicalSuburb:    &row[13],
				PhysicalCity:      &row[14],
				PhysicalPostcode:  &row[15],
				PhysicalState:     &row[16],
				PhysicalCountryID: &row[17],
				PostalAddress1:    &row[18],
				PostalAddress2:    &row[19],
				PostalSuburb:      &row[20],
				PostalCity:        &row[21],
				PostalPostcode:    &row[22],
				PostalState:       &row[23],
				PostalCountryID:   &row[24],
			},
		}

		// Append each supplier type to our slice of all suppliers.
		suppliers = append(suppliers, supplier)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return suppliers, err
}

// Post each Supplier to Vend
func postSuppliers(suppliers []vend.SupplierBase) (int, error) {
	var err error

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(suppliers), "Posting Suppliers")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	// Posting Suppliers to Vend
	var count int = 0
	for _, supplier := range suppliers {
		bar.Increment()
		// Create the Vend URL
		url := fmt.Sprintf("https://%s.vendhq.com/api/supplier", DomainPrefix)

		// Make the request to Vend
		res, err := vendClient.MakeRequest("POST", url, supplier)
		if err != nil {
			err = fmt.Errorf("something went wrong trying to post supplier: %s, %s", err, string(res))
			failedSupplierImportRequests = append(failedSupplierImportRequests, FailedSupplierImportRequest{
				Name:   *supplier.Name,
				Reason: err.Error(),
			})
			continue
		}
		count += 1
	}
	p.Wait()
	return count, err
}
