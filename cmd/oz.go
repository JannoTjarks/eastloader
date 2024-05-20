package cmd

import (
	"jannotjarks/eastloader/visiolink"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ozCmd)
}

var ozCmd = &cobra.Command{
	Use:   "oz",
	Short: "Downloads the \"Ostfriesischen Zeitung\"",
	Run: func(cmd *cobra.Command, args []string) {
		creds := visiolink.Credentials{
			Username: os.Getenv("OZ_DOWNLOADER_USERNAME"),
			Password: os.Getenv("OZ_DOWNLOADER_PASSWORD"),
		}

		paper := visiolink.GetOstfriesenZeitungMetadata()

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}

		handler := visiolink.VisiolinkHandler{Client: client, Paper: paper, Creds: creds}
		handler.RunDownloadRoutine()
	},
}
