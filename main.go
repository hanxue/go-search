package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	a "github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

type GoogleResult struct {
	ResultRank  int
	ResultURL   string
	ResultTitle string
	ResultDesc  string
}

func main() {
	app := cli.NewApp()
	app.Name = "s"
	app.Usage = "Search the internet from your command line."

	app.Action = func(c *cli.Context) error {
		query := c.Args().Get(0)
		googleURL := buildGoogleURL(query, "en")
		res, err := googleRequest(googleURL)
		if err != nil {
			log.Fatal(err)
		}
		results, err := googleResultParser(res)
		if err != nil {
			log.Fatal(err)
		} else {
			for i, v := range results {
				r, err := regexp.Compile(`.*\://?([^\/]+)`)
				if err != nil {
					fmt.Printf("There is a problem with your regexp.\n")
					os.Exit(1)
				}
				domain := r.FindAllStringSubmatch(v.ResultURL, -1)[0][1]
				fmt.Println(a.Cyan(" ("+strconv.Itoa(i)+")"),
					a.Brown("["+domain+"]"), a.Green(v.ResultTitle))
				fmt.Println(v.ResultDesc)
			}

		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Taken from https://gist.github.com/EdmundMartin/eaea4aaa5d231078cb433b89878dbecf
func buildGoogleURL(searchTerm string, languageCode string) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	return fmt.Sprintf("https://www.google.com/search?q=%s&num=10&hl=%s", searchTerm, languageCode)
}

func googleRequest(searchURL string) (*http.Response, error) {

	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("Referer", "https://www.google.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)")

	res, err := baseClient.Do(req)

	if err != nil {
		return nil, err
	}
	return res, nil
}

func googleResultParser(response *http.Response) ([]GoogleResult, error) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}
	results := []GoogleResult{}
	sel := doc.Find("div.g")
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h3.r")
		descTag := item.Find("span.st")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			result := GoogleResult{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
			rank++
		}
	}
	return results, err
}
