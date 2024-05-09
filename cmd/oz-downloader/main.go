package main

import (
	"jannotjarks/on-downloader/internal/visiolink"
	"net/http"
	"net/http/cookiejar"
	"os"
)

func main() {
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
}
