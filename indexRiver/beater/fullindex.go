package beater

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/olivere/elastic.v5"

	"strings"

	"github.com/elastic/beats/libbeat/beat"
)

//!+
type supplier struct {
	Index string
	Type  string
	stmt  string
}

//!-

func (bt *Mysqlbeat) fullIndexhandler(b *beat.Beat, cmd *RedisCmd) error {
	fmt.Println("start  call delIndexHandler function...")

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
	}

	// Build the MySQL connection string
	connString := fmt.Sprintf("%v:%v@tcp(%v:%v)/db_product", bt.username, bt.password, bt.hostname, bt.port)

	//get sql from supplier

	//execute sql
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	defer db.Close()

	// 检查 map xx 对象中是否存在某个 key 索引的元素，如果存在获取该 key 索引的 value
	//check 元素是否存在  *index is a map key.
	if _, ok := sups[cmd.Index]; !ok {
		fmt.Printf("Get supplier [%v] sql not exist, pls check yml suppliers and supquerys config.", cmd.Index)
		logp.Info("Please input a exist supported key in map. supplier[%v]", cmd.Index)
		return err
	}

	// 检查 map xx 对象中是否存在某个 key 索引的元素，如果存在获取该 key 索引的 value
	//get sql if key exist
	if _, ok := bt.incrSqlMap[cmd.Index]; !ok {
		fmt.Printf("Get supplier [%v] sql not exist, pls check yml suppliers and supquerys config.", cmd.Index)
		logp.Info("Get supplier [%v] sql not exist, pls check yml suppliers and supquerys config.", cmd.Index)
		return err
	}

	// in := sups[cmd.Index].Index
	// ty := sups[cmd.Index].Type
	stmt := sups[cmd.Index].stmt
	//fmt.Println("supplier index= %v, type=%v, stmt=%v", i, t, stmt)

	//rows,_ := db.Query("SELECT uid,username FROM USER")
	rows, err := db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	defer db.Close()

	fmt.Println("Msqldb query aii rows size=")

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	// Setup a group of goroutines from the excellent errgroup package
	g, ctx := errgroup.WithContext(context.TODO())

	//make a channel
	docsc := make(chan common.MapStr)

	begin := time.Now()
	dtNow := begin

	// Goroutine to creae documents from mysql
	g.Go(func() error {
		defer close(docsc)

	LoopRows:
		for rows.Next() {

			//event, err := bt.generateEventFromRow(rows, columns, bt.queryTypes[index], dtNow)
			event, err := bt.generateEventFromRow(rows, columns, "queryMysql", dtNow)

			if err != nil {
				logp.Err("Query #%v error generating event from rows:", err)
				continue LoopRows
			} else if event != nil {
				//b.Events.PublishEvent(event)
			}

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

		index := strings.ToLower(cmd.Index)
		typ := strings.ToLower(cmd.Type)

		bulk := bt.esClient.Bulk().Index(index).Type(typ)

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

			if bulk.NumberOfActions() >= bt.esBulkSize {
				// Commit
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					// Loo up the failed documents with res.Failed(), and e.g. recommit
					logp.Err("bulk insert error doc: #%v rows:  res.failed:  %v", d, res.Failed())
					fmt.Println("event:  [[[[[[[" + doc + "]]]]]]]]]]]]]]]]]]")

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

	return nil
}
