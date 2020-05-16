package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// voidSalesCmd represents the voidSales command
var voidSalesCmd = &cobra.Command{
	Use:   "void-sales",
	Short: "Void Sales",
	Long: fmt.Sprintf(`
This tool requires a CSV of Sale IDs, no headers.

Example:
%s`, color.GreenString("vendcli void-sales -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		voidSales()
	},
}

func init() {
	// Flag
	voidSalesCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	voidSalesCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(voidSalesCmd)
}

func voidSales() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Printf("\nReading CSV...\n")
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Fatalf(color.RedString("Failed to get ids from the file: %s", FilePath))
	}

	// Get sale payload
	for _, id := range ids {
		fmt.Printf("Sale %v", id)
		sale, err := getSale(id)
		if err != nil {
			log.Printf("Failed to get a sale: %v", err)
			continue
		}

		// Add adjustments to the sale
		sale = adjustments(sale)

		// Make the requests
		url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)
		_, err = vendClient.MakeRequest("POST", url, sale)
		fmt.Printf(color.GreenString(" - VOIDED\n"))
		if err != nil {
			fmt.Printf("Error maing request: %v", err)
		}
	}
	fmt.Println(color.GreenString("\nFinished!\n"))
}

// GetSale pulls sales information from Vend
func getSale(id string) (vend.Sale, error) {

	sale := vend.Sale{}
	saleResponse := vend.RegisterSales{}

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)

	// Make the request
	res, err := vendClient.MakeRequest("GET", url, nil)

	// Unmarshal JSON Response
	err = json.Unmarshal(res, &saleResponse)
	if err != nil {
		return sale, err
	}

	return saleResponse.RegisterSales[0], nil
}

// Add adjustments before closing the transaction
func adjustments(sale vend.Sale) vend.Sale {

	*sale.Status = "VOIDED"

	return sale
}
