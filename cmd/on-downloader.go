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

	issue := visiolink.GetNewestIssue(paper, client)
	fmt.Println(issue.PublicationDate)

	loginUrl, err := visiolink.GetLoginUrl(paper, client, creds)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := visiolink.ExtractSecretFromLoginUrl(paper, client, loginUrl)
	if err != nil {
		log.Fatal(err)
	}

	accessUrl, err := visiolink.GetIssueAccessUrl(paper, client, secret, issue.Catalog)
	if err != nil {
		log.Fatal(err)
	}

	accessKey, err := visiolink.GetIssueAccessKey(client, accessUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := visiolink.GenerateFileName(issue)

	err = visiolink.DownloadIssue(paper, client, issue.Catalog, accessKey, fileName)
	if err != nil {
		log.Fatal(err)
	}
}
