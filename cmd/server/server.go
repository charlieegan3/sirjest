package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

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

var searchEngines = map[string]searchEngine{
	"google": searchEngine{
		name:       "Google",
		docQuery:   css.Convert("#center_col > div > div > a > b > i", 0) + "/text()[1]",
		baseURL:    "https://www.google.co.uk/search",
		queryParam: "q",
	},
}

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, getCorrection("google", r.FormValue("q")))
}

func buildUrl(engine string, queryString string) string {
	var Url *url.URL
	Url, _ = url.Parse(searchEngines[engine].baseURL)

	parameters := url.Values{}
	parameters.Add(searchEngines[engine].queryParam, queryString)
	Url.RawQuery = parameters.Encode()

	return Url.String()
}

func getCorrection(engine string, queryString string) string {
	resp, _ := http.Get(buildUrl(engine, queryString))
	page, _ := ioutil.ReadAll(resp.Body)
	doc, _ := gokogiri.ParseHtml(page)

	result, err := doc.Root().Search(searchEngines[engine].docQuery)

	correction := "NULL"
	if err == nil && len(result) == 1 {
		correction = fmt.Sprintf("%v", result[0])
	}
	doc.Free()
	return correction
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/", index)
	http.ListenAndServe(":"+port, nil)
}
