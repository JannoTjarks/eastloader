package main

import (
	"jannotjarks/on-downloader/internal/auth"
	"jannotjarks/on-downloader/internal/visiolink"
	"net/http"
	"net/http/cookiejar"
)

func main() {
	creds := auth.GetCredentials()
	paper := visiolink.GetOstfriesischeNachrichtenMetadata()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	handler := visiolink.VisiolinkHandler{Client: client, Paper: paper, Creds: creds}
	handler.RunDownloadRoutine()
}
