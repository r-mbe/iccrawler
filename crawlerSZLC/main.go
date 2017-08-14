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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() //cancel when we are finished conxuming string list.!

	Pages := make(chan string)
	Storages := make(chan interface{})

	// defer close(Pages)
	// defer close(Storages)

	// ctx, cancel := context.WithCancel(context.Background())

	List := l.ListURLS(ctx, seeds)

	go l.DetailURLS(ctx, Pages, List)
	go l.DetailPage(ctx, Storages, Pages)

	l.StorageCockDB(ctx, Storages)
	close(Storages)

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
	stop := time.After(20 * 4 * time.Second)
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
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
