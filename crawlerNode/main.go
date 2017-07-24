package main

import (
	"fmt"
	"log"

	"github.com/robfig/cron"
	"techtoolkit.ickey.cn/crawlerNode/links"
)

//AJob struct
type AJob struct {
	l *links.Links
}

//Run interface
func (m *AJob) Run() {
	fmt.Println("I am runnning task.")

	u := "http://www.anglia-live.com"
	seeds, err := m.l.GetSeedURLS(u)
	if err != nil {
		fmt.Println("get init seed error")
		log.Fatal(err)
	}
	m.l.LinksList(seeds)
}

func taskWithParams(a int, b string) {
	fmt.Println(a, b)
}

func main() {

	m := &AJob{}

	m.l = links.NewLinks()

	defer m.l.Close()

	///////////////
	//var job AJob

	done := make(chan struct{})
	c := cron.New()
	defer c.Stop()

	//seconds - minutes - hours - day of mounth - month - day of week (0-6 , SUN to SAT)
	//will run a lot of times running task times.
	//c.AddJob("* 44 20  * * 2", m)

	// run once
	c.AddJob("0 01 09  * * 6", m)

	c.Start()

	done <- struct{}{}

}
