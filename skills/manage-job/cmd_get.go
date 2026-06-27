package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve all tracked job applications",
	Long:  "Fetches all job applications from the spreadsheet. Returns JSON to stdout. Pass key-value pairs as arguments to filter.",
	Example: `  manage-job get
  manage-job get page 1 pageSize 10 search "Acme" industry "Tech" status "Applied Only" order "desc"`,
	Run: func(cmd *cobra.Command, args []string) {
		scriptURL := loadScriptURL()

		if len(args) > 0 {
			params := url.Values{}
			for i := 0; i+1 < len(args); i += 2 {
				params.Set(args[i], args[i+1])
			}
			if len(params) > 0 {
				scriptURL = scriptURL + "?" + params.Encode()
			}
		}

		resp, err := http.Get(scriptURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(string(body))
	},
}
