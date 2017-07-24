package links

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	ftp "github.com/jlaffaye/ftp"
	"techtoolkit.ickey.cn/crawlerNode/ocsv"
	"techtoolkit.ickey.cn/crawlerNode/request"
	"techtoolkit.ickey.cn/crawlerNode/seed"
)

//Links export Links to main Cralwer
type Links struct {
	Cat    string
	s      *seed.Seed
	c      *ocsv.Ocsv
	out    chan string
	finish chan int
	Wg     *sync.WaitGroup
}

// type Data struct {
// 	Part  string   `json:"part"`
// 	Stock string   `json:"stock"`
// 	Pl    []string `json:"pl"`
// 	Pr    []string `json:"pt"`
// }
//
// type JsonRes struct {
// 	Status  int64  `json:"status"`
// 	Keyword int64  `json:"keyword"`
// 	data    []Data `json:"data"`
// }

//NewLinks newlinks
func NewLinks() *Links {
	ret := &Links{
		out:    make(chan string),
		finish: make(chan int),
	}
	ret.init()
	return ret
}

//Close release
func (l *Links) Close() {
	defer l.c.Close()
}

func (l *Links) init() {
	// Initialize the internal hosts map
	// c.hosts = make(map[string]struct{}, len(ctxs))
	l.s = new(seed.Seed)
	l.Wg = new(sync.WaitGroup)
	l.c = ocsv.NewOcsv()
	err := l.c.Init()
	if err != nil {
		log.Fatal(err)
	}

}

//GetSeedURLS get init seeds.
func (l *Links) GetSeedURLS(u string) ([]string, error) {
	//init data
	seeds, err := l.s.RootLinksGet(u)
	l.s.RootLinksGet(u)
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range seeds {
		fmt.Printf("a[%d] = %s\n", i, v)
	}
	return seeds, err

}

//parseListPage
func (l *Links) parseListPage(doc *goquery.Document) ([]string, error) {

	var list []string
	//fmt.Println("parse Html list page" + doc.Text())
	doc.Find("#ucProductList_pnlResults table#ucProductList_tbltop ul.product-list.products-plain > li").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("get li html=%s \n", s.Text())
		href, exist := s.Find(".description-container > h3.productNameList a").First().Attr("href")
		if exist {
			fmt.Printf("final get detail link=[%s]\n", href)
			list = append(list, href)
		}
	})

	return list, nil

}

//CrawlerCatListFromNode page
func (l *Links) CrawlerCatListFromNode(url string, out chan<- []string) error {
	// l.Wg.Add(1)

	fmt.Printf("###url From node.js list-data--list url = %s\n", url)
	list, err := request.PostParseAngliaList(url)

	if err != nil || list == nil {
		fmt.Printf("http reques err: %v\n", err)
		return err
	}

	fmt.Printf("List len=== %d\n", len(list))
	if len(list) > 0 {
		for i, v := range list {
			fmt.Printf("######## >>>>>> out<-data[%d] = %s\n", i, v)
		}
		out <- list
		fmt.Printf("###after send to l.out<- list")
	}

	return nil
}

func GetListPagesNum(href string, out chan<- string) error {
	//get pages num.
	pages, err := request.GetAngliaListPageNums(href)
	if err != nil || pages < 0 {
		fmt.Println("getListPagesNum err:", err)
		return err
	} else if pages == 1 {
		out <- href
	} else {
		//pages numb > 1.
		for i := 2; i <= pages; i++ {
			//construct pages  http://www.anglia-live.com/products/connectors/dc-power/plugs-sockets#resultspage,2|
			hrefPage := fmt.Sprintf("%s#resultspage,%d|", href, i)
			out <- hrefPage
		}
	}
	return nil
}

//CrawlerCatLinks crawler a cat inclue subcat page
func (l *Links) CrawlerCatLinks(url string, out chan<- string) error {
	doc, err := goquery.NewDocument(url)

	if err != nil {
		fmt.Printf("url===%s, err: %v\n", url, err)
		return err
	}

	subcat := doc.Find("td.ProductTextCell a.product-nav-name")
	if subcat != nil {
		fmt.Printf("url: %s crawlerCatLinks cat page items %d \n ", url, subcat.Length())

		subcat.Each(func(i int, s *goquery.Selection) {
			// title := strings.Tri mSpace(s.Text())
			href, exist := s.Attr("href")
			if exist {
				fmt.Println(">>>Wish List Crawler...url in Goroutine." + href)

				//retry 8 times
				var i int
				for i = 0; i < 8; i++ {
					err := GetListPagesNum(href, out)
					if err != nil {
						fmt.Println("SubCat get pagesnum err retry", i, href)
						continue
					} else {
						break
					}
				}
				if i >= 8 {
					//max
					fmt.Printf(">>>XXXXXAfter retry get max page err url=%s\n", href)
				}
			}
			// fmt.Printf("Review %d: %s - %s\n", i, title, href)
		})
		//wait group goroutine
		fmt.Println("after... wg.wait subcat link.")

	} else {
		//retry 8 times
		var i int
		for i = 0; i < 8; i++ {
			err := GetListPagesNum(url, out)
			if err != nil {
				fmt.Println("no subcat get directlist get pagesnum err retry", i, url)
				continue
			} else {
				break
			}
		}
		if i >= 8 {
			//max
			fmt.Printf(">>>XXXXXAfter retry get max page err url=%s\n", url)
		}
		// l.Wg.Done()
	}

	//close channel all
	return nil

}

//CrawlerDetailPageFromNode crawler detali
// func (l *Links) CrawlerDetailPageFromNode(urls []string, out chan<- interface{}) {
//
// 	var wg sync.WaitGroup
//
// 	for _, v := range urls {
// 		go func(href string) {
// 			wg.Add(1)
// 			defer wg.Done()
// 			fmt.Printf("###href From = %s\n", href)
// 			data, err := request.ParseAnglia(href)
// 			if err != nil {
// 				fmt.Printf("http reques detail page:%s from nodejs err: %v\n", href, err)
// 			}
// 			//fmt.Printf("###>>>>>>>---->>>>get one page data json %s\n", data)
// 			out <- data
// 		}(v)
// 	}
// 	wg.Wait()
// }

func (l *Links) CrawlerDetailPageFromNode(urls []string, out chan<- interface{}) {

	for _, href := range urls {
		fmt.Printf("###href From = %s\n", href)

		//retry 8 times
		var i int
		var err error
		var data interface{}
		for i = 0; i < 8; i++ {
			data, err = request.ParseAnglia(href)
			if err != nil {
				fmt.Println("Err CrawlerDetailPageFromNode  retry", i, href)
				continue
			} else {
				break
			}
		}
		if i >= 8 {
			//max
			fmt.Printf(">>>XXXXXAfter retry %d times ERr CrawlerDetailPageFromNode url=%s\n", i, href)
		} else {
			out <- data
		}

		//fmt.Printf("###>>>>>>>---->>>>get one page data json %s\n", data)
	}
}

func (l *Links) convertAndSave(d interface{}) error {
	var err error
	o := new(ocsv.CSVPartNumber)

	in, ok := d.(request.PartNumber)
	if !ok {
		return errors.New("Err data error")
	}
	o.Part = in.Part
	o.Comments = in.Keyword
	o.Description = in.Description
	o.Promaf = in.Promaf
	o.Package = in.Pkgtype
	o.Rosh = in.Rohs
	o.KuCunDi = in.Tcode
	o.Stock, err = strconv.ParseInt(in.Stock, 10, 64)
	if err != nil {
		fmt.Printf("strconv.Parset in.sTock:%s conver int64 error, err:%v", in.Stock, err)
		o.Stock = 0
	}
	o.StockBiaoStock = in.ReserveStock
	o.NotUse1 = in.TotalStockDue
	o.NotUse2 = in.TotalResStockDue
	o.Spq, err = strconv.ParseInt(in.Spq, 10, 64)
	if err != nil {
		fmt.Printf("strconv.Parset in.sTock:%s conver int64 error, err:%v", in.Stock, err)
		o.Spq = 0
	}
	o.Moq, err = strconv.ParseInt(in.Moq, 10, 64)
	if err != nil {
		fmt.Printf("strconv.Parset in.sTock:%s conver int64 error, err:%v", in.Stock, err)
		o.Moq = 0
	}

	//非数字
	pattern := `[\\d+$]`
	for i, v := range in.Steps {
		if i < 10 {
			if len(v) <= 0 {
				fmt.Println("strlen <=0")
				continue
			}
			//"1-42" or 43+. get last number first
			// arr := strings.SplitAfter(v, "-")
			arr := strings.Split(v, "-")
			// b := arr[len(arr)-1]
			//get the first number changed
			b := arr[0]
			reg := regexp.MustCompile(pattern)
			s := reg.ReplaceAllString(b, "")
			fmt.Println("befort parse int s=", s)

			iv, err2 := strconv.ParseInt(s, 10, 64)
			if err2 != nil {
				fmt.Println("conver to int64 err ,data=", v)
			}
			if 0 == i {
				o.PurchaseNum1 = iv
			} else if 1 == i {
				o.PurchaseNum2 = iv
			} else if 2 == i {
				o.PurchaseNum3 = iv
			} else if 3 == i {
				o.PurchaseNum4 = iv
			} else if 4 == i {
				o.PurchaseNum5 = iv
			} else if 5 == i {
				o.PurchaseNum6 = iv
			} else if 6 == i {
				o.PurchaseNum7 = iv
			} else if 7 == i {
				o.PurchaseNum8 = iv
			} else if 8 == i {
				o.PurchaseNum9 = iv
			} else if 9 == i {
				o.PurchaseNum10 = iv
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
	return err
}

//Storages channel one the last channel close done channal for singal all channal done
func (l *Links) Storages(in <-chan interface{}, done chan<- struct{}) {
	//consurmer
	defer close(done)
	for {
		select {
		case v, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Printf("receive  one chan storage %v.....", v)
				l.convertAndSave(v)
			} else {
				fmt.Println("recieve all chan storage....")
				return
			}
		}
	}

}

//ListPage out channel for list page. first output channel
func (l *Links) detailPage(out chan<- interface{}, in <-chan []string) {
	//consurmer
	defer close(out)
	for {
		select {
		case pages, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Println("received one page: ", pages)
				l.CrawlerDetailPageFromNode(pages, out)
			} else {
				fmt.Println("received all chan pages")
				return
			}
		}
	}
}

//ListPage out channel for list page. first output channel
func (l *Links) detailURLS(out chan<- []string, in <-chan string) {
	//consurmer
	defer close(out)
	for {
		select {
		case href, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Printf("received chan: list url= %s\n", href)
				// go l.CrawlerCatListFromNode(href, out)

				//retry 8 times
				var i int
				for i = 0; i < 8; i++ {
					err := l.CrawlerCatListFromNode(href, out)

					if err != nil {
						fmt.Println("Err CrawlerCatListFromNode  retry", i, href)
						continue
					} else {
						break
					}
				}
				if i >= 8 {
					//max
					fmt.Printf(">>>XXXXXAfter retry %d times err url=%s\n", i, href)
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
func (l *Links) ListURLS(urls []string, out chan<- string) {
	var err error

	defer close(out)

	for _, url := range urls {
		// fmt.Println("XXXX first url=" + url)
		err = l.CrawlerCatLinks(url, out)
		if err != nil {
			fmt.Printf("crawlerCatLinks err:%s\n", err)
			//return
			continue
		}
	}

}

/*LinksList  crawler list links from root seed []string  array.
  urls root seeds.
  out chan  put real per ic url into out channal
*/
func (l *Links) LinksList(urls []string) {
	start := time.Now().Unix()

	list := make(chan string)
	pages := make(chan []string)
	storages := make(chan interface{})

	done := make(chan struct{})

	//input, out-chan, out-chan

	go l.detailURLS(pages, list)
	go l.detailPage(storages, pages)
	go l.Storages(storages, done)
	l.ListURLS(urls, list)

	//wait all finished.

	defer close(storages)
	<-done

	//close csv file
	l.Close()

	//upload to ftp filename.
	fbegin := time.Now()
	ftpUploadFile("feed.data.ickey.cn:21", "anglia-live", "c6LfZthyVBy45tiB", l.c.Name, "/", l.c.Name)
	dur := time.Since(fbegin).Seconds()
	fmt.Printf("Upload file spend time: %v  seconds\n", dur)

	end := time.Now().Unix()
	fmt.Printf("All Done:  spend - time %d\n", end-start)

}

func ftpUploadFile(ftpserver, ftpuser, pw, localFile, remoteSavePath, saveName string) {
	ftp, err := ftp.Connect(ftpserver)
	if err != nil {
		fmt.Println(err)
	}
	err = ftp.Login(ftpuser, pw)
	if err != nil {
		fmt.Println(err)
	}
	//注意是 pub/log，不能带“/”开头
	ftp.ChangeDir("pub/log")
	dir, err := ftp.CurrentDir()
	fmt.Println(dir)
	ftp.MakeDir(remoteSavePath)
	ftp.ChangeDir(remoteSavePath)
	dir, _ = ftp.CurrentDir()
	fmt.Println(dir)
	file, err := os.Open(localFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	err = ftp.Stor(saveName, file)
	if err != nil {
		fmt.Println(err)
	}
	ftp.Logout()
	ftp.Quit()
	fmt.Println("success upload file:", localFile)
}

//!-
//!-
