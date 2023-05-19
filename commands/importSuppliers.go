package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

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
		log.Printf("Couldnt read Supplier CSV file, %s", err)
		panic(vend.Exit{1})
	}

	// Post Suppliers to Vend
	err = postSuppliers(suppliers)
	if err != nil {
		log.Printf("Failed to post Suppliers, %s", err)
		panic(vend.Exit{1})
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

// Read passed CSV, returns a slice of suppliers
func readSupplierCSV(filePath string) ([]vend.SupplierBase, error) {

	headers := []string{"name", "description", "first_name", "last_name", "company_name",
		"phone", "mobile", "fax", "email", "twitter", "website", "physical_address1",
		"physical_address2", "physical_suburb", "physical_city",
		"physical_postcode", "physical_state", "physical_country_id",
		"postal_address1", "postal_address2", "postal_suburb", "postal_city",
		"postal_postcode", "postal_state", "postal_country_id"}

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		fmt.Println("Could not read from CSV file")
		return nil, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			fmt.Println("Found error in header rows.")
			log.Printf("No header match for: %s Instead got: %s.",
				string(headers[i]), string(headerRow[i]))
			panic(vend.Exit{1})

		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()

	var suppliers []vend.SupplierBase

	// Loop through rows and assign them to supplier type.
	for _, row := range rawData {
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

	return suppliers, err
}

// Post each Supplier to Vend
func postSuppliers(suppliers []vend.SupplierBase) error {
	var err error

	// Posting Suppliers to Vend
	fmt.Printf("%d Suppliers to post.\n \n", len(suppliers))
	for _, supplier := range suppliers {
		fmt.Printf("Posting: %v \n", *supplier.Name)
		// Create the Vend URL
		url := fmt.Sprintf("https://%s.vendhq.com/api/supplier", DomainPrefix)

		// Make the request to Vend
		res, err := vendClient.MakeRequest("POST", url, supplier)
		if err != nil {
			return fmt.Errorf("Something went wrong trying to post supplier: %s, %s", err, string(res))
		}

	}
	fmt.Printf("\nFinished! Succesfully created %d Suppliers", len(suppliers))

	return err
}
