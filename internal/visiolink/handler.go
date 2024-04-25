package visiolink

import (
	"encoding/json"
	"fmt"
	"io"
	"jannotjarks/on-downloader/internal/auth"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type VisiolinkHandler struct {
	Creds  auth.Credentials
	Client *http.Client
	Paper  Paper
}

func GetOstfriesischeNachrichtenMetadata() Paper {
	return Paper{
		CatalogId:    12968,
		Customer:     "ostfriesischenachrichten",
		Domain:       "epaper.on-online.de",
		LoginDomain:  "www.on-online.de",
		ReaderDomain: "reader.on-online.de",
	}
}

func GetOstfriesenZeitungMetadata() Paper {
	return Paper{
		CatalogId:    12966,
		Customer:     "ostfriesenzeitung",
		Domain:       "epaper.oz-online.de",
		LoginDomain:  "www.oz-online.de",
		ReaderDomain: "reader.oz-online.de",
	}
}

func (h VisiolinkHandler) GetNewestIssue() Catalog {
	t := time.Now()
	year := fmt.Sprintf("%d", t.Year())
	month := fmt.Sprintf("%d", t.Month())

	issues := h.GetIssues(year, month)
	return issues[len(issues)-1]
}

func (h VisiolinkHandler) GetSpecificIssue(date string) Catalog {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatal(err)
	}

	year := fmt.Sprintf("%d", t.Year())
	month := fmt.Sprintf("%d", t.Month())

	issues := h.GetIssues(year, month)

	publicationDate := t.Format(time.DateOnly)
	fmt.Printf("Searching the issue from the following date: %s\n", publicationDate)

	var specificIssue Catalog
	for _, issue := range issues {
		if issue.PublicationDate == publicationDate {
			specificIssue = issue
			break
		}
	}

	if specificIssue.Catalog == 0 {
		log.Fatal("There was no issue published on this date!")
	}

	return specificIssue
}

func (h VisiolinkHandler) GetIssues(year string, month string) []Catalog {
	endpoint := "http://device.e-pages.dk/content/desktop/available.php"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("customer", h.Paper.Customer)
	q.Add("folder_id", fmt.Sprintf("%d", h.Paper.CatalogId))
	q.Add("year", year)
	q.Add("month", month)
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := h.Client.Do(req)
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
	var issues Content
	err = json.Unmarshal(body, &issues)
	if err != nil {
		log.Fatalln(err)
	}

	return issues.Catalogs
}

func (h VisiolinkHandler) GetLoginUrl() (string, error) {
	endpoint := fmt.Sprintf("https://%s/benutzer/loginVisiolink", h.Paper.LoginDomain)

	redirectUrl := fmt.Sprintf("https://%s/titles/%s/%d/?token=[OneTimeToken]", h.Paper.Domain, h.Paper.Customer, h.Paper.CatalogId)
	form := url.Values{}
	form.Add("_method", "POST")
	form.Add("redirect-url", redirectUrl)
	form.Add("data[Benutzer][username]", h.Creds.Username)
	form.Add("data[Benutzer][passwort]", h.Creds.Password)
	form.Add("stay", "1")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
	fmt.Printf("The loginUrl is \"%s\"\n", resp.Request.URL.String())
	return resp.Request.URL.String(), nil
}

func (h VisiolinkHandler) ExtractSecretFromLoginUrl(loginUrl string) (string, error) {
	urlPattern := fmt.Sprintf(regexp.QuoteMeta(fmt.Sprintf("https://%s/titles/%s/%d/publications/", h.Paper.Domain, h.Paper.Customer, h.Paper.CatalogId)) + `(\d*)/\?secret=(.*)`)

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

func (h VisiolinkHandler) GetIssueAccessUrl(secret string, newestIssueId int) (string, error) {
	endpoint := fmt.Sprintf("https://login-api.e-pages.dk/v1/%s/private/validate/prefix/%s/publication/%d/token", h.Paper.Domain, h.Paper.Customer, newestIssueId)

	data := url.Values{}
	data.Add("referrer_url", "POST")
	data.Add("token", secret)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.Client.Do(req)
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
	var accessUrl TokenResponse
	err = json.Unmarshal(body, &accessUrl)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("The access_url from the issue is \"%s\"\n", accessUrl.AccessURL)

	return accessUrl.AccessURL, nil
}

func (h VisiolinkHandler) GetIssueAccessKey(accessUrl string) (string, error) {
	endpoint := accessUrl

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := h.Client.Do(req)
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

func (h VisiolinkHandler) GenerateFileName(issue Catalog) string {
	return fmt.Sprintf("%s-%s.pdf", issue.Customer, issue.PublicationDate)
}

func (h VisiolinkHandler) DownloadIssue(issueId int, accessKey string, fileName string) error {
	endpoint := fmt.Sprintf("https://front.e-pages.dk/session-cc/%s/%s/%d/pdf/download_pdf.php", accessKey, h.Paper.Customer, issueId)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("domain", h.Paper.ReaderDomain)
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := h.Client.Do(req)
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
