package ocsv

import (
	"fmt"
	"log"
	"testing"
)

func TestWritetoFile(t *testing.T) {
	s := []CSVPassivePTNumber{
		{Cat: "MELF 电阻", Part: "OMM02040000000B300", Description: "0Ω(0000) 编带", Package: "SMM0204", Promaf: "VISHAY(威世)", PanPian: 3000, RecentSell: 220, Stock: 2780},
		{Cat: "MELF 电阻", Part: "SMM02040C1009FB300", Description: "10Ω(10R0) ±1% 编带", Package: "SMM0204", Promaf: "VISHAY(威世)", PanPian: 3000, RecentSell: 220, Stock: 2490},
		{Cat: "MELF 电阻", Part: "SMM02040C1002FB300", Description: "10Ω(10R0) ±1% 编带", Package: "SMM0204", Promaf: "VISHAY(威世)", PanPian: 3000, RecentSell: 220, Stock: 2340},
	}

	csv := NewOcsv()
	err := csv.Init()
	if err != nil {
		fmt.Println("init err: ", err)
		log.Fatal(err)
	}
	defer csv.Close()

	for _, v := range s {
		csv.Append(&v)
	}

	fmt.Println("s=", s)
}
