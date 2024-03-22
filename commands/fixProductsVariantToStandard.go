package cmd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"
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

type FailedConvertVariantToStandardRequest struct {
	VariantID string
	Reason    string
}

var failedRequests []FailedConvertVariantToStandardRequest

//var url = fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)

func fixProductsVariantToStandard() {
	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	if Reason == "" {
		log.Fatalf("Please specify a reason: -r '<REASON or ticket reference FOR RUNNING THIS>'")
	}

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := csvparser.ReadIdCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("Failed to get IDs from the file: %s\nError:%s", FilePath, err)
		messenger.ExitWithError(err)
	}

	// Make the requests
	fmt.Println("\nConverting variants to standard products...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Converting")
	if err != nil {
		fmt.Println("Error creating progress bar:", err)
	}

	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)

	maxProductsPerRequest := 10
	reqBody := make([]ConvertVariantToStandardRequest, 0, len(ids))
	for i, id := range ids {
		bar.Increment()
		// NOTE: This action is internal and could potentially change in a non-backward compatible manner.
		body := ConvertVariantToStandardRequest{
			Action:    "product.classification.convert_to_standard",
			Reason:    Reason,
			VariantID: id,
		}
		reqBody = append(reqBody, body)

		if i%maxProductsPerRequest == 0 {
			_, err = vendClient.MakeRequest(http.MethodPost, url, reqBody)
			if err != nil {
				retryConvertVariantToStandardRequests(reqBody)
			}
			reqBody = make([]ConvertVariantToStandardRequest, 0, maxProductsPerRequest)
		}
	}
	p.Wait()

	if len(reqBody) > 0 {
		_, err = vendClient.MakeRequest(http.MethodPost, url, reqBody)
		if err != nil {
			retryConvertVariantToStandardRequests(reqBody)
		}
	}

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_convert_variant_to_standard_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedRequests)
		if err != nil {
			messenger.ExitWithError(err)
		}
	}

	fmt.Println(color.GreenString("\n\nFinished! 🎉\n"))
}

// if the group fails a request, seperate the individual requests and retry
// log the failed requests
func retryConvertVariantToStandardRequests(failedGroup []ConvertVariantToStandardRequest) {

	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)

	for _, body := range failedGroup {
		_, err := vendClient.MakeRequest(http.MethodPost, url, body)
		if err != nil {
			failedRequests = append(
				failedRequests,
				FailedConvertVariantToStandardRequest{
					VariantID: body.VariantID, Reason: err.Error(),
				})
			continue
		}
	}
}
