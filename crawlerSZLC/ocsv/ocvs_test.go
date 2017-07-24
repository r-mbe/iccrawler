package ocsv

import (
	"fmt"
	"log"
	"testing"
)

func TestWritetoFile(t *testing.T) {
	s := []CSVPartNumber{
		{Part: "BTA08-800BWRG", Promaf: "STMicroelectronics", Stock: 46000, BasePurchasePrice: 26.666},
		{Part: "BTA10-600BRG", Promaf: "STMicroelectronics", Stock: 36000, BasePurchasePrice: 2.1673},
		{Part: "BTA16-700BRG", Promaf: "STMicroelectronics", Stock: 74000, BasePurchasePrice: 0.0501},
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
