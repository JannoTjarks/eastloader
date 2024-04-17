package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type catalog struct {
	Customer        string `json:"customer"`
	PublicationDate string `json:"publication_date"`
	Title           string `json:"title"`
	Sections        []struct {
		FrontPage int `json:"front_page"`
	} `json:"sections"`
	FolderID int `json:"folder_id"`
	Catalog  int `json:"catalog"`
	Pages    int `json:"pages"`
}

type visiolinkContent struct {
	Generated      string    `json:"generated"`
	TeaserImageURL string    `json:"teaser_image_url"`
	CatalogURL     string    `json:"catalog_url"`
	Catalogs       []catalog `json:"catalogs"`
}

type tokenResponse struct {
	AccessURL string `json:"access_url"`
	Success   bool   `json:"success"`
}

type visiolinkPaper struct {
	customer     string
	domain       string
	loginDomain  string
	loginUrl     string
	readerDomain string
	catalogId    int16
}

type credentials struct {
	username string
	password string
}

var on visiolinkPaper
var client *http.Client

func main() {
	on = visiolinkPaper{
		catalogId:    12968,
		customer:     "ostfriesischenachrichten",
		domain:       "epaper.on-online.de",
		loginDomain:  "www.on-online.de",
		readerDomain: "reader.on-online.de",
	}

	creds := credentials{
		username: os.Getenv("ON_DOWNLOADER_USERNAME"),
		password: os.Getenv("ON_DOWNLOADER_PASSWORD"),
	}

	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: jar,
	}

	issue := getNewestIssue()
	fmt.Println(issue.PublicationDate)

	loginUrl, err := getLoginUrl(creds)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := extractSecretFromLoginUrl(loginUrl)
	if err != nil {
		log.Fatal(err)
	}

	accessUrl, err := getIssueAccessUrl(secret, issue.Catalog)
	if err != nil {
		log.Fatal(err)
	}

	accessKey, err := getIssueAccessKey(accessUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := generateFileName(issue)

	err = downloadIssue(issue.Catalog, accessKey, fileName)
	if err != nil {
		log.Fatal(err)
	}
}

func getNewestIssue() catalog {
	t := time.Now()
	year := fmt.Sprintf("%d", t.Year())
	month := fmt.Sprintf("%d", t.Month())

	issues := getIssues(year, month)
	return issues[len(issues)-1]
}

func getIssues(year string, month string) []catalog {
	endpoint := "http://device.e-pages.dk/content/desktop/available.php"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("customer", on.customer)
	q.Add("folder_id", fmt.Sprintf("%d", on.catalogId))
	q.Add("year", year)
	q.Add("month", month)
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(body))
	var issues visiolinkContent
	err = json.Unmarshal(body, &issues)
	if err != nil {
		log.Fatalln(err)
	}

	return issues.Catalogs
}

func getLoginUrl(creds credentials) (string, error) {
	endpoint := fmt.Sprintf("https://%s/benutzer/loginVisiolink", on.loginDomain)

	redirectUrl := fmt.Sprintf("https://%s/titles/%s/%d/?token=[OneTimeToken]", on.domain, on.customer, on.catalogId)
	form := url.Values{}
	form.Add("_method", "POST")
	form.Add("redirect-url", redirectUrl)
	form.Add("data[Benutzer][username]", creds.username)
	form.Add("data[Benutzer][passwort]", creds.password)
	form.Add("stay", "1")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
	fmt.Printf("The loginUrl is \"%s\"\n", resp.Request.URL.String())
	return resp.Request.URL.String(), nil
}

func extractSecretFromLoginUrl(loginUrl string) (string, error) {
	urlPattern := fmt.Sprintf(regexp.QuoteMeta(fmt.Sprintf("https://%s/titles/%s/%d/publications/", on.domain, on.customer, on.catalogId)) + `(\d*)/\?secret=(.*)`)

	fmt.Println(urlPattern)
	re := regexp.MustCompile(urlPattern)

	matches := re.FindStringSubmatch(loginUrl)
	if len(matches) != 3 {
		log.Fatal("Secret extraction failed")
	}

	issue := matches[1]
	fmt.Printf("The extracted issue from the loginUrl is \"%s\"\n", issue)
	secret := matches[2]
	fmt.Printf("The extracted secret from the loginUrl is \"%s\"\n", secret)

	return secret, nil
}

func getIssueAccessUrl(secret string, newestIssueId int) (string, error) {
	endpoint := fmt.Sprintf("https://login-api.e-pages.dk/v1/%s/private/validate/prefix/%s/publication/%d/token", on.domain, on.customer, newestIssueId)

	data := url.Values{}
	data.Add("referrer_url", "POST")
	data.Add("token", secret)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(body))
	var accessUrl tokenResponse
	err = json.Unmarshal(body, &accessUrl)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("The access_url from the issue is \"%s\"\n", accessUrl.AccessURL)

	return accessUrl.AccessURL, nil
}

func getIssueAccessKey(accessUrl string) (string, error) {
	endpoint := accessUrl

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(body))
	re := regexp.MustCompile(`key4: '(.*)'`)

	matches := re.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		log.Fatal("Key extraction failed")
	}

	accessKey := matches[1]

	fmt.Printf("The access_key for the issue is \"%s\"\n", accessKey)
	return accessKey, nil
}

func generateFileName(issue catalog) string {
	return fmt.Sprintf("%s-%s.pdf", issue.Customer, issue.PublicationDate)
}

func downloadIssue(issueId int, accessKey string, fileName string) error {
	endpoint := fmt.Sprintf("https://front.e-pages.dk/session-cc/%s/%s/%d/pdf/download_pdf.php", accessKey, on.customer, issueId)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("domain", on.readerDomain)
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)

	out, fileErr := os.Create(fileName)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer out.Close()

	_, writeErr := io.Copy(out, resp.Body)
	if writeErr != nil {
		log.Fatal(writeErr)
	}

	return nil
}
