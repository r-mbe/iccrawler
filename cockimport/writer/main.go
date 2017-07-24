package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/codahale/hdrhistogram"
	"github.com/satori/go.uuid"
	// Import postgres driver.

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	insertBlockStmt = `INSERT INTO blocks (block_id, writer_id, block_num, raw_bytes) VALUES`
)

var createDB = flag.Bool("create-db", true, "Attempt to create the database (root user only)")

// concurrency = number of concurrent insertion processes.
var concurrency = flag.Int("concurrency", 2*runtime.NumCPU(), "Number of concurrent writers inserting blocks")

// batch = number of blocks to insert in a single SQL statement.
var batch = flag.Int("batch", 1, "Number of blocks to insert in a single SQL statement")

var splits = flag.Int("splits", 0, "Number of splits to perform before starting normal operations")

var tolerateErrors = flag.Bool("tolerate-errors", false, "Keep running on error")

// outputInterval = interval at which information is output to console.
var outputInterval = flag.Duration("output-interval", 1*time.Second, "Interval of output")

// Minimum and maximum size of inserted blocks.
var minBlockSizeBytes = flag.Int("min-block-bytes", 256, "Minimum amount of raw data written with each insertion")
var maxBlockSizeBytes = flag.Int("max-block-bytes", 1024, "Maximum amount of raw data written with each insertion")

var maxBlocks = flag.Uint64("max-blocks", 0, "Maximum number of blocks to write")
var duration = flag.Duration("duration", 0, "The duration to run. If 0, run forever.")
var benchmarkName = flag.String("benchmark-name", "BenchmarkBlockWriter", "Test name to report "+
	"for Go benchmark results.")

// numBlocks keeps a global count of successfully written blocks.
var numBlocks uint64

const (
	minLatency = 100 * time.Microsecond
	maxLatency = 10 * time.Second
)

func clampLatency(d, min, max time.Duration) time.Duration {
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

type blockWriter struct {
	db      *gorm.DB
	rand    *rand.Rand
	latency struct {
		sync.Mutex
		*hdrhistogram.WindowedHistogram
	}
}

type TProSellPrice struct {
	// gorm.Model
	ID         int64
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

type Message struct {
	Cmd  string `json:"cmd"`
	Data json.RawMessage
}

func newBlockWriter(db *gorm.DB) *blockWriter {
	bw := &blockWriter{
		db:   db,
		rand: rand.New(rand.NewSource(int64(time.Now().UnixNano()))),
	}
	bw.latency.WindowedHistogram = hdrhistogram.NewWindowed(1,
		minLatency.Nanoseconds(), maxLatency.Nanoseconds(), 1)
	return bw
}

// run is an infinite loop in which the blockWriter continuously attempts to
// write blocks of random data into a table in cockroach DB.
func (bw *blockWriter) run(errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	id := uuid.NewV4().String()
	var blockCount uint64

	for {
		var buf bytes.Buffer
		var args [][]byte
		fmt.Fprintf(&buf, "%s", insertBlockStmt)

		for i := 0; i < *batch; i++ {
			blockID := bw.rand.Int63()
			blockCount++
			args = append(args, bw.randomBlock())
			if i > 0 {
				fmt.Fprintf(&buf, ",")
			}
			fmt.Fprintf(&buf, ` (%d, '%s', %d, $%d)`, blockID, id, blockCount, i+1)
		}

		start := time.Now()

		for _, data := range args {
			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				fmt.Println("#### json.Unmashal rawMessage err", err, string(data))
			}

			var p TProSellPrice
			if err := json.Unmarshal([]byte(msg.Data), &p); err != nil {
				fmt.Print("json convert data error ", err, msg.Data)
			} else {
				bw.DoPriceStorage(&p)
			}

		}

		dur := time.Since(start).Seconds()
		fmt.Println("all write spend time", dur)
	}
}

func (bw *blockWriter) randomBlock() []byte {
	blockData := []byte(`{"cmd":"c_price","Data":{"MysqlID":87113256,"ProSellID":26612,"Sku":1003026612,"PriceType":1,"CurrencyID":1,"Price1":0,"Number1":3550,"Price2":0,"Number2":0,"Price3":0,"Number3":0,"Price4":0,"Number4":0,"Price5":0,"Number5":0,"Price6":0,"Number6":0,"Price7":0,"Number7":0,"Price8":0,"Number8":0,"Price9":0,"Number9":0,"Price10":0,"Number10":0,"Status":1,"LastUpTime":1499393237,"OpAdminID":1015,"LUptime":"2017-07-18T14:46:02.893886049+08:00"}}`)

	return blockData
}

func (bw *blockWriter) DoPriceStorage(p *TProSellPrice) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("recoverd in Price DoStorage. err:", err)

			fmt.Println("NSQ receive price", *p)
			//log.Info("error: %v", err)
		}
	}()

	price := *p

	if price.Price1 >= -0.0000001 && price.Price1 <= 0.0000001 {
		price.Price1 = 0.0000001
	}
	if price.Price2 >= -0.0000001 && price.Price2 <= 0.0000001 {
		price.Price2 = 0.0000001
	}
	if price.Price3 >= -0.0000001 && price.Price3 <= 0.0000001 {
		price.Price3 = 0.0000001
	}
	if price.Price4 >= -0.0000001 && price.Price4 <= 0.0000001 {
		price.Price4 = 0.0000001
	}
	if price.Price5 >= -0.0000001 && price.Price5 <= 0.0000001 {
		price.Price5 = 0.0000001
	}
	if price.Price6 >= -0.0000001 && price.Price6 <= 0.0000001 {
		price.Price6 = 0.0000001
	}
	if price.Price7 >= -0.0000001 && price.Price7 <= 0.0000001 {
		price.Price7 = 0.0000001
	}
	if price.Price8 >= -0.0000001 && price.Price8 <= 0.0000001 {
		price.Price8 = 0.0000001
	}
	if price.Price9 >= -0.0000001 && price.Price9 <= 0.0000001 {
		price.Price9 = 0.0000001
	}
	if price.Price10 >= -0.0000001 && price.Price10 <= 0.0000001 {
		price.Price10 = 0.0000001
	}

	price.LUptime = time.Now()
	err := bw.db.Create(&price)
	if err != nil {
		fmt.Printf("Err busy buff####XXXXX 333333 create price err: %v  , cock price: %v    \n", err.Error.Error(), price)
		return err.Error
	}
	return nil
}

// setupDatabase performs initial setup for the example, creating a database and
// with a single table. If the desired table already exists on the cluster, the
// existing table will be dropped.
func setupDatabase(dbURL string) (*gorm.DB, error) {

	fmt.Println("XXdbstring path", dbURL)

	// Open connection to server and create a database.
	db, err := gorm.Open("postgres", dbURL)
	if err != nil {
		return db, err
	}

	// Allow a maximum of concurrency+1 connections to the database.
	db.DB().SetMaxOpenConns(*concurrency + 1)
	db.DB().SetMaxIdleConns(*concurrency + 1)

	return db, nil
}

var usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <db URL>\n\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	//dbURL := "postgresql://root@localhost:26257/photos?sslmode=disable"
	dbURL := "postgresql://stan:888888@10.8.15.167:26257/db_product?sslmode=disable"

	if *concurrency < 1 {
		log.Fatalf("Value of 'concurrency' flag (%d) must be greater than or equal to 1", *concurrency)
	}

	if max, min := *maxBlockSizeBytes, *minBlockSizeBytes; max < min {
		log.Fatalf("Value of 'max-block-bytes' (%d) must be greater than or equal to value of 'min-block-bytes' (%d)", max, min)
	}

	var db *gorm.DB
	var err error
	db, err = setupDatabase(dbURL)
	if err != nil {
		panic(err)
	}
	fmt.Println("db=", db)
	// panic(err)

	lastNow := time.Now()
	start := lastNow
	var lastBlocks uint64
	writers := make([]*blockWriter, *concurrency)

	errCh := make(chan error)
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		writers[i] = newBlockWriter(db)
		go writers[i].run(errCh, &wg)
	}

	var numErr int
	tick := time.Tick(*outputInterval)
	done := make(chan os.Signal, 3)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		wg.Wait()
		done <- syscall.Signal(0)
	}()

	if *duration > 0 {
		go func() {
			time.Sleep(*duration)
			done <- syscall.Signal(0)
		}()
	}

	defer func() {
		// Output results that mimic Go's built-in benchmark format.
		elapsed := time.Since(start)
		fmt.Printf("%s\t%8d\t%12.1f ns/op\n",
			*benchmarkName, numBlocks, float64(elapsed.Nanoseconds())/float64(numBlocks))
	}()

	for i := 0; ; {
		select {
		case err := <-errCh:
			numErr++
			if !*tolerateErrors {
				log.Fatal(err)
			} else {
				log.Print(err)
			}
			continue

		case <-tick:
			var h *hdrhistogram.Histogram
			for _, w := range writers {
				w.latency.Lock()
				m := w.latency.Merge()
				w.latency.Rotate()
				w.latency.Unlock()
				if h == nil {
					h = m
				} else {
					h.Merge(m)
				}
			}

			p50 := h.ValueAtQuantile(50)
			p95 := h.ValueAtQuantile(95)
			p99 := h.ValueAtQuantile(99)
			pMax := h.ValueAtQuantile(100)

			now := time.Now()
			elapsed := time.Since(lastNow)
			blocks := atomic.LoadUint64(&numBlocks)
			if i%20 == 0 {
				fmt.Println("_elapsed___errors__ops/sec(inst)___ops/sec(cum)__p50(ms)__p95(ms)__p99(ms)_pMax(ms)")
			}
			i++
			fmt.Printf("%8s %8d %14.1f %14.1f %8.1f %8.1f %8.1f %8.1f\n",
				time.Duration(time.Since(start).Seconds()+0.5)*time.Second,
				numErr,
				float64(blocks-lastBlocks)/elapsed.Seconds(),
				float64(blocks)/time.Since(start).Seconds(),
				time.Duration(p50).Seconds()*1000,
				time.Duration(p95).Seconds()*1000,
				time.Duration(p99).Seconds()*1000,
				time.Duration(pMax).Seconds()*1000)
			lastBlocks = blocks
			lastNow = now

		case <-done:
			blocks := atomic.LoadUint64(&numBlocks)
			elapsed := time.Since(start).Seconds()
			fmt.Println("\n_elapsed___errors_________blocks___ops/sec(cum)")
			fmt.Printf("%7.1fs %8d %14d %14.1f\n\n",
				time.Since(start).Seconds(), numErr,
				blocks, float64(blocks)/elapsed)
			return
		}
	}
}
