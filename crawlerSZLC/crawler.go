package main

import (
	"fmt"
	"log"

	"github.com/stanxii/iccrawler/crawlerSZLC/links"
)

//u := "http://www.szlcsc.com/product/catalog.html"
func main() {
	l := links.NewLinks()
	defer l.Close()

	u := "http://127.0.0.1:8001/szlcsccat?keyword=2222"
	// u := "http://10.8.15.9:8001/szlcsccat?keyword=2222"
	seeds, err := l.GetSeedURLS(u)
	if err != nil {
		fmt.Println("get init seed error")
		log.Fatal(err)
	}
	fmt.Println("seeds len=", len(seeds))
	l.CrawlerSZLC(seeds)

}
