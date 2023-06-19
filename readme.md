<p align="center">
  <img src="https://cl.ly/qurB/greyicon.png" height="64">
  <h3 align="center">Vend</h3>
  <p align="center">A CLI tool to interact with the Vend API.<p></p>


## Installation

```
$ git clone https://github.com/vend/vend-cli.git
$ cd vend-cli
```
Then vend-cli can be run through
```
$ go run main.go [command name]
```
e.g.
```
$ go run main.go --help

                             _ 
 __   __   ___   _ __     __| |
 \ \ / /  / _ \ | '_ \   / _  |
  \ V /  |  __/ | | | | | (_| |
   \_/    \___| |_| |_|  \__,_|

Usage:
  vendcli [command]

Available Commands:
  delete-customers                      Delete Customers
  delete-products                       Delete Products
  export-auditlog                       Export Audit Log
  export-customers                      Export Customers
  export-giftcards                      Export Gift Cards
  export-images                         Export Product Images
  export-sales                          Export Sales
  export-storecredits                   Export Store Credits
  export-suppliers                      Export Suppliers 
  export-users                          Export Users
  fix-products-variant-to-standard      Convert variant product to standard
  help                                  Help about any command
  import-images                         Import Product Images
  import-product-codes                  Import Product Codes
  import-storecredits                   Import Store Credits
  import-suppliers                      Import Suppliers
  loyalty-adjustment                    Customer Loyalty Adjustment
  void-giftcards                        Void Gift Cards
  void-sales                            Void Sales

Flags:
  -d, --Domain string   The Vend store name (prefix in xxxx.vendhq.com)
  -t, --Token string    API Access Token for the store, Setup -> Personal Tokens.
  -h, --help            help for vendcli

Use "vendcli [command] --help" for more information about a command.
```


## Commands

- Delete Customers
- Delete Products
- Export Audit Log
- Export Sales Ledger
- Export Customers
- Export Gift Cards
- Export Store Credits
- Export Suppliers
- Export Audit Log
- Export Images
- Fix Products - Converting variant to standard
- Import Images
- Import Product Codes
- Import Suppliers
- Import Store Credits
- Adjust Customer Loyalty
- Void Gift Cards
- Void Sales

## Usage Examples

When running a command you need to pass the flags that specify the parameters for that tool. There are two sets of flags, global flags and command flags. Global flags such as domain prefix and token are required on all commands and command flags are passed depending on the tool.

#### Delete Customers

	$ vendcli delete-customers -d domainprefix -t token -f filename.csv

#### Delete Products

	$ vendcli delete-products -d domainprefix -t token -f filename.csv

#### Export Audit Log

	$ vendcli audit-log -d domainprefix -t token -F 2018-03-15T16:30:30 -T 2018-04-01T18:30:00

#### Export Sales Ledger

	$ vendcli export-sales -d domainprefix -t token -z timezone

#### Export Customers

	$ vendcli export-customers -d domainprefix -t token

#### Export Gift Cards

	$ vendcli export-giftcards -d domainprefix -t token

#### Export Store Credits

	$ vendcli export-storecredits -d domainprefix -t token

#### Export Suppliers

	$ vendcli export-suppliers -d domainprefix -t token	

#### Export Audit Log

	$ vendcli export-auditlog -d domainprefix -t token -F 2018-03-01T16:30:30 -T 2018-04-01T18:30:00	

#### Export Images

	$ vendcli export-images -d domainprefix -t token

#### Fix Products - Converting variant to standard

	$ vendcli fix-products-variant-to-standard -d domainprefix -t TOKEN -f FILENAME.csv -r ''

#### Import Images

	$ vendcli import-images -d domainprefix -t token -f filename.csv
	
#### Import Product Codes

	$ vendcli import-product-codes -d domainprefix -t token -f filename.csv

#### Import Suppliers

	$ vendcli import-suppliers -d domainprefix -t token -f filename.csv

#### Void Gift Cards

	$ vendcli void-giftcards -d domainprefix -t token -f filename.csv

#### Void Sales

	$ vendcli void-sales -d domainprefix -t token -f filename.csv

## Need Help?

If you are unsure which flags are needed for the command just type the command followed by --help, which will show you a breakdown of the required flags and a download link if a template file is needed.

	$ vendcli command-name --help

## Related Repositories

- [GoVend](https://github.com/jackharrisonsherlock/govend)
