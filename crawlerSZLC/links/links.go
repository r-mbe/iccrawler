package links

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/seiflotfy/cuckoofilter"
	"github.com/stanxii/iccrawler/crawlerSZLC/cockroach"
	"github.com/stanxii/iccrawler/crawlerSZLC/mylog"
	"github.com/stanxii/iccrawler/crawlerSZLC/ocsv"
	"github.com/stanxii/iccrawler/crawlerSZLC/request"
	"github.com/stanxii/iccrawler/crawlerSZLC/seed"
)

//Links export Links to main Cralwer
type Links struct {
	cock   *cockroach.Client
	s      *seed.Seed
	c      *ocsv.Ocsv
	l      *mylog.Log
	cf     *cuckoofilter.CuckooFilter
	out    chan string
	finish chan int
	Wg     *sync.WaitGroup
}

//NewLinks newlinks
func NewLinks() *Links {
	ret := &Links{
		out:    make(chan string),
		finish: make(chan int),
	}
	ret.init()
	return ret
}

func (l *Links) init() {
	// Initialize the internal hosts map
	// c.hosts = make(map[string]struct{}, len(ctxs))
	l.s = new(seed.Seed)
	l.Wg = new(sync.WaitGroup)

	//cuckoofilter size = 5000w  bucket slot size = 4
	// l.cf = cuckoofilter.NewCuckooFilter(50000000)
	l.cf = cuckoofilter.NewCuckooFilter(50)
	// l.c = ocsv.NewOcsv()
	// err := l.c.Init()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	isDebug := true
	//init cockroachdb
	// var dbURL, nsqaddr, ntopic, nchannel string
	var dbURL string
	if isDebug {
		// nsqaddr = "10.8.15.9:4161"
		// ntopic = "topic_cock"
		// nchannel = "channel_price"
		dbURL = "postgresql://stan:888888@172.31.225.122:26257/db_product?sslmode=disable"
	} else {
		// nsqaddr = "10.8.51.50:4161"
		// ntopic = "topic_cock"
		// nchannel = "channel_price"
		dbURL = "postgresql://stan@10.8.51.69:26257/db_product?sslcert=/usr/local/ickey-certs/client-stan/client.stan.crt&sslkey=/usr/local/ickey-certs/client-stan/client.stan.key&sslrootcert =/usr/local/ickey-certs/client-stan/ca.crt&sslmode=require"
	}

	l.cock = cockroach.NewClient(dbURL)
	l.l = mylog.NewLog()
	l.l.Init("szlcsc.log")

}

//GetSeedURLS get init seeds.
func (l *Links) GetSeedURLS(u string) ([]string, error) {
	//init data

	seeds, err := l.s.RootLinksGet(u)
	if err != nil {
		log.Fatal(err)
	}

	return seeds, err

}

func (l *Links) CrawlerDetailPageFromNode(href string, out chan<- interface{}) {

	//get list page from seed and request module
	fmt.Printf("###href From = %s\n", href)
	var data []request.PartNumber
	var err error

	// fmt.Println("Coockfilter counting", l.cf.Count())
	// b := l.cf.Lookup([]byte(href))
	// if b {
	// 	fmt.Println("ERR########## href exist. do crawler twice.", href)
	// 	return
	// }

	var i int
	for i = 0; i < 3; i++ {

		data, err = l.s.GetPageListDetail(href)
		if err != nil && data != nil {
			fmt.Printf("http reques detail page:%s from nodejs err: %v, data:%v \n", href, err, data)
			continue
		} else {
			for i, v := range data {
				fmt.Printf("###>>>>>>>---->>>>ok-ok-ok will save cs data i=%v, href=%v, v=%v \n", i, href, v)

				out <- v
			}
			//i < 3 break ok

			// l.cf.Insert([]byte(href))

			break
		}

	}

	fmt.Printf(">>>XXXXXAfter retry get detail page i=%d url=%s\n", i, href)
	if i >= 15 {
		fmt.Printf(">>>XXXXXAfter retry get detail page err url=%s\n", href)
		l.l.Error("err get page max" + href)
	}

}

func (l *Links) convertAndSave(d interface{}) error {
	o := new(ocsv.CSVPartNumber)

	in, ok := d.(request.PartNumber)
	if !ok {
		return errors.New("Err data error")
	}

	o.Part = in.Part
	o.Comments = in.Keyword
	o.Promaf = in.Promaf
	o.Stock = in.Stock

	o.Cat = in.Cat
	o.ProductDetail = in.Detail
	o.Package = in.Pkg
	o.Description = in.Desc

	//非数字
	// pattern := `[\\d+$]`
	for i, v := range in.Steps {
		if i < 10 {
			if 0 == i {
				o.PurchaseNum1 = v
			} else if 1 == i {
				o.PurchaseNum2 = v
			} else if 2 == i {
				o.PurchaseNum3 = v
			} else if 3 == i {
				o.PurchaseNum4 = v
			} else if 4 == i {
				o.PurchaseNum5 = v
			} else if 5 == i {
				o.PurchaseNum6 = v
			} else if 6 == i {
				o.PurchaseNum7 = v
			} else if 7 == i {
				o.PurchaseNum8 = v
			} else if 8 == i {
				o.PurchaseNum9 = v
			} else if 9 == i {
				o.PurchaseNum10 = v
			}
		}
	}

	for i, v := range in.Prices {
		if i < 10 {
			if 0 == i {
				o.PurchaseUnitPrice1 = v
			} else if 1 == i {
				o.PurchaseUnitPrice2 = v
			} else if 2 == i {
				o.PurchaseUnitPrice3 = v
			} else if 3 == i {
				o.PurchaseUnitPrice4 = v
			} else if 4 == i {
				o.PurchaseUnitPrice5 = v
			} else if 5 == i {
				o.PurchaseUnitPrice6 = v
			} else if 6 == i {
				o.PurchaseUnitPrice7 = v
			} else if 7 == i {
				o.PurchaseUnitPrice8 = v
			} else if 8 == i {
				o.PurchaseUnitPrice9 = v
			} else if 9 == i {
				o.PurchaseUnitPrice10 = v
			}
		}
	}
	l.c.Append(o)
	return nil
}

func (l *Links) DoCockStorage(d interface{}) error {

	in, ok := d.(request.PartNumber)
	if !ok {
		return errors.New("Err data error")
	}

	err := l.cock.DoSave(in)
	if err != nil {
		return errors.New("db save err.")
	}

	return nil
}

//StorageCockDB channel one the last channel close done channal for singal all channal done
func (l *Links) StorageCockDB(ctx context.Context, in <-chan interface{}) {
	//consurmer

	queue := []interface{}{}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("StorageCockDB list finished.")
			return
		case v, ok := <-in:
			if ok {
				fmt.Println("XXXXOOOOO##### len(queue) storage channel==", len(queue))
				if len(queue) >= 10 {
					//save to db
					fmt.Println(">>>>>>>>>>>>>>>>>>>>500000 ##### len(queue) storage channel==", len(queue))

					for _, item := range queue {
						l.DoCockStorage(item)
					}
					queue = nil
					fmt.Println("XXXXOOOOO##### after Nil len(queue)  channel==", len(queue))

				} else {
					queue = append(queue, v)
				}
			} else {
				// channel is empty all finished.
				fmt.Println("all storage done. in StorageCockDB .")
				return
			}

		}
	}
}

//Storages channel one the last channel close done channal for singal all channal done
// func (l *Links) Storages(in <-chan interface{}, done chan<- struct{}) {
// 	//consurmer
// 	for {
// 		select {
// 		case v, ok := <-in:
// 			//do resualt.\
// 			if ok {
// 				fmt.Printf("receive  one chan storage %v.....", v)
// 				l.convertAndSave(v)
// 			} else {
// 				fmt.Println("recieve all chan storage....")
// 				return
// 			}
// 		}
// 	}
// }

//DetailPage out channel for list page. first output channel
func (l *Links) DetailPage(ctx context.Context, out chan<- interface{}, in <-chan string) {
	//consurmer

	defer close(out)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("DetailPage list finished.")
			return
		case page, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Println("received one page: ", page)

				// b := l.cf.Lookup([]byte(page))
				// if b {
				// 	fmt.Println("ERR Bloom filter ########## check detailPage.", page)
				// 	return
				// }

				l.CrawlerDetailPageFromNode(page, out)

				// l.cf.Insert([]byte(page))
			} else {
				fmt.Println("received all chan pages")
				return
			}
		}
	}
}

//CrawlerCatListFromNode page
func (l *Links) CrawlerCatListFromNode(url string) ([]string, error) {
	// l.Wg.Add(1)
	var res []string
	pURL, max, sURL, err := l.s.GetPagesFromNodeJS(url)
	if err != nil {
		fmt.Printf("CrawlerCatListFromNode url:%s from nodejs err: %v\n", url, err)
		return nil, err
	}

	fmt.Printf("######## >>>> GetPagesFromNodeJS pre: %s, max: %d, sufix: %s len\n", pURL, max, sURL)
	if max > 1 {
		// fmt.Printf("######## >>>>>> out<-data[%d] = %s\n", max, url)
		for i := 1; i <= max; i++ {
			s := fmt.Sprintf("%s%d%s", pURL, i, sURL)
			fmt.Printf("########### >>>>After send to l.out<- list: %s,  %v\n", s, max)
			res = append(res, s)
		}
	} else if max == 1 {
		s := fmt.Sprintf("%s%d%s", pURL, max, sURL)
		// fmt.Printf("###After send to l.out<- list: %s  %v\n", s, max)
		res = append(res, s)
	} else {
		//error
		return nil, errors.New("max=0")
	}

	return res, nil
}

//DetailURLS out channel for list page. first output channel
func (l *Links) DetailURLS(ctx context.Context, out chan<- string, in <-chan string) {
	//consurmer

	for {
		select {
		case <-ctx.Done():
			fmt.Println("DetailURLS list finished.")
			return
		case href, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Printf(">>>detailURLS Received chan: list url= %s\n", href)
				// go l.CrawlerCatListFromNode(href, out)
				//if crawler max page num err retry 3 times.
				var data []string
				var err error
				var i int
				for i = 0; i < 5; i++ {

					data, err = l.CrawlerCatListFromNode(href)
					if len(data) > 0 {
						fmt.Println(">>>>>>>>>>>>>>>>Get list from node len >0 ", len(data))
						break
					}
					if err != nil {
						fmt.Println("l.CrawlerCatListFromNode: err", err)
					}
				}
				if i >= 5 {
					fmt.Printf(">>>XXXXXAfter retry get max page err url=%s\n", href)
					l.l.Error("err get page max" + href)
				}

				fmt.Println("List page url len=", href, len(data))
				for i, u := range data {
					fmt.Printf(">>>XXXXX Go detailURLS  sending  i=%d,url=%s\n", i, u)
					out <- u
				}

			} else {
				//channel closed
				fmt.Println("received all chan list")
				return
			}
		}
	}
}

//ListURLS out channel for list page. first output channel
func (l *Links) ListURLS(ctx context.Context, urls []string) <-chan string {

	dst := make(chan string)
	n := 1
	go func(ctx context.Context, urls []string) {
		for _, url := range urls {
			select {
			case <-ctx.Done():
				return //returning not to leak the goroutine.
			case dst <- string(url):
				n++
			}
		}
	}(ctx, urls)

	return dst
}

func (l *Links) Stop() {
	defer l.l.Close()
	defer l.cock.Close()
}
