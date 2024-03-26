package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedGiftCardVoidRequest struct {
	GiftCardID string
	Reason     string
}

// Command config
var (
	includeRedeemedStr         string
	failedGiftCardVoidRequests []FailedGiftCardVoidRequest

	voidGiftcardsCmd = &cobra.Command{
		Use:   "void-giftcards",
		Short: "Void Gift Cards",
		Long: fmt.Sprintf(`
This tool requires a CSV of Gift Card Numbers, no headers.
"Numbers" is column two of the export-giftcard command

Note: Due to an API limitation Redeemed Gift Cards are not able to be voided. To overcome this, vendcli presents the following workaround:
If "Include Redeemed" is set to "TRUE", redeemed gift cards will first have a penny added to them and then voided.
If "Include Redeemed" is set to "FALSE", redeemed gift cards will be skipped.

FAQ:
Q: Why does this command need to fetch data from Vend?
A: Good question! We are gathering gift card numbers from Vend in order to verify that the number in the provided csv is valid before posting

Example Usage:
%s`, color.GreenString("vendcli void-giftcards -d DOMAINPREFIX -t TOKEN -r TRUE/FALSE -f FILENAME.csv")),

		Run: func(cmd *cobra.Command, args []string) {
			voidGiftCards()
		},
	}
)

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
		fmt.Printf("\nRunning command %s redeemed gift cards included\n", color.GreenString("WITH"))
	} else {
		fmt.Printf("\nRunning command %s redeemed gift cards included\n", color.GreenString("WITHOUT"))
	}

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get Gift Card Numbers from CSV
	fmt.Println("\nReading Gift Card CSV")
	ids, err := csvparser.ReadIdCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("failed to get gift card numbers from the file: %s, error: %w", FilePath, err)
		messenger.ExitWithError(err)
	}

	fmt.Println("\nRetrieving Info from Vend...")
	userID, giftCardBalances, err := fetchDataForGiftCardVoid()
	if err != nil {
		err = fmt.Errorf("failed to retrieve info from Vend: %w", err)
		messenger.ExitWithError(err)
	}

	// Voiding Gift Cards
	fmt.Printf("\nVoiding %d Gift Cards...\n", len(ids))
	succesfulPosts := postGiftCardDeleteRequets(ids, userID, includeRedeemed, giftCardBalances)

	if len(failedGiftCardVoidRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_void_gift_card_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedGiftCardVoidRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\nVoided %d out of %d gift-cards", succesfulPosts, len(ids)))

}

func fetchDataForGiftCardVoid() (string, map[string]float64, error) {

	var user vend.User
	var giftCards []vend.GiftCard
	var userID string

	p, err := pbar.CreateMultiBarGroup(2, Token, DomainPrefix)
	if err != nil {
		fmt.Println("error creating progress bar: ", err)
	}

	p.FetchDataWithProgressBar("gift-cards")
	p.FetchDataWithProgressBar("user")
	p.MultiBarGroupWait()
	for err = range p.ErrorChannel {
		return "", nil, err
	}
	for data := range p.DataChannel {
		switch d := data.(type) {
		case vend.User:
			user = d
		case []vend.GiftCard:
			giftCards = d
		}
	}

	giftCardBalances := makeGCHash(giftCards)
	if user.ID != nil {
		userID = *user.ID
	} else {
		err = fmt.Errorf("failed to get user ID from token - check your token")
		messenger.ExitWithError(err)
	}

	return userID, giftCardBalances, nil
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

func postGiftCardDeleteRequets(ids []string, userID string, includeRedeemed bool, giftCardBalances map[string]float64) int {
	var err error

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Deleting")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	var count int = 0
	for _, id := range ids {
		bar.Increment()
		balance, exists := giftCardBalances[id]
		if !exists {
			failedGiftCardVoidRequests = append(failedGiftCardVoidRequests,
				FailedGiftCardVoidRequest{
					GiftCardID: id,
					Reason:     "Gift Card ID not found in Vend",
				})
			continue
		}
		if balance == 0 {
			if includeRedeemed {
				// Make a POST request to add $0.01 to the gift card
				err = addTransaction(id, userID)
				if err != nil {
					err = fmt.Errorf("failed to add transaction for gift card: %v", err)
					failedGiftCardVoidRequests = append(failedGiftCardVoidRequests,
						FailedGiftCardVoidRequest{
							GiftCardID: id,
							Reason:     err.Error(),
						})
					continue
				}
			} else {
				err = fmt.Errorf("gift card %s has a balance of zero", id)
				failedGiftCardVoidRequests = append(failedGiftCardVoidRequests,
					FailedGiftCardVoidRequest{
						GiftCardID: id,
						Reason:     err.Error(),
					})
				continue
			}
		}

		err = postGiftCardDelete(id)
		if err != nil {
			err = fmt.Errorf("failed to void Gift Card: %v", err)
			failedGiftCardVoidRequests = append(failedGiftCardVoidRequests,
				FailedGiftCardVoidRequest{
					GiftCardID: id,
					Reason:     err.Error(),
				})
			continue
		}
		count += 1
	}
	p.Wait()
	return count
}

// POST a giftcard transaction
func addTransaction(id string, userID string) error {
	clientID := generateUniqueClientID()
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/gift_cards/%s/transactions", DomainPrefix, id)
	data := map[string]interface{}{
		"amount":    0.01,
		"type":      "RELOADING",
		"user_id":   userID,
		"client_id": clientID,
	}

	resp, err := vendClient.MakeRequest("POST", url, data)
	if err != nil {
		err = fmt.Errorf("error making request: %w response: %s", err, string(resp))
		return err
	}
	return nil
}

func generateUniqueClientID() string {
	return uuid.New().String()
}

// DELETEs a giftcard
func postGiftCardDelete(id string) error {
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/balances/gift_cards/%s", DomainPrefix, id)
	resp, err := vendClient.MakeRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w response: %s", err, string(resp))
	}
	return nil
}
