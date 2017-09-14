package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/stanxii/iccrawler/crawlerMouser/config"
	"github.com/stanxii/iccrawler/crawlerMouser/links"
)

func worker(l *links.Links, seeds []string) {

	//real task worker very long time need be cancel up code.
	fmt.Println("Looping...working............... do work worker## ", time.Now())
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	defer cancel() //cancel when we are finished conxuming string list.!

	// defer close(Pages)

	// ctx, cancel := context.WithCancel(context.Background())

	List := l.ListURLS(ctx, seeds)

	Pages := l.DetailURLS(ctx, List)

	Storages := l.DetailPage(ctx, Pages)

	l.StorageCSV(ctx, Storages)
	//wait for pages close then close storage channel_price

	elapsed := time.Since(start)
	fmt.Println("Looping...working End End......... All storaged finish consumming save to db...  It took: ", elapsed)
}

func main() {

	fmt.Println("hello ..")

	configFile := flag.String("config", "./etc/crawler.passive.toml", "crawler passive config file.")
	flag.Parse()

	fmt.Printf("configFile =: %v\n", *configFile)

	c, err := config.NewConfigWithFile(*configFile)
	if err != nil {
		log.Fatal("Err config file error.", err)
	}

	l := links.NewLinks(c)
	if err != nil {
		log.Fatal("Err crawler Passive error.", err)
	}

	defer l.Stop()

	//passive seeds
	//page = 25
	//http://www.mouser.cn/Passive-Components/Antennas/_/N-8w0fa/?No=75
	seeds := []string{
		"http://www.mouser.cn/Passive-Components/Antennas/_/N-8w0fa/",
     "http://www.mouser.cn/Passive-Components/Ferrites/_/N-fb8t2/",
		 "http://www.mouser.cn/Passive-Components/Signal-Conditioning/_/N-8bzui/",
		 "http://www.mouser.cn/Passive-Components/Audio-Transformers-Signal-Transformers/_/N-5gbg/",
		 "http://www.mouser.cn/Passive-Components/Frequency-Control-Timing-Devices/_/N-6zu9e/",
		 "http://www.mouser.cn/Passive-Components/Thermistors-NTC/_/N-6g7mw/",
		 "http://www.mouser.cn/Passive-Components/Capacitors/_/N-5g7r/",
		 "http://www.mouser.cn/Passive-Components/Inductors-Chokes-Coils/_/N-5gb4/",
		 "http://www.mouser.cn/Passive-Components/Thermistors-PTC/_/N-796na/",
		 "http://www.mouser.cn/Passive-Components/EMI-Filters-EMI-Suppression/_/N-18v9d/",
		 "http://www.mouser.cn/Passive-Components/Potentiometers-Trimmers-Rheostats/_/N-9q0yi/",
		 "http://www.mouser.cn/Passive-Components/Varistors/_/N-6g7mv/",
		 "http://www.mouser.cn/Passive-Components/Encoders/_/N-6g7nx/",
		 "http://www.mouser.cn/Passive-Components/Resistors/_/N-5g9n/"
	}

	fmt.Println("seeds len=", len(seeds))
	//first time
	worker(l, seeds)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	fmt.Println()
	fmt.Println(sig)
	l.Stop()
	fmt.Println("awaiting signal")

	//send to nsq exist for loop.
	fmt.Println("exiting")
}
