package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type AverageCostRequestBody struct {
	ProductCosts []ProductCost `json:"average_costs"`
}

type ProductCost struct {
	ProductID string `json:"product_id"`
	OutletID  string `json:"outlet_id"`
	Cost      string `json:"cost"`
}

type FailedRequest struct {
	ProductCost
	Reason string
}

var (
	updateAverageCostCmdmMode string
	avgCostFilePath           string
	batchSize                 = 50
	averageCostEndpoint       = "api/_internal/twirp/vend.polecat.Polecat/SetAverageCosts"

	updateAverageCostCmd = &cobra.Command{
		Use:   "update-average-cost",
		Short: "Update Average Cost",
		Long: fmt.Sprintf(`

update-average-cost will update the average cost of products in Vend, based on the CSV file you provide.

The CSV file should have the following format:
+------------+---------------+------------+---------------+------------+-----+---------------+------------+
| product_id | outlet_1 name | avg_cost_1 | outlet_2 name | avg_cost_2 | ... | outlet_n name | avg_cost_n |
+------------+---------------+------------+---------------+------------+-----+---------------+------------+
| <uuid>     | <uuid>        | <amount>   | <uuid>        | <amount>   | ... | <uuid>        | <amount>   |
+------------+---------------+------------+---------------+------------+-----+---------------+------------+

The first column must be "product_id", but columns B onwards can be named anything you'd like. 
For ease, it's recommended to keep the column names the same as the outlet names.

To facilate creating the CSV file, you can use the print-template mode. 
This will print a worksheet that you can share with the retailer to fill out. 
Once the retailer has filled out the worksheet, remove the extra columns to fit 
the above pattern and then use the update mode to update the average costs in Vend.

Example:
%s`, color.GreenString("vendcli update-average-cost -d DOMAINPREFIX -t TOKEN -m MODE -f FILENAME.csv")),
		Run: func(cmd *cobra.Command, args []string) {
			vc := vend.NewClient(Token, DomainPrefix, "")
			vendClient = &vc

			if parseUpdateAverageCostMode(updateAverageCostCmdmMode) {
				updateAverageCost()
			} else {
				printAverageCostTemplate()
			}
		},
	}
)

func init() {
	// Flag
	updateAverageCostCmd.Flags().StringVarP(&avgCostFilePath, "filename", "f", "", "The name of your file: filename.csv")

	updateAverageCostCmd.Flags().StringVarP(&updateAverageCostCmdmMode, "mode", "m", "update", "modes: print-template, update")

	rootCmd.AddCommand(updateAverageCostCmd)
}

func printAverageCostTemplate() {

	fmt.Println("Template mode chosen...")

	outlets, products, recordsMap, maxSupplier := getInfoForTemplate()

	file, err := createAverageCostTemplate(DomainPrefix)
	if err != nil {
		log.Printf("Failed creating CSV file %v", err)
		panic(vend.Exit{1})
	}
	file = addHeadeToTemplate(file, outlets, maxSupplier)
	file = writeAverageCostTemplate(file, outlets, products, recordsMap, maxSupplier)

}

func getInfoForTemplate() ([]vend.Outlet, []vend.Product, map[string]map[string]vend.InventoryRecord, int) {
	vc := *vendClient

	fmt.Println("Getting info for template...")

	var wg sync.WaitGroup
	var outlets []vend.Outlet
	var products []vend.Product
	var inventoryRecords []vend.InventoryRecord
	var outletsErr, productsErr, InventoryRecordsErr error
	var recordsMap map[string]map[string]vend.InventoryRecord
	var maxSupplier int

	// Use goroutines to fetch outlets and products concurrently
	wg.Add(3)
	go func() {
		outlets, _, outletsErr = vc.Outlets()
		wg.Done()
	}()

	go func() {
		products, _, productsErr = vc.Products()
		wg.Done()
	}()

	go func() {
		inventoryRecords, InventoryRecordsErr = vc.Inventory()
		wg.Done()
	}()

	// Wait for goroutines to finish
	wg.Wait()

	// Check for errors in fetching outlets and products
	if outletsErr != nil {
		log.Printf("Failed retrieving outlets: %v", outletsErr)
		panic(vend.Exit{1})
	}

	if productsErr != nil {
		log.Printf("Failed retrieving products: %v", productsErr)
		panic(vend.Exit{1})
	}

	if InventoryRecordsErr != nil {
		fmt.Println("Error fetching inventory records: %v", InventoryRecordsErr)
	}

	wg.Add(2)

	go func() {
		maxSupplier = checkMaxSupplier(products)
		wg.Done()
	}()

	go func() {
		recordsMap = buildRecordsMap(inventoryRecords, outlets)
		wg.Done()
	}()

	wg.Wait()

	return outlets, products, recordsMap, maxSupplier
}

func createAverageCostTemplate(domainPrefix string) (*os.File, error) {

	fileName := fmt.Sprintf("%s_average_cost_worksheet.csv", DomainPrefix)
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		return nil, err
	}

	return file, err
}

func addHeadeToTemplate(file *os.File, outlets []vend.Outlet, maxSupplier int) *os.File {

	writer := csv.NewWriter(file)

	// Set header values.
	var headerLine []string
	headerLine = append(headerLine, "product_id")      //0
	headerLine = append(headerLine, "product_name")    //1
	headerLine = append(headerLine, "track_inventory") //2

	switch maxSupplier {
	case 1:
		headerLine = append(headerLine, "supply_price") // 3 & 4
	default:
		for i := 1; i <= maxSupplier; i++ {
			headerLine = append(headerLine, fmt.Sprintf("supplier_name_%v", i)) // 3
			headerLine = append(headerLine, fmt.Sprintf("supply_price_%v", i))  // 4
		}
	}

	for _, outlet := range outlets {
		headerLine = append(headerLine, fmt.Sprintf("outlet_id_%s", *outlet.Name))    // 5
		headerLine = append(headerLine, fmt.Sprintf("average_cost_%s", *outlet.Name)) // 6
	}

	// Write headerline to file.
	writer.Write(headerLine)
	writer.Flush()

	return file
}

func writeAverageCostTemplate(file *os.File, outlets []vend.Outlet, products []vend.Product, recordsMap map[string]map[string]vend.InventoryRecord, maxSupplier int) *os.File {

	writer := csv.NewWriter(file)

	for _, product := range products {

		var productID, productName, trackInventory string

		if product.ID != nil {
			productID = *product.ID
		} else {
			continue // if product has no ID, skip it
		}

		if product.Name != nil {
			productName = *product.Name
		}

		trackInventory = strconv.FormatBool(product.TrackInventory)

		var record []string
		record = append(record, productID)      //0
		record = append(record, productName)    //1
		record = append(record, trackInventory) //2

		switch maxSupplier {
		case 1:
			if len(product.ProductSuppliers) > 0 {
				supplier := product.ProductSuppliers[0]
				if supplier.Price != nil {
					record = append(record, fmt.Sprintf("%.2f", *supplier.Price)) // 3 & 4
				} else {
					record = append(record, "") // 3 & 4
				}
			} else {
				record = append(record, "") // 3 & 4
			}
		default:
			numSuppliers := len(product.ProductSuppliers)
			for s := 0; s < maxSupplier; s++ {
				switch {
				case s < numSuppliers:
					supplier := product.ProductSuppliers[s]
					if supplier.SupplierName != nil {
						record = append(record, *supplier.SupplierName) // 3
					} else {
						record = append(record, "") // 3
					}
					if supplier.Price != nil {
						record = append(record, fmt.Sprintf("%.2f", *supplier.Price)) // 4
					} else {
						record = append(record, "") // 4
					}
				default:
					record = append(record, "") // 3
					record = append(record, "") // 4
				}
			}
		}

		for _, outlet := range outlets {
			var outletID, averageCost string

			if outlet.ID != nil {
				outletID = *outlet.ID
				record = append(record, outletID) // 5
			} else {
				record = append(record, "") // 5
			}

			if invRecord, ok := recordsMap[outletID][productID]; ok {
				if invRecord.AverageCost != nil {
					averageCost = fmt.Sprintf("%.2f", *invRecord.AverageCost)
					record = append(record, averageCost) // 6
				} else {
					record = append(record, "") // 6
				}
			} else {
				record = append(record, "") // 6
			}
		}
		writer.Write(record)
	}
	writer.Flush()
	return file
}

func updateAverageCost() {
	// Read the CSV file
	productCosts := readAverageCostCSVFile(avgCostFilePath)

	fmt.Printf("Updating %v products\n", len(productCosts))
	failedProducts := postAverageCosts(productCosts)
	if len(failedProducts) > 0 {
		fmt.Printf("There were some errors. Writing failures to csv.. \n")
		saveFailedRequestsToCSV(failedProducts)
	}
	fmt.Println("Done!")
}

func parseUpdateAverageCostMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))

	switch mode {
	case "update", "":
		if len(avgCostFilePath) > 0 {
			return true
		} else {
			fmt.Println("Please provide a filename")
			fmt.Printf("Example:\n%s\n", color.GreenString("vendcli update-average-cost -d DOMAINPREFIX -t TOKEN -m update -f FILENAME.csv"))
			panic(vend.Exit{1})
		}
	case "print-template":
		return false
	default:
		fmt.Println("Invalid mode. Please use either 'update' or 'print-template'")
		panic(vend.Exit{1})
	}
}

func readAverageCostCSVFile(pathToFile string) []ProductCost {
	records, err := openAverageCostCSVFile(pathToFile)
	if err != nil {
		fmt.Println(err)
		panic(vend.Exit{1})
	}

	checkAverageCostCSVHeader(records)
	products := []ProductCost{}

	// Loop through the rows, skip the first row (header)
	for _, record := range records[1:] {
		productID := record[0]

		// Zero index is productID, then outletID and cost are in pairs so we increment by 2 from index 1
		for j := 1; j < len(record); j += 2 {

			product := ProductCost{
				ProductID: productID,
				OutletID:  record[j],
				Cost:      record[j+1],
			}

			// skip the empty requests
			if product.Cost == "" {
				continue
			}

			products = append(products, product)
		}
	}

	return products

}

func openAverageCostCSVFile(pathToFile string) ([][]string, error) {
	// Open the file
	csvFile, err := os.Open(pathToFile)
	if err != nil {
		fmt.Println(err)
		panic(vend.Exit{1})
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		panic(vend.Exit{1})
	}

	return records, nil
}

// checkHeader checks if the header has the correct format
func checkAverageCostCSVHeader(records [][]string) {
	// check if the first column has "product_id", and the number of columns is odd
	if len(records) == 0 || len(records[0]) < 2 || records[0][0] != "product_id" || len(records[0])%2 == 0 {
		fmt.Println("Warning: Incorrect header format. Expected format: 'product_id, outlet_id, cost, ...'")
		fmt.Println("Example header: 'product_id, outlet_id1, cost_1, outlet_id2, cost_2, ...'")
		panic(vend.Exit{1})
	}
}

func postAverageCosts(productCosts []ProductCost) [][]FailedRequest {
	vc := *vendClient

	url := fmt.Sprintf("https://%s.retail.lightspeed.app/%s", DomainPrefix, averageCostEndpoint)

	failedRequests := [][]FailedRequest{}

	// Split the ProductCosts into batches
	for i := 0; i < len(productCosts); i += batchSize {
		endIndex := i + batchSize
		if endIndex > len(productCosts) {
			endIndex = len(productCosts)
		}

		// Extract the current batch
		currentBatch := productCosts[i:endIndex]

		fmt.Println("Posting products: ", i, " - ", endIndex-1)

		// Create the request body
		body := AverageCostRequestBody{
			ProductCosts: currentBatch,
		}

		// Make the request
		_, err := vc.MakeRequest("POST", url, body)
		if err != nil {
			fmt.Println("Error posting batch")
			failedRequestsFromBatch := retryBatch(currentBatch, url, i)
			failedRequests = append(failedRequests, failedRequestsFromBatch)
		}

	}
	return failedRequests

}

func retryBatch(productCosts []ProductCost, url string, index int) []FailedRequest {
	fmt.Println("Retrying products in batch individually...")
	vc := *vendClient
	failedRequests := []FailedRequest{}

	for i, request := range productCosts {
		fmt.Println("Retrying product:", i+index)

		body := AverageCostRequestBody{
			ProductCosts: []ProductCost{request},
		}

		_, err := vc.MakeRequest("POST", url, body)
		if err != nil {
			failedRequest := FailedRequest{
				ProductCost: request,
				Reason:      err.Error(),
			}
			failedRequests = append(failedRequests, failedRequest)
		}
	}

	return failedRequests
}

func saveFailedRequestsToCSV(failedRequests [][]FailedRequest) {

	fileName := fmt.Sprintf("%s_failed_requests__%v.csv", DomainPrefix, time.Now().Unix())
	// Create a new file
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the header
	header := []string{"product_id", "outlet_id", "cost", "reason"}
	writer := csv.NewWriter(file)
	err = writer.Write(header)
	if err != nil {
		fmt.Println("Error writing failed requests to file:", err)
		return
	}

	// Write the failed requests
	for _, batch := range failedRequests {
		for _, request := range batch {
			record := []string{request.ProductID, request.OutletID, fmt.Sprintf("%v", request.Cost), request.Reason}
			err := writer.Write(record)
			if err != nil {
				fmt.Println("Error writing failed requests to file:", err)
				return
			}
		}
	}
	writer.Flush()
}
