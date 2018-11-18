/* package indexer

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/fedemengo/crawlit"
)

const URL = "https://fedemengo.github.io/"

const COLL = "abcdefghijklmnopqrstuvwxyz*"

const dbURL = "mongodb://127.0.0.1"

type Element struct {
	Keyword string
	Urls    []string
}

func main() {

	fmt.Println("Opening DB connection...")
	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connection opened!")
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	database := session.DB("index")

	collections := make([]*mgo.Collection, 27)
	for i := range COLL {
		collName := string(COLL[i])
		collections[i] = database.C(collName)
	}

	fmt.Println("Initialized collections")

	c := crawlit.NewCrawler()

	config := crawlit.CrawlConfig{
		SeedURLs:    []string{URL},
		MaxURLs:     20,
		MaxDistance: 3,
		Timeout:     3,
		Restrict:    false,
	}

	index := make(map[string][]string)
	c.Crawl(config, func(res crawlit.CrawlitResponse) error {

		keywords := ExtractKeywords(res.Body.Text())
		for _, kword := range keywords {
			index[kword] = append(index[kword], res.URL)
		}

		if err != nil {
			fmt.Println("Couldn't load document")
		}

		err = Index(res.Body.Text())

		if err != nil {
			fmt.Print("Couldn't index page", res.URL)
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
		err = collections[collID].Find(bson.M{"keyword": keyword}).One(&result)
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

	var test string
	fmt.Scanln(&test)
	result := Element{}

	collID := int(test[0]) - 'a'
	if collID < 0 || collID > 26 {
		collID = 26
	}

	fmt.Println("searchin", test, "in collection", collID)
	err = collections[collID].Find(bson.M{"keyword": test}).One(&result)

	fmt.Println(test, "found in", len(result.Urls), "pages")
	for _, url := range result.Urls {
		fmt.Println(url)
	}
}
*/
package indexer
