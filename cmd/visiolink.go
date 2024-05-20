package cmd

import (
	"fmt"
	"jannotjarks/eastloader/visiolink"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(visiolinkCmd)
	visiolinkCmd.PersistentFlags().String("name", "", "The (short) name of the wanted paper")
}

var visiolinkCmd = &cobra.Command{
	Use:   "visiolink",
	Short: "Downloads an epaper based on Visiolink ApS",
	Run: func(cmd *cobra.Command, args []string) {
		var creds visiolink.Credentials

		var paper visiolink.Paper

		name, _ := cmd.Flags().GetString("name")
		switch name {
		case "":
			fmt.Println("You have to specify a newspaper!")
			fmt.Println("Currently there is support for:")
			fmt.Println("\t1. Ostfriesische Nachrichten \t(on)")
			fmt.Println("\t2. Ostfriesischen Zeitung \t(oz)")
			os.Exit(1)
		case "oz":
			creds = visiolink.Credentials{
				Username: os.Getenv("OZ_DOWNLOADER_USERNAME"),
				Password: os.Getenv("OZ_DOWNLOADER_PASSWORD"),
			}
			paper = visiolink.GetOstfriesenZeitungMetadata()
		case "on":
			creds = visiolink.Credentials{
				Username: os.Getenv("ON_DOWNLOADER_USERNAME"),
				Password: os.Getenv("ON_DOWNLOADER_PASSWORD"),
			}

			paper = visiolink.GetOstfriesischeNachrichtenMetadata()
		}

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}

		handler := visiolink.VisiolinkHandler{Client: client, Paper: paper, Creds: creds}
		handler.RunDownloadRoutine()
	},
}
