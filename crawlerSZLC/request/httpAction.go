package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

//CatsResponse json response from ndoe.js struct
type CatsResponse struct {
	Status  int      `json:"status"`
	Keyword string   `json:"keyword"`
	Cats    []string `json:"cats"`
}

//PageResponse json response from ndoe.js struct
type PageResponse struct {
	Status  int    `json:"status"`
	Pagemax int    `json:"pagemax"`
	Purl    string `json:"preurl"`
	Surl    string `json:"suxurl"`
}

//PartNumber str
type PartNumber struct {
	Status  int       `json:"status"`
	Keyword string    `json:"keyword"`
	Steps   []int64   `json:"steps"`
	Prices  []float64 `json:"prices"`
	Part    string    `json:"part"`
	Promaf  string    `json:"pro_maf"`
	Stock   int64     `json:"stock"`
}

//PartNumberResponse part
type PartNumberResponse struct {
	Status      int          `json:"status"`
	Keyword     string       `json:"pagemax"`
	PartNumbers []PartNumber `json:"data"`
}

//StanHTTPGetURL go request url use http get.
func StanHTTPGetURL(ur string) ([]byte, error) {

	u, _ := url.Parse(ur)
	fmt.Println("XXXXXnow real get node.js url", u.String())
	res, err := http.Get(u.String())
	if err != nil {
		// log.Fatal(err)
		return nil, err
	}

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// log.Fatal(err)
		return nil, err
	}
	defer res.Body.Close()
	// fmt.Printf("result: %s\n", string(result))
	return result, nil
}

//GetCatListFromNodeJS http reques from node.js and parse
func GetCatListFromNodeJS(ur string) ([]string, error) {
	body, err := StanHTTPGetURL(ur)
	if err != nil {
		fmt.Println("http get from node.js error.")
		return nil, err
	}

	res := CatsResponse{}

	fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return []string{}, err
	}

	fmt.Printf("Get Cats data from node.js xxxx= %d\n", len(res.Cats))
	for i, v := range res.Cats {
		fmt.Printf("a[%d]=%s\n", i, v)
	}

	return res.Cats, nil
}

/* GetPagesFromNodeJS http reques from node.js and parse
**  http://127.0.0.1:8001/szlcscpages?keyword=http://www.szlcsc.com/catalog/924.html
** ur = http://www.szlcsc.com/catalog/924.html
 */
func GetPagesFromNodeJS(ur string) (string, int, string, error) {
	// baseURL := "http://127.0.0.1:8001/szlcscpages?keyword="
	baseURL := "http://127.0.0.1:8001/szlcscpages?keyword="
	u := baseURL + ur
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js error.")
		return "", 0, "", err
	}

	res := PageResponse{}

	fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return "", 0, "", err
	}

	fmt.Printf("Get Pages status from node.js xxxx= %d\n", res.Status)
	fmt.Printf("Get pagets %s %d %s\n", res.Purl, res.Pagemax, res.Surl)

	return res.Purl, res.Pagemax, res.Surl, nil
}

/* GetPageListDetail http reques from node.js and parse
**  http://127.0.0.1:8001/szlcsclist?keyword=http://www.szlcsc.com/search/catalog_603_1_0_1-0-0-3-1_0.html&queryBeginPrice=null&queryEndPrice=null
** ur = http://www.szlcsc.com/search/catalog_603_1_0_1-0-0-3-1_0.html&queryBeginPrice=null&queryEndPrice=null
 */
func GetPageListDetail(ur string) ([]PartNumber, error) {
	// baseUrl = "http://127.0.0.1:8001/szlcsclist?keyword="
	baseURL := "http://127.0.0.1:8001/szlcsclist?keyword="
	u := baseURL + ur
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js error.")
		return nil, err
	}

	res := PartNumberResponse{}

	fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return nil, err
	}

	fmt.Printf("Get Pages status from node.js xxxx= %d\n", res.Status)

	for i, p := range res.PartNumbers {
		fmt.Printf("a[%d] = name=%v,  stock=%v\n", i, p.Part, p.Stock)
	}
	return res.PartNumbers, nil
}
