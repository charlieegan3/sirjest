package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/kennygrant/sanitize"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/css"
)

var googleDocQuery string = css.Convert("#center_col > div > div > a > b > i", 0) + "/text()[1]"

type searchEngine struct {
	name       string
	docQuery   string
	baseURL    string
	queryParam string
}

type correctionResult struct {
	engine     string
	correction string
}

var searchEngines = map[string]searchEngine{
	"google": searchEngine{
		name:       "Google",
		docQuery:   css.Convert("#center_col > div > div > a", 0),
		baseURL:    "https://www.google.co.uk/search",
		queryParam: "q",
	},
	"yahoo": searchEngine{
		name:       "Yahoo",
		docQuery:   css.Convert(".compTitle > span > a", 0),
		baseURL:    "https://uk.search.yahoo.com/search",
		queryParam: "p",
	},
	"bing": searchEngine{
		name:       "Bing",
		docQuery:   css.Convert("#sp_requery > h2 > a", 0),
		baseURL:    "https://www.bing.com/search",
		queryParam: "q",
	},
	"duckduckgo": searchEngine{
		name:       "DuckDuckGo",
		docQuery:   css.Convert("#did_you_mean > a", 0),
		baseURL:    "https://duckduckgo.com/html",
		queryParam: "q",
	},
}

func index(w http.ResponseWriter, r *http.Request) {
	c := make(chan correctionResult)
	for k := range searchEngines {
		go getCorrection(k, r.FormValue("q"), c)
	}

	var corrections = map[string]string{}
	for i := 0; i < len(searchEngines); i++ {
		r := <-c
		corrections[r.engine] = r.correction
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

	result, err := doc.Root().Search(searchEngines[engine].docQuery)

	correction := "NULL"
	if err == nil && len(result) > 0 {
		correction = fmt.Sprintf("%v", result[0])
	} else if err != nil {
		fmt.Println(err)
	}
	doc.Free()
	c <- correctionResult{engine: engine, correction: sanitize.HTML(correction)}
}

func main() {
	port := os.Getenv("PORT")
	fmt.Printf("Listening on port %v\n", port)

	http.HandleFunc("/", index)
	http.ListenAndServe(":"+port, nil)
}
