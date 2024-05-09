package main

import (
	"jannotjarks/on-downloader/internal/visiolink"
	"net/http"
	"net/http/cookiejar"
	"os"
)

func main() {
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
}
