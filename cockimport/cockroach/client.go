package cockroach

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	mainlog "github.com/cihub/seelog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/juju/errors"
)

// Although there are many cockroachdb clients with Go, I still want to implement one by myself.
// Because we only need some very simple usages.
type Client struct {
	CockDBName   string
	CockAddr     string
	CockUser     string
	CockPassword string

	Seelog *mainlog.LoggerInterface
	DB     *gorm.DB
	// ctx    context.Context
}

type ClientConfig struct {
	CockDBName   string
	CockAddr     string
	CockUser     string
	CockPassword string
	SeelogFile   string
}

func NewClient(conf *ClientConfig) *Client {
	c := new(Client)

	c.CockDBName = conf.CockDBName
	c.CockAddr = conf.CockAddr
	c.CockUser = conf.CockUser
	c.CockPassword = conf.CockPassword

	Seelog, _ := mainlog.LoggerFromConfigAsFile(conf.SeelogFile)
	c.Seelog = &Seelog

	conn := "postgresql://" + c.CockUser + ":" + c.CockPassword + "@" + c.CockAddr + "/" + c.CockDBName + "?sslmode=disable"

	// c.DB, err := sql.Open("postgres", "postgresql://stan:888888@localhost:26257/bank?sslmode=disable")

	DB, err := gorm.Open("postgres", conn)
	c.DB = DB
	if err != nil {
		log.Fatal("error connecting to the database", err.Error())
		Seelog.Errorf("error connection cockroachdb error: %v", err)
		return nil
	}

	return c
}

type ResponseItem struct {
	ID      string                 `json:"_id"`
	Index   string                 `json:"_index"`
	Type    string                 `json:"_type"`
	Version int                    `json:"_version"`
	Found   bool                   `json:"found"`
	Source  map[string]interface{} `json:"_source"`
}

type Response struct {
	Code int
	ResponseItem
}

// See http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/bulk.html
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionIndex  = "index"
)

type BulkRequest struct {
	Action string
	Index  string
	Type   string
	ID     string
	Parent string

	Data map[string]interface{}
}

func (c *Client) CockroachBulk(r *BulkRequest) error {

	switch r.Action {
	case ActionDelete:
		//nothing to do
		if strings.EqualFold(r.Type, "t_pro_sell_stock") {
			c.DoDeleteStock(r.Data)
		} else if strings.EqualFold(r.Type, "t_pro_sell_price") {
			c.DoDeletePrice(r.Data)
		} else {
			fmt.Printf("client.go:CockroachBulk unknow:%v ActionDelete \n", r.Type)
		}
	case ActionUpdate:
		//update db
		if strings.EqualFold(r.Type, "t_pro_sell_stock") {
			c.DoUpdateStock(r.Data)
		} else if strings.EqualFold(r.Type, "t_pro_sell_price") {
			c.DoUpdatePrice(r.Data)
		} else {
			fmt.Printf("client.go:CockroachBulk unknow: %v ActionUpdate \n", r.Type)
		}

	default:
		//for create and index
		fmt.Println("create new one")

		if strings.EqualFold(r.Type, "t_pro_sell_stock") {
			c.DoCreateStock(r.Data)
		} else if strings.EqualFold(r.Type, "t_pro_sell_price") {
			c.DoUpdatePrice(r.Data)
		} else {
			fmt.Printf("client.go:CockroachBulk unknow:%v ActionDefault \n", r.Type)
		}
	}

	return nil
}

type BulkResponse struct {
	Code   int
	Took   int  `json:"took"`
	Errors bool `json:"errors"`

	Items []map[string]*BulkResponseItem `json:"items"`
}

type BulkResponseItem struct {
	Index   string          `json:"_index"`
	Type    string          `json:"_type"`
	ID      string          `json:"_id"`
	Version int             `json:"_version"`
	Status  int             `json:"status"`
	Error   json.RawMessage `json:"error"`
	Found   bool            `json:"found"`
}

type TProSellStock struct {
	// gorm.Model
	Sku        int64
	Stocknum   int64     `gorm:"column:stock_num"`
	Frozennum  int64     `gorm:"column:frozen_num"`
	Virtualnum int64     `gorm:"column:virtual_num"`
	LastUpTime int64     `gorm:"column:last_update_time"`
	LUptime    time.Time `gorm:"column:luptime"`
}

type TProSellPrice struct {
	// gorm.Model
	MysqlID    int64     `gorm:"column:mysql_id"`
	ProSellID  int64     `gorm:"column:pro_sell_id"`
	Sku        int64     `gorm:"column:sku"`
	PriceType  int64     `gorm:"column:price_type"`
	CurrencyID int64     `gorm:"column:currency_id"`
	Price1     float64   `gorm:"type:decimal(15,6);"`
	Number1    int64     `gorm:"column:number1"`
	Price2     float64   `gorm:"type:decimal(15,6);"`
	Number2    int64     `gorm:"column:number2"`
	Price3     float64   `gorm:"type:decimal(15,6);"`
	Number3    int64     `gorm:"column:number3"`
	Price4     float64   `gorm:"type:decimal(15,6);"`
	Number4    int64     `gorm:"column:number4"`
	Price5     float64   `gorm:"type:decimal(15,6);"`
	Number5    int64     `gorm:"column:number5"`
	Price6     float64   `gorm:"type:decimal(15,6);"`
	Number6    int64     `gorm:"column:number6"`
	Price7     float64   `gorm:"type:decimal(15,6);"`
	Number7    int64     `gorm:"column:number7"`
	Price8     float64   `gorm:"type:decimal(15,6);"`
	Number8    int64     `gorm:"column:number8"`
	Price9     float64   `gorm:"type:decimal(15,6);"`
	Number9    int64     `gorm:"column:number9"`
	Price10    float64   `gorm:"type:decimal(15,6);"`
	Number10   int64     `gorm:"column:number10"`
	Status     int64     `gorm:"column:status"`
	LastUpTime int64     `gorm:"column:last_update_time"`
	OpAdminID  int64     `gorm:"column:op_admin_id"`
	LUptime    time.Time `gorm:"column:luptime"`
}

//CSVPartNumber csvpart
type CockPartNumber struct {
	ID                  int64
	Part                string  `gorm:"column:pro_sno"`
	Cat                 string  `gorm:"column:pro_cat"`
	Imag                string  `gorm:"column:pro_img"`
	Promaf              string  `gorm:"column:pro_maf"`
	BaseSalePrice       float64 `gorm:"column:base_sale_price"`
	Stock               int64   `gorm:"column:stock"`
	SupPart             string  `gorm:"column:suppart"`
	Moq                 int64   `gorm:"column:moq"`
	Mbuy                float64 `gorm:"column:mbuy"`
	Currency            string  `gorm:"column:currency"`
	KuCunDi             string  `gorm:"column:kucundi"`
	CunHuoWeiZhi        string  `gorm:"column:kucunweizi"`
	BasePurchasePrice   float64 `gorm:"column:base_pur_price"`
	PiHao               string  `gorm:"column:pihao"`
	Package             string  `gorm:"column:pkg"`
	Spq                 int64   `gorm:"column:spq"`
	FutureCN            string  `gorm:"column:futurecn"`
	futureHK            string  `gorm:"column:futurehk"`
	Pname               string  `gorm:"column:pname"`
	PurchaseNum1        int64   `gorm:"column:purnum1"`
	PurchaseNum2        int64   `gorm:"column:purnum2"`
	PurchaseNum3        int64   `gorm:"column:purnum3"`
	PurchaseNum4        int64   `gorm:"column:purnum4"`
	PurchaseNum5        int64   `gorm:"column:purnum5"`
	PurchaseNum6        int64   `gorm:"column:purnum6"`
	PurchaseNum7        int64   `gorm:"column:purnum7"`
	PurchaseNum8        int64   `gorm:"column:purnum8"`
	PurchaseNum9        int64   `gorm:"column:purnum9"`
	PurchaseNum10       int64   `gorm:"column:purnum10"`
	SalesUnitPrice1     float64 `gorm:"column:saluprice1"`
	SalesUnitPrice2     float64 `gorm:"column:saluprice2"`
	SalesUnitPrice3     float64 `gorm:"column:saluprice3"`
	SalesUnitPrice4     float64 `gorm:"column:saluprice4"`
	SalesUnitPrice5     float64 `gorm:"column:saluprice5"`
	SalesUnitPrice6     float64 `gorm:"column:saluprice6"`
	SalesUnitPrice7     float64 `gorm:"column:saluprice7"`
	SalesUnitPrice8     float64 `gorm:"column:saluprice8"`
	SalesUnitPrice9     float64 `gorm:"column:saluprice9"`
	SalesUnitPrice10    float64 `gorm:"column:saluprice10"`
	PurchaseUnitPrice1  float64 `gorm:"column:purunit_price1"`
	PurchaseUnitPrice2  float64 `gorm:"column:purunit_price2"`
	PurchaseUnitPrice3  float64 `gorm:"column:purunit_price3"`
	PurchaseUnitPrice4  float64 `gorm:"column:purunit_price4"`
	PurchaseUnitPrice5  float64 `gorm:"column:purunit_price5"`
	PurchaseUnitPrice6  float64 `gorm:"column:purunit_price6"`
	PurchaseUnitPrice7  float64 `gorm:"column:purunit_price7"`
	PurchaseUnitPrice8  float64 `gorm:"column:purunit_price8"`
	PurchaseUnitPrice9  float64 `gorm:"column:purunit_price9"`
	PurchaseUnitPrice10 float64 `gorm:"column:purunit_price10"`
	Comments            string  `gorm:"column:comments"`
	Description         string  `gorm:"column:pro_desc"`
	SaleChannel         string  `gorm:"column:sale_channel"`
	ProductDetail       string  `gorm:"column:pro_detail"`
	Datasheet           string  `gorm:"column:datasheet"`
	Rosh                string  `gorm:"column:rosh"`
	StockBiaoStock      string  `gorm:"column:stock_biao"`
	NotUse1             string  `gorm:"column:not_use1"`
	NotUse2             string  `gorm:"column:not_use2"`
	PartEN              string  `gorm:"column:parten"`
	PromatEN            string  `gorm:"column:promaten"`
	FutureEN            string  `gorm:"column:futureen"`
	DescriptionEN       string  `gorm:"column:descen"`
}

func (c *Client) Close() {
	c.DB.Close()
}

func (c *Client) DoUpdateStock(body map[string]interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recoverd in DoUpdateStock")
			seelog := *c.Seelog
			seelog.Errorf("db find error %v", err)
		}
	}()

	var oldstock TProSellStock
	var stock TProSellStock

	fmt.Printf("change update DoUpdateStock stock body: %v \n", body)
	if v, ok := body["sku"]; ok {
		stock.Sku = v.(int64)
	}
	if _, ok := body["stock_num"]; ok {
		x := body["stock_num"]
		var v1 = reflect.TypeOf(body["stock_num"])
		fmt.Printf("stock num type val = %v type=%T\n", x, x)
		fmt.Printf("stock num type val = %v type=%T\n", v1, v1)

		stock.Stocknum = body["stock_num"].(int64)
	}
	if _, ok := body["frozen_num"]; ok {
		stock.Frozennum = body["frozen_num"].(int64)
	}
	if _, ok := body["virtual_num"]; ok {
		stock.Virtualnum = body["virtual_num"].(int64)
	}
	stock.LUptime = time.Now()

	if c.DB.Where(map[string]interface{}{"sku": stock.Sku}).First(&oldstock).RecordNotFound() {
		//create new one
		fmt.Printf("XXXX NNNNot found crate NWWW found  stock= %+v\n", stock)

		c.DB.Create(&stock)
		return nil
	}
	//update
	fmt.Printf("XXXX Update oldstock= %+v\n", oldstock)
	fmt.Printf("XXXX Update new stock= %+v\n", stock)
	c.DB.Where(map[string]interface{}{"sku": stock.Sku}).Updates(stock)

	return nil

}

func (c *Client) DoUpdatePrice(body map[string]interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recoverd in DoUpdatePrice")
			seelog := *c.Seelog
			seelog.Info("db find error %v", err)
		}
	}()

	var oldPrice TProSellPrice
	var price TProSellPrice

	if v, ok := body["id"]; ok {
		price.MysqlID = v.(int64)
	} else {
		fmt.Println("Err .. have no mysql id")
		return errors.New("no mysql id")
	}

	if v, ok := body["sku"]; ok {
		price.Sku = v.(int64)
	}
	if v, ok := body["price1"]; ok {
		price.Price1 = v.(float64)
	}
	if v, ok := body["number1"]; ok {
		price.Number1 = v.(int64)
	}
	price.LUptime = time.Now()

	if c.DB.Where(map[string]interface{}{"mysql": price.Sku}).First(&oldPrice).RecordNotFound() {
		c.DB.Model(&oldPrice).Update(price)
	} else {
		fmt.Println("data alread exist is db")
	}

	if c.DB.Where(map[string]interface{}{"mysql_id": price.MysqlID}).First(&oldPrice).RecordNotFound() {
		//create new one
		fmt.Printf("XXXX NNNNot found crate Price found  new price= %+v\n", price)

		c.DB.Create(&price)
		return nil
	}
	//update
	fmt.Printf("XXXX Update oldPrice = %+v\n", oldPrice)
	fmt.Printf("XXXX Update new price = %+v\n", price)
	c.DB.Where(map[string]interface{}{"mysql_id": price.MysqlID}).Updates(price)
	return nil

}
