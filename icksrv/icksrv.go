package main

import (
	"fmt"
	"github.com/larspensjo/config"
	"./events"
	"os"
	"time"
	"strings"
)

func main() {
	args := os.Args
	if args == nil || len(args) < 2 {
		fmt.Println("usage: proxy [ini file]")
		os.Exit(1)
	}
	cnf, err := config.ReadDefault(args[1])
	if err != nil {
		fmt.Println("conferr:", err)
		os.Exit(1)
	}
	if !cnf.HasSection("main") {
		fmt.Println("conferr: missing main")
		os.Exit(1)
  }
	if !cnf.HasOption("main", "check-list") {
		fmt.Println("conferr: missing check-list")
		os.Exit(1)
  }

	checklist, err := cnf.String("main", "check-list")
	if err != nil {
		fmt.Println("conferr: ", err)
		os.Exit(1)
  }

	//Logger
	events.DefLogger = events.NewLogger(cnf)

	//rsync log
	go events.DefLogger.Rsync()

	//Check
	list := strings.Split(checklist, ",")
	chSync := make(chan events.CheckEntry, len(list))
	for _, val := range list {
		go events.Check(chSync, events.CheckEntry{args[1], val})
	}

	//Listen
	go events.Listen()

	//Watch
	events.Watch(chSync, 60 * time.Second)
}
