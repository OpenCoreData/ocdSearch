package main

// recreate the sample index
//go:generate rm -rf indexes/test.bleve
//go:generate bleve_create -index indexes/test.bleve -store goleveldb
//go:generate bleve_index -index indexes/test.bleve a.json

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/config"
	bleveHttp "github.com/blevesearch/bleve/http"
	"github.com/gorilla/mux"
)

const indexDir = "indexes"

func init() {

	router := mux.NewRouter()
	router.StrictSlash(true)

	listIndexesHandler := bleveHttp.NewListIndexesHandler()
	router.Handle("/ocdsearchapi", listIndexesHandler).Methods("GET")

	docCountHandler := bleveHttp.NewDocCountHandler("")
	docCountHandler.IndexNameLookup = indexNameLookup
	router.Handle("/ocdsearchapi/{indexName}/_count", docCountHandler).Methods("GET")

	searchHandler := bleveHttp.NewSearchHandler("")
	searchHandler.IndexNameLookup = indexNameLookup
	router.Handle("/ocdsearchapi/{indexName}/_search", searchHandler).Methods("POST")

	http.Handle("/", &CORSWrapper{router})

	log.Printf("opening indexes")
	// walk the data dir and register index names
	dirEntries, err := ioutil.ReadDir(indexDir)
	if err != nil {
		log.Printf("error reading data dir: %v", err)
		return
	}

	for _, dirInfo := range dirEntries {
		indexPath := indexDir + string(os.PathSeparator) + dirInfo.Name()

		// skip single files in data dir since a valid index is a directory that
		// contains multiple files
		if !dirInfo.IsDir() {
			log.Printf("not registering %s, skipping", indexPath)
			continue
		}

		i, err := bleve.OpenUsing(indexPath, map[string]interface{}{
			"read_only": true,
		})

		if err != nil {
			log.Printf("error opening index %s: %v", indexPath, err)
		} else {
			log.Printf("registered index: %s at %s", dirInfo.Name(), indexPath)
			bleveHttp.RegisterIndexName(dirInfo.Name(), i)
		}
	}

	// Playing with index aliases
	// Open all indexes in an alias and use this in a named call
	log.Printf("Start building Codex index \n")

	index1, err := bleve.OpenUsing("indexes/abstracts", map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		log.Printf("Error with index alias: %v", err)
		return
	}
	index2, err := bleve.OpenUsing("indexes/compositIndex", map[string]interface{}{
		"read_only": true,
	})
	if err != nil {
		log.Printf("Error with index alias: %v", err)
		return
	}
	everything := bleve.NewIndexAlias(index1, index2)
	log.Printf("Codex index built\n")

	// add := []string{"indexes/abstracts", "indexes/compositIndex"}
	// remove := []string{}
	// err = bleveHttp.UpdateAlias("codex", add, remove)
	// if err != nil {
	// 	log.Printf("Error with index alias: %v ", err)
	// 	return
	// }

	bleveHttp.RegisterIndexName("codex", everything) // search codex for all
	log.Printf("registered index:  codex\n")

}

func muxVariableLookup(req *http.Request, name string) string {
	return mux.Vars(req)[name]
}

func indexNameLookup(req *http.Request) string {
	return muxVariableLookup(req, "indexName")
}
