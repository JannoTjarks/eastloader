package cmd

import (
	"jannotjarks/eastloader/internal/visiolink"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(onCmd)
}

var onCmd = &cobra.Command{
	Use:   "on",
	Short: "Downloads the \"Ostfriesische Nachrichten\"",
	Run: func(cmd *cobra.Command, args []string) {
		creds := visiolink.Credentials{
			Username: os.Getenv("ON_DOWNLOADER_USERNAME"),
			Password: os.Getenv("ON_DOWNLOADER_PASSWORD"),
		}

		paper := visiolink.GetOstfriesischeNachrichtenMetadata()

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}

		handler := visiolink.VisiolinkHandler{Client: client, Paper: paper, Creds: creds}
		handler.RunDownloadRoutine()
	},
}
