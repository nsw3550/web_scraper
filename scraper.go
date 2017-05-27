package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	"path"
	"os"
)

// getBody pulls the html body from a given URL.
func getBody(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func parseBody(resp *http.Response) ([]*url.URL, error) {
	z := html.NewTokenizer(resp.Body)
	results := []*url.URL{}

	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return results, nil
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						u, err := resp.Request.URL.Parse(a.Val)
						if err != nil {
							return nil, err
						}
						results = append(results, u)
					}
				}
			}
		}
	}
}

func buildDirs(links []*url.URL) []string {
	result := []string{}
	//fReg := regexp.MustCompile(".*\\..*")
	for _, link := range links {
		if link.Host == "" {
			continue
		}
		if link.Path == "" {
			result = append(result, link.Host)
			continue
		}
		result = append(result, link.Host+link.Path)
	}

	return result
}

func main() {
	s := "http://www.rit.edu/myrit/"
	resp, err := getBody(s)
	if err != nil {
		log.Fatalf("Could not get HTTP data: %v", err)
	}

	links, err := parseBody(resp)
	if err != nil {
		log.Fatalf("Could not parse body: %v", err)
	}

	results := buildDirs(links)
	for i, r := range results {
		if i < 5 {
			d, f := path.Split(r)
			if d == "" {
				fmt.Printf("No directory for %v\n", r)
				continue
			}
			os.MkdirAll(d, 0777)
			if f != "" {
				newF, err := os.Create(d+f)
				if err != nil {
					fmt.Println("Could not create file")
				}
				newF.Close()
			}
		}
	}
}