package indexer

import (
	"fmt"

	"github.com/fedemengo/crawlit"
	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

// ResponseData represent the structure of the data that is sent back to the client
type ResponseData struct {
	URL  string
	Name string
}

// DBElement represent the structure of the data that is store in the database
type DBElement struct {
	Keyword string
	URLs    []string
}

const coll = "abcdefghijklmnopqrstuvwxyz*"
const dbURL = "mongodb://127.0.0.1"

var session *mgo.Session
var collections = make([]*mgo.Collection, 27)

func init() {
	fmt.Println("Opening database connection...")
	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connection successfull")
	}

	session.SetMode(mgo.Monotonic, true)
	database := session.DB("index")

	for i := range coll {
		collName := string(coll[i])
		collections[i] = database.C(collName)
	}
}

// Close database connection
func Close() {
	fmt.Println("Closing database connection...")
	session.Close()
}

// GetData receive a slice of keyword and return the result of the search
func GetData(keywords []string) []ResponseData {
	var result []ResponseData
	var stored DBElement

	for _, k := range keywords {
		collID := getCollectionIndex(k)
		err := collections[collID].Find(bson.M{"keyword": k}).One(&stored)
		if err != nil {
			fmt.Println("Couldn't find", keywords, "in index")
			continue
		}
		for _, url := range stored.URLs {
			result = append(result, ResponseData{URL: url, Name: url})
		}
	}

	// Here could be a good point for ranking the results

	return result
}

func getCollectionIndex(x string) int {
	collID := int(x[0]) - 'a'
	if collID < 0 || collID > 26 {
		collID = 26
	}
	return collID
}

// NewCrawlReq start a new crawling session
func NewCrawlReq(timeout int, maxurls int, maxdist int, restrict string, urls string) {
	isRestrict := (restrict == "true")
	seedURLs := ExtractURLs(urls)

	go StartCrawling(crawlit.CrawlConfig{
		SeedURLs:    seedURLs,
		MaxURLs:     maxurls,
		MaxDistance: maxdist,
		Timeout:     timeout,
		Restrict:    isRestrict,
	})
}

// StartCrawling starts a new crawl session with a specific configuration
func StartCrawling(config crawlit.CrawlConfig) {

	c := crawlit.NewCrawler()
	index := make(map[string][]string)
	c.Crawl(config, func(res crawlit.CrawlitResponse) error {
		keywords := ExtractKeywords(res.Body.Text())

		for _, keyword := range keywords {
			index[keyword] = append(index[keyword], res.URL)
		}
		return nil
	})
	c.Result()

	var stored DBElement
	for keyword, urls := range index {
		collID := getCollectionIndex(keyword)

		err := collections[collID].Find(bson.M{"keyword": keyword}).One(&stored)
		if err != nil {
			fmt.Println("Couldn't find", keyword, "in index")
		}

		if len(stored.URLs) > 0 {
			urls = append(stored.URLs, urls...)
		}

		err = collections[collID].Insert(&DBElement{Keyword: keyword, URLs: urls})
		if err != nil {
			fmt.Println("Couldn't insert", keyword, "in index")
		}
	}
}
