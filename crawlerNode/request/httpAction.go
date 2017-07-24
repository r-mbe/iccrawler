package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

//Response json response from ndoe.js struct
type Response struct {
	Status  int      `json:"status"`
	Keyword string   `json:"keyword"`
	Sups    []string `json:"sups"`
}

/*
{
  "status": 0,
  "keyword": "2222",
  "sups": {
    "steps": [
      "1-23",
      "24-46",
      "47+"
    ],
    "prices": [
      3.78523,
      3.01795,
      2.97771
    ],
    "description": "SNAPHAT BATTERY FOR TIMEKEEPER",
    "part": "M4T28-BR12SH1",
    "pro_maf": "STMICROELECTRONICS",
    "pkgtype": "Package Type: Tube",
    "rohs": "RoHS:",
    "tcode": "Tariff Code: 8542324500",
    "stock": "2997",
    "reserveStock": "0",
    "totalStockDue": "0",
    "totalResStockDue": "0",
    "supplierLeadTime": "20",
    "spq": "23",
    "moq": "1",
    "orderMultiple": "1"
  }
}
*/

//PartNumber str
type PartNumber struct {
	Status           int       `json:"status"`
	Keyword          string    `json:"keyword"`
	Steps            []string  `json:"steps"`
	Prices           []float64 `json:"prices"`
	Description      string    `json:"description"`
	Part             string    `json:"part"`
	Promaf           string    `json:"pro_maf"`
	Pkgtype          string    `json:"pkgtype"`
	Rohs             string    `json:"rohs"`
	Tcode            string    `json:"tcode"`
	Stock            string    `json:"stock"`
	ReserveStock     string    `json:"reserveStock"`
	TotalStockDue    string    `json:"totalStockDue"`
	TotalResStockDue string    `json:"totalResStockDue"`
	SupplierLeadTime string    `json:"supplierLeadTime"`
	Spq              string    `json:"spq"`
	Moq              string    `json:"moq"`
	OrderMultiple    string    `json:"orderMultiple"`
}

//PagesResponse json response from ndoe.js struct
type PagesResponse struct {
	Status  int    `json:"status"`
	Keyword string `json:"keyword"`
	Pages   int    `json:"pages"`
}

type AngliaPostResponse struct {
	Status  int      `json:"status"`
	Keyword string   `json:"keyword"`
	Sups    []string `json:"sups"`
}

// PagesResponse
type BestProxyIP struct {
	Status int     `json:"status"`
	ID     string  `json:"id"`
	IP     string  `json:"ip"`
	Dur    float64 `json:"dur"`
	Err    string  `json:"err,omitempty"`
}

type BestIPResponse struct {
	Bestip *BestProxyIP `json:"bestip"`
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

//St   var jsonprep string = `{"username":"`+username+`","password":"`+password+`"}`
//     var jsonStr = []byte(jsonprep)
/*
u := "http://47.91.138.212:8001/anglialist"
k := "http://www.anglia-live.com/products/circuit-protection/fuses-breakers-holders/resettable-fuses#resultspage,4|"
var jsonrep string = `{"keyword":"` + k + `"}`
*/
func StanHTTPPostURL(url string, jsonStr []byte) ([]byte, error) {

	client := http.Client{}
	fmt.Println("post params url=", url, string(jsonStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Printf("Parse Json param err, %v\n", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Post http request  err, %v\n", err)
		return nil, err
	}
	// defer resp.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Parse Json response body err: %v\n", err)
		return nil, err
	}
	return body, nil
}

func PostParseAngliaList(keyword string) ([]string, error) {

	u := "http://127.0.0.1:8001/anglialist"
	var jsonrep string = `{"keyword":"` + keyword + `"}`
	jsonStr := []byte(jsonrep)
	body, err := StanHTTPPostURL(u, jsonStr)
	if err != nil {
		fmt.Printf("Stan http Post err %v\n", err)
		return nil, err
	}
	res := AngliaPostResponse{}

	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return nil, err
	}

	if res.Status != 0 {
		fmt.Println("Err Parse AngliaList  Post response status = 0")
		return nil, errors.New("No List Data")
	}

	fmt.Printf("Post get responsed  data xxxx= %d\n", len(res.Sups))
	for i, v := range res.Sups {
		fmt.Printf("a[%d]=%s\n", i, v)
	}
	return res.Sups, nil
}

//ParseAngliaList http reques from node.js and parse
func ParseAngliaList(ur string) ([]string, error) {
	baseURL := "http://127.0.0.1:8001/anglialist?keyword="
	u := baseURL + ur
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js error.")
		return nil, err
	}

	res := Response{}

	// fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return nil, err
	}

	if res.Status != 0 {
		fmt.Println("Err ParseAngliaList status = 0")
		return nil, errors.New("No List Data")
	}

	fmt.Printf("data xxxx= %d\n", len(res.Sups))
	for i, v := range res.Sups {
		fmt.Printf("a[%d]=%s\n", i, v)
	}
	return res.Sups, nil
}

//ParseAnglia pare anglia detail
func ParseAnglia(ur string) (interface{}, error) {
	//baseURL := "http://10.8.15.9:8001/anglia?keyword="
	baseURL := "http://127.0.0.1:8001/anglia?keyword="
	u := baseURL + ur
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js one detail page error.")
		return nil, err
	}

	res := PartNumber{}

	// fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return nil, err
	}

	if res.Status != 0 {
		// fmt.Println("Err ParseAnglia response get data error")
		return nil, errors.New("No data")
	}
	// for i, v := range res.Steps {
	// 	fmt.Printf("a[%d]=%s\n", i, v)
	// }
	return res, nil
}

//GetAngliaListPageNums pare anglia detail
func GetAngliaListPageNums(ur string) (int, error) {
	baseURL := "http://127.0.0.1:8001/angliapages?keyword="
	u := baseURL + ur
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js one detail page error.")
		return 0, err
	}

	res := PagesResponse{}

	// fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return 0, err
	}

	if res.Status != 0 {
		fmt.Printf("Err:GetAngliaListPageNums  response get data error %d", res.Status)
		return 0, errors.New("No data")
	}
	// for i, v := range res.Steps {
	// 	fmt.Printf("a[%d]=%s\n", i, v)
	// }
	return res.Pages, nil
}

//47.91.138.212:9999/api/v1/getip/1
//GetAngliaListPageNums pare anglia detail
func GetBestIPfromMicroService(ur string) (string, error) {
	suffixURL := ":9999/api/v1/getip/1"
	u := "http://" + ur + suffixURL
	body, err := StanHTTPGetURL(u)
	if err != nil {
		fmt.Println("http get from node.js one detail page error.")
		return "", err
	}

	res := BestIPResponse{}

	// fmt.Printf("http get body: %s \n", string(body))
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		return "", err
	}

	if res.Bestip.Status != 0 {
		fmt.Printf("Err ParseAnglia response get data error %d\n", res.Bestip.Status)
		return "", errors.New("No data")
	}
	// for i, v := range res.Steps {
	// 	fmt.Printf("a[%d]=%s\n", i, v)
	// }
	return res.Bestip.IP, nil
}
