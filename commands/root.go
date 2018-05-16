package cmd

import (
	"fmt"
	"os"

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
)

// importSuppliersCmd represents the importSuppliers command
var rootCmd = &cobra.Command{
	Use:   "vendcli",
	Short: "A CLI tool for Vend",
	Long: `
Vend is a CLI tool to interact with the Vend API.
Commands represent tools and flags are parameters for those tool to run.

There are two sets of flags, global flags and command flags. Global flags such as 
Domain Prefix and Token are required on all commands and then command flags 
are passed depending on the tool. 
	
To run a tool:
Type vendcli followed by the command you wish to run and then the flags
[vend] [command] [flags]

Example: 
vendcli export-customers -d domainprefix -t token`,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Get store info from command line flags.
	rootCmd.PersistentFlags().StringVarP(&DomainPrefix, "Domain", "d", "", "The Vend store name (prefix in xxxx.vendhq.com)")
	rootCmd.PersistentFlags().StringVarP(&Token, "Token", "t", "", "Personal API Access Token for the store, generated from Setup -> Personal Tokens.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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

		// Search config in home directory with name ".vendcli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vendcli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
