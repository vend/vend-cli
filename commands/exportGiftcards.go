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
var exportGiftcardsCmd = &cobra.Command{
	Use:   "export-giftcards",
	Short: "Export Gift Cards",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-giftcards -d DOMAINPREFIX -t TOKEN")),

	Run: func(cmd *cobra.Command, args []string) {
		getGiftCards()
	},
}

func init() {
	rootCmd.AddCommand(exportGiftcardsCmd)
}

// Run executes the process of exporting Gift Cards then writing them to CSV.
func getGiftCards() {

	// Create new Vend Client
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Gift Cards
	fmt.Println("\nRetrieving Gift Cards from Vend...")
	giftCards, err := vc.GiftCards()
	if err != nil {
		log.Printf(color.RedString("Failed while retrieving Gift Cards: %v", err))
		panic(vend.Exit{1})
	}

	// Write Gift Cards to CSV
	fmt.Println("Writing Gift Cards to CSV file...")
	err = gcWriterFile(giftCards)
	if err != nil {
		log.Printf(color.RedString("Failed while writing Gift Cards to CSV: %v", err))
		panic(vend.Exit{1})
	}

	fmt.Println(color.GreenString("\nExported %v Gift Cards 🎉\n", len(giftCards)))
}

// WriteFile writes Gift Cards to CSV
func gcWriterFile(giftCards []vend.GiftCard) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_giftcard_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Printf(color.RedString("Failed to create CSV: %v", err))
		panic(vend.Exit{1})
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "ID")
	header = append(header, "number")
	header = append(header, "sale_ID")
	header = append(header, "created_at")
	header = append(header, "Expires_at")
	header = append(header, "Status")
	header = append(header, "Balance")
	header = append(header, "Total_Sold")
	header = append(header, "Total_Redeemed")

	// Commit the header.
	writer.Write(header)

	// Now loop through each gift card object and populate the CSV.
	for _, giftcard := range giftCards {

		var ID, number, saleID, createdAt, expiresAt, status, balance, totalSold, totalRedeemed string

		if giftcard.ID != nil {
			ID = *giftcard.ID
		}
		if giftcard.Number != nil {
			number = *giftcard.Number
		}
		if giftcard.SaleID != nil {
			saleID = *giftcard.SaleID
		}
		if giftcard.CreatedAt != nil {
			createdAt = *giftcard.CreatedAt
		}
		if giftcard.ExpiresAt != nil {
			expiresAt = *giftcard.ExpiresAt
		}
		if giftcard.Status != nil {
			status = *giftcard.Status
		}
		if giftcard.Balance != nil {
			balance = fmt.Sprintf("%v", *giftcard.Balance)
		}
		if giftcard.TotalSold != nil {
			totalSold = fmt.Sprintf("%v", *giftcard.TotalSold)
		}
		if giftcard.TotalRedeemed != nil {
			totalRedeemed = fmt.Sprintf("%v", *giftcard.TotalRedeemed)
		}

		var record []string
		record = append(record, ID)
		record = append(record, number)
		record = append(record, saleID)
		record = append(record, createdAt)
		record = append(record, expiresAt)
		record = append(record, status)
		record = append(record, balance)
		record = append(record, totalSold)
		record = append(record, totalRedeemed)
		writer.Write(record)
	}

	writer.Flush()
	return err
}
