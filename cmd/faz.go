package cmd

import (
	"jannotjarks/eastloader/faz"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fazCmd)
	fazCmd.PersistentFlags().String("date", "", "The release date of the wanted paper")
}

var fazCmd = &cobra.Command{
	Use:   "faz",
	Short: "Downloads an epaper based on faz ApS",
	Run: func(cmd *cobra.Command, args []string) {
		var creds = faz.Credentials{
			Email:      os.Getenv("FAZ_DOWNLOADER_USERNAME"),
			Password:   os.Getenv("FAZ_DOWNLOADER_PASSWORD"),
		}

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}

		// date, _ := cmd.Flags().GetString("date")
		handler := faz.FazHandler{Client: client, Creds: creds}
		faz.Login(handler)
        body, _ := faz.GetKioskHtml(handler)
        faz.GetFazPaper(body)
	},
}
