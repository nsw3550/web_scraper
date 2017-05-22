package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"io"
)

var root string
var site string
var cCase []*regexp.Regexp

// getBody pulls the html body from a given URL.
func getBody(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// cornerCase generates regex objects to iterate through.
func cornerCase() {
	// Corner cases to ignore
	badCases := []string{"^#.*", "mailto:.*", ".*pdf"}
	for _, re := range badCases {
		cCase = append(cCase, regexp.MustCompile(re))
	}
}

// checkCorner checks a string against multiple regex objects.
func checkCorner(s string) bool {
	// Iterate through corner case regexes
	for _, re := range cCase {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// parseBody goes through an HTML page and parses all hyperlinks.
func parseBody(body io.Reader) ([]string, error) {
	z := html.NewTokenizer(body)
	results := []string{}

	// Starts with http
	sReg := regexp.MustCompile("^http.*")
	// Starts with a /
	rReg := regexp.MustCompile("^/.*")
	// Check for files
	fReg := regexp.MustCompile("^.*\\..*")

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
						// Check for corner cases to ignore
						if checkCorner(a.Val) {
							break
						}
						// Check for absolute links
						if sReg.MatchString(a.Val) {
							u, err := url.Parse(a.Val)
							if err != nil {
								return nil, fmt.Errorf("could not parse url %v", a.Val)
							}
							if strings.HasSuffix(u.Path, "/") {
								results = append(results, u.Host+u.Path)
							} else {
								if !fReg.MatchString(u.Path) {
									u.Path = u.Path + "/"
								}
								results = append(results, u.Host+u.Path)
							}
							break
						}
						// Check for relative links from root
						if rReg.MatchString(a.Val) {
							if strings.HasSuffix(a.Val, "/") {
								results = append(results, root+a.Val)
							} else {
								if !fReg.MatchString(a.Val) {
									a.Val = a.Val + "/"
								}
								results = append(results, root+a.Val)
							}
							break
						}
						if strings.HasSuffix(a.Val, "/") {
							results = append(results, site+a.Val)
						} else {
							if !fReg.MatchString(a.Val) {
								a.Val = a.Val + "/"
							}
							results = append(results, site+a.Val)
						}
						break
					}
				}
			}
		}
	}
}

func main() {
	cornerCase()
	s := "http://www.rit.edu/myrit/"
	tempSite, err := url.Parse(s)
	if err != nil {
		log.Fatalf("%v is not a valid url", s)
	}
	site = tempSite.Host + tempSite.Path
	data, err := getBody(s)
	if err != nil {
		log.Fatalf("Could not get HTTP data: %v", err)
	}

	// Find the root site up to the first / after http(s)
	r, err := url.Parse(s)
	if err != nil {
		log.Fatalf("Invalid site name")
	}
	root = r.Host

	urls, err := parseBody(data)
	if err != nil {
		log.Fatalf("Could not parse HTTP data: %v", err)
	}

	for i, u := range urls {
		if i < 5 {
			d, f := path.Split(u)
			if d == "" {
				fmt.Printf("No directory for %v\n", u)
				continue
			}
			os.MkdirAll(d, 0777)
			if f != "" {
				newF, err := os.Create(d + f)
				if err != nil {
					fmt.Println("Could not create file")
				}
				newF.Close()
			}
		}
	}

	fmt.Println("Adding a test line")
}
