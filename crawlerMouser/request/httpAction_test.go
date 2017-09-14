package request

import (
	"fmt"
	"testing"
)

//Test crawler detail page from node.js
func TestGetCatListFromNodeJS(t *testing.T) {
	// u := "http://127.0.0.1:8001/szlcsccat?keyword=2222"
	u := "http://127.0.0.1:8001/szlcsccat?keyword=2222"
	urls, err := GetCatListFromNodeJS(u)
	if err != nil {
		fmt.Printf("ParseAnglia error:%s from nodejs err: %v\n", u, err)
	}

	fmt.Printf("http get body: %v \n", len(urls))

}

//Test crawler detail page from node.js
func TestGetPagesFromNodeJS(t *testing.T) {
	// u := "http://www.szlcsc.com/catalog/924.html"
	u := "http://www.szlcsc.com/catalog/924.html"
	pURL, max, sURL, err := GetPagesFromNodeJS(u)
	if err != nil {
		fmt.Printf("TestGetPagesFromNodeJS url:%s from nodejs err: %v\n", u, err)
	}

	fmt.Printf("get pages pURL:%s,  max:%d, sURL: %s, err: %v \n", pURL, max, sURL, err)

}

//Test crawler detail page from node.js
func TestGetPageListDetail(t *testing.T) {
	// u := "http://www.szlcsc.com/search/catalog_603_1_0_1-0-0-3-1_0.html&queryBeginPrice=null&queryEndPrice=null"
	u := "http://www.szlcsc.com/so/catalog_913_12_1-0-0-0-1_0.html?queryProductArrange=0&keyword=&queryBeginPrice=&queryEndPrice=&queryProductStandard=0"
	PartNumbers, err := GetPageListDetail(u)
	if err != nil {
		fmt.Printf("TestGetPagesFromNodeJS url:%s from nodejs err: %v\n", u, err)
	}

	fmt.Printf("get  partnums %d \n", len(PartNumbers))

}

/*  get page prefix , max pagenum, sufix
http://127.0.0.1:8001/szlcscpages?keyword=http://www.szlcsc.com/catalog/924.html
*/

//test for request http.
// func TestStanHTTPGetURL(t *testing.T) {
// 	u := "http://www.anglia-live.com/products/batteries-accessories/batteries/ic-batteries-memory-backup"
// 	data, err := StanHTTPGetURL(u)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	fmt.Printf("http get body: %s \n", string(data))
//
// }
