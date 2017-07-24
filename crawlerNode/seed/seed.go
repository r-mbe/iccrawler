package seed

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Seed seed struct
type Seed struct {
}

//RootLinksGet find root seed urs []string seed.
func (s *Seed) RootLinksGet(url string) ([]string, error) {

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	var list []string
	doc.Find("#ctl00_ucTopNavigation_navcontainerdiv .tier3  .container .text a").Each(func(i int, s *goquery.Selection) {
		// title := strings.TrimSpace(s.Text())
		href, _ := s.Attr("href")
		href = strings.TrimSpace(href)
		list = append(list, href)
		// fmt.Printf("Review %d: %s - %s\n", i, title, href)
	})

	return list, nil
}
