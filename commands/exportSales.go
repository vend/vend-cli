package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

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

	// Parse date input for errors. Sample: 2017-11-20
	layout := "2006-01-02"
	_, err := time.Parse(layout, dateFrom)
	if err != nil {
		fmt.Printf("incorrect date from: %v, %v", dateFrom, err)
		os.Exit(1)
	}

	_, err = time.Parse(layout, dateTo)
	if err != nil {
		fmt.Printf("incorrect date to: %v, %v", dateTo, err)
		os.Exit(1)
	}

	// prevent further processing by checking provided timezone to be valid
	_, err = getUtcTime(dateTo+"T00:00:00Z", timeZone)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Pull data from Vend
	fmt.Println("\nRetrieving data from Vend...")

	// Get outlets first to check provided outlet first before expensive sales pull
	outlets, _, err := vc.Outlets()
	if err != nil {
		log.Fatalf(color.RedString("Failed to get outlets: %v", err))
	}

	// lookup outlet name by id
	oidToOutletName := getOidToOutletName(outlets)

	// prevent unnecessary processing to retrieve if providing wrong outlet name
	if outlet != "all" && !validOutlet(outlet, oidToOutletName) {
		fmt.Printf(color.RedString("\n'%s' Outlet does not exist in the '%s' account\n\n", outlet, DomainPrefix))
		return
	}

	// Filter the sales by date range and outlet
	utcDateFrom, _ := getUtcTime(dateFrom+"T00:00:00Z", vc.TimeZone)
	utcDateTo, _ := getUtcTime(dateTo+"T23:59:59Z", vc.TimeZone)

	versionAfter, _ := vc.GetStartVersion(getTime(utcDateFrom), utcDateFrom)

	// Get Sale data.
	//sales, err := vc.Sales()
	sales, err := vc.SalesAfter(versionAfter)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	// Get registers
	registers, err := vc.Registers()
	if err != nil {
		log.Fatalf("Failed to get registers: %v", err)
	}

	// Get users.
	users, err := vc.Users()
	if err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	// Get customers.
	customers, err := vc.Customers()
	if err != nil {
		log.Fatalf("Failed to get customers: %v", err)
	}

	// Get products.
	products, _, err := vc.Products()
	if err != nil {
		log.Fatalf("Failed to get products: %v", err)
	}

	fmt.Printf("\nFiltering sales by outlet and date range...\n")

	var allOutletsName []string

	// add the provided outlet by default
	allOutletsName = append(allOutletsName, outlet)

	if outlet == "all" {
		allOutletsName = getAllOutletNames(oidToOutletName)
	}

	// go through outlets to filter by date range and write to CSV
	for _, outlet := range allOutletsName {
		//filteredSales := getFilteredSales(sales, utcDateFrom, utcDateTo, outlets, outlet)
		filteredSales := getFilteredSales(sales, utcDateFrom, utcDateTo, oidToOutletName, outlet)

		//sort the sales asc by saledate since the pull may be out of order - version
		sortBySaleDate(filteredSales)

		// Create template report to be written to.
		file, err := createReport(vc.DomainPrefix, outlet)
		if err != nil {
			log.Fatalf("Failed creating template CSV: %v", err)
		}
		defer file.Close()

		// Write sale to CSV.
		fmt.Printf("Writing Sales to CSV file - %s...\n", outlet)
		// file = writeReport(file, registers, users, customers, products, sales, vc.DomainPrefix, vc.TimeZone)
		file = writeReport(file, registers, users, customers, products, filteredSales, vc.DomainPrefix, vc.TimeZone)

		fmt.Printf(color.GreenString("\nExported %v sales - %s\n\n", len(filteredSales), outlet))
	}

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
	fileName := fmt.Sprintf("%s_%s_sales_history_%v.csv", DomainPrefix, outlet, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatalf("Error creating CSV file: %s", err)
	}

	// Start CSV writer.
	writer := csv.NewWriter(file)

	// Set header values.
	var headerLine []string
	headerLine = append(headerLine, "Sale Date")
	headerLine = append(headerLine, "Sale Time")
	headerLine = append(headerLine, "Invoice Number")
	headerLine = append(headerLine, "Line Type")
	headerLine = append(headerLine, "Customer Code")
	headerLine = append(headerLine, "Customer Name")
	headerLine = append(headerLine, "Customer Email")
	headerLine = append(headerLine, "Do not email")
	headerLine = append(headerLine, "Sale Note")
	headerLine = append(headerLine, "Quantity")
	headerLine = append(headerLine, "Price")
	headerLine = append(headerLine, "Tax")
	headerLine = append(headerLine, "Discount")
	headerLine = append(headerLine, "Loyalty")
	headerLine = append(headerLine, "Total")
	headerLine = append(headerLine, "Paid")
	headerLine = append(headerLine, "Details")
	headerLine = append(headerLine, "Register")
	headerLine = append(headerLine, "User")
	headerLine = append(headerLine, "Status")
	headerLine = append(headerLine, "Product Sku")
	// Write headerline to file.
	writer.Write(headerLine)
	writer.Flush()

	return file, err
}

// writeReport aims to mimic the report generated by exporting Vend sales history
func writeReport(file *os.File, registers []vend.Register, users []vend.User,
	customers []vend.Customer, products []vend.Product, sales []vend.Sale, domainPrefix,
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

		// Takes a Vend timestamp string as input and converts it to a Go Time.time value.
		dateTimeInLocation := vend.ParseVendDT(*sale.SaleDate, timeZone)
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
		for _, customer := range customers {
			// Make sure we only use info from customer on our sale.
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
				if customer.DoNotEmail != nil {
					doNotEmail = fmt.Sprint(*customer.DoNotEmail)
				}
				customerName = strings.Join(customerFullName, " ")
				break
			}
		}

		// Sale note wrapped in quote marks.
		var saleNote string
		if sale.Note != nil {
			saleNote = fmt.Sprintf("%q", *sale.Note)
		}

		// Add up the total quantities of each product line item.
		var totalQuantity, totalDiscount float64
		var saleItems []string
		for _, lineitem := range *sale.LineItems {
			totalQuantity += *lineitem.Quantity
			totalDiscount += *lineitem.DiscountTotal

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
		// Show items sold separated by + sign.
		saleDetails := strings.Join(saleItems, " + ")

		// Sale subtotal.
		totalPrice := strconv.FormatFloat(*sale.TotalPrice, 'f', -1, 64)
		// Sale tax.
		totalTax := strconv.FormatFloat(*sale.TotalTax, 'f', -1, 64)
		// Sale total (subtotal plus tax).
		total := strconv.FormatFloat((*sale.TotalPrice + *sale.TotalTax), 'f', -1, 64)

		// Total loyalty on sale.
		totalLoyaltyStr := strconv.FormatFloat(*sale.TotalLoyalty, 'f', -1, 64)

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
				userName = *user.DisplayName
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
		record = append(record, dateStr)             // Date
		record = append(record, timeStr)             // Time
		record = append(record, invoiceNumber)       // Receipt Number
		record = append(record, "Sale")              // Line Type
		record = append(record, customerCode)        // Customer Code
		record = append(record, customerName) // Customer Name
		record = append(record, customerEmail)        // Customer Email
		record = append(record, doNotEmail)        // Marketing Opt in/out
		record = append(record, saleNote)            // Note
		record = append(record, totalQuantityStr)    // Quantity
		record = append(record, totalPrice)          // Subtotal
		record = append(record, totalTax)            // Sales Tax
		record = append(record, totalDiscountStr)    // Discount
		record = append(record, totalLoyaltyStr)     // Loyalty
		record = append(record, total)               // Sale total
		record = append(record, "")                  // Paid
		record = append(record, saleDetails)         // Details
		record = append(record, registerName)        // Register
		record = append(record, userName)            // User
		record = append(record, saleStatus)          // Status
		record = append(record, "")                  // Sku
		record = append(record, "")                  // AccountCodeSale
		record = append(record, "")                  // AccountCodePurchase
		writer.Write(record)

		for _, lineitem := range *sale.LineItems {

			quantity := strconv.FormatFloat(*lineitem.Quantity, 'f', -1, 64)
			price := strconv.FormatFloat(*lineitem.Price, 'f', -1, 64)
			tax := strconv.FormatFloat(*lineitem.Tax, 'f', -1, 64)
			discount := strconv.FormatFloat(*lineitem.Discount, 'f', -1, 64)
			loyalty := strconv.FormatFloat(*lineitem.LoyaltyValue, 'f', -1, 64)
			total := strconv.FormatFloat(((*lineitem.Price + *lineitem.Tax) * *lineitem.Quantity), 'f', -1, 64)

			var productName, productSKU string
			for _, product := range products {
				if *product.ID == *lineitem.ProductID {
					productName = *product.VariantName
					productSKU = *product.SKU
				}
			}

			// Write product records for given sale to file.
			productRecord := record
			productRecord[0] = dateStr      // Sale Date
			productRecord[1] = timeStr      // Sale Time
			productRecord[2] = ""           // Invoice Number
			productRecord[3] = "Sale Line"  // Line Type
			productRecord[4] = ""           // Customer Code
			productRecord[5] = ""           // Customer Name
			productRecord[6] = ""           // Customer Name
			productRecord[7] = ""           // TODO: line note from the product?
			productRecord[8] = quantity     // Quantity
			productRecord[9] = price        // Subtotal
			productRecord[10] = tax         // Sales Tax
			productRecord[11] = discount    // Discount
			productRecord[12] = loyalty     // Loyalty
			productRecord[13] = total       // Total
			productRecord[14] = ""          // Paid
			productRecord[15] = productName // Details
			productRecord[16] = ""          // Register
			productRecord[17] = ""          // User
			productRecord[18] = ""          // Status
			productRecord[19] = productSKU  // Sku
			// productRecord[19] = ""     // AccountCodeSale
			// productRecord[20] = "" // AccountCodePurchase
			writer.Write(productRecord)
		}

		payments := *sale.Payments
		for _, payment := range payments {

			paid := strconv.FormatFloat(*payment.Amount, 'f', -1, 64)
			name := fmt.Sprintf("%s", *payment.Name)

			paymentRecord := record
			paymentRecord[0] = dateStr   // Sale Date
			paymentRecord[1] = timeStr   // Sale Time
			paymentRecord[2] = ""        // Invoice Number
			paymentRecord[3] = "Payment" // Line Type
			paymentRecord[4] = ""        // Customer Code
			paymentRecord[5] = ""        // Customer Name
			paymentRecord[6] = ""        // Customer Name
			paymentRecord[7] = ""        // TODO: line note
			paymentRecord[8] = ""        // Quantity
			paymentRecord[9] = ""        // Subtotal
			paymentRecord[10] = ""       // Sales Tax
			paymentRecord[11] = ""       // Discount
			paymentRecord[12] = ""       // Loyalty
			paymentRecord[13] = ""       // Total
			paymentRecord[14] = paid     // Paid
			paymentRecord[15] = name     //  Details
			paymentRecord[16] = ""       // Register
			paymentRecord[17] = ""       // User
			paymentRecord[18] = ""       // Status
			paymentRecord[19] = ""       // Sku
			// paymentRecord[19] = ""       // AccountCodeSale
			// paymentRecord[20] = ""       // AccountCodePurchase

			writer.Write(paymentRecord)
		}
	}
	writer.Flush()
	return file
}
