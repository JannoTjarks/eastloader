package visiolink

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type VisiolinkHandler struct {
	Creds  Credentials
	Client *http.Client
	Paper  Paper
}

func GetOstfriesischeNachrichtenMetadata() Paper {
	return Paper{
		catalogId:    12968,
		customer:     "ostfriesischenachrichten",
		domain:       "epaper.on-online.de",
		loginDomain:  "www.on-online.de",
		readerDomain: "reader.on-online.de",
	}
}

func GetOstfriesenZeitungMetadata() Paper {
	return Paper{
		catalogId:    12966,
		customer:     "ostfriesenzeitung",
		domain:       "epaper.oz-online.de",
		loginDomain:  "www.oz-online.de",
		readerDomain: "reader.oz-online.de",
	}
}

func (h VisiolinkHandler) RunDownloadRoutine(date string) {
	var issue Catalog
	if date == "" {
		issue = h.getNewestIssue()
	} else {
		issue = h.GetSpecificIssue(date)
	}

	fmt.Println(issue.PublicationDate)

	loginUrl, err := h.getLoginUrl()
	if err != nil {
		log.Fatal(err)
	}

	secret, err := h.extractSecretFromLoginUrl(loginUrl)
	if err != nil {
		log.Fatal(err)
	}

	accessUrl, err := h.getIssueAccessUrl(secret, issue.Catalog)
	if err != nil {
		log.Fatal(err)
	}

	accessKey, err := h.getIssueAccessKey(accessUrl)
	if err != nil {
		log.Fatal(err)
	}

	fileName := h.generateFileName(issue)

	done := make(chan bool, 1)
	go h.downloadIssue(done, issue.Catalog, accessKey, fileName)
	WaitForHttpResponse(done)
}

func (h VisiolinkHandler) getNewestIssue() Catalog {
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
	q.Add("customer", h.Paper.customer)
	q.Add("folder_id", fmt.Sprintf("%d", h.Paper.catalogId))
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

func (h VisiolinkHandler) getLoginUrl() (string, error) {
	endpoint := fmt.Sprintf("https://%s/benutzer/loginVisiolink", h.Paper.loginDomain)

	redirectUrl := fmt.Sprintf("https://%s/titles/%s/%d/?token=[OneTimeToken]", h.Paper.domain, h.Paper.customer, h.Paper.catalogId)
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

func (h VisiolinkHandler) extractSecretFromLoginUrl(loginUrl string) (string, error) {
	urlPattern := fmt.Sprintf(regexp.QuoteMeta(fmt.Sprintf("https://%s/titles/%s/%d/publications/", h.Paper.domain, h.Paper.customer, h.Paper.catalogId)) + `(\d*)/\?secret=(.*)`)

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

func (h VisiolinkHandler) getIssueAccessUrl(secret string, newestIssueId int) (string, error) {
	endpoint := fmt.Sprintf("https://login-api.e-pages.dk/v1/%s/private/validate/prefix/%s/publication/%d/token", h.Paper.domain, h.Paper.customer, newestIssueId)

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

func (h VisiolinkHandler) getIssueAccessKey(accessUrl string) (string, error) {
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

func (h VisiolinkHandler) generateFileName(issue Catalog) string {
	return fmt.Sprintf("%s-%s.pdf", issue.Customer, issue.PublicationDate)
}

func (h VisiolinkHandler) downloadIssue(done chan bool, issueId int, accessKey string, fileName string) error {
	endpoint := fmt.Sprintf("https://front.e-pages.dk/session-cc/%s/%s/%d/pdf/download_pdf.php", accessKey, h.Paper.customer, issueId)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("domain", h.Paper.readerDomain)
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

	done <- true

	return nil
}
