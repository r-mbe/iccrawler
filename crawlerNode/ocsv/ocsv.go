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
	Name string
	f    *os.File
	w    *csv.Writer
}

//CSVPartNumber csvpart
type CSVPartNumber struct {
	Part                string  `csv:"*产品型号"`
	Imag                string  `csv:"图片名称"`
	Promaf              string  `csv:"*厂牌"`
	BaseSalePrice       float64 `csv:"基础销售价"`
	Stock               int64   `csv:"库存"`
	SupPart             string  `csv:"供应商型号"`
	Moq                 int64   `csv:"最小起订量"`
	Mbuy                float64 `csv:"最小购买金额"`
	Currency            string  `csv:"币种"`
	KuCunDi             string  `csv:"库存地"`
	CunHuoWeiZhi        string  `csv:"存货位置"`
	BasePurchasePrice   float64 `csv:"基础采购价"`
	PiHao               string  `csv:"批号"`
	Package             string  `csv:"封装"`
	Spq                 int64   `csv:"标准包装数"`
	FutureCN            string  `csv:"货期(国内)"`
	futureHK            string  `csv:"货期(香港)"`
	Pname               string  `csv:"产品名称"`
	PurchaseNum1        int64   `csv:"购买数量1"`
	PurchaseNum2        int64   `csv:"购买数量2"`
	PurchaseNum3        int64   `csv:"购买数量3"`
	PurchaseNum4        int64   `csv:"购买数量4"`
	PurchaseNum5        int64   `csv:"购买数量5"`
	PurchaseNum6        int64   `csv:"购买数量6"`
	PurchaseNum7        int64   `csv:"购买数量7"`
	PurchaseNum8        int64   `csv:"购买数量8"`
	PurchaseNum9        int64   `csv:"购买数量9"`
	PurchaseNum10       int64   `csv:"购买数量10"`
	SalesUnitPrice1     float64 `csv:"销售单价4"`
	SalesUnitPrice2     float64 `csv:"销售单价3"`
	SalesUnitPrice3     float64 `csv:"销售单价1"`
	SalesUnitPrice4     float64 `csv:"销售单价2"`
	SalesUnitPrice5     float64 `csv:"销售单价5"`
	SalesUnitPrice6     float64 `csv:"销售单价6"`
	SalesUnitPrice7     float64 `csv:"销售单价7"`
	SalesUnitPrice8     float64 `csv:"销售单价8"`
	SalesUnitPrice9     float64 `csv:"销售单价9"`
	SalesUnitPrice10    float64 `csv:"销售单价10"`
	PurchaseUnitPrice1  float64 `csv:"采购单价1"`
	PurchaseUnitPrice2  float64 `csv:"采购单价2"`
	PurchaseUnitPrice3  float64 `csv:"采购单价3"`
	PurchaseUnitPrice4  float64 `csv:"采购单价4"`
	PurchaseUnitPrice5  float64 `csv:"采购单价5"`
	PurchaseUnitPrice6  float64 `csv:"采购单价6"`
	PurchaseUnitPrice7  float64 `csv:"采购单价7"`
	PurchaseUnitPrice8  float64 `csv:"采购单价8"`
	PurchaseUnitPrice9  float64 `csv:"采购单价9"`
	PurchaseUnitPrice10 float64 `csv:"采购单价10"`
	Comments            string  `csv:"备注"`
	Description         string  `csv:"描述"`
	SaleChannel         string  `csv:"销售渠道"`
	ProductDetail       string  `csv:"商品详情"`
	Datasheet           string  `csv:"DataSheet"`
	Rosh                string  `csv:"是否通过Rohs认证"`
	StockBiaoStock      string  `csv:"库存表库存"`
	NotUse1             string  `csv:"Not Use1"`
	NotUse2             string  `csv:"Not Use2"`
	PartEN              string  `csv:"产品型号(英文)"`
	PromatEN            string  `csv:"厂牌(英文)"`
	FutureEN            string  `csv:"货期(英文)"`
	DescriptionEN       string  `csv:"描述(英文)"`
}

// type PartNumber struct {
// 	Part          string  `csv:"*产品型号"`
// 	Imag          string  `csv:"图片名称"`
// 	Promaf        string  `csv:"*厂牌"`
// 	BaseSalePrice string  `csv:"基础销售价"`
// 	Stock         int64   `csv:"库存"`
// 	Price         float64 `csv:"价格"`
// }

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

//NewOcsv new
func NewOcsv() *Ocsv {

	o := &Ocsv{}
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
	c.Name = fname
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// defer f.Close()
	c.f = f
	c.w = initOutWriter(f)

	gocsv.SetCSVWriter(initOutWriter)

	s := []CSVPartNumber{}

	//set header for csv file
	gocsv.MarshalFile(s, f)

	return err
}

//Append append a data into
func (c *Ocsv) Append(in interface{}) error {
	v, ok := in.(*CSVPartNumber)
	if !ok {
		fmt.Println("input == err.", ok)
		return errors.New("Err csvPart error")
	}

	fmt.Println("Now CSv... saving. will append csv part:", v)
	//convert part to csv parnumber
	s := []*CSVPartNumber{}
	s = nil
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
