package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// Command config
var exportSuppliersCmd = &cobra.Command{
	Use:   "export-suppliers",
	Short: "Export Suppliers ",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-suppliers -d DOMAINPREFIX -t TOKEN")),

	Run: func(cmd *cobra.Command, args []string) {
		getAllSuppliers()
	},
}

func init() {
	rootCmd.AddCommand(exportSuppliersCmd)
}

func getAllSuppliers() {

	// Create new Vend Client
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Suppliers.
	fmt.Println("Retrieving Suppliers from Vend...")
	suppliers, err := vc.Suppliers()
	if err != nil {
		log.Fatalf("Failed while retrieving Suppliers: %v", err)
	}

	// Write Suppliers to CSV
	fmt.Println("Writing Suppliers to CSV file...")
	err = sWriteFile(suppliers)
	if err != nil {
		log.Fatalf("Failed while writing Suppliers to CSV: %v", err)
	}

	fmt.Println("Finished!")
}

// WriteFile writes suppliers info to file.
func sWriteFile(suppliers []vend.SupplierBase) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_supplier_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatalf("Failed while creating CSV %v", err)
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "name")
	header = append(header, "description")
	header = append(header, "first_name")
	header = append(header, "last_name")
	header = append(header, "company_name")
	header = append(header, "phone")
	header = append(header, "mobile")
	header = append(header, "fax")
	header = append(header, "email")
	header = append(header, "twitter")
	header = append(header, "website")
	header = append(header, "physical_address1")
	header = append(header, "physical_address2")
	header = append(header, "physical_suburb")
	header = append(header, "physical_city")
	header = append(header, "physical_postcode")
	header = append(header, "physical_state")
	header = append(header, "physical_country_id")
	header = append(header, "postal_address1")
	header = append(header, "postal_address2")
	header = append(header, "postal_suburb")
	header = append(header, "postal_city")
	header = append(header, "postal_postcode")
	header = append(header, "postal_state")
	header = append(header, "postal_country_id")

	// Commit the header.
	writer.Write(header)

	// Now loop through each supplier object and populate the CSV.
	for _, supplier := range suppliers {

		var name, description, firstName, lastName, email, companyName, twitter, phone, mobile, fax, website,
			physicalSuburb, physicalCity, physicalPostcode, physicalState, postalSuburb, postalCity, postalState,
			postalPostcode, physicalAddress1, physicalAddress2, postalAddress1, postalAddress2, postalCountryID string

		if supplier.Name != nil {
			name = *supplier.Name
		}
		if supplier.Description != nil {
			description = *supplier.Description
		}
		if supplier.Contact.FirstName != nil {
			firstName = *supplier.Contact.FirstName
		}
		if supplier.Contact.LastName != nil {
			lastName = *supplier.Contact.LastName
		}
		if supplier.Contact.Email != nil {
			email = *supplier.Contact.Email
		}
		if supplier.Contact.CompanyName != nil {
			companyName = *supplier.Contact.CompanyName
		}
		if supplier.Contact.Phone != nil {
			phone = *supplier.Contact.Phone
		}
		if supplier.Contact.Mobile != nil {
			mobile = *supplier.Contact.Mobile
		}
		if supplier.Contact.Fax != nil {
			fax = *supplier.Contact.Fax
		}
		if supplier.Contact.Twitter != nil {
			twitter = *supplier.Contact.Twitter
		}
		if supplier.Contact.Website != nil {
			website = *supplier.Contact.Website
		}
		if supplier.Contact.PhysicalSuburb != nil {
			physicalSuburb = *supplier.Contact.PhysicalSuburb
		}
		if supplier.Contact.PhysicalCity != nil {
			physicalCity = *supplier.Contact.PhysicalCity
		}
		if supplier.Contact.PhysicalPostcode != nil {
			physicalPostcode = *supplier.Contact.PhysicalPostcode
		}
		if supplier.Contact.PhysicalState != nil {
			physicalState = *supplier.Contact.PhysicalState
		}
		if supplier.Contact.PostalSuburb != nil {
			postalSuburb = *supplier.Contact.PostalSuburb
		}
		if supplier.Contact.PostalCity != nil {
			postalCity = *supplier.Contact.PostalCity
		}
		if supplier.Contact.PostalState != nil {
			postalState = *supplier.Contact.PostalState
		}
		if supplier.Contact.PostalPostcode != nil {
			postalPostcode = *supplier.Contact.PostalPostcode
		}
		if supplier.Contact.PhysicalAddress1 != nil {
			physicalAddress1 = *supplier.Contact.PhysicalAddress1
		}
		if supplier.Contact.PhysicalAddress2 != nil {
			physicalAddress2 = *supplier.Contact.PhysicalAddress2
		}
		if supplier.Contact.PostalAddress1 != nil {
			postalAddress1 = *supplier.Contact.PostalAddress1
		}
		if supplier.Contact.PostalAddress2 != nil {
			postalAddress2 = *supplier.Contact.PostalAddress2
		}
		if supplier.Contact.PostalCountryID != nil {
			postalCountryID = *supplier.Contact.PostalCountryID
		}

		var record []string
		record = append(record, name)
		record = append(record, description)
		record = append(record, firstName)
		record = append(record, lastName)
		record = append(record, companyName)
		record = append(record, phone)
		record = append(record, mobile)
		record = append(record, fax)
		record = append(record, email)
		record = append(record, twitter)
		record = append(record, website)
		record = append(record, physicalAddress1)
		record = append(record, physicalAddress2)
		record = append(record, physicalSuburb)
		record = append(record, physicalCity)
		record = append(record, physicalPostcode)
		record = append(record, physicalState)
		record = append(record, postalAddress1)
		record = append(record, postalAddress2)
		record = append(record, postalSuburb)
		record = append(record, postalCity)
		record = append(record, postalPostcode)
		record = append(record, postalState)
		record = append(record, postalCountryID)
		writer.Write(record)
	}

	writer.Flush()
	return err
}
