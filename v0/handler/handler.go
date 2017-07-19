package handler

import (
	"log"
	"net/http"
	// "strings"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/blevesearch/bleve"
	"github.com/parnurzeal/gorequest"
)

// DoSearch handler
func DoSearch(w http.ResponseWriter, r *http.Request) {
	log.Printf("r path: %s\n", r.URL.Query())
	queryterm := r.URL.Query().Get("q")

	index := r.URL.Query().Get("i")
	defaultindex := "compositIndex"
	if index == "" {
		index = defaultindex
	}

	// make a string for the various templates I would use....
	// Perhaps not the way....
	// rather have the template for each.. have one that responds to
	// the data in it...
	templateFile := ""
	if index == "compositIndex" {
		templateFile = "./static/index.html"
	}
	if index == "csdcoFX.bleve" {
		templateFile = "./static/indexFX.html"
	}
	if index == "abstracts" {
		templateFile = "./static/indexAbs.html"
	}

	ht, err := template.New("Template").ParseFiles(templateFile) //open and parse a template text file
	if err != nil {
		log.Printf("template parse failed: %s", err)
	}

	content := fmt.Sprintf(`{"size":150,"from":0,"query":{"conjuncts":[{"boost":1,"query":"%s"}]},"fields":["*"],"highlight":{"fields":["content"]},"facets":{"Types":{"field":"type","size":5}}}`, queryterm)
	// content := `{"size":20,"from":0,"query":{"conjuncts":[{"boost":1,"query":"JanusCoreSummary"}]},"fields":["*"],"highlight":{"fields":["content"]},"facets":{"Types":{"field":"type","size":5}}}`

	url := fmt.Sprintf("http://localhost:9800/ocdsearchapi/%s/_search", index)

	// REST call to Bleve (POINTLESS...    just open and work with the local index?)
	// it is more usefull for putting the UI in other places though....
	// url := "http://localhost:9800/ocdsearchapi/jrso/_search"
	// url := "http://localhost:9800/ocdsearchapi/abstracts/_search"
	// url := "http://localhost:9800/ocdsearchapi/compositIndex/_search"
	// url := "http://localhost:9800/ocdsearchapi/codex/_search"
	// url := "/ocdsearchapi/jrso/_search"

	log.Println(url)
	log.Println(content)

	request := gorequest.New()
	resp, body, errs := request.Post(url).Set("Accept", "text/plain").Send(content).End()
	if errs != nil {
		log.Printf("Response is an error: %s", errs)
	}

	results := bleve.SearchResult{}
	json.Unmarshal([]byte(body), &results)

	// Some print outs to see results in console
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	// fmt.Println("response Body:", body)
	// fmt.Print(hits)
	fmt.Printf("Total is %d \n", results.Total)
	// for _, v := range results.Hits {
	// 	fmt.Printf("%v \n\n", v)
	// 	fmt.Printf("%v \n\n", v.Fields)
	// 	fmt.Printf("%v \n\n", v.Fields["OCDSOURCE"])
	// }

	// FUNCTION call here to replace the REST call above?

	err = ht.ExecuteTemplate(w, "T", r.URL.Query().Get("q")) //substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		log.Printf("htemplate execution failed: %s", err)
	}

	err = ht.ExecuteTemplate(w, "R", results.Hits) //substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		log.Printf("htemplate execution failed: %s", err)
	}

}

// QueryStringSearch put this function in a search package
// ref http://studygolang.com/articles/2537
func QueryStringSearch(index bleve.Index) {
	qString := `+description:text summary:"text indexing" summary:believe~2 -description:lucene duration:<30`
	q := bleve.NewQueryStringQuery(qString)
	req := bleve.NewSearchRequest(q)
	req.Highlight = bleve.NewHighlightWithStyle("ansi")
	req.Fields = []string{"summary", "speaker", "description", "duration"}
	res, err := index.Search(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
