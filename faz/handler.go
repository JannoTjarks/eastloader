package faz

// A german documentation for the FAZ api can be found here: https://www.peterhofmann.me/2013/11/faz-to-kindle-in-python/

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
    "regexp"

	"golang.org/x/net/html"
)

type FazHandler struct {
	Client *http.Client
	Creds  Credentials
}

type FazPaper struct {
	Name string
	URL  string
}

func Login(handler FazHandler) {
	endpoint := "https://www.faz.net/membership/loginNoScript?nomobile=1&redirectUrl=`/"

	dataLogin := fmt.Sprintf("loginName=%s&password=%s", handler.Creds.Email, handler.Creds.Password)
	// "loginName=" + $config.username + "&password=" + $config.password

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(dataLogin))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://zeitung.faz.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")

	resp, err := handler.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("The http status code is \"%s\"\n", resp.Status)
}

func GetKioskHtml(handler FazHandler) (string, error) {
	endpoint := "https://zeitung.faz.net/meine-ausgaben/faz"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(req.URL.String())

	resp, err := handler.Client.Do(req)
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

	return string(body), nil
}

func GetIssues(handler FazHandler) {
	endpoint := "https://zeitung.faz.net/meine-ausgaben/faz"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(req.URL.String())

	resp, err := handler.Client.Do(req)
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
}

func GetFazPaper(body string) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}

    m := make(map[string]FazPaper)

    var f func(*html.Node)
    f = func(n *html.Node) {
        paper, err := parseEpaperHref(n)
        // Add error handling
        if err == nil {
            re := regexp.MustCompile(`(?m)\d{2}.\d{2}.\d{4}`)
            date := re.FindString(paper.Name)
            m[date] = FazPaper {
                Name: paper.Name,
                URL: paper.URL,
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            f(c)
        }
    }
    f(doc)

    fmt.Println(m)
}

func parseEpaperHref(n *html.Node) (FazPaper, error){
    m := make(map[string]string)
    fmt.Println("1")

    if n.Type == html.ElementNode && n.Data == "a" {
    fmt.Println("2")
        elementPaper := false
        for _, a := range n.Attr {
            if a.Key == "track-element" && strings.Contains(a.Val, "E-Paper+FAZ+n/a") {
    fmt.Println("3")
                elementPaper = true
            }
        }
        if elementPaper {
            for _, a := range n.Attr {
                if a.Key == "track-element" {
    fmt.Println("4")
                    fmt.Printf("Key: %s, Value: %s\n", a.Key, a.Val)
                    m[a.Key] = a.Val
                }
                if a.Key == "href" {
                    fmt.Printf("Key: %s, Value: %s\n", a.Key, a.Val)
                    m[a.Key] = a.Val
                }
            }
        } else {
            paper := FazPaper{}
            return paper, fmt.Errorf("the node does not contain any href to an epaper")
        }
    }

    paper := FazPaper {
        Name: m["track-element"],
        URL: m["href"],
    }

    return paper, nil
}
