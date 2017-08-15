package main

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/stanxii/iccrawler/crawlerSZLC/links"
)

func worker(l *links.Links, seeds []string) {

	//real task worker very long time need be cancel up code.
	fmt.Println("Looping...working............... do work worker## ", time.Now())
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel() //cancel when we are finished conxuming string list.!

	// defer close(Pages)

	// ctx, cancel := context.WithCancel(context.Background())

	List := l.ListURLS(ctx, seeds)

	Pages := l.DetailURLS(ctx, List)

	Storages := l.DetailPage(ctx, Pages)

	l.StorageCockDB(ctx, Storages)
	//wait for pages close then close storage channel_price

	elapsed := time.Since(start)
	fmt.Println("Looping...working End End......... All storaged finish consumming save to db...  It took: ", elapsed)
}

func main() {

	//ten years.

	l := links.NewLinks()
	defer l.Stop()
	u := "http://127.0.0.1:8001/szlcsccat?keyword=2222"
	// u := "http://10.8.15.9:8001/szlcsccat?keyword=2222"
	seeds, err := l.GetSeedURLS(u)
	if err != nil {
		fmt.Println("get init seed error")
		log.Fatal(err)
	}
	fmt.Println("seeds len=", len(seeds))

	////////////////////////////////get list first.
	stop := time.After(6 * 4 * time.Minute)
	tick := time.NewTicker(6 * time.Minute)
	defer tick.Stop()

	//first time
	worker(l, seeds)

	for {
		select {
		case <-tick.C:

			worker(l, seeds)

		case <-stop:
			fmt.Println("#################All Loop done")
			return
		}
	}

}
