package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/css"
)

var googleDocQuery string = css.Convert("#center_col > div > div > a > b > i", 0) + "/text()[1]"

type searchEngine struct {
	name               string
	baseURL            string
	queryParam         string
	suggestedTermQuery string
	topLinkQuery       string
	topTitleQuery      string
}

type correctionResult struct {
	engine        string
	SuggestedTerm string
	TopLink       string
	TopTitle      string
}

var searchEngines = map[string]searchEngine{
	"google": searchEngine{
		name:               "Google",
		baseURL:            "https://www.google.co.uk/search",
		queryParam:         "q",
		suggestedTermQuery: css.Convert("#center_col > div > div > a", 0),
		topLinkQuery:       css.Convert("cite", 0),
		topTitleQuery:      css.Convert("h3.r", 0),
	},
	"yahoo": searchEngine{
		name:               "Yahoo",
		baseURL:            "https://uk.search.yahoo.com/search",
		queryParam:         "p",
		suggestedTermQuery: css.Convert(".compTitle > span > a", 0),
		topLinkQuery:       css.Convert("ol > li > div > div > div > span", 0),
		topTitleQuery:      css.Convert("li div.algo h3.title", 0),
	},
	"bing": searchEngine{
		name:               "Bing",
		baseURL:            "https://www.bing.com/search",
		queryParam:         "q",
		suggestedTermQuery: css.Convert("#sp_requery > h2 > a", 0),
		topLinkQuery:       css.Convert("cite", 0),
		topTitleQuery:      css.Convert("li.b_algo h2", 0),
	},
	"duckduckgo": searchEngine{
		name:               "DuckDuckGo",
		baseURL:            "https://duckduckgo.com/html",
		queryParam:         "q",
		suggestedTermQuery: css.Convert("#did_you_mean > a", 0),
		topLinkQuery:       css.Convert(".url", 0),
		topTitleQuery:      css.Convert("div.links_main a.large", 0),
	},
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := make(chan correctionResult)
	for k := range searchEngines {
		go getCorrection(k, r.FormValue("q"), c)
	}

	var corrections = map[string]correctionResult{}
	for i := 0; i < len(searchEngines); i++ {
		r := <-c
		corrections[r.engine] = r
	}
	jsonString, _ := json.Marshal(corrections)
	io.WriteString(w, string(jsonString))
}

func buildUrl(engine string, queryString string) string {
	var endpoint *url.URL
	endpoint, _ = url.Parse(searchEngines[engine].baseURL)

	parameters := url.Values{}
	parameters.Add(searchEngines[engine].queryParam, queryString)
	endpoint.RawQuery = parameters.Encode()

	return endpoint.String()
}

func getCorrection(engine string, queryString string, c chan correctionResult) {
	resp, _ := http.Get(buildUrl(engine, queryString))
	page, _ := ioutil.ReadAll(resp.Body)
	doc, _ := gokogiri.ParseHtml(page)

	suggestedTermResult, err := doc.Root().Search(searchEngines[engine].suggestedTermQuery)
	suggestedTerm := "NULL"
	if err == nil && len(suggestedTermResult) > 0 {
		suggestedTerm = fmt.Sprintf("%v", suggestedTermResult[0])
	} else if err != nil {
		fmt.Println(err)
	}

	topLinkResult, err := doc.Root().Search(searchEngines[engine].topLinkQuery)
	topLink := "NULL"
	if err == nil && len(topLinkResult) > 0 {
		topLink = fmt.Sprintf("%v", topLinkResult[0])
	} else if err != nil {
		fmt.Println(err)
	}

	topTitleResult, err := doc.Root().Search(searchEngines[engine].topTitleQuery)
	topTitle := "NULL"
	if err == nil && len(topTitleResult) > 0 {
		topTitle = fmt.Sprintf("%v", topTitleResult[0])
	} else if err != nil {
		fmt.Println(err)
	}

	doc.Free()
	c <- correctionResult{
		engine:        engine,
		SuggestedTerm: sanitize.HTML(suggestedTerm),
		TopLink:       strings.TrimSpace(sanitize.HTML(topLink)),
		TopTitle:      strings.TrimSpace(sanitize.HTML(topTitle)),
	}
}

func main() {
	port := os.Getenv("PORT")
	fmt.Printf("Listening on port %v\n", port)

	http.HandleFunc("/", index)
	http.ListenAndServe(":"+port, nil)
}
