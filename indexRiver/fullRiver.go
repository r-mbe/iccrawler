package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/olivere/elastic.v5"
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

var db = &sql.DB{}

func init() {

	//for test
	db, _ = sql.Open("mysql", "root:ickey_2016@tcp(10.8.11.225:3306)/db_product")

	//pre
	// db, _ = sql.Open("mysql", "xiyx:xiyx@2017@tcp(10.8.51.33:3306)/db_product")

	//for production master
	//db, _ = sql.Open("mysql", "dbproduct_master:RdEa211e8HPnK2zM@tcp(mysqldb.product.rw.01.ickey.cn:3306)/db_product")
}

func main() {
	//    insert()
	bt := New()
	query(bt)
	//    update()
	//    query()
	//    delete()
}

func update() {
	//方式1 update
	start := time.Now()
	for i := 1001; i <= 1100; i++ {
		db.Exec("UPdate user set age=? where uid=? ", i, i)
	}
	end := time.Now()
	fmt.Println("方式1 update total time:", end.Sub(start).Seconds())

	//方式2 update
	start = time.Now()
	for i := 1101; i <= 1200; i++ {
		stm, _ := db.Prepare("UPdate user set age=? where uid=? ")
		stm.Exec(i, i)
		stm.Close()
	}
	end = time.Now()
	fmt.Println("方式2 update total time:", end.Sub(start).Seconds())

	//方式3 update
	start = time.Now()
	stm, _ := db.Prepare("UPdate user set age=? where uid=?")
	for i := 1201; i <= 1300; i++ {
		stm.Exec(i, i)
	}
	stm.Close()
	end = time.Now()
	fmt.Println("方式3 update total time:", end.Sub(start).Seconds())

	//方式4 update
	start = time.Now()
	tx, _ := db.Begin()
	for i := 1301; i <= 1400; i++ {
		tx.Exec("UPdate user set age=? where uid=?", i, i)
	}
	tx.Commit()

	end = time.Now()
	fmt.Println("方式4 update total time:", end.Sub(start).Seconds())

	//方式5 update
	start = time.Now()
	for i := 1401; i <= 1500; i++ {
		tx, _ := db.Begin()
		tx.Exec("UPdate user set age=? where uid=?", i, i)
		tx.Commit()
	}
	end = time.Now()
	fmt.Println("方式5 update total time:", end.Sub(start).Seconds())

}

func delete() {
	//方式1 delete
	start := time.Now()
	for i := 1001; i <= 1100; i++ {
		db.Exec("DELETE FROM USER WHERE uid=?", i)
	}
	end := time.Now()
	fmt.Println("方式1 delete total time:", end.Sub(start).Seconds())

	//方式2 delete
	start = time.Now()
	for i := 1101; i <= 1200; i++ {
		stm, _ := db.Prepare("DELETE FROM USER WHERE uid=?")
		stm.Exec(i)
		stm.Close()
	}
	end = time.Now()
	fmt.Println("方式2 delete total time:", end.Sub(start).Seconds())

	//方式3 delete
	start = time.Now()
	stm, _ := db.Prepare("DELETE FROM USER WHERE uid=?")
	for i := 1201; i <= 1300; i++ {
		stm.Exec(i)
	}
	stm.Close()
	end = time.Now()
	fmt.Println("方式3 delete total time:", end.Sub(start).Seconds())

	//方式4 delete
	start = time.Now()
	tx, _ := db.Begin()
	for i := 1301; i <= 1400; i++ {
		tx.Exec("DELETE FROM USER WHERE uid=?", i)
	}
	tx.Commit()

	end = time.Now()
	fmt.Println("方式4 delete total time:", end.Sub(start).Seconds())

	//方式5 delete
	start = time.Now()
	for i := 1401; i <= 1500; i++ {
		tx, _ := db.Begin()
		tx.Exec("DELETE FROM USER WHERE uid=?", i)
		tx.Commit()
	}
	end = time.Now()
	fmt.Println("方式5 delete total time:", end.Sub(start).Seconds())

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

		if strings.EqualFold("pro_sno", strColName) || strings.EqualFold("p", strColName) || strings.EqualFold("pro_sup_sno", strColName) || strings.EqualFold("date_code", strColName) || strings.EqualFold("lead_time_cn", strColName) || strings.EqualFold("lead_time_hk", strColName) || strings.EqualFold("package", strColName) || strings.EqualFold("pro_name", strColName) || strings.EqualFold("pro_maf", strColName) || strings.EqualFold("img_url", strColName) || strings.EqualFold("pro_desc", strColName) || strings.EqualFold("keywords", strColName) || strings.EqualFold("chrono", strColName) || strings.EqualFold("pro_sno_md5", strColName) || strings.EqualFold("pro_sup_sno_md5", strColName) || strings.EqualFold("pro_maf_md5", strColName) {

			if strings.EqualFold(strEventColName, "p") || strings.EqualFold(strEventColName, "pro_sup_sno") || strings.EqualFold(strEventColName, "pro_name") || strings.EqualFold(strEventColName, "pro_maf") {
				strColValue = strings.ToUpper(strColValue)
			}
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
				fmt.Printf("strColname= %s  strcolvalue= %v ,  strcoltype:  %v", strColName, strColValue, strColType)
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

//!-

func query(bt *Mysqlbeat) {

	//####subset sql can 0.29s query 100000  offset 20000
	//select * from t_pro_sell where id >=( select id from t_pro_sell order by id limit 20000 ,1 ) limit 10000;

	//方式1 query
	start := time.Now()
	dtNow := start

	var (
		url      = flag.String("url", "http://10.8.15.168:9200", "Elasticsearch URL")
		index    = flag.String("index", "", "Elasticsearch index name")
		typ      = flag.String("type", "", "Elasticsearch type name")
		sniff    = flag.Bool("sniff", true, "Enable or disable sniffing")
		n        = flag.Int("n", 0, "Number of documents to bulk insert")
		bulkSize = flag.Int("bulk-size", 0, "Number of documents to collect before committing")
	)
	flag.Parse()
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())

	if *url == "" {
		log.Fatal("missing url parameter")
	}
	if *index == "" {
		log.Fatal("missing index parameter")
	}
	if *typ == "" {
		log.Fatal("missing type parameter")
	}
	if *n <= 0 {
		log.Fatal("n must be a positive number")
	}
	if *bulkSize <= 0 {
		log.Fatal("bulk-size must be a positive number")
	}

	sups := map[string]*supplier{
		"ickey": &supplier{Index: "ickey", Type: "product", stmt: "SELECT t1.id pid,t1.pro_sno AS p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package,t1. status, t1.pro_name,supplier_id,IF (pro_maf = '',maf_name, pro_maf) AS pro_maf, moq, currency_id, base_buy_amounts, buy_price, base_price, pro_num, t1.img_url, bus_type, pro_desc, t3.sku AS sku, unix_timestamp() AS chrono, md5(trim(t1.pro_sno)) AS pro_sno_md5, md5(trim(pro_sup_sno)) AS pro_sup_sno_md5, md5( 	trim( IF(pro_maf = '', maf_name, pro_maf))) AS pro_maf_md5, t4.cate_id FROM t_pro_sell t1, t_maf t2, t_sku t3, t_product t4 WHERE t1.maf_id = t2.id AND t1.pro_id = t4.id AND t1.id = t3.pro_sell_id AND t1.pro_sup_sno != '' AND bus_type IN (101, 102, 103) AND t1. status = 1"},
		"chip1stop":   &supplier{Index: "chip1stop", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status,  pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=1"},
		"digikey":     &supplier{Index: "digikey", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status,  pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=2"},
		"future":      &supplier{Index: "future", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=3"},
		"wpi":         &supplier{Index: "wpi", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=4"},
		"master":      &supplier{Index: "master", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=7"},
		"microchip":   &supplier{Index: "microchip", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=10"},
		"ps":          &supplier{Index: "ps", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=11"},
		"vicor":       &supplier{Index: "vicor", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=13"},
		"avnet":       &supplier{Index: "avnet", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=14"},
		"rs":          &supplier{Index: "rs", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=22"},
		"element14":   &supplier{Index: "element14", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=26"},
		"element14cn": &supplier{Index: "element14cn", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=34"},
		"aipco":       &supplier{Index: "aipco", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=27"},
		"rochester":   &supplier{Index: "rochester", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=33"},
		"arrow":       &supplier{Index: "arrow", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=23"},
		"pei":       &supplier{Index: "pei", Type: "product", stmt: "select t1.id pid,pro_sno as p,pro_sup_sno,pro_id,date_code,lead_time_cn,lead_time_hk,spq,package, t1.status, pro_name,supplier_id,t2.maf_name as pro_maf,moq,currency_id,base_buy_amounts,buy_price,base_price,pro_num,img_url,bus_type, pro_desc, t3.sku as sku, unix_timestamp() as chrono,md5(trim(pro_sno)) as pro_sno_md5,md5(trim(pro_sup_sno)) as pro_sup_sno_md5,md5(trim(if(pro_maf='',maf_name,pro_maf))) as pro_maf_md5 from t_pro_sell t1,t_maf t2,  t_sku t3  where t1.maf_id=t2.id AND t1.id = t3.pro_sell_id and t1.pro_sup_sno != '' and t1.bus_type=2 and t1.status=1 and t1.supplier_id=20"},
	}

	// 检查 map xx 对象中是否存在某个 key 索引的元素，如果存在获取该 key 索引的 value
	//*index is ickey_v1 | ickey_v2 | ickey_v3 | ickey_v4
	//check 元素是否存在  *index is a map key.
	suffex := ".es_v"
	i := strings.Index(*index, suffex)
	var myIndex string
	if -1 == i {
		//not include
		myIndex = *index
	} else {
		myIndex = (*index)[0:strings.Index(*index, suffex)]
	}

	//now myIndex is [[ickey]]   --------------not include [[ickey.es_v2]]

	if _, ok := sups[myIndex]; !ok {
		fmt.Printf("index[%v] not exist", *index)
		log.Fatal("Please input a exist supported key in map.")
	}

	//in := sups[*index].Index
	in := strings.ToLower(*index)
	ty := strings.ToLower(sups[myIndex].Type)
	stmt := sups[myIndex].stmt
	//fmt.Println("supplier index= %v, type=%v, stmt=%v", i, t, stmt)

	//rows,_ := db.Query("SELECT uid,username FROM USER")
	rows, err := db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	defer db.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	// Create an Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetSniff(*sniff))
	if err != nil || nil == client {
		log.Fatal(err)
	}

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())

	//make a channel
	docsc := make(chan common.MapStr)

	begin := time.Now()
	fmt.Println("Mysql read done: now will go to index", begin)

	// Goroutine to creae documents from mysql
	g.Go(func() error {
		defer close(docsc)

		//buf := make([]byte, 32)
	LoopRows:
		for rows.Next() {

			//event, err := bt.generateEventFromRow(rows, columns, bt.queryTypes[index], dtNow)
			event, err := bt.generateEventFromRow(rows, columns, "queryMysql", dtNow)

			if err != nil {
				logp.Err("Query #%v error generating event from rows:", err)
				continue LoopRows
			} else if event != nil {
				//b.Events.PublishEvent(event)
				//b.Events.PublishEvent(event)
				//fmt.Println("ok event result %v", event)
				//logp.Info("event sent" )
			}

			//fmt.Println("event is %v", event)

			// Generate a random ID
			// _, err = rand.Read(buf)
			// if err != nil {
			// 	return err
			// }
			//set es doc id
			//id := base64.URLEncoding.EncodeToString(buf)

			// Try to parse the value to an int64
			//pid, _ := event["pid"].(int64)
			//fmt.Println("int64 pid==:%v", pid)
			//nColValue, err := strconv.ParseInt(pid, 0, 64)
			//if err == nil {
			//    strColType = columnTypeInt
			//}
			//fmt.Println("id==: " + id)

			//event.Put("Id", id)

			//fmt.Println("will river id:== %v", id)
			//sid := strconv.FormatInt(int64(id), 10)
			//event.Put("isOnline", 1)

			// Send event over to 2nd goroutine, or chancel
			select {
			case docsc <- event:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		//end := time.Now()
		//fmt.Println("方式1 query total time:",end.Sub(start).Seconds())
		return nil
	})

	type IckeyIndexType struct{ index, typ string }

	// Second goroutine will consume the documents sent from the first an bulk insert into ES
	var total uint64
	g.Go(func() error {

		fmt.Println("index", in)
		fmt.Println("type", ty)

		bulk := client.Bulk().Index(in).Type(ty)

		//         LoopDocsc:
		for d := range docsc {
			// Simple progress
			current := atomic.AddUint64(&total, 1)
			dur := time.Since(begin).Seconds()
			sec := int(dur)
			pps := int64(float64(current) / dur)
			fmt.Printf("%10d | %6d req/s | %02d:%02d\r", current, pps, sec/60, sec%60)

			//Enqueue the document
			//id, err := d.GetValue("Id")
			//if err != nil {
			//    log.Fatal(err)
			//}

			//fmt.Println("will river id:== %v", id)
			//sid := strconv.FormatInt(int64(id), 10)
			//fmt.Println("will river string id:== " + id.(string))
			// String returns the MapStr as JSON

			doc := d.String()

			sku, err := d.GetValue("sku")
			if err != nil {
				fmt.Println("this row does not contains sku field err will not index!")
				fmt.Println("err %v", d)
			}
			isku := sku.(int64)
			ssku := strconv.FormatInt(isku, 10)
			bulk.Add(elastic.NewBulkIndexRequest().Id(ssku).Doc(doc))

			if bulk.NumberOfActions() >= *bulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					// Loo up the failed documents with res.Failed(), and e.g. recommit
					logp.Err("bulk insert error doc: #%v rows:  res.failed:  %v", d, res.Failed())
					fmt.Println("event:  [[[[[[[[[[[[" + doc + "]]]]]]]]]]]]]]]]]]]]")

					// return errors.New("bulk commit failed! stan" )
				}
				// "bulk" is reset after Do, so you can resuse it
			}

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		////finish range loop

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			_, err = bulk.Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	// Final results
	dur := time.Since(begin).Seconds()
	sec := int(dur)
	pps := int64(float64(total) / dur)
	fmt.Printf("%10d | %6d req/s | %02d:%02d\n", total, pps, sec/60, sec%60)
}

func insert() {

	//方式1 insert
	//strconv,int转string:strconv.Itoa(i)
	start := time.Now()
	for i := 1001; i <= 1100; i++ {
		//每次循环内部都会去连接池获取一个新的连接，效率低下
		db.Exec("INSERT INTO user(uid,username,age) values(?,?,?)", i, "user"+strconv.Itoa(i), i-1000)
	}
	end := time.Now()
	fmt.Println("方式1 insert total time:", end.Sub(start).Seconds())

	//方式2 insert
	start = time.Now()
	for i := 1101; i <= 1200; i++ {
		//Prepare函数每次循环内部都会去连接池获取一个新的连接，效率低下
		stm, _ := db.Prepare("INSERT INTO user(uid,username,age) values(?,?,?)")
		stm.Exec(i, "user"+strconv.Itoa(i), i-1000)
		stm.Close()
	}
	end = time.Now()
	fmt.Println("方式2 insert total time:", end.Sub(start).Seconds())

	//方式3 insert
	start = time.Now()
	stm, _ := db.Prepare("INSERT INTO user(uid,username,age) values(?,?,?)")
	for i := 1201; i <= 1300; i++ {
		//Exec内部并没有去获取连接，为什么效率还是低呢？
		stm.Exec(i, "user"+strconv.Itoa(i), i-1000)
	}
	stm.Close()
	end = time.Now()
	fmt.Println("方式3 insert total time:", end.Sub(start).Seconds())

	//方式4 insert
	start = time.Now()
	//Begin函数内部会去获取连接
	tx, _ := db.Begin()
	for i := 1301; i <= 1400; i++ {
		//每次循环用的都是tx内部的连接，没有新建连接，效率高
		tx.Exec("INSERT INTO user(uid,username,age) values(?,?,?)", i, "user"+strconv.Itoa(i), i-1000)
	}
	//最后释放tx内部的连接
	tx.Commit()

	end = time.Now()
	fmt.Println("方式4 insert total time:", end.Sub(start).Seconds())

	//方式5 insert
	start = time.Now()
	for i := 1401; i <= 1500; i++ {
		//Begin函数每次循环内部都会去连接池获取一个新的连接，效率低下
		tx, _ := db.Begin()
		tx.Exec("INSERT INTO user(uid,username,age) values(?,?,?)", i, "user"+strconv.Itoa(i), i-1000)
		//Commit执行后连接也释放了
		tx.Commit()
	}
	end = time.Now()
	fmt.Println("方式5 insert total time:", end.Sub(start).Seconds())
}
