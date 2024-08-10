package cmd

import (
	"fmt"
	"jannotjarks/eastloader/visiolink"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(visiolinkCmd)
	visiolinkCmd.PersistentFlags().String("name", "", "The (short) name of the wanted paper")
	visiolinkCmd.PersistentFlags().String("date", "", "The release date of the wanted paper")
}

var visiolinkCmd = &cobra.Command{
	Use:   "visiolink",
	Short: "Downloads an epaper based on Visiolink ApS",
	Run: func(cmd *cobra.Command, args []string) {
		var creds visiolink.Credentials

		var paper visiolink.Metadata

		metadataMap := visiolink.MakeVisiolinkMetadataMap()

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

			paper = metadataMap["OstfriesenZeitung"]
		case "on":
			creds = visiolink.Credentials{
				Username: os.Getenv("ON_DOWNLOADER_USERNAME"),
				Password: os.Getenv("ON_DOWNLOADER_PASSWORD"),
			}

			paper = metadataMap["OstfriesischeNachrichten"]
		}

		jar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: jar,
		}

		date, _ := cmd.Flags().GetString("date")

		handler := visiolink.VisiolinkHandler{Client: client, Meta: paper, Creds: creds}
		RunDownloadRoutine(date, handler)
	},
}

func RunDownloadRoutine(date string, handler visiolink.VisiolinkHandler) {
	var issue visiolink.Catalog
	if date == "" {
		issue = visiolink.GetNewestIssue(handler)
	} else {
		issue = visiolink.GetSpecificIssue(handler, date)
	}

	fileName := visiolink.GenerateFileName(issue)

	fileExists, errFileExists := checkIfFileExists(fileName)
	if fileExists {
		fmt.Printf("Download will be skipped, because there is already a file with the name \"%s\"\n", fileName)
		return
	}

	if errFileExists != nil {
		log.Fatal(errFileExists)
	}

	fmt.Println(issue.PublicationDate)

	loginUrl, err := visiolink.GetLoginUrl(handler)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := visiolink.ExtractSecretFromLoginUrl(handler, loginUrl)
	if err != nil {
		log.Fatal(err)
	}

	accessUrl, err := visiolink.GetIssueAccessUrl(handler, secret, issue.Catalog)
	if err != nil {
		log.Fatal(err)
	}

	accessKey, err := visiolink.GetIssueAccessKey(handler, accessUrl)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool, 1)
	go visiolink.DownloadIssue(handler, done, issue.Catalog, accessKey, fileName)
	waitForHttpResponse(done)
}
