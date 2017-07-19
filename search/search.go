package search

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/blevesearch/bleve"
	"opencoredata.org/ocdSearch/sparql"
)

type FreeTextResults struct {
	Place           int
	Index           string
	Score           float64
	ID              string
	Fragments       []Fragment
	IconName        string
	IconDescription string
}

type Fragment struct {
	Key   string
	Value []string
}

type SearchMetaData struct {
	Term      string
	Count     uint64
	StartAt   uint64
	EndAt     uint64
	NextStart uint64
	PrevStart uint64
	Message   string
}

type Qstring struct {
	Query      string
	Qualifiers map[string]string
}

// DoSearch is there to do searching..  (famous documentation style intact!)
func DoSearch(w http.ResponseWriter, r *http.Request) {
	log.Printf("r path: %s\n", r.URL.Query()) // need to log this better so I can filter out search terms later
	queryterm := r.URL.Query().Get("q")
	queryterm = strings.TrimSpace(queryterm) // remove leading and trailing white spaces a user might put in (not internal spaces though)

	// get the start at value or set to 0
	var startAt uint64
	startAt = 0
	if s, err := strconv.Atoi(r.URL.Query().Get("start")); err == nil {
		startAt = uint64(s)
	}

	// Make a var in case I want other templates I switch to later...
	templateFile := "./templates/ocdsearch.html"

	// parse the queryterm to get the colon based qualifiers
	qstring := parse(queryterm)

	// var queryResults DocumentMatchCollection{}
	distance := ""
	queryResults, sr := indexCall(qstring, startAt, distance)
	ql := len(queryResults)

	// TODO..  Yet Another Ugly Section (YAUS)  (I've named the pattern..  that is just sad)
	// check here..  if results are 0 then recursive call with ~1
	// check here and if 0 then try again with ~2
	fmt.Printf("The length is %d \n", ql)
	if ql == 0 {
		if strings.EqualFold(distance, "") {
			queryResults, _ = indexCall(qstring, startAt, "~1")
			ql = len(queryResults)
			fmt.Printf("The length in loop 1 %d \n", ql)

		}
	}
	if ql == 0 {
		if strings.Contains(distance, "~1") {
			queryResults, _ = indexCall(qstring, startAt, "~2")
			ql = len(queryResults)
			fmt.Printf("The length in loop 2 %d \n", ql)
		}
	}

	fmt.Printf("The length final is %d \n", ql)

	// Set up some metadata on the search results to return
	var searchmeta SearchMetaData
	searchmeta.Term = queryterm // We don't use qstring.Query here since we want the full string including qualifiers, returned to the page for rendering with results
	searchmeta.Count = sr.Total
	searchmeta.StartAt = startAt
	searchmeta.EndAt = startAt + 20 // TODO make this a var..   do not set statis!!!!!!
	searchmeta.NextStart = searchmeta.EndAt + 1
	searchmeta.PrevStart = searchmeta.StartAt - 20
	if ql == 0 {
		if qstring.Query == "" {
			searchmeta.Message = "Search EarthCube CDF RWG demo index"

		} else {
			searchmeta.Message = "No results found for this search"
		}
	}

	// If we have a term.. search the triplestore
	var spres sparql.SPres
	if ql > 0 {
		topResult := queryResults[0] // pass this as a new template section TR!
		fmt.Println(topResult.ID)
		var err error
		spres, err = sparql.DoCall(topResult.ID) // turn sparql call on / off
		if err != nil {
			log.Printf("SPARQL call failed: %s", err)
		}
		// fmt.Print(spres.Description)
	}

	ht, err := template.New("Template").ParseFiles(templateFile) //open and parse a template text file
	if err != nil {
		log.Printf("template parse failed: %s", err)
	}

	err = ht.ExecuteTemplate(w, "Q", searchmeta) //substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		log.Printf("Template execution failed: %s", err)
	}

	err = ht.ExecuteTemplate(w, "T", queryResults) //substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		log.Printf("Template execution failed: %s", err)
	}

	err = ht.ExecuteTemplate(w, "S", spres) //substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		log.Printf("Template execution failed: %s", err)
	}
}

// TODO  return to html/template and use the FUNCS register to put ina  "toHTMl"
// then use this in the "mark" section of the code.
// needed function to escape the HTML in the results
// func toHTML(s string) template.URL {
// 	return template.URL(s)
// }

func parse(qstring string) Qstring {
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`) // get rid of multiple spaces
	qstring = re_inside_whtsp.ReplaceAllString(qstring, " ")
	sa := strings.Split(qstring, " ")

	var buffer bytes.Buffer
	qpairs := make(map[string]string)
	for _, item := range sa {
		if strings.ContainsAny(item, ":") {
			qualpair := strings.Split(item, ":")
			qpairs[qualpair[0]] = qualpair[1]
		} else {
			buffer.WriteString(item)
			buffer.WriteString(" ")
		}
	}

	qs := Qstring{Query: buffer.String(), Qualifiers: qpairs}
	return qs
}

// termReWrite puts the bleve ~1 or ~2 term options on for fuzzy matching
func termReWrite(phrase string, distanceAppend string) string {
	terms := strings.Split(phrase, " ")

	for k, _ := range terms {
		var str bytes.Buffer
		str.WriteString(strings.TrimSpace(terms[k]))
		str.WriteString(distanceAppend)
		terms[k] = str.String()
	}

	fmt.Println(strings.Join(terms, " "))
	return strings.Join(terms, " ")
}

// return JSON string..  enables use of func for REST call too
func indexCall(qstruct Qstring, startAt uint64, distance string) ([]FreeTextResults, *bleve.SearchResult) {
	if qstruct.Query == "" {
		return nil, nil
	}

	// TODO ..  improve this..
	// Really need to check if it is ~1 or ~2.  If not, set to empty
	// if distance == "" {
	// 	distance = ""
	// }

	// Playing with index aliases
	// Open all indexes in an alias and use this in a named call
	log.Printf("Start building Codex index \n")

	index1, err := bleve.OpenUsing("/Users/dfils/Data/OCDDataVolumes/indexes/abstracts.bleve", map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		log.Printf("Error with index1 alias: %v", err)
	}
	index2, err := bleve.OpenUsing("/Users/dfils/Data/OCDDataVolumes/indexes/csdco.bleve", map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		log.Printf("Error with index2 alias: %v", err)
	}
	index3, err := bleve.OpenUsing("/Users/dfils/Data/OCDDataVolumes/indexes/janus.bleve", map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		log.Printf("Error with index3 alias: %v", err)
	}

	var index bleve.IndexAlias

	if _, ok := qstruct.Qualifiers["type"]; ok {
		//  TODO..  system needs to handle accepting N number type: qualifiers like type:jrso,csdco
		if strings.Contains(qstruct.Qualifiers["type"], "abstracts") {
			index = bleve.NewIndexAlias(index1)
			log.Println("Active index: 1")
		}
		if strings.Contains(qstruct.Qualifiers["type"], "csdco") {
			index = bleve.NewIndexAlias(index2)
			log.Println("Active index: 2")
		}
		if strings.Contains(qstruct.Qualifiers["type"], "jrso") {
			index = bleve.NewIndexAlias(index3)
			log.Println("Active index: 3")
		}
	} else {
		index = bleve.NewIndexAlias(index1, index2, index3)
		log.Println("Active index: 1,2,3")
	}

	log.Printf("Codex index built\n")

	// parse string and add ~2 to each term/word, then rebuild as a string.
	fmt.Printf("Ready to search with %s and distance: %s \n", qstruct.Query, distance)
	query := bleve.NewQueryStringQuery(termReWrite(qstruct.Query, distance))
	search := bleve.NewSearchRequestOptions(query, 20, int(startAt), false) // no explanation
	search.Highlight = bleve.NewHighlightWithStyle("html")                  // need Stored and IncludeTermVectors in index
	searchResults, err := index.Search(search)
	if err != nil {
		log.Printf("Error search results: %v", err)
	}

	hits := searchResults.Hits // array of struct DocumentMatch

	var results []FreeTextResults

	for k, item := range hits {
		// fmt.Printf("\n%d: %s, %f, %s, %v\n", k, item.Index, item.Score, item.ID, item.Fragments)
		// fmt.Printf("%v\n", item.Fields["potentialAction.target.description"])
		var frags []Fragment
		for key, frag := range item.Fragments {
			// fmt.Printf("%s   %s\n", key, frag)
			frags = append(frags, Fragment{key, frag})
		}

		// set up a material icon   ref:  https://material.io/icons/
		var iconName string
		var iconDescription string
		if strings.Contains(item.Index, "janus") {
			iconName = "file_download"                 // material design icon name used in template
			iconDescription = "JRSO Data landing page" // material design icon name used in template
		}
		if strings.Contains(item.Index, "csdco") {
			iconName = "file_download"                  // material design icon name used in template
			iconDescription = "CSDCO Data landing page" // material design icon name used in template
		}
		if strings.Contains(item.Index, "abstracts") {
			iconName = "http"                  // material design icon name used in template  alts:  web_asset or web
			iconDescription = "CSDCO Abstract" // material design icon name used in template  alts:  web_asset or web
		}

		results = append(results, FreeTextResults{k, item.Index, item.Score, item.ID, frags, iconName, iconDescription})
	}

	fmt.Printf("Looping status count:%d, distance:%s\n", len(results), distance)

	index.Close()
	return results, searchResults
}
