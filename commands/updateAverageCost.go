package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

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

type FailedUpdateAvgCostRequest struct {
	ProductID string
	OutletID  string
	Cost      string
	Reason    string
}

var (
	failedUpdateAvgCostRequests []FailedUpdateAvgCostRequest
	updateAverageCostCmdmMode   string
	avgCostFilePath             string
	batchSize                   = 50
	averageCostEndpoint         = "api/_internal/twirp/vend.polecat.Polecat/SetAverageCosts"

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
				fmt.Printf("\nRunning command in %s mode\n", color.YellowString("UPDATE"))
				updateAverageCost()
			} else {
				fmt.Printf("\nRunning command in %s mode\n", color.YellowString("PRINT-TEMPLATE"))
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

	fmt.Println("\nGetting info from vend...")
	outlets, products, recordsMap, maxSupplier := getInfoForTemplate()

	fmt.Println("\nCreating template CSV file...")
	err := writeAverageCostTemplate(DomainPrefix, outlets, products, recordsMap, maxSupplier)
	if err != nil {
		err = fmt.Errorf("failed creating CSV file %v", err)
		messenger.ExitWithError(err)
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\n"))
}

func getInfoForTemplate() ([]vend.Outlet, []vend.Product, map[string]map[string]vend.InventoryRecord, int) {

	p, err := pbar.CreateMultiBarGroup(3, Token, DomainPrefix)
	if err != nil {
		fmt.Println("error creating progress bar group: ", err)
	}
	p.FetchDataWithProgressBar("outlets")
	p.FetchDataWithProgressBar("products")
	p.FetchDataWithProgressBar("inventory")

	p.MultiBarGroupWait()

	var outlets []vend.Outlet
	var products []vend.Product
	var inventoryRecords []vend.InventoryRecord

	for err = range p.ErrorChannel {
		err = fmt.Errorf("error fetching data: %v", err)
		messenger.ExitWithError(err)
	}

	for data := range p.DataChannel {
		switch d := data.(type) {
		case []vend.Outlet:
			outlets = d
		case []vend.Product:
			products = d
		case []vend.InventoryRecord:
			inventoryRecords = d
		}
	}

	var wg sync.WaitGroup
	var recordsMap map[string]map[string]vend.InventoryRecord
	var maxSupplier int

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

	fileName := fmt.Sprintf("%s_average_cost_worksheet_%v.csv", domainPrefix, time.Now().Unix())
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

func writeAverageCostTemplate(domainPrefix string, outlets []vend.Outlet, products []vend.Product, recordsMap map[string]map[string]vend.InventoryRecord, maxSupplier int) error {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(products), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	file, err := createAverageCostTemplate(domainPrefix)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed creating CSV file %v", err)
		return err
	}
	file = addHeadeToTemplate(file, outlets, maxSupplier)

	writer := csv.NewWriter(file)

	for _, product := range products {
		bar.Increment()
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
	p.Wait()
	return nil
}

func updateAverageCost() {

	fmt.Println("\nReading CSV file...")
	productCosts, err := readAverageCostCSVFile(avgCostFilePath)
	if err != nil {
		err = fmt.Errorf("error reading CSV file: %v", err)
		messenger.ExitWithError(err)
	}

	fmt.Printf("\nUpdating %v products\n", len(productCosts))
	count := postAverageCosts(productCosts)

	if len(failedUpdateAvgCostRequests) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_update_average_cost_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedUpdateAvgCostRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\nUpdated %d out of %d requests", count, len(productCosts)))
}

func parseUpdateAverageCostMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))

	switch mode {
	case "update", "":
		if len(avgCostFilePath) > 0 {
			return true
		} else {
			err := fmt.Errorf("please provide a filename Example: -%s", color.GreenString("vendcli update-average-cost -d DOMAINPREFIX -t TOKEN -m update -f FILENAME.csv"))
			messenger.ExitWithError(err)
		}
	case "print-template":
		return false
	default:
		err := fmt.Errorf("invalid mode. Please use either 'update' or 'print-template'")
		messenger.ExitWithError(err)
	}
	return false
}

func readAverageCostCSVFile(pathToFile string) ([]ProductCost, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	records, err := openAverageCostCSVFile(pathToFile)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	err = checkAverageCostCSVHeader(records)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

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
			if product.ProductID == "" || product.OutletID == "" || product.Cost == "" {
				failedUpdateAvgCostRequests = append(failedUpdateAvgCostRequests, FailedUpdateAvgCostRequest{
					ProductID: product.ProductID,
					OutletID:  product.OutletID,
					Cost:      product.Cost,
					Reason:    "missing fields",
				})
				continue
			}
			products = append(products, product)
		}
	}
	bar.SetIndeterminateBarComplete()
	p.Wait()

	return products, nil
}

func openAverageCostCSVFile(pathToFile string) ([][]string, error) {
	// Open the file
	csvFile, err := os.Open(pathToFile)
	if err != nil {
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// checkHeader checks if the header has the correct format
func checkAverageCostCSVHeader(records [][]string) error {
	// check if the first column has "product_id", and the number of columns is odd
	if len(records) == 0 || len(records[0]) < 2 || records[0][0] != "product_id" || len(records[0])%2 == 0 {
		err := fmt.Errorf("warning: Incorrect header format. Expected format: 'product_id, outlet_id, cost, ...'")
		return err
	}
	return nil
}

func postAverageCosts(productCosts []ProductCost) int {

	vc := *vendClient
	url := fmt.Sprintf("https://%s.retail.lightspeed.app/%s", DomainPrefix, averageCostEndpoint)

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(productCosts), "Updating average costs")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	var count int = 0
	for i := 0; i < len(productCosts); i += batchSize {
		endIndex := i + batchSize
		if endIndex > len(productCosts) {
			endIndex = len(productCosts)
		}
		currentBatch := productCosts[i:endIndex]

		// Create the request body
		body := AverageCostRequestBody{
			ProductCosts: currentBatch,
		}
		_, err := vc.MakeRequest("POST", url, body)
		if err != nil {
			successes := retryBatch(currentBatch, url, bar)
			count += successes
			continue
		}
		bar.IncBy(len(currentBatch))
		count += len(currentBatch)
	}
	p.Wait()
	return count
}

func retryBatch(productCosts []ProductCost, url string, bar *pbar.CustomBar) int {
	vc := *vendClient

	var count int = 0
	for _, request := range productCosts {
		bar.Increment()
		body := AverageCostRequestBody{
			ProductCosts: []ProductCost{request},
		}
		resp, err := vc.MakeRequest("POST", url, body)
		if err != nil {
			err = fmt.Errorf("failure when making request: %v response: %s", err, string(resp))
			failedUpdateAvgCostRequests = append(failedUpdateAvgCostRequests,
				FailedUpdateAvgCostRequest{
					ProductID: request.ProductID,
					OutletID:  request.OutletID,
					Cost:      request.Cost,
					Reason:    err.Error(),
				})
			continue
		}
		count += 1
	}
	return count
}
