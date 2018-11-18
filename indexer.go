package indexer

import (
	"fmt"

	"github.com/fedemengo/crawlit"
	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type ResponseData struct {
	URL  string
	Name string
}

var channel = make(chan crawlit.CrawlConfig)

const COLL = "abcdefghijklmnopqrstuvwxyz*"
const DB_URL = "mongodb://127.0.0.1"

var session *mgo.Session
var collections = make([]*mgo.Collection, 27)

type Element struct {
	Keyword string
	Urls    []string
}

func init() {
	fmt.Println("Opening DB connection...")
	session, err := mgo.Dial(DB_URL)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connection opened!")
	}

	session.SetMode(mgo.Monotonic, true)
	database := session.DB("index")

	for i := range COLL {
		collName := string(COLL[i])
		collections[i] = database.C(collName)
	}
	fmt.Println()
}

func Close() {
	fmt.Println("Closing database connection...")
	session.Close()
}

func GetData(keywords []string) []ResponseData {
	var result []ResponseData
	var qRes Element
	for _, k := range keywords {
		collID := int(k[0]) - 'a'
		if collID < 0 || collID > 26 {
			collID = 26
		}
		err := collections[collID].Find(bson.M{"keyword": k}).One(&qRes)
		if err != nil {
			continue
		}
		for _, url := range qRes.Urls {
			result = append(result, ResponseData{
				URL:  url,
				Name: url,
			})
		}
	}
	return result
}

func NewReq(timeout int, maxurls int, maxdist int, restrict string, urls string) {
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

func StartCrawling(config crawlit.CrawlConfig) {
	c := crawlit.NewCrawler()

	index := make(map[string][]string)
	c.Crawl(config, func(res crawlit.CrawlitResponse) error {

		keywords := ExtractKeywords(res.Body.Text())
		for _, kword := range keywords {
			index[kword] = append(index[kword], res.URL)
		}
		return nil
	})

	c.Result()

	for keyword, urls := range index {
		collID := int(keyword[0]) - 'a'
		if collID < 0 || collID > 26 {
			collID = 26
		}
		fmt.Println("Insert", keyword, "in collection", collID)
		for _, url := range urls {
			fmt.Println(url)
		}

		result := Element{}
		err := collections[collID].Find(bson.M{"keyword": keyword}).One(&result)
		if err != nil {
			fmt.Println("Couldn't find element")
		}

		if len(result.Urls) > 0 {
			urls = append(result.Urls, urls...)
		}

		err = collections[collID].Insert(&Element{
			Keyword: keyword,
			Urls:    urls,
		})
		if err != nil {
			fmt.Println("Couldn't insert", keyword, "in index")
		}
	}
}
