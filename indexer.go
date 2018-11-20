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

type multipleKeywordHandler func(curr, next []ResponseData) []ResponseData

const coll = "abcdefghijklmnopqrstuvwxyz*"

//const dbURL = "mongodb://gosearch:recv12@ds211504.mlab.com:11504/gosearch-db"
const dbURL = "mongodb://127.0.0.1"

var session *mgo.Session
var collections = make([]*mgo.Collection, 27)

func init() {
	fmt.Println("Opening database connection...")
	fmt.Println("Connecting to", dbURL)
	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connection successfull")
	}

	session.SetMode(mgo.Monotonic, true)
	database := session.DB("gosearch-db")

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

	handle := andHandler

	for _, k := range keywords {
		if k == "or" {
			handle = orHandler
			continue
		} else if k == "and" {
			handle = andHandler
			continue
		}

		collID := getCollectionIndex(k)
		fmt.Println("search", k)
		err := collections[collID].Find(bson.M{"keyword": k}).One(&stored)
		if err != nil {
			fmt.Println("Couldn't find", keywords, "in index")
			continue
		}

		var data []ResponseData
		for _, url := range stored.URLs {
			data = append(data, ResponseData{URL: url, Name: url})
		}

		fmt.Println(data)
		result = handle(result, data)
	}

	// Here could be a good point for ranking the results

	return result
}

func orHandler(curr, next []ResponseData) []ResponseData {
	return append(curr, next...)
}

func andHandler(curr, next []ResponseData) []ResponseData {
	if len(next) == 0 {
		return curr
	}

	if len(curr) == 0 {
		return next
	}

	var res []ResponseData
	data1 := make(map[string]bool)
	data2 := make(map[string]bool)
	union := make(map[ResponseData]bool)

	for _, w := range curr {
		data1[w.URL] = true
		union[w] = true
	}

	for _, w := range next {
		data2[w.URL] = true
		union[w] = true
	}

	for w := range union {
		if data1[w.URL] && data2[w.URL] {
			res = append(res, w)
		}
	}

	return res
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
	isRestrict := (restrict == "on")
	seedURLs := ExtractURLs(urls)

	fmt.Println("[API]", timeout, maxurls, maxdist, isRestrict, seedURLs)

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

	fmt.Println("[INDEXER] Start crawling", config.SeedURLs)
	c := crawlit.NewCrawler()
	index := make(map[string][]string)
	c.Crawl(config, func(res crawlit.CrawlitResponse) error {
		keywords := ExtractKeywords(res.Body.Text(), 4)

		for _, keyword := range keywords {
			index[keyword] = append(index[keyword], res.URL)
		}
		return nil
	})
	c.Result()
	fmt.Println("[INDEXER] Done crawling", config.SeedURLs)

	for keyword, urls := range index {
		collID := getCollectionIndex(keyword)

		var stored DBElement
		err := collections[collID].Find(bson.M{"keyword": keyword}).One(&stored)
		if err == mgo.ErrNotFound {
			//fmt.Println("Couldn't find", keyword, "in index")
			err = collections[collID].Remove(bson.M{"keyword": keyword})
			if err != nil {
				//fmt.Println("Couldn't remove", keyword, "from index")
			}
		}

		if len(stored.URLs) > 0 {
			urls = merge(stored.URLs, urls)
		}

		err = collections[collID].Insert(&DBElement{Keyword: keyword, URLs: urls})
		if err != nil {
			fmt.Println("Couldn't insert", keyword, "in index")
		} else {
			fmt.Println("Inserting", keyword, "in index")
		}
	}
	fmt.Println("[INDEXER] Done inserting", config.SeedURLs)
}
