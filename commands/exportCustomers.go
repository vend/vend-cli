package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var exportCustomersCmd = &cobra.Command{
	Use:   "export-customers",
	Short: "Export Customers",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-customers -d DOMAINPREFIX -t TOKEN")),

	Run: func(cmd *cobra.Command, args []string) {
		getAllCustomers()
	},
}

func init() {
	rootCmd.AddCommand(exportCustomersCmd)
}

// Run executes the process of grabbing customers then writing them to CSV.
func getAllCustomers() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get customers.
	fmt.Println("\nRetrieving Customers from Vend...")
	customers, err := vc.Customers()
	if err != nil {
		log.Fatalf("Failed retrieving customers from Vend %v", err)
	}

	customerGroupMap, err := vc.CustomerGroups()
	if err != nil {
		log.Fatalf("Failed retrieving customer groups from Vend %v", err)
	}

	// Write Customers to CSV
	fmt.Println("Writing customers to CSV file...")
	err = cWriteFile(customers, customerGroupMap)
	if err != nil {
		log.Fatalf(color.RedString("Failed writing customers to CSV: %v", err))
	}

	fmt.Println(color.GreenString("\nExported %v customers  🎉\n", len(customers)))
}

// WriteFile writes customer info to file.
func cWriteFile(customers []vend.Customer, customerGroupMap map[string]string) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_customer_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		return err
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "id")
	header = append(header, "customer_code")
	header = append(header, "first_name")
	header = append(header, "last_name")
	header = append(header, "email")
	header = append(header, "customer_group")
	header = append(header, "year_to_date")
	header = append(header, "balance")
	header = append(header, "loyalty_balance")
	header = append(header, "note")
	header = append(header, "gender")
	header = append(header, "date_of_birth")
	header = append(header, "company_name")
	header = append(header, "phone")
	header = append(header, "mobile")
	header = append(header, "fax")
	header = append(header, "twitter")
	header = append(header, "website")
	header = append(header, "do_not_email")
	header = append(header, "created_at")
	header = append(header, "physical_address1")
	header = append(header, "physical_address2")
	header = append(header, "physical_suburb")
	header = append(header, "physical_city")
	header = append(header, "physical_postcode")
	header = append(header, "physical_state")
	header = append(header, "postal_address1")
	header = append(header, "postal_address2")
	header = append(header, "postal_suburb")
	header = append(header, "postal_city")
	header = append(header, "postal_postcode")
	header = append(header, "postal_state")
	header = append(header, "postal_country_id")
	header = append(header, "custom_field_1")
	header = append(header, "custom_field_2")
	header = append(header, "custom_field_3")
	header = append(header, "custom_field_4")

	// Commit the header.
	writer.Write(header)

	// Now loop through each customer object and populate the CSV.
	for _, customer := range customers {

		var id, code, firstName, lastName, email, customerGroup, yearToDate, balance, loyaltyBalance, note, gender, dateOfBirth, companyName, phone, mobile, fax, twitter,
			website, doNotEmail, physicalSuburb, physicalCity, physicalPostcode, physicalState, postalSuburb, postalCity, postalState, createdAt, postalPostcode, physicalAddress1, physicalAddress2, postalAddress1, postalAddress2, postalCountryID, customField1, customField2, customField3, customField4 string

		// Moving before ID since the loop can continue
		// after the anonymous code check
		if customer.Code != nil {
			// Moving in here to prevent seg fault
			if *customer.Code == "Anonymous Customer" {
				continue
			}
			code = *customer.Code
		}

		if customer.ID != nil {
			id = *customer.ID
		}

		if customer.FirstName != nil {
			firstName = *customer.FirstName
		}
		if customer.LastName != nil {
			lastName = *customer.LastName
		}
		if customer.Email != nil {
			email = *customer.Email
		}
		if customer.GroupId != nil {
			customerGroup = customerGroupMap[*customer.GroupId]
		}
		if customer.YearToDate != nil {
			yearToDate = fmt.Sprintf("%f", *customer.YearToDate)
		}
		if customer.Balance != nil {
			balance = fmt.Sprintf("%f", *customer.Balance)
		}
		if customer.LoyaltyBalance != nil {
			loyaltyBalance = fmt.Sprintf("%f", *customer.LoyaltyBalance)
		}
		if customer.Note != nil {
			note = *customer.Note
		}
		if customer.Gender != nil {
			gender = *customer.Gender
		}
		if customer.DateOfBirth != nil {
			dateOfBirth = *customer.DateOfBirth
		}
		if customer.CompanyName != nil {
			companyName = *customer.CompanyName
		}
		if customer.Phone != nil {
			phone = *customer.Phone
		}
		if customer.Mobile != nil {
			mobile = *customer.Mobile
		}
		if customer.Fax != nil {
			fax = *customer.Fax
		}
		if customer.Twitter != nil {
			twitter = *customer.Twitter
		}
		if customer.Website != nil {
			website = *customer.Website
		}
		if customer.DoNotEmail != nil {
			if *customer.DoNotEmail == false {
				doNotEmail = "0"
			} else if *customer.DoNotEmail == true {
				doNotEmail = "1"
			}
		}

		if customer.PhysicalSuburb != nil {
			physicalSuburb = *customer.PhysicalSuburb
		}
		if customer.PhysicalCity != nil {
			physicalCity = *customer.PhysicalCity
		}
		if customer.PhysicalPostcode != nil {
			physicalPostcode = *customer.PhysicalPostcode
		}
		if customer.PhysicalState != nil {
			physicalState = *customer.PhysicalState
		}
		if customer.PostalSuburb != nil {
			postalSuburb = *customer.PostalSuburb
		}
		if customer.PostalCity != nil {
			postalCity = *customer.PostalCity
		}
		if customer.PostalState != nil {
			postalState = *customer.PostalState
		}
		if customer.CreatedAt != nil {
			createdAt = *customer.CreatedAt
		}
		if customer.PostalPostcode != nil {
			postalPostcode = *customer.PostalPostcode
		}
		if customer.PhysicalAddress1 != nil {
			physicalAddress1 = *customer.PhysicalAddress1
		}
		if customer.PhysicalAddress2 != nil {
			physicalAddress2 = *customer.PhysicalAddress2
		}
		if customer.PostalAddress1 != nil {
			postalAddress1 = *customer.PostalAddress1
		}
		if customer.PostalAddress2 != nil {
			postalAddress2 = *customer.PostalAddress2
		}
		if customer.PostalCountryID != nil {
			postalCountryID = *customer.PostalCountryID
		}
		if customer.CustomField1 != nil {
			customField1 = *customer.CustomField1
		}
		if customer.CustomField2 != nil {
			customField2 = *customer.CustomField2
		}
		if customer.CustomField3 != nil {
			customField3 = *customer.CustomField3
		}
		if customer.CustomField4 != nil {
			customField4 = *customer.CustomField4
		}

		var record []string
		record = append(record, id)
		record = append(record, code)
		record = append(record, firstName)
		record = append(record, lastName)
		record = append(record, email)
		record = append(record, customerGroup)
		record = append(record, yearToDate)
		record = append(record, balance)
		record = append(record, loyaltyBalance)
		record = append(record, note)
		record = append(record, gender)
		record = append(record, dateOfBirth)
		record = append(record, companyName)
		record = append(record, phone)
		record = append(record, mobile)
		record = append(record, fax)
		record = append(record, twitter)
		record = append(record, website)
		record = append(record, doNotEmail)
		record = append(record, createdAt)
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
		record = append(record, customField1)
		record = append(record, customField2)
		record = append(record, customField3)
		record = append(record, customField4)

		writer.Write(record)
	}

	writer.Flush()
	return err
}
