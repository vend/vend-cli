package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var voidGiftcardsCmd = &cobra.Command{
	Use:   "void-giftcards",
	Short: "Void Gift Cards",
	Long: fmt.Sprintf(`
This tool requires the Gift Card CSV template, you can download it here: http://bit.ly/vendclitemplates,

Example:
%s`, color.GreenString("vendcli void-giftcards -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		voidGiftCards()
	},
}

func init() {
	// Flags
	voidGiftcardsCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	voidGiftcardsCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(voidGiftcardsCmd)
}

func voidGiftCards() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get Gift Card Numbers from CSV
	fmt.Printf("\nReading Gift Card CSV\n")
	ids, err := readGiftCardCSV(FilePath)
	if err != nil {
		log.Fatalf("Failed to get gift card numbers from the file: %s", FilePath)
	}

	// Voiding Gift Cards
	fmt.Printf("%d Gift Card to Void.\n", len(ids))
	for _, id := range ids {
		err = requester(id)
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

// Read passed CSV, returns a slice of Gift Cards
func readGiftCardCSV(FilePath string) ([]string, error) {

	headers := []string{"number"}

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		fmt.Println("Could not read from CSV file")
		return nil, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, _ := reader.Read()

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			fmt.Println("Found error in header rows.")
			log.Fatalf("No header match for: %s Instead got: %s.",
				string(headers[i]), string(headerRow[i]))
		}
	}

	// Read the rest of the data from the CSV.
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rowNumber int
	giftCardIDs := []string{}

	// Loop through rows and assign them to Gift Card.
	for _, row := range rows {
		rowNumber++
		giftcardNo := row[0]
		giftcardNo = strings.Trim(giftcardNo, "\u00a0 ") // removes nonbreaking space, if present. Support has been seeing these in some xlsx exports
		giftCardIDs = append(giftCardIDs, giftcardNo)
	}

	return giftCardIDs, err
}

func requester(id string) error {

	// Create the Vend URL to delete Gift Card
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/balances/gift_cards/%s", DomainPrefix, id)

	err, _ := vendClient.MakeRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to void Gift Card: %s", err)
	}

	return nil
}
