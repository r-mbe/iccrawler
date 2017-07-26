package cockroach

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/stanxii/iccrawler/crawlerSZLC/request"
)

// Although there are many cockroachdb clients with Go, I still want to implement one by myself.
// Because we only need some very simple usages.
type Client struct {
	DB *gorm.DB
	// ctx    context.Context
}

func NewClient(dbURL string) *Client {
	c := new(Client)

	db, err := setupDatabase(dbURL)
	if err != nil {
		log.Fatal("error connecting to the database", err.Error())
		return nil
	}
	c.DB = db

	fmt.Println("now will create table cockpark")
	//if table not exist create table
	hasTable := c.DB.HasTable(&CockPartNumber{})
	fmt.Println("hasTable not hasTable. == .", hasTable)
	if !hasTable {
		fmt.Println("Table not exist. create it.")
		c.DB.CreateTable(&CockPartNumber{})
	} else {
		fmt.Println("Table exist. true it.")
	}

	return c
}

func setupDatabase(dbURL string) (*gorm.DB, error) {

	fmt.Println("XXdbstring path", dbURL)

	// Open connection to server and create a database.
	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Cockroach db open error. Xdbstring path", dbURL)
		return db, err
	}

	// Allow a maximum of concurrency+1 connections to the database.
	db.DB().SetMaxOpenConns(100 + 1)
	db.DB().SetMaxIdleConns(100 + 1)

	return db, nil
}

//CSVPartNumber csvpart
type CockPartNumber struct {
	ID                  int64
	Part                string    `gorm:"column:pro_sno"`
	Cat                 string    `gorm:"column:pro_cat"`
	Imag                string    `gorm:"column:pro_img"`
	Promaf              string    `gorm:"column:pro_maf"`
	BaseSalePrice       float64   `gorm:"column:base_sale_price"`
	Stock               int64     `gorm:"column:stock"`
	SupPart             string    `gorm:"column:suppart"`
	Moq                 int64     `gorm:"column:moq"`
	Mbuy                float64   `gorm:"column:mbuy"`
	Currency            string    `gorm:"column:currency"`
	KuCunDi             string    `gorm:"column:kucundi"`
	CunHuoWeiZhi        string    `gorm:"column:kucunweizi"`
	BasePurchasePrice   float64   `gorm:"column:base_pur_price"`
	PiHao               string    `gorm:"column:pihao"`
	Package             string    `gorm:"column:pkg"`
	Spq                 int64     `gorm:"column:spq"`
	FutureCN            string    `gorm:"column:futurecn"`
	futureHK            string    `gorm:"column:futurehk"`
	Pname               string    `gorm:"column:pname"`
	PurchaseNum1        int64     `gorm:"column:purnum1"`
	PurchaseNum2        int64     `gorm:"column:purnum2"`
	PurchaseNum3        int64     `gorm:"column:purnum3"`
	PurchaseNum4        int64     `gorm:"column:purnum4"`
	PurchaseNum5        int64     `gorm:"column:purnum5"`
	PurchaseNum6        int64     `gorm:"column:purnum6"`
	PurchaseNum7        int64     `gorm:"column:purnum7"`
	PurchaseNum8        int64     `gorm:"column:purnum8"`
	PurchaseNum9        int64     `gorm:"column:purnum9"`
	PurchaseNum10       int64     `gorm:"column:purnum10"`
	SalesUnitPrice1     float64   `gorm:"column:saluprice1"`
	SalesUnitPrice2     float64   `gorm:"column:saluprice2"`
	SalesUnitPrice3     float64   `gorm:"column:saluprice3"`
	SalesUnitPrice4     float64   `gorm:"column:saluprice4"`
	SalesUnitPrice5     float64   `gorm:"column:saluprice5"`
	SalesUnitPrice6     float64   `gorm:"column:saluprice6"`
	SalesUnitPrice7     float64   `gorm:"column:saluprice7"`
	SalesUnitPrice8     float64   `gorm:"column:saluprice8"`
	SalesUnitPrice9     float64   `gorm:"column:saluprice9"`
	SalesUnitPrice10    float64   `gorm:"column:saluprice10"`
	PurchaseUnitPrice1  float64   `gorm:"column:purunit_price1"`
	PurchaseUnitPrice2  float64   `gorm:"column:purunit_price2"`
	PurchaseUnitPrice3  float64   `gorm:"column:purunit_price3"`
	PurchaseUnitPrice4  float64   `gorm:"column:purunit_price4"`
	PurchaseUnitPrice5  float64   `gorm:"column:purunit_price5"`
	PurchaseUnitPrice6  float64   `gorm:"column:purunit_price6"`
	PurchaseUnitPrice7  float64   `gorm:"column:purunit_price7"`
	PurchaseUnitPrice8  float64   `gorm:"column:purunit_price8"`
	PurchaseUnitPrice9  float64   `gorm:"column:purunit_price9"`
	PurchaseUnitPrice10 float64   `gorm:"column:purunit_price10"`
	Comments            string    `gorm:"column:comments"`
	Description         string    `gorm:"column:pro_desc"`
	SaleChannel         string    `gorm:"column:sale_channel"`
	ProductDetail       string    `gorm:"column:pro_detail"`
	Datasheet           string    `gorm:"column:datasheet"`
	Rosh                string    `gorm:"column:rosh"`
	StockBiaoStock      string    `gorm:"column:stock_biao"`
	NotUse1             string    `gorm:"column:not_use1"`
	NotUse2             string    `gorm:"column:not_use2"`
	PartEN              string    `gorm:"column:parten"`
	PromatEN            string    `gorm:"column:promaten"`
	FutureEN            string    `gorm:"column:futureen"`
	DescriptionEN       string    `gorm:"column:descen"`
	LUptime             time.Time `gorm:"column:luptime"`
}

func (c CockPartNumber) TableName() string {
	return "t_pro_sell"
}

func (c *Client) Close() {
	c.DB.Close()
}

func convertReqToCock(in request.PartNumber, o *CockPartNumber) error {

	o.Part = in.Part
	o.Comments = in.Keyword
	o.Promaf = in.Promaf
	o.Stock = in.Stock

	o.Cat = in.Cat
	o.ProductDetail = in.Detail
	o.Package = in.Pkg
	o.Description = in.Desc

	//非数字
	// pattern := `[\\d+$]`
	for i, v := range in.Steps {
		if i < 10 {
			if 0 == i {
				o.PurchaseNum1 = v
			} else if 1 == i {
				o.PurchaseNum2 = v
			} else if 2 == i {
				o.PurchaseNum3 = v
			} else if 3 == i {
				o.PurchaseNum4 = v
			} else if 4 == i {
				o.PurchaseNum5 = v
			} else if 5 == i {
				o.PurchaseNum6 = v
			} else if 6 == i {
				o.PurchaseNum7 = v
			} else if 7 == i {
				o.PurchaseNum8 = v
			} else if 8 == i {
				o.PurchaseNum9 = v
			} else if 9 == i {
				o.PurchaseNum10 = v
			}
		}
	}

	for i, v := range in.Prices {
		if i < 10 {
			if 0 == i {
				o.PurchaseUnitPrice1 = v
			} else if 1 == i {
				o.PurchaseUnitPrice2 = v
			} else if 2 == i {
				o.PurchaseUnitPrice3 = v
			} else if 3 == i {
				o.PurchaseUnitPrice4 = v
			} else if 4 == i {
				o.PurchaseUnitPrice5 = v
			} else if 5 == i {
				o.PurchaseUnitPrice6 = v
			} else if 6 == i {
				o.PurchaseUnitPrice7 = v
			} else if 7 == i {
				o.PurchaseUnitPrice8 = v
			} else if 8 == i {
				o.PurchaseUnitPrice9 = v
			} else if 9 == i {
				o.PurchaseUnitPrice10 = v
			}
		}
	}

	o.LUptime = time.Now()
	return nil
}

func (c *Client) DoSave(p request.PartNumber) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recoverd in DoUpdatePrice")
		}
	}()

	pt := CockPartNumber{}

	convertReqToCock(p, &pt)

	c.DB.Create(&pt)

	//update
	// fmt.Printf("XXXX Update new itme = %+v\n", pt)
	return nil

}

//do stan
