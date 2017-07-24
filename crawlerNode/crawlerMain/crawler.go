package main

import (
	"fmt"
	"log"

	"techtoolkit.ickey.cn/crawlerNode/links"
)

func main() {
	u := "http://www.anglia-live.com"
	l := links.NewLinks()

	seeds, err := l.GetSeedURLS(u)
	if err != nil {
		fmt.Println("get init seed error")
		log.Fatal(err)
	}
	l.LinksList(seeds)

}
