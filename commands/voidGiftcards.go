package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var voidGiftcardsCmd = &cobra.Command{
	Use:   "void-giftcards",
	Short: "Void Gift Cards",
	Long: fmt.Sprintf(`
This tool requires the Gift Card CSV template, you can download it here: http://bit.ly/vendclitemplates,

Note: Due to an API limitation Redeemed Gift Cards are not able to be voided. 
If "Include Redeemed" is set to "TRUE", redeemed gift cards will first have a penny added to them and then voided.
If "Include Redeemed" is set to "FALSE", redeemed gift cards will be skipped.

Example Usage:
%s`, color.GreenString("vendcli void-giftcards -d DOMAINPREFIX -t TOKEN -r TRUE/FALSE -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		voidGiftCards()
	},
}

var includeRedeemedStr string

func init() {
	// Flags
	voidGiftcardsCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	voidGiftcardsCmd.MarkFlagRequired("Filename")
	voidGiftcardsCmd.Flags().StringVarP(&includeRedeemedStr, "include-redeemed", "r", "", "include redeemed Gift Cards: true or false")
	voidGiftcardsCmd.MarkFlagRequired("include-redeemed")
	rootCmd.AddCommand(voidGiftcardsCmd)
}

func voidGiftCards() {

	// Fix the includeRedeemed flag to either true or false
	includeRedeemed, err := strconv.ParseBool(includeRedeemedStr)
	if err != nil {
		// default to false
		includeRedeemed = false
	}

	if includeRedeemed {
		fmt.Printf("\nRunning command with redeemed gift cards included\n")
	} else {
		fmt.Printf("\nRunning void without including redeemed gift cards\n")
	}

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get user ID from token. This also checks if the token is valid.
	user, err := vc.User()
	if err != nil {
		fmt.Printf("Failed to get user ID from token - check your token")
		panic(vend.Exit{1})
	}

	// Fetch gift card data
	giftCards, err := vc.GiftCards()
	if err != nil {
		fmt.Printf("Failed to fetch gift card data: %v", err)
		panic(vend.Exit{1})
	}

	// Get Gift Card Numbers from CSV
	fmt.Printf("\nReading Gift Card CSV\n")
	ids, err := readGiftCardCSV(FilePath)
	if err != nil {
		fmt.Printf("Failed to get gift card numbers from the file: %s", FilePath)
		panic(vend.Exit{1})
	}

	giftCardBalances := makeGCHash(giftCards)

	var userID string
	if user.ID != nil {
		userID = *user.ID
	} else {
		fmt.Printf("Failed to get user ID from token - check your token")
		panic(vend.Exit{1})
	}

	// Voiding Gift Cards
	fmt.Printf("\nVoiding %d Gift Cards...\n", len(ids))

	for _, id := range ids {
		balance, exists := giftCardBalances[id]
		if !exists {
			fmt.Printf("Gift card number %s not found in Vend, skipping..\n", id)
			continue
		}

		if balance == 0 {
			if includeRedeemed {
				// Make a POST request to add $0.01 to the gift card
				err = addTransaction(id, userID)
				if err != nil {
					fmt.Printf("Failed to add transaction for gift card %s: %v\n", id, err)
				}
			} else {
				fmt.Printf("Gift card %s has a balance of zero, skipping..\n", id)
				continue
			}
		}

		// Send a DELETE request to void the gift card
		err = deleteGiftCard(id)
		if err != nil {
			fmt.Printf("Failed to void Gift Card %s: %v\n", id, err)
		}
	}

	fmt.Println(color.GreenString("\nFinished!\n"))

}

// make a gift card hash
func makeGCHash(giftCards []vend.GiftCard) map[string]float64 {
	gcHash := make(map[string]float64)
	for _, card := range giftCards {

		if card.Number != nil && card.Balance != nil {
			gcHash[*card.Number] = *card.Balance
		}
		// the "else" is handled elsewhere. We check the ids in the csv against the gift cards this hash.
		// If there is an error we let the user know
	}
	return gcHash
}

// POST a giftcard transaction
func addTransaction(id string, userID string) error {
	fmt.Printf("Adding transaction for Gift Card %s\n", id)
	clientID := generateUniqueClientID()
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/gift_cards/%s/transactions", DomainPrefix, id)
	data := map[string]interface{}{
		"amount":    0.01,
		"type":      "RELOADING",
		"user_id":   userID,
		"client_id": clientID,
	}

	_, err := vendClient.MakeRequest("POST", url, data)
	if err != nil {
		return fmt.Errorf("failed to add transaction for Gift Card %s: %v\n", id, err)
	}

	return nil
}

// DELETEs a giftcard
func deleteGiftCard(id string) error {
	fmt.Printf("Voiding Gift Card %s\n", id)
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/balances/gift_cards/%s", DomainPrefix, id)
	_, err := vendClient.MakeRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to void Gift Card %s: %v\n", id, err)
	}
	return nil
}

func generateUniqueClientID() string {
	return uuid.New().String()
}

// Read passed CSV, returns a slice of Gift Cards
func readGiftCardCSV(FilePath string) ([]string, error) {

	headers := []string{"number"}

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		errorMsg := `error opening csv file - please check you've specified the right file

Tip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`
		fmt.Println(errorMsg, "\n")
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
			fmt.Printf("No header match for: %s Instead got: %s.\n",
				string(headers[i]), string(headerRow[i]))
			panic(vend.Exit{1})
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
