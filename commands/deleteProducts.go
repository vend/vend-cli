package cmd

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// deleteProductsCmd represents the deleteProducts command
var deleteProductsCmd = &cobra.Command{
	Use:   "delete-products",
	Short: "Delete Products",
	Long: fmt.Sprintf(`
This tool requires a CSV of Product IDs, no headers.

Example:
%s`, color.GreenString("vendcli delete-products -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		deleteProducts()
	},
}

func init() {
	// Flag
	deleteProductsCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	deleteProductsCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(deleteProductsCmd)
}

func deleteProducts() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Fatalf("Failed to get ids from the file: %s", FilePath)
	}

	// Make the requests
	for _, id := range ids {
		fmt.Printf("\nDeleting %v", id)
		url := fmt.Sprintf("https://%s.vendhq.com/api/products/%s", DomainPrefix, id)
		_, _, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf(" Failed to delete Products: %v", err)
		}
	}
	fmt.Println(color.GreenString("\n\nFinished!\n"))
}
