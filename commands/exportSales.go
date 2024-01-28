package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type SaleResults struct {
	sales            []vend.Sale
	registers        []vend.Register
	users            []vend.User
	customers        []vend.Customer
	customerGroupMap map[string]string
	products         []vend.Product
	err              error
}

// Command config
var (
	timeZone string
	dateFrom string
	dateTo   string
	outlet   string
	register string

	exportSalesCmd = &cobra.Command{
		Use:   "export-sales",
		Short: "Export Sales",
		Long: fmt.Sprintf(`
Exports all the Sales from an account, you can pass a single outlet to the command or export all outlets:
Single Outlet: -o Newmarket
Single Outlet, two words: -o 'Newmarket Outlet'
All Outlets: -o all

Example:
%s`, color.GreenString("vendcli export-sales -d DOMAINPREFIX -t TOKEN -z TIMEZONE -F DATEFROM -T DATETO -o all")),

		Run: func(cmd *cobra.Command, args []string) {
			getAllSales()
		},
	}
)

func init() {
	// Flags
	exportSalesCmd.Flags().StringVarP(&timeZone, "Timezone", "z", "", "Timezone of the store in zoneinfo format.")
	exportSalesCmd.Flags().StringVarP(&dateFrom, "DateFrom", "F", "", "Date from (YYYY-MM-DD)")
	exportSalesCmd.Flags().StringVarP(&dateTo, "DateTo", "T", "", "Date to (YYYY-MM-DD)")
	exportSalesCmd.Flags().StringVarP(&outlet, "Outlet", "o", "", "Outlet to export the sales from")
	exportSalesCmd.MarkFlagRequired("Timezone")
	exportSalesCmd.MarkFlagRequired("DateFrom")
	exportSalesCmd.MarkFlagRequired("DateTo")

	rootCmd.AddCommand(exportSalesCmd)
}

func getAllSales() {
	// Create a new Vend Client
	vc := vend.NewClient(Token, DomainPrefix, timeZone)

	// Validate date input
	validateDateInput(dateFrom, "date from")
	validateDateInput(dateTo, "date to")

	// Validate provided timezone
	validateTimeZone(dateTo+"T00:00:00Z", timeZone)

	// Pull data from Vend
	fmt.Println("\nRetrieving data from Vend...")

	// Get outlets and lookup outlet name by id
	oidToOutletName := getOutletsAndOutletNameMap(vc)

	// Check if the provided outlet exists
	if outlet != "all" && !validOutlet(outlet, oidToOutletName) {
		fmt.Printf(color.RedString("\n'%s' Outlet does not exist in the '%s' account\n\n", outlet, DomainPrefix))
		return
	}

	// Filter the sales by date range and outlet
	utcDateFrom, utcDateTo, versionAfter := prepareDateAndVersion(vc)

	// Get all sales data
	sales, registers, users, customers, customerGroupMap, products := getAllSalesData(vc, versionAfter)

	fmt.Printf("\nFiltering sales by outlet and date range...\n")

	// Process outlets
	processOutlets(vc, oidToOutletName, sales, utcDateFrom, utcDateTo, registers, users, customers, customerGroupMap, products)
}

func getAllSalesData(vc vend.Client, versionAfter int64) ([]vend.Sale, []vend.Register, []vend.User, []vend.Customer, map[string]string, []vend.Product) {
	var wg sync.WaitGroup
	saleResults := make(chan SaleResults, 2)
	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		sales, err := vc.SalesAfter(versionAfter)
		saleResults <- SaleResults{sales: sales, err: err}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		registers, users, customers, customerGroupMap, products := GetVendDataForSalesReport(vc)
		saleResults <- SaleResults{registers: registers, users: users, customers: customers, customerGroupMap: customerGroupMap, products: products}
	}()
	// Launch a goroutine to close the res channel after all other goroutines complete
	go func() {
		wg.Wait()
		close(saleResults)
	}()
	var sales []vend.Sale
	var registers []vend.Register
	var users []vend.User
	var customers []vend.Customer
	customerGroupMap := make(map[string]string)
	var products []vend.Product
	for i := 0; i < 2; i++ {
		s := <-saleResults
		if s.err != nil {
			log.Printf(color.RedString("Failed to get data: %v", s.err))
			return nil, nil, nil, nil, nil, nil
		}
		sales = append(sales, s.sales...)
		registers = append(registers, s.registers...)
		users = append(users, s.users...)
		customers = append(customers, s.customers...)
		for k, v := range s.customerGroupMap {
			customerGroupMap[k] = v
		}
		products = append(products, s.products...)
	}
	return sales, registers, users, customers, customerGroupMap, products
}

func processOutlets(vc vend.Client, oidToOutletName map[string]string, sales []vend.Sale, utcDateFrom, utcDateTo string, registers []vend.Register, users []vend.User, customers []vend.Customer, customerGroupMap map[string]string, products []vend.Product) {
	allOutletsName := getAllOutletsToProcess(oidToOutletName)

	var wg sync.WaitGroup
	for _, outlet := range allOutletsName {
		wg.Add(1)
		go func(outlet string) {
			defer wg.Done()
			filteredSales := getFilteredSales(sales, utcDateFrom, utcDateTo, oidToOutletName, outlet)
			processOutlet(vc, outlet, filteredSales, registers, users, customers, customerGroupMap, products)
		}(outlet)
	}

	wg.Wait()
}

func getAllOutletsToProcess(oidToOutletName map[string]string) []string {
	var allOutletsName []string

	// add the provided outlet by default
	allOutletsName = append(allOutletsName, outlet)

	if outlet == "all" {
		allOutletsName = getAllOutletNames(oidToOutletName)
	}

	return allOutletsName
}

func processOutlet(vc vend.Client, outlet string, filteredSales []vend.Sale, registers []vend.Register, users []vend.User, customers []vend.Customer, customerGroupMap map[string]string, products []vend.Product) {
	if len(filteredSales) > 0 {
		sortBySaleDate(filteredSales)

		file, err := createReport(vc.DomainPrefix, outlet)
		if err != nil {
			err = fmt.Errorf("Failed creating template CSV: %v", err)
			messenger.ExitWithError(err)
		}
		defer file.Close()

		file = addSalesReportHeader(file)

		fmt.Printf("Writing Sales to CSV file - %s...\n", outlet)
		file = writeSalesReport(file, registers, users, customers, customerGroupMap, products, filteredSales, vc.DomainPrefix, vc.TimeZone)

		fmt.Printf(color.GreenString("\nExported %v sales - %s\n\n", len(filteredSales), outlet))
	} else {
		fmt.Printf(color.GreenString("\n%s has no sales for the specified time period, skipping...\n", outlet))
	}
}

func validateDateInput(date string, label string) {
	layout := "2006-01-02"
	_, err := time.Parse(layout, date)
	if err != nil {
		err = fmt.Errorf("incorrect %s: %v, %v", label, date, err)
		messenger.ExitWithError(err)
	}
}

func validateTimeZone(date string, timeZone string) {
	_, err := getUtcTime(date, timeZone)
	if err != nil {
		err = fmt.Errorf("timezone invalid: %v\n", err)
		messenger.ExitWithError(err)
	}
}

func getOutletsAndOutletNameMap(vc vend.Client) map[string]string {
	outlets, _, err := vc.Outlets()
	if err != nil {
		err = fmt.Errorf("Failed to get outlets: %v", err)
		messenger.ExitWithError(err)
	}
	return getOidToOutletName(outlets)
}

func prepareDateAndVersion(vc vend.Client) (string, string, int64) {
	utcDateFrom, _ := getUtcTime(dateFrom+"T00:00:00Z", vc.TimeZone)
	utcDateTo, _ := getUtcTime(dateTo+"T23:59:59Z", vc.TimeZone)
	versionAfter, _ := vc.GetStartVersion(getTime(utcDateFrom), utcDateFrom)
	return utcDateFrom, utcDateTo, versionAfter
}

func GetVendDataForSalesReport(vc vend.Client) ([]vend.Register, []vend.User, []vend.Customer, map[string]string, []vend.Product) {
	// create a waitgroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	// create a channel to receive the results for each goroutine
	res := make(chan SaleResults, 5)

	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		registers, err := vc.Registers()
		res <- SaleResults{registers: registers, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		users, err := vc.Users()
		res <- SaleResults{users: users, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		customers, err := vc.Customers()
		res <- SaleResults{customers: customers, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		customerGroupMap, err := vc.CustomerGroups()
		res <- SaleResults{customerGroupMap: customerGroupMap, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done() // Finish the goroutine when we're done
		products, _, err := vc.Products()
		res <- SaleResults{products: products, err: err}
	}()

	// Launch a goroutine to close the res channel after all other goroutines complete
	go func() {
		wg.Wait()
		close(res)
	}()

	var registers []vend.Register
	var users []vend.User
	var customers []vend.Customer
	customerGroupMap := make(map[string]string)
	var products []vend.Product

	for i := 0; i < 5; i++ {
		s := <-res
		if s.err != nil {
			err := fmt.Errorf("Failed to get data: %v", s.err)
			messenger.ExitWithError(err)
		}
		registers = append(registers, s.registers...)
		users = append(users, s.users...)
		customers = append(customers, s.customers...)
		for k, v := range s.customerGroupMap {
			customerGroupMap[k] = v
		}
		products = append(products, s.products...)
	}

	return registers, users, customers, customerGroupMap, products
}

// getAllOutletNames returns the slice of outlet names based on the provided
// map[outletid] : outletName
func getAllOutletNames(oidToOutletName map[string]string) []string {
	var outletNames []string

	for oid := range oidToOutletName {
		currName := oidToOutletName[oid]

		outletNames = append(outletNames, currName)
	}

	return outletNames
}

// sortBySaleDate sorts the provided sale slice by sale_date asc with built in sort
// not sure exactly what algo is used but I know that the sales are already nearly sorted
// insertion sort is usually better in that case
func sortBySaleDate(sales []vend.Sale) {
	sort.SliceStable(sales, func(i, j int) bool {
		return getTime((*sales[i].SaleDate)[:19] + "Z").Before(getTime((*sales[j].SaleDate)[:19] + "Z"))
	})
}

// insertionSortSaleDate sorts the provided sale slice by sale_date asc
// since we know that the sales are usually sorted, insertion sort is probably best
// for those cases where a sale is updated within the date range will place it based on version
func insertionSortSaleDate(sales []vend.Sale) {
	var j int
	for i := 1; i < len(sales); i++ {
		currSale := sales[i]
		j = i - 1

		for j >= 0 && getTime((*sales[j].SaleDate)[:19]+"Z").After(getTime((*currSale.SaleDate)[:19]+"Z")) {
			sales[j+1] = sales[j]
			j = j - 1
		}
		sales[j+1] = currSale
	}
}

// validOutlet checks if outlet name exists in store
func validOutlet(outletName string, oidToName map[string]string) bool {
	for oid := range oidToName {
		currName := oidToName[oid]

		if strings.ToLower(currName) == strings.ToLower(outletName) {
			return true
		}
	}
	return false
}

// getFilteredSales returns the filtered sales based on provided outlet and utc datetime range
func getFilteredSales(sales []vend.Sale, utcdatefrom string, utcdateto string,
	oidToOutletName map[string]string, outlet string) []vend.Sale {
	var filteredSales []vend.Sale
	//oidToOutletName := getOidToOutletName(outlets)

	for _, sale := range sales {
		outletId := *sale.OutletID
		//outletName := oidToOutlet[outletId][0] // seems like the .Oultets returns a map outletid : []Outlet?
		outletName := oidToOutletName[outletId]

		// avoid any surprises with casing
		if strings.ToLower(outlet) != strings.ToLower(outletName) {
			continue
		}

		//.After and .Before does not seem inclusive
		dtFrom := getTime(utcdatefrom).Add(-1 * time.Second)
		dtTo := getTime(utcdateto).Add(1 * time.Second)
		saleDate := getTime((*sale.SaleDate)[:19] + "Z")

		if saleDate.After(dtFrom) && saleDate.Before(dtTo) {
			filteredSales = append(filteredSales, sale)
		}

	}

	return filteredSales
}

// getOidToOutletName returns a map[oid] string {outlet name}
func getOidToOutletName(outlets []vend.Outlet) map[string]string {
	oidToName := make(map[string]string)

	for _, o := range outlets {
		name := *o.Name
		name = strings.ReplaceAll(name, "/", "_") //  vend supports "/" in register name, but this breaks os.File
		id := *o.ID

		oidToName[id] = name
	}

	return oidToName
}

// getTime returns a time object of given dt string
func getTime(t string) time.Time {
	format := "2006-01-02T15:04:05Z"
	timeObj, _ := time.Parse(format, t)
	return timeObj
}

// getUtcTime converts local time to utc
func getUtcTime(localdt string, tz string) (string, error) {
	LOCAL, err := time.LoadLocation(tz)
	if err != nil {
		//fmt.Println(err)
		return "", err
	}

	const longForm = "2006-01-02T15:04:05Z"
	t, err := time.ParseInLocation(longForm, localdt, LOCAL)

	if err != nil {
		fmt.Println(err)
		return "Could not parse time.", err
	}

	utc := t.UTC()

	return utc.Format(longForm), err
}

// createReport creates a template CSV file with headers ready to be written to.
func createReport(domainPrefix string, outlet string) (*os.File, error) {

	// Create blank CSV file to be written to.
	fileName := fmt.Sprintf("sales_history_%s_%s_f%s_t%s.csv", DomainPrefix, outlet, dateFrom, dateTo)
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		err = fmt.Errorf("Error creating CSV file: %s", err)
		messenger.ExitWithError(err)
	}

	return file, err
}

func addSalesReportHeader(file *os.File) *os.File {
	// Start CSV writer.
	writer := csv.NewWriter(file)

	// Set header values.
	var headerLine []string
	headerLine = append(headerLine, "Sale UUID")          // 0
	headerLine = append(headerLine, "Sale Date")          // 1
	headerLine = append(headerLine, "Sale Time")          // 2
	headerLine = append(headerLine, "Invoice Number")     // 3
	headerLine = append(headerLine, "Line Type")          // 4
	headerLine = append(headerLine, "Customer Code")      // 5
	headerLine = append(headerLine, "Customer Name")      // 6
	headerLine = append(headerLine, "Customer Email")     // 7
	headerLine = append(headerLine, "Customer Group")     // 8
	headerLine = append(headerLine, "Customer Address1")  // 9
	headerLine = append(headerLine, "Customer Address2")  // 10
	headerLine = append(headerLine, "Customer City")      // 11
	headerLine = append(headerLine, "Customer State")     // 12
	headerLine = append(headerLine, "Customer Postcode")  // 13
	headerLine = append(headerLine, "Customer CountryID") // 14
	headerLine = append(headerLine, "Do not email")       // 15
	headerLine = append(headerLine, "Sale Note")          // 16
	headerLine = append(headerLine, "Quantity")           // 17
	headerLine = append(headerLine, "Cost")               // 18
	headerLine = append(headerLine, "Price")              // 19
	headerLine = append(headerLine, "Tax")                // 20
	headerLine = append(headerLine, "Discount")           // 21
	headerLine = append(headerLine, "Loyalty")            // 22
	headerLine = append(headerLine, "Total")              // 23
	headerLine = append(headerLine, "Paid")               // 24
	headerLine = append(headerLine, "Details")            // 25
	headerLine = append(headerLine, "Register")           // 26
	headerLine = append(headerLine, "User")               // 27
	headerLine = append(headerLine, "Status")             // 28
	headerLine = append(headerLine, "Product Sku")        // 29

	// Write headerline to file.
	writer.Write(headerLine)
	writer.Flush()

	return file
}

// writeReport aims to mimic the report generated by exporting Vend sales history
func writeSalesReport(file *os.File, registers []vend.Register, users []vend.User,
	customers []vend.Customer, customerGroupMap map[string]string, products []vend.Product, sales []vend.Sale, domainPrefix,
	timeZone string) *os.File {

	// Create CSV writer.
	writer := csv.NewWriter(file)

	// Prepare data to be written to CSV.
	for _, sale := range sales {

		// Do not include deleted sales in reports.
		if sale.DeletedAt != nil {
			continue
		}
		// Do not include sales with status of "OPEN"
		if sale.Status != nil && *sale.Status == "OPEN" {
			continue
		}

		var saleID string
		if sale.ID != nil {
			saleID = *sale.ID
		}

		// Takes a Vend timestamp string as input and converts it to a Go Time.time value.
		dateTimeInLocation, err := vend.ParseVendDT(*sale.SaleDate, timeZone)
		if err != nil {
			fmt.Printf("Error parsing date: %s\n", err)
			dateTimeInLocation = time.Unix(0, 0) // If we can't parse the date, set it to the Unix epoch.
		}
		// Time string with timezone removed.
		dateTimeStr := dateTimeInLocation.String()[0:19]
		// Split time and date on space.
		// Example date/time string: 2015-07-01 07:03:22
		var dateStr, timeStr string
		dateStr = dateTimeStr[0:10]
		timeStr = dateTimeStr[10:19]

		var invoiceNumber string
		if sale.InvoiceNumber != nil {
			invoiceNumber = *sale.InvoiceNumber
		}

		// Customer
		var customerName, customerFirstName, customerLastName,
			customerCode, customerEmail, doNotEmail string
		var customerFullName []string

		// extra customer info field based on feature request
		var customerPostalAddress1, customerPostalAddress2, customerPostalCity,
			customerPostalState, customerPostalPostcode, customerPostalCountryID, customerGroup string
		for _, customer := range customers {
			// Make sure we only use info from customer on our sale.
			if customer.ID != nil && sale.CustomerID != nil {
				if *customer.ID == *sale.CustomerID {
					if customer.FirstName != nil {
						customerFirstName = *customer.FirstName
						customerFullName = append(customerFullName, customerFirstName)
					}
					if customer.LastName != nil {
						customerLastName = *customer.LastName
						customerFullName = append(customerFullName, customerLastName)
					}
					if customer.Code != nil {
						customerCode = *customer.Code
					}
					if customer.Email != nil {
						customerEmail = *customer.Email
					}
					if customer.GroupId != nil {
						customerGroup = customerGroupMap[*customer.GroupId]
					}
					if customer.DoNotEmail != nil {
						doNotEmail = fmt.Sprint(*customer.DoNotEmail)
					}
					if customer.PostalAddress1 != nil {
						customerPostalAddress1 = *customer.PostalAddress1
					}
					if customer.PostalAddress2 != nil {
						customerPostalAddress2 = *customer.PostalAddress2
					}
					if customer.PostalCity != nil {
						customerPostalCity = *customer.PostalCity
					}
					if customer.PostalState != nil {
						customerPostalState = *customer.PostalState
					}
					if customer.PostalPostcode != nil {
						customerPostalPostcode = *customer.PostalPostcode
					}
					if customer.PostalCountryID != nil {
						customerPostalCountryID = *customer.PostalCountryID
					}

					customerName = strings.Join(customerFullName, " ")
					break
				}
			}
		}

		// Sale note wrapped in quote marks.
		var saleNote string
		if sale.Note != nil {
			saleNote = fmt.Sprintf("%q", *sale.Note)
		}

		// Add up the total quantities of each product line item.
		var totalQuantity, totalDiscount, totalTransactionCost float64
		var saleItems []string
		for _, lineitem := range *sale.LineItems {
			if lineitem.Quantity != nil && lineitem.DiscountTotal != nil {
				totalQuantity += *lineitem.Quantity
			}

			if lineitem.TotalCost != nil {
				totalTransactionCost += *lineitem.TotalCost
			}

			for _, product := range products {
				if *product.ID == *lineitem.ProductID {
					var productItems []string
					productItems = append(productItems, fmt.Sprintf("%v", *lineitem.Quantity))
					productItems = append(productItems, fmt.Sprintf("%s", *product.Name))

					prodItem := strings.Join(productItems, " X ")
					saleItems = append(saleItems, fmt.Sprintf("%v", prodItem))
					break
				}
			}
		}
		totalQuantityStr := strconv.FormatFloat(totalQuantity, 'f', -1, 64)
		totalDiscountStr := strconv.FormatFloat(totalDiscount, 'f', -1, 64)
		totalTransactionCostStr := strconv.FormatFloat(totalTransactionCost, 'f', 2, 64)
		// Show items sold separated by + sign.
		saleDetails := strings.Join(saleItems, " + ")

		var totalPrice, totalTax, total string
		// Sale subtotal.
		if sale.TotalPrice != nil {
			totalPrice = strconv.FormatFloat(*sale.TotalPrice, 'f', -1, 64)
		}

		// Sale tax.
		if sale.TotalTax != nil {
			totalTax = strconv.FormatFloat(*sale.TotalTax, 'f', -1, 64)
		}

		// Sale total (subtotal plus tax).
		if sale.TotalPrice != nil && sale.TotalTax != nil {
			total = strconv.FormatFloat((*sale.TotalPrice + *sale.TotalTax), 'f', -1, 64)
		}

		var totalLoyaltyStr string
		// Total loyalty on sale.
		if sale.TotalLoyalty != nil {
			totalLoyaltyStr = strconv.FormatFloat(*sale.TotalLoyalty, 'f', -1, 64)
		}

		var registerName string
		for _, register := range registers {
			if *sale.RegisterID == *register.ID {
				registerName = *register.Name
				// Append (Deleted) to name if register is deleted.
				if register.DeletedAt != nil {
					registerName += " (Deleted)"
				}
				break
			} else {
				// Should no longer reach this point as registers endpoint now returns
				// deleted registers. But if for whatever reason we do, write <deleted register>.
				registerName = "<Deleted Register>"
			}
		}

		var userName string
		for _, user := range users {
			if sale.UserID != nil && *sale.UserID == *user.ID {
				if user.DisplayName != nil {
					userName = *user.DisplayName
				} else if user.Username != nil {
					userName = *user.Username
				} else {
					userName = ""
				}
				break
			} else {
				userName = ""
			}
		}

		var saleStatus string
		if sale.Status != nil {
			saleStatus = *sale.Status
		}

		// Write first sale line to file.
		var record []string
		record = append(record, saleID)                  // 0. Sale UUID
		record = append(record, dateStr)                 // 1. Date
		record = append(record, timeStr)                 // 2. Time
		record = append(record, invoiceNumber)           // 3. Receipt Number
		record = append(record, "Sale")                  // 4. Line Type
		record = append(record, customerCode)            // 5. Customer Code
		record = append(record, customerName)            // 6. Customer Name
		record = append(record, customerEmail)           // 7. Customer Email
		record = append(record, customerGroup)           // 8. Customer Group
		record = append(record, customerPostalAddress1)  // 9. Customer PostalAddress 1
		record = append(record, customerPostalAddress2)  // 10. Customer PostalAddress 2
		record = append(record, customerPostalCity)      // 11. Customer PostalCity
		record = append(record, customerPostalState)     // 12. Customer Postal State
		record = append(record, customerPostalPostcode)  // 13. Customer Postal PostCode
		record = append(record, customerPostalCountryID) // 14. Customer Postal Country ID
		record = append(record, doNotEmail)              // 15. Marketing Opt in/out
		record = append(record, saleNote)                // 16. Note
		record = append(record, totalQuantityStr)        // 17. Quantity
		record = append(record, totalTransactionCostStr) // 18. Transaction Cost
		record = append(record, totalPrice)              // 19. Subtotal
		record = append(record, totalTax)                // 20. Sales Tax
		record = append(record, totalDiscountStr)        // 21. Discount
		record = append(record, totalLoyaltyStr)         // 22. Loyalty
		record = append(record, total)                   // 23. Sale total
		record = append(record, "")                      // 24. Paid
		record = append(record, saleDetails)             // 25. Details
		record = append(record, registerName)            // 26. Register
		record = append(record, userName)                // 27. User
		record = append(record, saleStatus)              // 28. Status
		record = append(record, "")                      // 29. Sku

		writer.Write(record)

		for _, lineitem := range *sale.LineItems {
			var quantity, unitCost, price, tax, discount, loyalty, total string
			if lineitem.Quantity != nil {
				quantity = strconv.FormatFloat(*lineitem.Quantity, 'f', -1, 64)
			}
			if lineitem.UnitCost != nil {
				unitCost = strconv.FormatFloat(*lineitem.UnitCost, 'f', 2, 64)
			}
			if lineitem.Price != nil {
				price = strconv.FormatFloat(*lineitem.Price, 'f', -1, 64)
			}
			if lineitem.Tax != nil {
				tax = strconv.FormatFloat(*lineitem.Tax, 'f', -1, 64)
			}
			if lineitem.Discount != nil {
				discount = strconv.FormatFloat(*lineitem.Discount, 'f', -1, 64)
			}
			if lineitem.LoyaltyValue != nil {
				loyalty = strconv.FormatFloat(*lineitem.LoyaltyValue, 'f', -1, 64)
			}
			if lineitem.Price != nil && lineitem.Tax != nil && lineitem.Quantity != nil {
				total = strconv.FormatFloat(((*lineitem.Price + *lineitem.Tax) * *lineitem.Quantity), 'f', -1, 64)
			}

			var productName, productSKU string
			for _, product := range products {
				if *product.ID == *lineitem.ProductID {
					productName = *product.VariantName
					productSKU = *product.SKU
				}
			}

			// Write product records for given sale to file.
			productRecord := record
			productRecord[0] = saleID        // 0. Sale UUID
			productRecord[1] = dateStr       // 1. Sale Date
			productRecord[2] = timeStr       // 2. Sale Time
			productRecord[3] = invoiceNumber // 3. Invoice Number
			productRecord[4] = "Sale Line"   // 4. Line Type
			productRecord[5] = ""            // 5. Customer Code
			productRecord[6] = ""            // 6. Customer Name
			productRecord[7] = ""            // 7. Customer Email
			productRecord[8] = ""            // 8. Customer Group
			productRecord[9] = ""            // 9. Customer Postal Address 1
			productRecord[10] = ""           // 10. Customer Postal Address 2
			productRecord[11] = ""           // 11. Customer Postal City
			productRecord[12] = ""           // 12. Customer Postal State
			productRecord[13] = ""           // 13. Customer Postal PostCode
			productRecord[14] = ""           // 14. Customer Postal Country ID
			productRecord[15] = ""           // 15. TODO: line note from the product?
			productRecord[16] = ""           // 16. Marketing Opt in/out
			productRecord[17] = quantity     // 17. Quantity
			productRecord[18] = unitCost     // 18. Unit Cost
			productRecord[19] = price        // 19. Subtotal
			productRecord[20] = tax          // 20. Sales Tax
			productRecord[21] = discount     // 21. Discount
			productRecord[22] = loyalty      // 22. Loyalty
			productRecord[23] = total        // 23. Total
			productRecord[24] = ""           // 24. Paid
			productRecord[25] = productName  // 25. Details
			productRecord[26] = ""           // 26. Register
			productRecord[27] = ""           // 27. User
			productRecord[28] = ""           // 28. Status
			productRecord[29] = productSKU   // 29. Sku

			writer.Write(productRecord)
		}

		payments := *sale.Payments
		for _, payment := range payments {

			paid := strconv.FormatFloat(*payment.Amount, 'f', -1, 64)
			name := fmt.Sprintf("%s", *payment.Name)

			paymentRecord := record
			paymentRecord[0] = saleID        // 0. Sale UUID
			paymentRecord[1] = dateStr       // 1. Sale Date
			paymentRecord[2] = timeStr       // 2. Sale Time
			paymentRecord[3] = invoiceNumber // 3. Invoice Number
			paymentRecord[4] = "Payment"     // 4. Line Type
			paymentRecord[5] = ""            // 5. Customer Code
			paymentRecord[6] = ""            // 6. Customer Name
			paymentRecord[7] = ""            // 7. Customer Email
			paymentRecord[8] = ""            // 8. Customer Group
			paymentRecord[9] = ""            // 9. Customer Postal Address 1
			paymentRecord[10] = ""           // 10. Customer Postal Address 2
			paymentRecord[11] = ""           // 11. Customer Postal City
			paymentRecord[12] = ""           // 12. Customer Postal State
			paymentRecord[13] = ""           // 13. Customer Postal PostalCode
			paymentRecord[14] = ""           // 14. Customer Postal Country ID
			paymentRecord[15] = ""           // 15. Marketing Opt in/out
			paymentRecord[16] = ""           // 16. TODO: line note
			paymentRecord[17] = ""           // 17. Quantity
			paymentRecord[18] = ""           // 18. Cost
			paymentRecord[19] = ""           // 19. Subtotal
			paymentRecord[20] = ""           // 20. Sales Tax
			paymentRecord[21] = ""           // 21. Discount
			paymentRecord[22] = ""           // 22. Loyalty
			paymentRecord[23] = ""           // 23. Total
			paymentRecord[24] = paid         // 24. Paid
			paymentRecord[25] = name         // 25. Details
			paymentRecord[26] = ""           // 26. Register
			paymentRecord[27] = ""           // 27. User
			paymentRecord[28] = ""           // 28. Status
			paymentRecord[29] = ""           // 29. Sku

			writer.Write(paymentRecord)
		}
	}
	writer.Flush()
	return file
}
