package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Variables for Client authentication details and flags
var (
	DomainPrefix string
	Token        string
	vendClient   *vend.Client
	FilePath     string
	cfgFile      string
	logo         = color.GreenString(`                             _ 
 __   __   ___   _ __     __| |
 \ \ / /  / _ \ | '_ \   / _  |
  \ V /  |  __/ | | | | | (_| |
   \_/    \___| |_| |_|  \__,_|`)
)

// Command Config
var rootCmd = &cobra.Command{
	Use: "vendcli",
	Short: fmt.Sprintf(`
%s`, logo)}

func init() {
	cobra.OnInitialize(initConfig)

	// Get store info from command line flags.
	rootCmd.PersistentFlags().StringVarP(&DomainPrefix, "Domain", "d", "", "The Vend store name (prefix in xxxx.vendhq.com)")
	rootCmd.PersistentFlags().StringVarP(&Token, "Token", "t", "", "API Access Token for the store, Setup -> Personal Tokens.")
	rootCmd.MarkPersistentFlagRequired("Domain")
	rootCmd.MarkPersistentFlagRequired("Token")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vend" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vendcli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Read passed CSV and returns the IDs
func readCSV(FilePath string) ([]string, error) {

	// Open our provided CSV file
	file, err := os.Open(FilePath)
	if err != nil {
		fmt.Println("Could not read from CSV file")
		return nil, err
	}

	// Make sure to close the file
	defer file.Close()

	// Create CSV read on our file
	reader := csv.NewReader(file)

	// Read the rest of the data from the CSV
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rowNumber int
	entities := []string{}

	// Loop through rows and assign them
	for _, row := range rows {
		rowNumber++
		entityIDs := row[0]
		entities = append(entities, entityIDs)
	}

	return entities, err
}
