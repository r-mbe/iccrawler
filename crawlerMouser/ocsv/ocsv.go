package ocsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

//Ocsv module
type Ocsv struct {
	name string
	f    *os.File
	w    *csv.Writer
}

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

type CSVPassivePTNumber struct {
	Cat            string `csv:"*产品类别"`
	Part           string `csv:"*产品型号"`
	Description    string `csv:"描述"`
	Package        string `csv:"封装/规格"`
	ICode          string `csv:"编号"`
	Promaf         string `csv:"*品牌"`
	SalesUnitPrice int64  `csv:"销售单价"`
	PanPian        int64  `csv:"1圆盘有*片"`
	RecentSell     int64  `csv:"近期约售"`
	Stock          int64  `csv:"现货库存"`
}

//NewOcsv new
func NewOcsv() *Ocsv {

	o := &Ocsv{
		name: "ocsv"}
	return o
}

func initOutWriter(f io.Writer) *csv.Writer {
	return csv.NewWriter(f)
}

//Init ocsv
func (c *Ocsv) Init() error {
	t := time.Now().Format("2006-01-02-15-04-05")
	// t := time.Now().Unix()
	// fname := strconv.FormatInt(t, 10)

	fname := t + ".csv"
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// defer f.Close()
	c.f = f
	c.w = initOutWriter(f)

	gocsv.SetCSVWriter(initOutWriter)

	fmt.Println("init ocsv....xxxx1")

	s := []CSVPassivePTNumber{}

	//set header for csv file
	gocsv.MarshalFile(s, f)

	fmt.Println("csv init ok.....")
	return err
}

//Append append a data into
func (c *Ocsv) Append(in interface{}) error {
	v, ok := in.(*CSVPassivePTNumber)
	if !ok {
		fmt.Println("input == err.", ok)
		return errors.New("Err csvPart error")
	}

	fmt.Println("Now CSv... saving. will append csv part:", v, c.f)
	//convert part to csv parnumber
	s := []*CSVPassivePTNumber{}
	s = append(s, v)
	err := gocsv.MarshalWithoutHeaders(s, c.f)
	if err != nil {
		fmt.Println("save csv without header error.", err)
	}

	return err
}

//Close oscv close
func (c *Ocsv) Close() {
	defer c.f.Close()

}
