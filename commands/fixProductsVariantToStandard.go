package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

var fixProductsVariantToStandardCmd = &cobra.Command{
	Use:   "fix-products-variant-to-standard",
	Short: "Fix Products - Converting variant to standard",
	Long: fmt.Sprintf(`
This tool requires a CSV of Product IDs, no headers.

Example:
%s`, color.GreenString("vendcli fix-products-variant-to-standard -d DOMAINPREFIX -t TOKEN -f FILENAME.csv -r ''")),
	Run: func(cmd *cobra.Command, args []string) {
		fixProductsVariantToStandard()
	},
}

var Reason string

func init() {
	// Flag
	fixProductsVariantToStandardCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	fixProductsVariantToStandardCmd.Flags().StringVarP(&Reason, "Reason", "r", "", "The reason for performing this action")

	fixProductsVariantToStandardCmd.MarkFlagRequired("Filename")
	fixProductsVariantToStandardCmd.MarkFlagRequired("Reason")

	rootCmd.AddCommand(fixProductsVariantToStandardCmd)
}

type ConvertVariantToStandardRequest struct {
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	VariantID string `json:"variant_id"`
}

func fixProductsVariantToStandard() {
	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	if Reason == "" {
		log.Fatalf("Please specify a reason: -r '<REASON or ticket reference FOR RUNNING THIS>'")
	}

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Fatalf(color.RedString("Failed to get ids from the file: %s", FilePath))
	}

	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)
	// Make the requests
	maxProductsPerRequest := 10
	reqBody := make([]ConvertVariantToStandardRequest, 0, len(ids))
	for i, id := range ids {
		// NOTE: This action is internal and could potentially change in a non-backward compatible manner.
		body := ConvertVariantToStandardRequest{
			Action:    "product.classification.convert_to_standard",
			Reason:    Reason,
			VariantID: id,
		}
		reqBody = append(reqBody, body)

		if i%maxProductsPerRequest == 0 {
			fmt.Printf("\nConverting %d variants to a standard product. \n ids: %v", len(reqBody), ids)
			_, err = vendClient.MakeRequest(http.MethodPost, url, reqBody)
			if err != nil {
				fmt.Printf(color.RedString("Failed to fix product: %v", err))
				return
			}
			reqBody = make([]ConvertVariantToStandardRequest, 0, maxProductsPerRequest)
		}
	}

	if len(reqBody) > 0 {
		_, err = vendClient.MakeRequest(http.MethodPost, url, reqBody)
		if err != nil {
			fmt.Printf(color.RedString("Failed to fix product: %v", err))
		}
	}

	fmt.Println(color.GreenString("\n\nFinished! 🎉\n"))
}
