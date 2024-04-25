package main

import (
	"fmt"
	"jannotjarks/on-downloader/internal/auth"
	"jannotjarks/on-downloader/internal/visiolink"
	"log"
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

	issue := handler.GetNewestIssue()
	fmt.Println(issue.PublicationDate)

	loginUrl, err := handler.GetLoginUrl()
	if err != nil {
		log.Fatal(err)
	}

	secret, err := handler.ExtractSecretFromLoginUrl(loginUrl)
	if err != nil {
		log.Fatal(err)
	}

	accessUrl, err := handler.GetIssueAccessUrl(secret, issue.Catalog)
	if err != nil {
		log.Fatal(err)
	}

	accessKey, err := handler.GetIssueAccessKey(accessUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := handler.GenerateFileName(issue)

	err = handler.DownloadIssue(issue.Catalog, accessKey, fileName)
	if err != nil {
		log.Fatal(err)
	}
}
