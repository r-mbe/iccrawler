package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"techtoolkit.ickey.cn/cockimport/job"
	"techtoolkit.ickey.cn/cockimport/worker"

	_ "github.com/go-sql-driver/mysql"

	// Import GORM-related packages.
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	// column types values
	columnTypeString = iota
	columnTypeInt
	columnTypeFloat
)

type Mysqlbeat struct {
	done chan struct{}
}

// New Creates beater
func New() *Mysqlbeat {
	return &Mysqlbeat{
		done: make(chan struct{}),
	}
}

var db *sql.DB
var cdb *gorm.DB
var begin time.Time
var total uint64

var isDebug bool

func init() {
	//for mysql
	begin = time.Now()
	total = 0
	isDebug = true

	//mysql
	db = GetConnection()
	//################################################cockroachdb#######################################
	//COCKROACHDB  for postgresql production
	var addr string
	if isDebug {
		fmt.Println("Debuging.. model.......")
		addr = "postgresql://stan:888888@10.8.15.167:26257/db_product?sslmode=disable"
	} else {
		fmt.Println("Production.. model.......")
		addr = "postgresql://stan@10.8.51.69:26257/db_product?sslcert=/usr/local/ickey-certs/client-stan/client.stan.crt&sslkey=/usr/local/ickey-certs/client-stan/client.stan.key&sslrootcert =/usr/local/ickey-certs/client-stan/ca.crt&sslmode=require"
	}

	var err error
	cdb, err = gorm.Open("postgres", addr)
	if err != nil {
		log.Fatal(err)
	}
	cdb.DB().SetMaxOpenConns(50)
}

func main() {
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var jobs = flag.Int("j", 100, "Number of jobs")
	var workers = flag.Int("w", runtime.NumCPU(), "Number of workers")
	var queues = flag.Int("q", 2, "Number of queues")
	// var fail = flag.Bool("f", false, "Fail randomly")
	// var retries = flag.Int("r", 1, "Number of retries for failed jobs")
	var fromdb = flag.String("fromdb", "t_pro_sell_stock", "from mysql")
	var todb = flag.String("todb", "", "to cockroach db name")

	flag.Parse()
	log.SetFlags(0)

	rand.Seed(time.Now().UnixNano())

	// dispatcher := worker.NewDispatcher(*workers)
	// JobQueue := dispatcher.Run(*queues)

	defer db.Close()
	defer cdb.Close()

	bt := New()
	bt.Query(*jobs, *workers, *queues, *fromdb, *todb)

	// //send
	// wg := &sync.WaitGroup{}
	// for i := 0; i < *jobs; i++ {
	// 	wg.Add(1)
	// 	fmt.Printf("Adding job %d to the queue\n", i)
	// 	JobQueue <- job.Job{Name: fmt.Sprintf("%d", i), Wg: wg, Rnd: r, JobQueue: JobQueue, RandomFail: *fail, Retries: *retries}
	// }
	//
	// wg.Wait()
	// dispatcher.Stop()
}

var derr error

func GetConnection() *sql.DB {
	if db != nil {
		fmt.Println("********** CHECKING PING")
		derr = db.Ping()
		if derr == nil {
			fmt.Println("************ CONNECTION STILL ACTIVE")
			return db
		} else {
			fmt.Println("********** PING ERROR: " + derr.Error())
			db.Close()
		}
	}

	if isDebug {
		fmt.Println("Debuging.. model.......")
		db, derr = sql.Open("mysql", "root:ickey_2016@tcp(10.8.11.225:3306)/db_product")
	} else {
		fmt.Println("Production.. model.......")
		db, _ = sql.Open("mysql", "product_repl:37rLWxKzDadxZNQF@tcp(mysqldb.product.rw.01.ickey.cn:3306)/db_product")
	}
	//for product
	if derr != nil {
		panic(derr)
	}

	db.SetConnMaxLifetime(time.Second * 30)

	return db
}

//add by stan
// generateEventFromRow creates a new event from the row data and returns it
func (bt *Mysqlbeat) generateEventFromRow(row *sql.Rows, columns []string, queryType string, rowAge time.Time) (common.MapStr, error) {
	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// Copy the references into such a []interface{} for row.Scan
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Create the event and populate it
	event := common.MapStr{
		"@timestamp": common.Time(rowAge),
		//"type":       queryType,
	}

	// Get RawBytes from data
	err := row.Scan(scanArgs...)
	if err != nil {
		return nil, err
	}

	isINF := false
	// var  jsonstring string
	//jsonstring = "{\"timestamp\": \"" + time.Now().Format("2006-01-02 15:04:05") + "\",\"data\":["
	//allcount :=0
	// Loop on all columns
	for i, col := range values {

		// Get column name and string value
		strColName := string(columns[i])
		strColValue := string(col)
		strColType := columnTypeString

		//stan f1 ftp-simple vsc remote edit.

		// Set the event column name to the original column name (as default)
		strEventColName := strings.ToLower(strColName)

		if strings.EqualFold("pro_sno", strColName) ||
			strings.EqualFold("pro_name", strColName) ||
			strings.EqualFold("img_url", strColName) ||
			strings.EqualFold("pro_url", strColName) ||
			strings.EqualFold("data_sheet", strColName) ||
			strings.EqualFold("remark", strColName) ||
			strings.EqualFold("op_admin_name", strColName) ||
			strings.EqualFold("supplier_category", strColName) ||
			strings.EqualFold("keywords", strColName) ||
			strings.EqualFold("create_time", strColName) ||
			strings.EqualFold("ip", strColName) ||
			strings.EqualFold("source_host", strColName) ||
			strings.EqualFold("source_referrer", strColName) {

			event[strEventColName] = strColValue
		} else {

			// Try to parse the value to an int64
			nColValue, err := strconv.ParseInt(strColValue, 0, 64)
			if err == nil {
				strColType = columnTypeInt
			}

			// Try to parse the value to a float64
			fColValue, err := strconv.ParseFloat(strColValue, 64)
			if err == nil {
				// If it's not already an established int64, set type to float
				if strColType == columnTypeString {
					strColType = columnTypeFloat
				}
			}

			if strColType == columnTypeString {
				// fmt.Println("XXXXXXXXXXXXXXXXXXXXXWhy you guess this is string?????")
				// fmt.Print("\n++++is infh  %v", fColValue)
				// fmt.Println("strColType= %v", strColType)
				//fmt.Printf("##++mysql field| strColname= %s  strcolvalue= %v ,  strcoltype:  %v\n", strColName, strColValue, strColType)
				strEventColName = strColName
				//event[strEventColName] = strColValue
			} else if strColType == columnTypeInt {
				event[strEventColName] = nColValue
			} else if strColType == columnTypeFloat {
				event[strEventColName] = fColValue

				if math.IsInf(fColValue, 0) {
					isINF = true
					fmt.Printf("\n+++is inF strColname=%v col=%v    strColType = %v type %T", strColName, col, strColType, strColType)
					fColValue = 0.0
				}
				if math.IsNaN(fColValue) {
					fmt.Printf("\n++++is Nan  %v  Type: %T\n", fColValue, fColValue)
					fColValue = 0.0
				}

			}
		}
	}
	//jsonstring += "},"

	// If the event has no data, set to nil
	if len(event) == 2 {
		fmt.Println("no event nil event len=3")
		event = nil
	}

	if isINF {
		fmt.Println(event)
		fmt.Println("+++Yeah+++ is inf")
		isINF = false
	}

	return event, nil
}

//!+
type supplier struct {
	Index string
	Type  string
	stmt  string
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

func (bt *Mysqlbeat) Query(jobs int, workers int, queues int, fromdb string, todb string) {

	//####subset sql can 0.29s query 100000  offset 20000
	//select * from t_pro_sell where id >=( select id from t_pro_sell order by id limit 20000 ,1 ) limit 10000;

	//方式1 query

	sups := map[string]*supplier{
		"ickey": &supplier{Index: "ickey", Type: "product", stmt: "SELECT t1.id pid,t1.pro_sno AS p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package,t1. status, t1.pro_name,supplier_id,IF (pro_maf = '',maf_name, pro_maf) AS pro_maf, moq, currency_id, base_buy_amounts, buy_price, base_price, pro_num, t1.img_url, bus_type, pro_desc, t3.sku AS sku, unix_timestamp() AS chrono, md5(trim(t1.pro_sno)) AS pro_sno_md5, md5(trim(pro_sup_sno)) AS pro_sup_sno_md5, md5( 	trim( IF(pro_maf = '', maf_name, pro_maf))) AS pro_maf_md5, t4.cate_id FROM t_pro_sell t1, t_maf t2, t_sku t3, t_product t4 WHERE t1.maf_id = t2.id AND t1.pro_id = t4.id AND t1.id = t3.pro_sell_id AND t1.pro_sup_sno != '' AND bus_type IN (101, 102, 103) AND t1. status = 1"},
		"chip1stop":        &supplier{Index: "chip1stop", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status,  pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=1"},
		"digikey":          &supplier{Index: "digikey", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status,  pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=2"},
		"future":           &supplier{Index: "future", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=3"},
		"wpi":              &supplier{Index: "wpi", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=4"},
		"master":           &supplier{Index: "master", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=7"},
		"microchip":        &supplier{Index: "microchip", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=10"},
		"ps":               &supplier{Index: "ps", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=11"},
		"vicor":            &supplier{Index: "vicor", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=13"},
		"avnet":            &supplier{Index: "avnet", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=14"},
		"rs":               &supplier{Index: "rs", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=22"},
		"element14":        &supplier{Index: "element14", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=26"},
		"element14cn":      &supplier{Index: "element14cn", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=34"},
		"aipco":            &supplier{Index: "aipco", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=27"},
		"rochester":        &supplier{Index: "rochester", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=33"},
		"product":          &supplier{Index: "product", Type: "product", stmt: "select id as pro_id, pro_sno,maf_id,pro_name,cate_id,pro_group,img_url,pro_url,data_sheet,remark,status,op_admin_id,op_admin_name,created_time,last_update_time,is_rohs,supplier_category from t_product"},
		"keywords":         &supplier{Index: "keywords", Type: "product", stmt: "SELECT w_id as id, keywords, uid, create_time, ip, fromid, guestid, status, source_type, source_host, source_referrer from t_keywords"},
		"t_pro_sell_stock": &supplier{Index: "t_pro_sell_stock", Type: "t_pro_sell_stock", stmt: "select sku, stock_num, frozen_num, virtual_num, last_update_time from t_pro_sell_stock order by sku limit %d, %d"},
		"t_pro_sell_price": &supplier{Index: "t_pro_sell_price", Type: "t_pro_sell_price", stmt: "SELECT c.id,c.pro_sell_id,c.sku,c.price_type,c.currency_id,c.number1,c.price1,c.number2,c.price2,c.number3,c.price3,c.number4,c.price4,c.number5,c.price5,c.number6,c.price6,c.number7,c.price7,c.number8,c.price8,c.number9,c.price9,c.number10,c.price10,c.status,c.last_update_time,c.op_admin_id FROM t_pro_sell a LEFT JOIN t_sku b ON a.id = b.pro_sell_id LEFT JOIN t_pro_sell_price c on b.sku=c.sku  where a.`status`=1 AND c.id is not NULL limit %d, %d"},
	}

	// 检查 map xx 对象中是否存在某个 key 索引的元素，如果存在获取该 key 索引的 value
	//*index is ickey_v1 | ickey_v2 | ickey_v3 | ickey_v4
	//check 元素是否存在  *index is a map key.

	b := false

	tables := [...]string{"t_pro_sell_stock", "t_pro_sell_price"}
	//suffexs := [...]string{".es_v", ".pro_v"}
	//suffexs := []string{".es_v", ".pro_v"}

	for _, table := range tables {
		if fromdb == table {
			b = true
			break
		}
	}

	if false == b {
		fmt.Println("Does not supptable river table ", fromdb)
		log.Fatal("err not inclue like *.es_v or *.pro_v ")
	}

	//fmt.Println("supplier index= %v, type=%v, stmt=%v", i, t, stmt)

	// Setup a group of goroutines from the excellent errgroup package
	//docsc := make(chan common.MapStr, 2000)
	// docsc := make(chan common.MapStr, 500)
	//
	// defer close(docsc)
	defer db.Close()

	//storage
	//go Storage(ctx, docsc)
	// for i := 0; i < 100; i++ {
	// 	go Storage(ctx, docsc)
	// }

	offset := 0
	limit := 5000

	//loop function
	var lerr error
	stmt := sups[fromdb].stmt

	sql := fmt.Sprintf(stmt, offset, limit)

	//send cockroachdb cdb
	var docs []common.MapStr
	var finish bool
	docs, finish, lerr = bt.DoMysqlLimitQuery(sql, todb)
	if lerr != nil {
		log.Printf("db op reconnection  err%v", lerr)
		db = GetConnection()
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dispatcher := worker.NewDispatcher(workers)
	JobQueue := dispatcher.Run(queues)
	wg := &sync.WaitGroup{}

	jid := 0

	for i, doc := range docs {
		wg.Add(1)
		jid += i
		fmt.Printf("Adding jo --- first b %d , event: %v to the queue\n", jid, doc)
		// fmt.Printf("Adding jo --- first b %d  to the queue\n", jid)
		JobQueue <- job.Job{Name: fmt.Sprintf("%d", jid), Wg: wg, Event: &doc, Rnd: r, JobQueue: JobQueue, RandomFail: false, Retries: 0}
	}
	wg.Wait()
	// dispatcher.Stop()

	defer dispatcher.Stop()

	if finish {
		fmt.Println("finish mysql done")
		return
	}

	fmt.Println("looping...")
	if len(docs) > 0 {
		for {
			offset += limit

			sql := fmt.Sprintf(stmt, offset, limit)
			//send cockroachdb cdb
			var docs []common.MapStr
			var finish bool

			// fmt.Println("Loopinnggg sql: ", sql)
			docs, finish, lerr = bt.DoMysqlLimitQuery(sql, todb)
			if lerr != nil {
				log.Printf("db op reconnection  err%v", lerr)
				db = GetConnection()
				select {
				case <-time.After(time.Second * 3):
					continue
				}
			}

			for i, doc := range docs {
				wg.Add(1)
				jid += i
				fmt.Printf("Adding ++++ job %d  to the queue\n", jid)
				// fmt.Printf("Adding ++++ job %d  event: %v to the queue\n", jid, doc)
				jb := job.Job{Name: fmt.Sprintf("%d", jid), Wg: wg, Event: &doc, Rnd: r, JobQueue: JobQueue, RandomFail: false, Retries: 1}
				JobQueue <- jb
			}
			wg.Wait()

			if finish {
				fmt.Println("finish mysql done")
				break
			}
		}
	}

	//select * from t_pro_sell where id >=( select id from t_pro_sell order by id limit 5000 ,1 ) limit  5000;
	//select * from t_pro_sell_stock order by sku where last_update_time > ? limit ?,?"
	// Send event over to 2nd goroutine, or chancel

}

func (bt *Mysqlbeat) DoMysqlLimitQuery(sql string, todb string) (res []common.MapStr, finish bool, err error) {
	docs := []common.MapStr{}

	defer func() (res []common.MapStr, finish bool, err error) {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			err, _ = r.(error)

			return docs, false, err
		}
		return docs, false, nil
	}()

	//rows,_ := db.Query("SELECT uid,username FROM USER")
	dtNow := time.Now()

	rows, err := db.Query(sql)
	if err != nil {
		logp.Err("Err mysql Query #%v error generating event from rows:", err)
		return docs, false, err
	}
	// Goroutine to creae documents from mysql

	columns, err := rows.Columns()
	if err != nil {
		logp.Err("Err rows.Columns() #%v error ws:", err)
		return docs, false, err
	}
	defer rows.Close()
	//buf := make([]byte, 32)
	ret := 0
LoopRows:
	for rows.Next() {
		ret++
		//event, err := bt.generateEventFromRow(rows, columns, bt.queryTypes[index], dtNow)
		event, err := bt.generateEventFromRow(rows, columns, "queryMysql", dtNow)

		if err != nil {
			logp.Err("Query #%v generateEventFromRow continue one row event: %v\n", err, event)
			continue LoopRows
		} else if event != nil {
			//b.Events.PublishEvent(event)
			//b.Events.PublishEvent(event)
			//fmt.Println("ok event result:", event)
			//logp.Info("event sent" )
			event["table"] = todb
			if v, ok := event["sku"]; ok {
				if v.(int64) > 0 {
					// docsc <- event
					docs = append(docs, event)
				}
			}
		}

	}

	if ret == 0 {
		//db have no rows. so termnal all.
		fmt.Println("all done in mysql select.")
		return docs, true, nil
	}
	return docs, false, nil

}

func doStorage(db *gorm.DB, d common.MapStr) error {
	defer func() {
		if err := recover(); err != nil {
			// fmt.Println("recoverd in DoStorage. err:", err)
			//log.Info("error: %v", err)
		}
	}()

	// Simple progress
	current := atomic.AddUint64(&total, 1)
	dur := time.Since(begin).Seconds()
	sec := int(dur)
	pps := int64(float64(current) / dur)
	fmt.Printf("%10d | %6d req/s | %02d:%02d\r", current, pps, sec/60, sec%60)

	if _, ok := d["table"]; ok {
		//table name exist

		//for product index use id specia
		if strings.EqualFold(d["table"].(string), "t_pro_sell_stock") {
			//Enqueue the document

			stock := TProSellStock{}
			if v, ok := d["sku"]; ok {
				stock.Sku = v.(int64)
			}
			if v, ok := d["stock_num"]; ok {
				stock.Stocknum = v.(int64)
			}

			if v, ok := d["frozen_num"]; ok {
				stock.Frozennum = v.(int64)
			}
			if v, ok := d["virtual_num"]; ok {
				stock.Virtualnum = v.(int64)
			}
			if v, ok := d["last_update_time"]; ok {
				stock.LastUpTime = v.(int64)
			}
			stock.LUptime = time.Now()

			if db.Where(map[string]interface{}{"sku": stock.Sku}).Find(&TProSellStock{}).RecordNotFound() {
				err := db.Create(&stock)
				if err != nil {
					return err.Error
				}
			}

		} else if strings.EqualFold(d["table"].(string), "t_pro_sell_price") {
			// id = strconv.FormatInt(d["id"].(int64), 10)
			price := TProSellPrice{}

			//mysql save to cockroach mysqlid
			if v, ok := d["id"]; ok {
				price.MysqlID = v.(int64)
			} else {
				fmt.Println("Haa old mysql not exist  not save", d)
				return errors.New("Have no mysqlID")
			}
			if v, ok := d["pro_sell_id"]; ok {
				price.ProSellID = v.(int64)
			}

			if v, ok := d["sku"]; ok {
				price.Sku = v.(int64)
			} else {
				fmt.Println("Haa old mysql not exist  not save", d)
				return errors.New("Have no Sku")
			}
			if v, ok := d["price_type"]; ok {
				price.PriceType = v.(int64)
			}
			if v, ok := d["currency_id"]; ok {
				price.CurrencyID = v.(int64)
			}
			if v, ok := d["number1"]; ok {
				price.Number1 = v.(int64)
			}
			if v, ok := d["price1"]; ok {
				price.Price1 = v.(float64)
			}
			if v, ok := d["number2"]; ok {
				price.Number2 = v.(int64)
			}
			if v, ok := d["price2"]; ok {
				price.Price2 = v.(float64)
			}
			if v, ok := d["number3"]; ok {
				price.Number3 = v.(int64)
			}
			if v, ok := d["price3"]; ok {
				price.Price3 = v.(float64)
			}
			if v, ok := d["number4"]; ok {
				price.Number4 = v.(int64)
			}
			if v, ok := d["price4"]; ok {
				price.Price4 = v.(float64)
			}
			if v, ok := d["number5"]; ok {
				price.Number5 = v.(int64)
			}
			if v, ok := d["price5"]; ok {
				price.Price5 = v.(float64)
			}
			if v, ok := d["number6"]; ok {
				price.Number6 = v.(int64)
			}
			if v, ok := d["price6"]; ok {
				price.Price6 = v.(float64)
			}
			if v, ok := d["number7"]; ok {
				price.Number7 = v.(int64)
			}
			if v, ok := d["price7"]; ok {
				price.Price7 = v.(float64)
			}
			if v, ok := d["number8"]; ok {
				price.Number8 = v.(int64)
			}
			if v, ok := d["price8"]; ok {
				price.Price8 = v.(float64)
			}
			if v, ok := d["number9"]; ok {
				price.Number9 = v.(int64)
			}
			if v, ok := d["price9"]; ok {
				price.Price9 = v.(float64)
			}
			if v, ok := d["number10"]; ok {
				price.Number10 = v.(int64)
			}
			if v, ok := d["price10"]; ok {
				price.Price10 = v.(float64)
			}
			if v, ok := d["status"]; ok {
				price.Status = v.(int64)
			}
			if v, ok := d["last_update_time"]; ok {
				price.LastUpTime = v.(int64)
			}
			if v, ok := d["op_admin_id"]; ok {
				price.OpAdminID = v.(int64)
			}

			price.LUptime = time.Now()

			if db.Where(map[string]interface{}{"mysql_id": price.MysqlID}).Find(&TProSellPrice{}).RecordNotFound() {
				// fmt.Println("Haa ###2233  mysql not exist  not old mysql", price)
				err := db.Create(&price)
				if err != nil {
					fmt.Printf("Err busy buff####XXXXX 333333 create price err: %v  , cock price: %v    old mysql: %v \n", err.Error.Error(), price, d)
					return err.Error
				}
			} else {
				// log .Println("have alread exist mysqlid", price.MysqlID)
				logp.Err("Err rows.Columns() #%v error ws:", price.MysqlID)
			}

		} else {
			fmt.Println("un support tables.")
		}
	}
	return nil
}

//ContextSotrage child storage goroutine
func ContextSotrage(db *gorm.DB, ctx context.Context, v common.MapStr) error {
	ctx, cCancel := context.WithCancel(ctx)
	err := doStorage(db, v)

	cCancel()
	select {
	case <-ctx.Done():
		return err
	}

}

//Storage storage main goroutine
func Storage(db *gorm.DB, ctx context.Context, in <-chan common.MapStr) error {
	// Second goroutine will consume the documents sent from the first an bulk insert into cockroach

	fmt.Println("Storage goroutine ..working....")

	for {
		select {
		case <-ctx.Done():
			//finish all then return
			fmt.Println("recieve all chan storage....")
			return ctx.Err()
		case v, ok := <-in:
			//do resualt.\
			if ok {
				// fmt.Printf("receive  one chan storage %v.....\n", v)
				//must sync processing...

				//child ctx and ctrx.
				go ContextSotrage(db, ctx, v)

				// doStorage(v)
				// ContextSotrage(ctx, v)

			}

			// else {
			// 	fmt.Println("recieve all chan storage....")
			// 	return errors.New("finish")
			// }
		}
	}

}
