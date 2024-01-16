package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
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
	avgCostFilePath     string
	batchSize           = 50
	averageCostEndpoint = "api/_internal/twirp/vend.polecat.Polecat/SetAverageCosts"

	updateAverageCostCmd = &cobra.Command{
		Use:   "update-average-cost",
		Short: "Update Average Cost",
		Long: fmt.Sprintf(`
	Example:
	%s`, color.GreenString("vendcli update-average-cost -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
		Run: func(cmd *cobra.Command, args []string) {
			updateAverageCost()
		},
	}
)

func init() {
	// Flag
	updateAverageCostCmd.Flags().StringVarP(&avgCostFilePath, "filename", "f", "", "The name of your file: filename.csv")
	updateAverageCostCmd.MarkFlagRequired("filename")

	rootCmd.AddCommand(updateAverageCostCmd)
}

func updateAverageCost() {
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

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
