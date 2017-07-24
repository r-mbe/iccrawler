package beater

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	//	"github.com/elastic/beats/libbeat/publisher"

	"context"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"
	"techtoolkit.ickey.cn/indexRiver/config"
	"techtoolkit.ickey.cn/indexRiver/redis"
        "gopkg.in/olivere/elastic.v5"
)

type Mysqlbeat struct {
	done       chan struct{}
	beatConfig *config.MysqlbeatConfig

	rClient  *redis.RedisClient
	esClient *elastic.Client
	//	client           publisher.Client
	period           time.Duration
	hostname         string
	port             string
	username         string
	password         string
	passwordAES      string
	queries          []string
	queryTypes       []string
	deltaWildcard    string
	deltaKeyWildcard string

	redisServer   string
	redisUsername string
	redisPassword string
	redisRiverkey string

	esServer   string
	esUsername string
	esPassword string
	esSniff    bool
	esBulkSize int

	suppliers []string

	//supplier sql
	incrSqlMap map[string]string

	//redis cmd map
	cmdMap map[string]interface{}
	//cmdMap      map[string]IndexRiverFunc

	//mysql xunsou index log db
	logdb *sql.DB

	oldValues    common.MapStr
	oldValuesAge common.MapStr
}

type RedisCmd struct {
	Cmd   string `json:"cmd"`
	Index string `json:"index"`
	Type  string `json:"type"`
	ID    string `json:"id"`
	Doc   string `json:"doc"`
}

type IndexRiverFunc func(b *beat.Beat, cmd *RedisCmd) error

var (
	commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
)

const (
	// secret length must be 16, 24 or 32, corresponding to the AES-128, AES-192 or AES-256 algorithms
	// you should compile your mysqlbeat with a unique secret and hide it (don't leave it in the code after compiled)
	// you can encrypt your password with github.com/adibendahan/mysqlbeat-password-encrypter just update your secret
	// (and commonIV if you choose to change it) and compile.
	secret = "github.com/adibendahan/mysqlbeat"

	// default values
	defaultPeriod           = "10s"
	defaultHostname         = "10.8.11.225"
	defaultPort             = "3306"
	defaultUsername         = "root"
	defaultPassword         = "ickey_2016"
	defaultDeltaWildcard    = "__DELTA"
	defaultDeltaKeyWildcard = "__DELTAKEY"

	defaultRedisServer   = "127.0.0.1:6379"
	defaultRedisUsername = ""
	defaultRedisPassword = ""
	defaultRedisRiverkey = "es:index_river:0"

	defaultEsServer   = "10.8.15.9:9200"
	defaultEsUsername = ""
	defaultEsPassword = ""
	defaultEsSniff    = true
	defaultEsBulkSize = 2000

	// query types values
	queryTypeSingleRow    = "single-row"
	queryTypeMultipleRows = "multiple-rows"
	queryTypeTwoColumns   = "two-columns"
	queryTypeSlaveDelay   = "show-slave-delay"

	// special column names values
	columnNameSlaveDelay = "Seconds_Behind_Master"

	// column types values
	columnTypeString = iota
	columnTypeInt
	columnTypeFloat
)

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := &config.MysqlbeatConfig{}
	if err := cfg.Unpack(config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Mysqlbeat{
		done:       make(chan struct{}),
		beatConfig: config,
	}
	bt.setup(b)

	return bt, nil
}

func (bt *Mysqlbeat) Run(b *beat.Beat) error {
	logp.Info("mysqlbeat is running! Hit CTRL-C to stop it.")

	///#####################################mysql conn#########################################
	///#####################################es conn#########################################
	//bt.client = b.Publisher.Connect()

	rclient := redis.NewClient(bt.redisServer, bt.redisPassword)
	bt.rClient = rclient
	defer rclient.Close()

	fmt.Printf("es connectionsstring=%s,   sniffer=%v", bt.esServer, bt.esSniff)
	// esClient, err := elastic.NewClient(elastic.SetURL(bt.esServer), elastic.SetSniff(bt.esSniff))
	esClient, err := elastic.NewClient(elastic.SetURL(bt.esServer))
	if err != nil || nil == esClient {
		err := fmt.Errorf("connection xx es error")
		return err
	}

	bt.esClient = esClient
	//defer bt.esClient.

	//defer client.Close()
	//first create connection redis mysql and Elasticsearch

	//ticker := time.NewTicker(bt.period)
	// for {
	// 	select {
	// 	case <-bt.done:
	// 		return nil
	// 	case <-ticker.C:
	// 	}

	// 	err := bt.beat(b)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

LoopBeat:
	for {

		///#####################################redis conn#########################################

		//blocked read redis list queue
		//read a cmd from redis.
		// rCount := rclient.ActiveCount()
		// if rCount <= 0 {
		// 	logp.Err("Redis Connection pool num : %v <=0. will recreate pool", rCount)
		// 	rclient.Close()
		// 	rclient := redis.NewClient(bt.redisServer, bt.redisPassword)
		// 	bt.rClient = rclient
		// 	continue LoopBeat
		// }

		rStr, err := rclient.BRPop(bt.redisRiverkey, 10)
		if err != nil {

			if strings.EqualFold(rStr, "timeout") {
				logp.Info("BRpop redis  #timeout.")
				continue LoopBeat
			} else {
				//maybe redis connection error not brpop
				logp.Info("BRpop redis  #(error: %v): %v", err, rStr)
				rclient.Close()
				rclient := redis.NewClient(bt.redisServer, bt.redisPassword)
				bt.rClient = rclient
			}

			continue LoopBeat
		}

		//json parse
		//str1 := '{"cmd": "delindex", "index": ickey", "type": "product", "id": "1"}'
		//str := '{"cmd": "incrindex", "index": "ickey", "type": "product"}'
		/* addindex cmd to redis json
				 '{"cmd": "addindex", "index": "ickey", "type": "product", "id": "10030029764289", "doc":  {
		               "@timestamp": "2017-02-14T06:47:13.219Z",
		               "Id": "04swEgKOnyztM2w1gY3ywXjIwfzK-AaOe5AOflud2yc=",
		               "STATUS": 1,
		               "base_buy_amounts": 0,
		               "base_price": 0,
		               "bus_type": 102,
		               "buy_price": 0,
		               "cate_id": 1074,
		               "chrono": 1487054833,
		               "currency_id": 1,
		               "date_code": "",
		               "img_url": "/images/home/nophotodetail.jpg",
		               "lead_time_cn": "2工作日",
		               "lead_time_hk": "",
		               "moq": 1,
		               "p": "技术服务Part Search Lite",
		               "package": "",
		               "pid": 47959862,
		               "pro_desc": "试用期限3个月；还税6%",
		               "pro_id": 23458402,
		               "pro_maf": "SiliconExpert",
		               "pro_maf_md5": "c04ff8c8d8a42e4c0694172f09e7707b",
		               "pro_name": "技术服务Part Search Lite",
		               "pro_num": 20,
		               "pro_sno_md5": "895075759dd970a16c655360a72e688b",
		               "pro_sup_sno": "18-技术服务Part Search Lite",
		               "pro_sup_sno_md5": "7a350d2de498eab0587b349d355327b7",
		               "sku": 10030029764289,
		               "spq": 1,
		               "supplier_id": 3708
		            }}'
		*/

		rcmd := RedisCmd{}

		//json.Unmarshal([]byte(str), &rcmd)
		json.Unmarshal([]byte(rStr), &rcmd)
		fmt.Println(rcmd)
		fmt.Println(rcmd.Cmd)

		//conver to lower
		rcmd.Cmd = strings.ToLower(rcmd.Cmd)
		rcmd.Type = strings.ToLower(rcmd.Type)
		rcmd.ID = strings.ToLower(rcmd.ID)
		rcmd.Doc = strings.ToLower(rcmd.Doc)

		//check cmd 命令 元素是否存在  *index is a map key. 是否是我们认识的命令
		if _, ok := bt.cmdMap[rcmd.Cmd]; !ok {
			fmt.Printf("index[%v] Unknow how to call handler", rcmd.Cmd)
			logp.Info("Please input a exist supported key in map. %s cmd not know", rcmd.Cmd)
			continue LoopBeat
		}

		logp.Info("Redis RPOP cmd:", rcmd)
		err = bt.cmdMap[rcmd.Cmd].(func(*beat.Beat, *RedisCmd) error)(b, &rcmd)
		if err != nil {
			logp.Info("Error call bt.cmdMap", rcmd)
		}

		// bterr := bt.beat(b)
		// if bterr != nil {
		// 	logp.Info("Beat  #(error: %v): %s", b, rStr)
		// 	continue LoopBeat
		// }

	}

}

func (bt *Mysqlbeat) Stop() {
	//bt.client.Close()
	bt.logdb.Close()
	bt.rClient.Close()

	close(bt.done)
}

// func (bt *Mysqlbeat) beat(b *beat.Beat) error {
// 	return nil
// }

// beat is a function that iterate over the query array, generate and publish events
func (bt *Mysqlbeat) incrIndexHandler(b *beat.Beat, cmd *RedisCmd) error {

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
	//cmd.Index  is ickey_v1   need truncat _v1 ----> to ickey
	suffex := ".es_v"
	i := strings.Index(cmd.Index, suffex)
	var myIndex string
	if -1 == i {
		//not include
		myIndex = cmd.Index
	} else {
		myIndex = cmd.Index[0:strings.Index(cmd.Index, suffex)]
	}

	//myIndex  is [[ickey]] .......................... not include [[ickey.es_v2]]

	//get sql if key exist
	if _, ok := bt.incrSqlMap[myIndex]; !ok {
		fmt.Printf("Get supplier [%v] sql not exist, pls check yml suppliers and supquerys config.", cmd.Index)
		logp.Info("Get supplier [%v] sql not exist, pls check yml suppliers and supquerys config.", cmd.Index)
		return err
	}

	//log sql
	//logp.Info("Rpop redis cmd #v% ", cmd)

	sql := bt.incrSqlMap[myIndex]
	if strings.EqualFold(myIndex, "ickey") {
		sql = fmt.Sprintf(bt.incrSqlMap[myIndex], time.Now().Unix()-10*3600)
		//fmt.Println("########################increment Now execute sql" + sql)
	}

	//log sql
	logp.Info("Redis exec sql #v%", sql)

	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	defer db.Close()

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

	a := makeTimestamp()

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

		//fmt.Println("index", cmd.Index)
		//fmt.Println("type", cmd.Type)

		bulk := bt.esClient.Bulk().Index(cmd.Index).Type(cmd.Type)

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
				fmt.Println("this row does not contains sku field err will not index!", d)
				logp.Info("index bulk err '%v' ", d)
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
					logp.Err("411：bulk insert error doc: #%v rows:  res.failed:  %v", d, res.Failed())
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

	logp.Info("Redis Bulk increament index : %10d | %6d req/s | %02d:%02d\n", total, pps, sec/60, sec%60)

	//log result into xunsou mysql index log db
	logp.Info("myindex=%v total, %v   bigin %v", myIndex, total, begin)
	if !strings.EqualFold(myIndex, "ickey") {
		logp.Info("before prepare stmtins")
		//first  field is project name not include id field
		stmtIns, logerr := bt.logdb.Prepare("INSERT INTO t_xunsou_index_log  (project_name, create_total_num, create_succ_num, create_fail_num, start_time, end_time, spend_time, server_ip, act_type)  VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ? )") // ? = placeholder
		if logerr != nil {
			logp.Info("prepare stmt err %v", logerr)
			log.Fatal(err)
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates micro second

		b := makeTimestamp()
		_, logerr = stmtIns.Exec(myIndex, total, total, 0, a, b, b-a, "10.8.51.96", 1)
		if logerr != nil {
			logp.Info("exec stmt err %v", logerr)
			log.Fatal(err)
		}
	}

	return nil
}

// appendRowToEvent appends the two-column event the current row data
func (bt *Mysqlbeat) appendRowToEvent(event common.MapStr, row *sql.Rows, columns []string, rowAge time.Time) error {

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// Copy the references into such a []interface{} for row.Scan
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Get RawBytes from data
	err := row.Scan(scanArgs...)
	if err != nil {
		return err
	}

	// First column is the name, second is the value
	strColName := string(values[0])
	strColValue := string(values[1])
	strColType := columnTypeString
	strEventColName := strings.Replace(strColName, bt.deltaWildcard, "_PERSECOND", 1)

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

	// If the column name ends with the deltaWildcard
	if strings.HasSuffix(strColName, bt.deltaWildcard) {
		var exists bool
		_, exists = bt.oldValues[strColName]

		// If an older value doesn't exist
		if !exists {
			// Save the current value in the oldValues array
			bt.oldValuesAge[strColName] = rowAge

			if strColType == columnTypeString {
				bt.oldValues[strColName] = strColValue
			} else if strColType == columnTypeInt {
				bt.oldValues[strColName] = nColValue
			} else if strColType == columnTypeFloat {
				bt.oldValues[strColName] = fColValue
			}
		} else {
			// If found the old value's age
			if dtOldAge, ok := bt.oldValuesAge[strColName].(time.Time); ok {
				delta := rowAge.Sub(dtOldAge)

				if strColType == columnTypeInt {
					var calcVal int64

					// Get old value
					oldVal, _ := bt.oldValues[strColName].(int64)
					if nColValue > oldVal {
						// Calculate the delta
						devResult := float64((nColValue - oldVal)) / float64(delta.Seconds())
						// Round the calculated result back to an int64
						calcVal = roundF2I(devResult, .5)
					} else {
						calcVal = 0
					}

					// Add the delta value to the event
					event[strEventColName] = calcVal

					// Save current values as old values
					bt.oldValues[strColName] = nColValue
					bt.oldValuesAge[strColName] = rowAge
				} else if strColType == columnTypeFloat {
					var calcVal float64

					// Get old value
					oldVal, _ := bt.oldValues[strColName].(float64)
					if fColValue > oldVal {
						// Calculate the delta
						calcVal = (fColValue - oldVal) / float64(delta.Seconds())
					} else {
						calcVal = 0
					}

					// Add the delta value to the event
					event[strEventColName] = calcVal

					// Save current values as old values
					bt.oldValues[strColName] = fColValue
					bt.oldValuesAge[strColName] = rowAge
				} else {
					event[strEventColName] = strColValue
				}
			}
		}
	} else { // Not a delta column, add the value to the event as is
		if strColType == columnTypeString {
			event[strEventColName] = strColValue
		} else if strColType == columnTypeInt {
			event[strEventColName] = nColValue
		} else if strColType == columnTypeFloat {
			event[strEventColName] = fColValue
		}
	}

	// Great success!
	return nil
}

func makeTimestamp() float64 {
	//return time.Now().UnixNano() / float64(time.Microsecond*1000000)

	return float64(time.Now().UnixNano()) / float64(time.Microsecond*1000000)

	// float64(d) / float64(time.Millisecond)
}

// generateEventFromRow creates a new event from the row data and returns it
func (bt *Mysqlbeat) generateEventFromRow(row *sql.Rows, columns []string, queryType string, rowAge time.Time) (common.MapStr, error) {
	//logp.Info("generateEventFromRow: %v", queryType)

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
		logp.Err("scan error: %v", err)
		return nil, err
	}

	isINF := false
	// Loop on all columns
	for i, col := range values {
		// Get column name and string value
		strColName := string(columns[i])
		var strColValue string
		if nil != col {
			strColValue = string(col)
		}

		strColType := columnTypeString

		// Skip column proccessing when query type is show-slave-delay and the column isn't Seconds_Behind_Master
		if queryType == queryTypeSlaveDelay && strColName != columnNameSlaveDelay {
			continue
		}

		// Set the event column name to the original column name (as default)
		strEventColName := strColName

		// Remove unneeded suffix, add _PERSECOND to calculated columns
		if strings.HasSuffix(strColName, bt.deltaKeyWildcard) {
			strEventColName = strings.Replace(strColName, bt.deltaKeyWildcard, "", 1)
		} else if strings.HasSuffix(strColName, bt.deltaWildcard) {
			strEventColName = strings.Replace(strColName, bt.deltaWildcard, "_PERSECOND", 1)
		}

		strEventColName = strings.ToLower(strEventColName)

		if strings.EqualFold("pro_sno", strColName) || strings.EqualFold("p", strColName) || strings.EqualFold("pro_sup_sno", strColName) || strings.EqualFold("date_code", strColName) || strings.EqualFold("lead_time_cn", strColName) || strings.EqualFold("lead_time_hk", strColName) || strings.EqualFold("package", strColName) || strings.EqualFold("pro_name", strColName) || strings.EqualFold("pro_maf", strColName) || strings.EqualFold("img_url", strColName) || strings.EqualFold("pro_desc", strColName) || strings.EqualFold("keywords", strColName) || strings.EqualFold("chrono", strColName) || strings.EqualFold("pro_sno_md5", strColName) || strings.EqualFold("pro_sup_sno_md5", strColName) || strings.EqualFold("pro_maf_md5", strColName) {

			if strings.EqualFold(strEventColName, "p") || strings.EqualFold(strEventColName, "pro_sup_sno") || strings.EqualFold(strEventColName, "pro_name") || strings.EqualFold(strEventColName, "pro_maf") {
				strColValue = strings.ToUpper(strColValue)
			}
			event[strEventColName] = strColValue
		} else {

			//convert value
			// Try to parse the value to an int64
			nColValue, err := strconv.ParseInt(strColValue, 0, 64)
			if err == nil {
				strColType = columnTypeInt
			}

			// Try to parse the value to a float64
			fColValue, err := strconv.ParseFloat(strColValue, 64)
			if err == nil {

				//big bug if string field is INF or Nan string it will treat be to fload INf or nan!!! fix by stan.
				if math.IsInf(fColValue, 0) {
				} else if math.IsNaN(fColValue) {
				} else {
					// Now will real convert field type to float64
					// If it's not already an established int64, set type to float
					if strColType == columnTypeString {
						strColType = columnTypeFloat
					}

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

				var inv interface{}
				inv = col
				if math.IsInf(fColValue, 0) {
					isINF = true
					fmt.Println("\n ##############OOOOOOOOO############ isINf  XXXXXXX val=", col, "Type= ", reflect.TypeOf(inv))
					fmt.Printf("\n ++++is Nan  ##############OOOOOOOOO############ strColname= %s  strcolvalue= %v ,  strcoltype:  %v, type  %T,  fColValue:%v", strColName, strColValue, strColType, col, fColValue)

				} else if math.IsNaN(fColValue) {
					isINF = true
					fmt.Println("\n isNan XXXXXXX val=", col, "Type= ", reflect.TypeOf(inv))
					fmt.Printf("\n ++++is Nan   ##############OOOOOOOOO############ strColname= %s  strcolvalue= %v ,  strcoltype:  %v, type  %T,  fColValue:%v", strColName, strColValue, strColType, col, fColValue)
				}
			}
		}

		//}

		//logp.Info("key=%v, value=%v", strEventColName, event[strEventColName])
	}

	// If the event has no data, set to nil
	if len(event) == 2 {
		fmt.Println("no event nil event len=3")
		event = nil
	}

	if isINF {
		isINF = false
		for k, v := range event {
			// log err kv event
			logp.Info("key=%s, value=%v", k, v)
		}
	}

	return event, nil
}

// getKeyFromRow is a function that returns a unique key from row
func getKeyFromRow(bt *Mysqlbeat, values []sql.RawBytes, columns []string) (strKey string, err error) {

	keyFound := false

	// Loop on all columns
	for i, col := range values {
		// Get column name and string value
		if strings.HasSuffix(string(columns[i]), bt.deltaKeyWildcard) {
			strKey += string(col)
			keyFound = true
		}
	}

	if !keyFound {
		err = fmt.Errorf("query type multiple-rows requires at least one delta key column")
	}

	return strKey, err
}

// roundF2I is a function that returns a rounded int64 from a float64
func roundF2I(val float64, roundOn float64) (newVal int64) {
	var round float64

	digit := val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return int64(round)
}

func (bt *Mysqlbeat) setup(b *beat.Beat) error {

	if len(bt.beatConfig.Queries) < 1 {
		err := fmt.Errorf("there are no queries to execute")
		return err
	}

	if len(bt.beatConfig.Queries) != len(bt.beatConfig.QueryTypes) {
		err := fmt.Errorf("error on config file, queries array length != queryTypes array length (each query should have a corresponding type on the same index)")
		return err
	}

	// Setting defaults for missing config
	if bt.beatConfig.Period == "" {
		logp.Info("Period not selected, proceeding with '%v' as default", defaultPeriod)
		bt.beatConfig.Period = defaultPeriod
	}

	if bt.beatConfig.Hostname == "" {
		logp.Info("Hostname not selected, proceeding with '%v' as default", defaultHostname)
		bt.beatConfig.Hostname = defaultHostname
	}

	if bt.beatConfig.Port == "" {
		logp.Info("Port not selected, proceeding with '%v' as default", defaultPort)
		bt.beatConfig.Port = defaultPort
	}

	if bt.beatConfig.Username == "" {
		logp.Info("Username not selected, proceeding with '%v' as default", defaultUsername)
		bt.beatConfig.Username = defaultUsername
	}

	if bt.beatConfig.Password == "" && bt.beatConfig.EncryptedPassword == "" {
		logp.Info("Password not selected, proceeding with default password")
		bt.beatConfig.Password = defaultPassword
	}

	if bt.beatConfig.DeltaWildcard == "" {
		logp.Info("DeltaWildcard not selected, proceeding with '%v' as default", defaultDeltaWildcard)
		bt.beatConfig.DeltaWildcard = defaultDeltaWildcard
	}

	if bt.beatConfig.DeltaKeyWildcard == "" {
		logp.Info("DeltaKeyWildcard not selected, proceeding with '%v' as default", defaultDeltaKeyWildcard)
		bt.beatConfig.DeltaKeyWildcard = defaultDeltaKeyWildcard
	}

	if bt.beatConfig.RedisServer == "" {
		logp.Info("redisServer not selected , proceeding with '%v' as default", defaultRedisServer)
	}

	if bt.beatConfig.RedisUsername == "" {
		logp.Info("redisUsername not selected, proceeding with '%v' as default", defaultRedisUsername)
	}

	if bt.beatConfig.RedisPassword == "" {
		logp.Info("redisPassword not selected, proceeding with '%v' as default", defaultRedisPassword)
	}

	if bt.beatConfig.RedisRiverkey == "" {
		logp.Info("RedisRiverkey not selected, proceeding with '%v' as default", defaultRedisRiverkey)
	}

	if bt.beatConfig.EsServer == "" {
		logp.Info("EsServer not selected, proceeding with '%v' as default", defaultEsServer)
	}

	if bt.beatConfig.EsUsername == "" {
		logp.Info("EsUsername not selected, proceeding with '%v' as default", defaultEsUsername)
	}

	if bt.beatConfig.EsPassword == "" {
		logp.Info("EsPassword not selected, proceeding with '%v' as default", defaultEsPassword)
	}

	if bt.beatConfig.EsBulkSize <= 0 {
		logp.Info("EsbulkSize not selected, proceeding with '%v' as default", defaultEsBulkSize)
	}

	if len(bt.beatConfig.Suppliers) <= 0 {
		logp.Info("suppliers not selected in yaml, must feed some supplier could index to es.")
		// bt.beatConfig.Suppliers = []string{"ickey","chip1stop"}
		log.Fatal(nil)
	}

	if len(bt.beatConfig.SupQueries) <= 0 {
		logp.Info("supplier query sqls not selected in yaml, must feed some supplier could index to es.")
		log.Fatal(nil)
	}

	if len(bt.beatConfig.Suppliers) != len(bt.beatConfig.SupQueries) {
		logp.Info("suppliers array len != suppliers query sqls array len . will crreate map suppler to sql err.")
		log.Fatal(nil)
	}

	// Parse the Period string
	var durationParseError error
	bt.period, durationParseError = time.ParseDuration(bt.beatConfig.Period)
	if durationParseError != nil {
		return durationParseError
	}

	// Handle password decryption and save in the bt
	if bt.beatConfig.Password != "" {
		bt.password = bt.beatConfig.Password
	} else if bt.beatConfig.EncryptedPassword != "" {
		aesCipher, err := aes.NewCipher([]byte(secret))
		if err != nil {
			return err
		}
		cfbDecrypter := cipher.NewCFBDecrypter(aesCipher, commonIV)
		chiperText, err := hex.DecodeString(bt.beatConfig.EncryptedPassword)
		if err != nil {
			return err
		}
		plaintextCopy := make([]byte, len(chiperText))
		cfbDecrypter.XORKeyStream(plaintextCopy, chiperText)
		bt.password = string(plaintextCopy)
	}

	// init the oldValues and oldValuesAge array
	bt.oldValues = common.MapStr{"mysqlbeat": "init"}
	bt.oldValuesAge = common.MapStr{"mysqlbeat": "init"}

	// Save config values to the bt
	bt.hostname = bt.beatConfig.Hostname
	bt.port = bt.beatConfig.Port
	bt.username = bt.beatConfig.Username
	bt.queries = bt.beatConfig.Queries
	bt.queryTypes = bt.beatConfig.QueryTypes
	bt.deltaWildcard = bt.beatConfig.DeltaWildcard
	bt.deltaKeyWildcard = bt.beatConfig.DeltaKeyWildcard

	bt.redisServer = bt.beatConfig.RedisServer
	bt.redisUsername = bt.beatConfig.RedisUsername
	bt.redisPassword = bt.beatConfig.RedisPassword
	bt.redisRiverkey = bt.beatConfig.RedisRiverkey

	bt.esServer = bt.beatConfig.EsServer
	bt.esUsername = bt.beatConfig.EsUsername
	bt.esPassword = bt.beatConfig.EsPassword
	bt.esSniff = bt.beatConfig.EsSniff
	bt.esBulkSize = bt.beatConfig.EsBulkSize

	bt.suppliers = bt.beatConfig.Suppliers

	//set sql
	//convert array to map
	bt.incrSqlMap = make(map[string]string)
	for i, v := range bt.suppliers {
		bt.incrSqlMap[v] = bt.beatConfig.SupQueries[i]
	}

	//init map cmd handle function to call
	bt.cmdMap = make(map[string]interface{})
	bt.cmdMap["incrindex"] = bt.incrIndexHandler
	bt.cmdMap["delindex"] = bt.delIndexHandler
	bt.cmdMap["addindex"] = bt.addIndexHandler
	bt.cmdMap["fullindex"] = bt.fullIndexhandler

	safeQueries := true

	logp.Info("Total # of queries to execute: %d", len(bt.queries))
	for index, queryStr := range bt.queries {

		strCleanQuery := strings.TrimSpace(strings.ToUpper(queryStr))

		if !strings.HasPrefix(strCleanQuery, "SELECT") && !strings.HasPrefix(strCleanQuery, "SHOW") || strings.ContainsAny(strCleanQuery, ";") {
			safeQueries = false
		}

		logp.Info("Query #%d (type: %s): %s", index+1, bt.queryTypes[index], queryStr)
	}

	if !safeQueries {
		err := fmt.Errorf("Only SELECT/SHOW queries are allowed (the char ; is forbidden)")
		return err
	}

	fmt.Println("hostname=" + bt.username)
	// fmt.Println("redisserver=" + bt.redisserver)

	// db_xunsou_index_log	// Build the MySQL connection string
	connString := fmt.Sprintf("%v:%v@tcp(%v:%v)/db_log", "dblog_master", "RdEa211e8HPnK2zM", "mysqldb.log.rw.01.ickey.cn", 3306)

	//get sql from supplier

	//execute sql
	var err error
	bt.logdb, err = sql.Open("mysql", connString)
	if err != nil {
		logp.Info("connect db_log error: err%v", err)
		return err
	}

	return nil
}
