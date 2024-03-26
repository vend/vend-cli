package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedUpdateStoreCreditRequests struct {
	CustomerID   string
	CustomerCode string
	Amount       string
	Reason       string
}

// Command config
var (
	submitMode                      string
	failedUpdateStoreCreditRequests []FailedUpdateStoreCreditRequests

	updateStorecreditCmd = &cobra.Command{
		Use:   "update-storecredits",
		Short: "Update Store Credits",
		Long: fmt.Sprintf(`
update-storecredits will update the credit balances for a given list of customers

It offers two modes:
- "replace" [default]
	replace mode REPLACES the current customer balance with a new, desired balance

	csv format:
	+-------------+---------------+-------------+
	| customer_id | customer_code | new_balance |
	+-------------+---------------+-------------+
	| <id>        | <code>        |         0.0 |
	+-------------+---------------+-------------+

- "adjust"
	adjust will ADJUST the current customer balance by a given +/- amount. 
	To add value specify a positive amount
	To subtract value specify a negative amount

	csv format:
	+-------------+---------------+--------+
	| customer_id | customer_code | amount |
	+-------------+---------------+--------+
	| <id>        | <code>        |    0.0 |
	+-------------+---------------+--------+

*Note: both modes must have the requisite headers. 
However, values are only required in customer_id OR customer_code not both

Example:
%s`, color.GreenString("vendcli update-storecredits -d DOMAINPREFIX -t TOKEN -f FILENAME.csv -m replace")),

		Run: func(cmd *cobra.Command, args []string) {
			updateStoreCredit()
		},
	}
)

func init() {
	// Flag
	updateStorecreditCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateStorecreditCmd.MarkFlagRequired("Filename")

	updateStorecreditCmd.Flags().StringVarP(&submitMode, "mode", "m", "replace", "the method used for updating: replace, adjust")

	rootCmd.AddCommand(updateStorecreditCmd)
}

func updateStoreCredit() {

	// Test mode option is set correctly
	submitMode = strings.ToLower(submitMode)
	if !(submitMode == "adjust" || submitMode == "replace") {
		err := fmt.Errorf("'%s' is not a valid option for -m mode. Mode should be 'adjust' or replace'", submitMode)
		messenger.ExitWithError(err)
	}
	fmt.Printf("\nRunning command in %s mode\n", color.YellowString(strings.ToUpper(submitMode)))

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	fmt.Println("\nReading Store Credits CSV...")
	csvRows, usesCustomerCodes, err := readStoreCreditCSV(FilePath, submitMode)
	if err != nil {
		err = fmt.Errorf("couldnt read Store Credits CSV file,  %s", err)
		messenger.ExitWithError(err)
	}

	fmt.Println("\nMake Transactions...")
	transactions, err := makeTransactions(csvRows, usesCustomerCodes)
	if err != nil {
		err = fmt.Errorf("couldnt make transactions,  %s", err)
		messenger.ExitWithError(err)
	}

	numTransactions := len(transactions)
	fmt.Printf("\n Posting %v Store Credits..\n", numTransactions)
	numPosted := postStoreCredit(transactions)

	if len(failedUpdateStoreCreditRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_update_storecredit_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedUpdateStoreCreditRequests)
		if err != nil {
			err = fmt.Errorf("failed to write error csv: %w", err)
			messenger.ExitWithError(err)
		}
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\nSuccesfully Posted %s of %s Store Credits \n",
		strconv.Itoa(numPosted), strconv.Itoa(numTransactions)))
}

// Read passed CSV, returns a slice of Store Credits
func readStoreCreditCSV(filePath string, submitMode string) ([]vend.StoreCreditCsv, bool, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, false, err
	}
	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file
	file, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		bar.AbortBar()
		p.Wait()
		return nil, false, err
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("error reading header row")
		return nil, false, err
	}

	// Check each header in the row is same as our template. Fail if not
	if err = checkHeaders(submitMode, headerRow); err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, false, err
	}

	// Read the rest of the data from the CSV
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("error reading data from csv: %w", err)
		messenger.ExitWithError(err)
	}

	var csvStructs []vend.StoreCreditCsv
	usesCustomerCodes := false

	// Loop through rows and assign them to a StoreCreditCsv struct.
	for _, row := range rawData {
		amount, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			err = errors.New("invalid amount")
			failedUpdateStoreCreditRequests = append(failedUpdateStoreCreditRequests,
				FailedUpdateStoreCreditRequests{
					CustomerID:   row[0],
					CustomerCode: row[1],
					Amount:       row[2],
					Reason:       err.Error(),
				})
			continue
		}

		rowStruct := vend.StoreCreditCsv{
			CustomerID:   &row[0],
			CustomerCode: &row[1],
			Amount:       &amount,
		}

		// make sure there is at least a customer id or customer code present
		if *rowStruct.CustomerCode != "" {
			usesCustomerCodes = true
		} else if *rowStruct.CustomerID == "" {
			err = errors.New("you must have at least customer_id or customer_code")
			failedUpdateStoreCreditRequests = append(failedUpdateStoreCreditRequests,
				FailedUpdateStoreCreditRequests{
					CustomerID:   row[0],
					CustomerCode: row[1],
					Amount:       row[2],
					Reason:       err.Error(),
				})
			continue
		}
		csvStructs = append(csvStructs, rowStruct)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return csvStructs, usesCustomerCodes, err
}

// Post each Store Credits to Vend
func postStoreCredit(transactions []vend.StoreCreditTransaction) int {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(transactions), "Posting Store Credits")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	var count int = 0
	for _, transaction := range transactions {
		bar.Increment()
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/store_credits/%s/transactions", DomainPrefix, transaction.CustomerID)
		resp, err := vendClient.MakeRequest("POST", url, transaction)
		if err != nil {
			err = fmt.Errorf("error posting store credit transaction: %s response: %s", err, string(resp))
			failedUpdateStoreCreditRequests = append(failedUpdateStoreCreditRequests,
				FailedUpdateStoreCreditRequests{
					CustomerID:   transaction.CustomerID,
					CustomerCode: "",
					Amount:       strconv.FormatFloat(transaction.Amount, 'f', -1, 64),
					Reason:       err.Error(),
				})
			continue
		}
		count += 1
	}
	p.Wait()
	return count
}

// checkHeaders makes sure the submitted CSV headers matches our desired format
func checkHeaders(submitMode string, headerRow []string) error {
	headers := [3]string{"customer_id", "customer_code", ""}
	if submitMode == "replace" {
		headers[2] = "new_balance"
	} else if submitMode == "adjust" {
		headers[2] = "amount"
	}

	for i := range headerRow {
		if headerRow[i] != headers[i] {
			err := fmt.Errorf("mismatch in headers, this mode (%s) needs three headers: customer_id, customer_code, %s. No header match for: %s instead got: %s",
				submitMode, string(headers[2]), string(headers[i]), string(headerRow[i]))
			return err
		}
	}
	return nil
}

// makeTransactions takes the info from csvStructs and makes bodys for our POST requests
func makeTransactions(csvRows []vend.StoreCreditCsv, usesCustomerCodes bool) ([]vend.StoreCreditTransaction, error) {
	var transactions []vend.StoreCreditTransaction
	var userID string

	// fetch data from vend
	user, customers, storeCredits, err := fetchDataForTransactions(usesCustomerCodes)
	if err != nil {
		err = fmt.Errorf("failed to fetch data for transactions: %v", err)
		return nil, err
	}

	// parse data - can't make this a wg since they all update the same data
	if usesCustomerCodes {
		csvRows, err = addCustomerIDsToStruct(csvRows, customers)
		if err != nil {
			err = fmt.Errorf("not able to map customer ids to codes: %w", err)
			return nil, err
		}
	}
	if submitMode == "replace" {
		csvRows, err = updateAmounts(csvRows, storeCredits)
		if err != nil {
			err = fmt.Errorf("not able to update amounts: %w", err)
			return nil, err
		}
	}
	if user.ID != nil {
		userID = *user.ID
	} else {
		err = fmt.Errorf("failed to get user ID from token - check your token")
		return nil, err
	}

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(csvRows), "Make Transactions")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	for _, rowStruct := range csvRows {
		bar.Increment()
		var transType string
		clientID := uuid.New().String()

		if *rowStruct.Amount < 0 {
			transType = "REDEMPTION"
		} else {
			transType = "ISSUE"
		}
		transaction := vend.StoreCreditTransaction{
			CustomerID: *rowStruct.CustomerID,
			Amount:     *rowStruct.Amount,
			Type:       transType,
			ClientID:   &clientID,
			UserID:     &userID,
		}
		transactions = append(transactions, transaction)
	}

	p.Wait()
	return transactions, nil
}

func fetchDataForTransactions(usesCustomerCodes bool) (vend.User, []vend.Customer, []vend.StoreCredit, error) {
	var err error
	var user vend.User
	var customers []vend.Customer
	var storeCredits []vend.StoreCredit

	// fetch data from vend
	p, err := pbar.CreateMultiBarGroup(3, Token, DomainPrefix)
	if err != nil {
		fmt.Println("error creating progress bar: ", err)
	}

	p.FetchDataWithProgressBar("user")
	if usesCustomerCodes {
		p.FetchDataWithProgressBar("customers")
	}
	if submitMode == "replace" {
		p.FetchDataWithProgressBar("store-credits")
	}
	p.MultiBarGroupWait()

	for err = range p.ErrorChannel {
		return user, customers, storeCredits, err
	}
	for data := range p.DataChannel {
		switch d := data.(type) {
		case vend.User:
			user = d
		case []vend.Customer:
			customers = d
		case []vend.StoreCredit:
			storeCredits = d
		}
	}

	return user, customers, storeCredits, nil
}

// addCustomerIDsToStruct finds the associated customer_ids for a given customer_code
func addCustomerIDsToStruct(csvRows []vend.StoreCreditCsv, customers []vend.Customer) ([]vend.StoreCreditCsv, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(csvRows), "Mapping Customer IDs")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	// Build Customer Map
	customerMap := vend.CustomerMap(customers)
	// Loop through our csv row structs and attach a customer_id if one doesn't exist,
	// if succesful include that in an updated struct array
	var updatedCSVRows []vend.StoreCreditCsv
	for _, rowStruct := range csvRows {
		bar.Increment()
		// skip if we've already got a customer_id
		if len(*rowStruct.CustomerID) > 0 {
			updatedCSVRows = append(updatedCSVRows, rowStruct)
			continue
		} else {
			// check the customer_code has a valid map to a customer_id
			// if so, set the customer_id
			if customerID, ok := customerMap[*rowStruct.CustomerCode]; ok {
				*rowStruct.CustomerID = customerID
				updatedCSVRows = append(updatedCSVRows, rowStruct)
			} else {
				err = fmt.Errorf("invalid customer code. can not match to customer id")
				failedUpdateStoreCreditRequests = append(failedUpdateStoreCreditRequests,
					FailedUpdateStoreCreditRequests{
						CustomerID:   *rowStruct.CustomerID,
						CustomerCode: *rowStruct.CustomerCode,
						Amount:       strconv.FormatFloat(*rowStruct.Amount, 'f', -1, 64),
						Reason:       err.Error(),
					})
			}
		}
	}
	p.Wait()
	return updatedCSVRows, nil
}

// if mode is replace the transaction amount should be newBalance - currentBalance
func updateAmounts(csvRows []vend.StoreCreditCsv, storeCredits []vend.StoreCredit) ([]vend.StoreCreditCsv, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(csvRows), "Updating Amounts")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}
	// make customer id -> customer balance map
	creditMap := vend.CreditMap(storeCredits)
	// loop through structs and update the amount
	var updatedCSVRows []vend.StoreCreditCsv
	for _, rowStruct := range csvRows {
		bar.Increment()
		if currentBalance, ok := creditMap[*rowStruct.CustomerID]; ok {
			*rowStruct.Amount = *rowStruct.Amount - currentBalance
			updatedCSVRows = append(updatedCSVRows, rowStruct)
		} else {
			err = fmt.Errorf("invalid customer id. can not match to customer balance")
			failedUpdateStoreCreditRequests = append(failedUpdateStoreCreditRequests,
				FailedUpdateStoreCreditRequests{
					CustomerID:   *rowStruct.CustomerID,
					CustomerCode: *rowStruct.CustomerCode,
					Amount:       strconv.FormatFloat(*rowStruct.Amount, 'f', -1, 64),
					Reason:       err.Error(),
				})
		}
	}
	p.Wait()

	if len(updatedCSVRows) == 0 {
		err = fmt.Errorf("no valid customer ids found")
		return nil, err
	}
	return updatedCSVRows, nil
}
