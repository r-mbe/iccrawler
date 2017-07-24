package main

import (
	"fmt"
	"log"
  "time"

	"github.com/PuerkitoBio/goquery"
)

// Redis root seeds sets key:  root:seeds:set
func AngliaScrapeRootSeed(out chan<- string) {

	doc, err := goquery.NewDocument("http://www.anglia-live.com")
	if err != nil {
		log.Fatal(err)
	}

	//Find the review items
	doc.Find("#ctl00_ucTopNavigation_navcontainerdiv .tier3  .container .text a").Each(func(i int, s *goquery.Selection) {
		//For each item found get the detail seed url
		href, _ := s.Attr("href")
		out <- href
	})
	close(out)
}

//for save to redis
func SeedtoRedis(out chan<- string, in chan<- string) {

  //timeTick reloop and select so low cpu load
  ticker := time.NewTicker(bt.period)
  for {
  	select {
  	case <-bt.done:
  		return nil
  	case <-ticker.
  	}

  	err := bt.beat(b)
  	if err != nil {
  		return err
  	}
  }

		//save to redis
		fmt.Println(v)
		//out for print or log
		out <- v
	}
	close(out)
}

//for log or print data
func SeedLog(in chan<- string) {
	xx := make(chan int)
	for v := range xx {
		fmt.Println(v)
	}
}

func saveToRedis()

func main() {

	seeds := make(chan []byte)
	logs := make(chan []byte)

	go AngliaScrapeRootSeed(seeds)
	go SeedtoRedis(logs, seeds)
	SeedLog(logs)
}
