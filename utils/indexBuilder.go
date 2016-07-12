package main

import (
	// "fmt"
	"log"
	"os"

	"github.com/blevesearch/bleve"
	"gopkg.in/mgo.v2"
)

type Mdoc struct {
	ProfileID      string   `json:"profile_id"`
	GroupID        string   `json:"group_id"`
	LastModified   string   `json:"last_modified"`
	Tags           []string `json:"tags"`
	Read           bool     `json:"read"`
	Starred        bool     `json:"starred"`
	Authored       bool     `json:"authored"`
	Confirmed      bool     `json:"confirmed"`
	Hidden         bool     `json:"hidden"`
	CitationKey    string   `json:"citation_key"`
	SourceType     string   `json:"source_type"`
	Language       string   `json:"language"`
	ShortTitle     string   `json:"short_title"`
	ReprintEdition string   `json:"reprint_edition"`
	Genre          string   `json:"genre"`
	Country        string   `json:"country"`
	Translators    []struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"translators"`
	SeriesEditor            string `json:"series_editor"`
	Code                    string `json:"code"`
	Medium                  string `json:"medium"`
	UserContext             string `json:"user_context"`
	PatentOwner             string `json:"patent_owner"`
	PatentApplicationNumber string `json:"patent_application_number"`
	PatentLegalStatus       string `json:"patent_legal_status"`
	Notes                   string `json:"notes"`
	Accessed                string `json:"accessed"`
	FileAttached            bool   `json:"file_attached"`
	Created                 string `json:"created"`
	ID                      string `json:"id"`
	Year                    int    `json:"year"`
	Month                   int    `json:"month"`
	Day                     int    `json:"day"`
	Source                  string `json:"source"`
	Edition                 string `json:"edition"`
	Authors                 []struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"authors"`
	Keywords     []string `json:"keywords"`
	Pages        string   `json:"pages"`
	Volume       string   `json:"volume"`
	Issue        string   `json:"issue"`
	Websites     []string `json:"websites"`
	Publisher    string   `json:"publisher"`
	City         string   `json:"city"`
	Institution  string   `json:"institution"`
	Department   string   `json:"department"`
	Series       string   `json:"series"`
	SeriesNumber string   `json:"series_number"`
	Chapter      string   `json:"chapter"`
	Editors      []struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"editors"`
	Title       string `json:"title"`
	Revision    string `json:"revision"`
	Identifiers string `json:"identifiers"`
	Abstract    string `json:"abstract"`
	Type        string `json:"type"`
	OCDSOURCE   string `json:ocdsource`
}

type SchemaOrgMetadata struct {
	Context             Context      `json:"@context"`
	Type                string       `json:"@type"`
	Author              Author       `json:"author"`
	Description         string       `json:"description"`
	Distribution        Distribution `json:"distribution"`
	GlviewDataset       string       `json:"glview:dataset"`
	GlviewKeywords      string       `json:"glview:keywords"`
	GlviewMD5           string       `json:"glview:md5"`
	OpenCoreLeg         string       `json:"opencore:leg"`
	OpenCoreSite        string       `json:"opencore:site"`
	OpenCoreHole        string       `json:"opencore:hole"`
	OpenCoreProgram     string       `json:"opencore:program"`
	OpenCoreMeasurement string       `json:"opencore:measurement"`
	Keywords            string       `json:"keywords"`
	Name                string       `json:"name"`
	Spatial             Spatial      `json:"spatial"`
	URL                 string       `json:"url"`
	OCDSOURCE           string       `json:ocdsource`
}

type Context struct {
	Vocab    string `json:"@vocab"`
	GeoLink  string `json:"glview"`
	OpenCore string `json:"opencore"`
}

type Author struct {
	Type        string `json:"@type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	URL         string `json:"url"`
}

type Distribution struct {
	Type           string `json:"@type"`
	ContentURL     string `json:"contentUrl"`
	DatePublished  string `json:"datePublished"`
	EncodingFormat string `json:"encodingFormat"`
	InLanguage     string `json:"inLanguage"`
}

type Spatial struct {
	Type string `json:"@type"`
	Geo  Geo    `json:"geo"`
}

type Geo struct {
	Type      string `json:"@type"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

func main() {
	// open a new index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("compositIndex.bleve", mapping)

	// Open mongo and read out a record...   then index it.
	session, err := GetMongoCon()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//Do the abstracts
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("abstracts").C("csdco")

	var results []Mdoc
	err = c.Find(nil).All(&results)
	if err != nil {
		log.Printf("Error calling CSDCO abstract collection : %v", err)
	}

	for _, elem := range results {
		elem.OCDSOURCE = "CSDCO"
		err = index.Index(elem.ID, elem)
	}

	// Do schema.org
	d := session.DB("test").C("schemaorg")

	var results2 []SchemaOrgMetadata
	err = d.Find(nil).All(&results2)
	if err != nil {
		log.Printf("Error calling test schema.org collection : %v", err)
	}

	for _, elem2 := range results2 {
		elem2.OCDSOURCE = "JRSO"
		err = index.Index(elem2.URL, elem2)
	}

	// search for some text
	// query := bleve.NewMatchQuery("GLAD7")
	// search := bleve.NewSearchRequest(query)
	// searchResults, err := index.Search(search)

	// fmt.Println(searchResults)

}

func GetMongoCon() (*mgo.Session, error) {
	host := os.Getenv("MONGO_HOST")

	return mgo.Dial(host)
}
