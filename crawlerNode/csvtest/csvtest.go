package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

//PartNumber p
type PartNumber struct {
	Part          string  `csv:"*产品型号"`
	Imag          string  `csv:"图片名称"`
	Promaf        string  `csv:"*厂牌"`
	BaseSalePrice string  `csv:"基础销售价"`
	Stock         int64   `csv:"库存"`
	Price         float64 `csv:"价格"`
}

func initOutWriter(f io.Writer) *csv.Writer {
	return csv.NewWriter(f)
}

func main() {

	// t := time.Now().Format("2006-01-02-15:04:05") + ".csv"
	t := time.Now().Format("2006-01-02-15-04-05")
	// t := time.Now().Unix()
	// fname := strconv.FormatInt(t, 10)

	fname := t + ".csv"
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	initOutWriter(f)
	gocsv.SetCSVWriter(initOutWriter)

	s := []*PartNumber{}

	gocsv.MarshalFile(s, f)

	p1 := &PartNumber{Part: "BTA08-800BWRG", Promaf: "STMicroelectronics", Stock: 46000, Price: 26.666}
	p2 := &PartNumber{Part: "BTA10-600BRG", Promaf: "STMicroelectronics", Stock: 36000, Price: 2.1673}
	// p3 =	{Part: "BTA16-700BRG", Promaf: "STMicroelectronics", Stock: 74000, Price: 0.0501},

	s = nil
	s = append(s, p1)
	gocsv.MarshalWithoutHeaders(s, f)

	s = nil
	s = append(s, p2)
	gocsv.MarshalWithoutHeaders(s, f)
}
