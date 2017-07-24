package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
)

//Test crawler detail page from node.js
func TestGetBestIPfromMicroService(t *testing.T) {
	//ip := "47.91.138.212"
	ip := "127.0.0.1"
	data, err := GetBestIPfromMicroService(ip)
	if err != nil {
		fmt.Printf("ParseAnglia error:%s from nodejs err: %v\n", ip, err)
	}

	fmt.Printf("http get body: %v \n", data)

}

//test for request http.
func TestStanHTTPGetURL(t *testing.T) {
	u := "http://www.anglia-live.com/products/batteries-accessories/batteries/ic-batteries-memory-backup"
	data, err := StanHTTPGetURL(u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("http get body: %s \n", string(data))

}

//StanHTTPPostURL
func TestStanHTTPPostURL(t *testing.T) {
	u := "http://47.91.138.212:8001/anglialist"
	k := "http://www.anglia-live.com/products/circuit-protection/fuses-breakers-holders/resettable-fuses#resultspage,4|"
	var jsonrep string = `{"keyword":"` + k + `"}`
	jsonStr := []byte(jsonrep)
	body, err := StanHTTPPostURL(u, jsonStr)
	if err != nil {
		fmt.Printf("I go err. %v\n", err)
		t.Fatal(err)
	}

	fmt.Printf("http post res body: %s \n", string(body))

	res := AngliaPostResponse{}

	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Err:json.Unmarshal err:%v.", err)
		t.Fatal(err)
	}

	if res.Status != 0 {
		fmt.Println("Err Parse AngliaList  Post response status = 0")
		t.Fatal(errors.New("res.status error"))
	}

	fmt.Printf("Post get responsed  data xxxx= %d\n", len(res.Sups))
	for i, v := range res.Sups {
		fmt.Printf("a[%d]=%s\n", i, v)
	}

	fmt.Println("Yes it is ok.")
}

//Test crawler list page from node.js
func TestParseAngliaList(t *testing.T) {
	u := "http://www.anglia-live.com/products/batteries-accessories/batteries/ic-batteries-memory-backup"
	data, err := ParseAngliaList(u)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("http get body: %d \n", len(data))

}

//TestGetAngliaListPageNums crawler detail page from node.js
func TestGetAngliaListPageNums(t *testing.T) {
	u := "http://www.anglia-live.com/products/circuit-protection/suppression-emi-filters/x-y-safety-capacitors"
	pages, err := GetAngliaListPageNums(u)
	if err != nil {
		fmt.Printf("TestGetAngliaListPageNums url:%s from nodejs err: %v\n", u, err)
	}

	fmt.Printf("get pages pURL:%s,  pages:%d, err: %v \n", u, pages, err)

}

//Test crawler detail page from node.js
func TestParseAnglia(t *testing.T) {
	u := "http://www.anglia-live.com/products/sensors-transducers/encoders/contactless/116332001_enc1j-optical-enc-128-pulses"
	data, err := ParseAnglia(u)
	if err != nil {
		fmt.Printf("ParseAnglia error:%s from nodejs err: %v\n", u, err)
	}

	fmt.Printf("http get body: %v \n", data)

}
