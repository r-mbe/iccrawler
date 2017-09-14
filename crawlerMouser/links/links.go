package links

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/stanxii/iccrawler/crawlerMouser/config"
	"github.com/stanxii/iccrawler/crawlerMouser/logp"
	"github.com/stanxii/iccrawler/crawlerMouser/ocsv"
	"github.com/stanxii/iccrawler/crawlerMouser/request"
	"github.com/stanxii/iccrawler/crawlerMouser/seed"
)

//Links export Links to main Cralwer
type Links struct {
	cfg    *config.Config
	s      *seed.Seed
	c      *ocsv.Ocsv
	l      logp.Log
	out    chan string
	finish chan int
	Wg     *sync.WaitGroup
}

//NewLinks newlinks
func NewLinks(c *config.Config) *Links {

	fmt.Println("init links csv")

	csv := ocsv.NewOcsv()
	err := csv.Init()
	if err != nil {
		fmt.Println("init err: ", err)
		log.Fatal(err)
	}

	fmt.Println("init links csv")

	ret := &Links{
		cfg:    c,
		c:      csv,
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

	logp := logp.NewLog()

	l.l = *logp
	l.l.Init("crawer-passive.log")

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

type CSVPassivePTNumber struct {
	Cat            string  `csv:"*产品类别"`
	Part           string  `csv:"*产品型号"`
	Description    string  `csv:"描述"`
	Package        string  `csv:"封装/规格"`
	Icode          string  `csv:"编号"`
	Promaf         string  `csv:"*品牌"`
	SalesUnitPrice float64 `csv:"销售单价"`
	PanPian        int64   `csv:"1圆盘有*片"`
	RecentSell     int64   `csv:"近期约售"`
	Stock          int64   `csv:"现货库存"`
}

func (l *Links) convertAndSave(d interface{}) error {

	o := new(ocsv.CSVPassivePTNumber)

	in, ok := d.(request.PartNumber)
	if !ok {
		return errors.New("Err data error")
	}

	o.Cat = in.Cat
	o.Part = in.Part
	o.Description = in.Detail
	o.Package = in.Pkg
	o.ICode = in.Icode
	o.Promaf = in.Promaf

	o.SalesUnitPrice = 0.0
	o.PanPian = in.YuanPan
	o.RecentSell = in.RecentSell

	o.Stock = in.Stock

	l.c.Append(o)
	return nil
}

//DetailPage out channel for list page. first output channel
func (l *Links) DetailPage(ctx context.Context, in <-chan string) <-chan interface{} {
	//consurmer

	dst := make(chan interface{})

	go func(ctx context.Context, in <-chan string) {
		for page := range in {
			// select {
			// case <-ctx.Done():
			// 	fmt.Println(">>>>>>DetailPage goroutine context Done")
			// 	return
			// default:
			l.CrawlerDetailPageFromNode(page, dst)
			// }
		}
	}(ctx, in)

	return dst

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
func (l *Links) DetailURLS(ctx context.Context, in <-chan string) <-chan string {
	//consurmer
	dst := make(chan string)

	go func(ctx context.Context, in <-chan string) {
		for href := range in {
			select {
			// case <-ctx.Done():
			// 	fmt.Println("close  detailURLS done detail chanlel")
			// 	return
			default:
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
					dst <- u
				}
			}
		}

	}(ctx, in)

	return dst
}

//ListURLS out channel for list page. first output channel
func (l *Links) ListURLS(ctx context.Context, urls []string) <-chan string {

	dst := make(chan string)
	n := 1
	go func(ctx context.Context, urls []string) {
		for _, url := range urls {
			select {
			case <-ctx.Done():
				fmt.Println("ListURLS goroutine done.")
				return //returning not to leak the goroutine.
			case dst <- string(url):
				n++
			}
		}
	}(ctx, urls)

	return dst
}

//StorageCSV channel one the last channel close done channal for singal all channal done
func (l *Links) StorageCSV(ctx context.Context, in <-chan interface{}) {
	//consurmer
	for {
		select {
		case <-ctx.Done():
			fmt.Println("sorage finished by cancel.")
			return
		case v, ok := <-in:
			//do resualt.\
			if ok {
				// os.Exit(0)
				fmt.Printf("receive  one chan storage %v.....", v)
				l.convertAndSave(v)
			} else {
				fmt.Println("recieve all chan storage....")
				return
			}
		}
	}
}

func (l *Links) Stop() {
	defer l.l.Close()
	defer l.c.Close()
}
