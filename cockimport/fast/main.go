package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	nsq "github.com/nsqio/go-nsq"

	// Import GORM-related packages.
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type TProSellStock struct {
	// gorm.Model
	ID         int64
	Sku        int64     `gorm:"column:sku"`
	Stocknum   int64     `gorm:"column:stock_num"`
	Frozennum  int64     `gorm:"column:frozen_num"`
	Virtualnum int64     `gorm:"column:virtual_num"`
	LastUpTime int64     `gorm:"column:last_update_time"`
	LUptime    time.Time `gorm:"column:luptime"`
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

//Handle h
type Handle struct {
	msgchan chan *nsq.Message
	db      *gorm.DB
	done    chan bool
}

type Message struct {
	Cmd  string `json:"cmd"`
	Data json.RawMessage
}

//DoProcess doprocess.
func (h *Handle) DoProcess(ctx context.Context, iin <-chan []byte) error {

	slen := 400

	var s [][]byte
	for {
		select {
		case <-ctx.Done():
			h.Stop()
			return ctx.Err()
		case v := <-iin:
			if len(s) >= slen {
				//bulk save db see how fast to save. with waitgourp
				h.doSave(&s)
			} else {
				s = append(s, v)
			}
		case <-time.After(time.Second * 10):
			if len(s) >= slen {
				//bulk save db see how fast to save. with waitgourp
				h.doSave(&s)
			}
		}
	}

	return nil
}

func getconnection(isDebug bool) (*gorm.DB, error) {
	var addr string
	if isDebug {
		fmt.Println("Debuging.. model.......")
		addr = "postgresql://stan:888888@10.8.15.167:26257/db_product?sslmode=disable"
	} else {
		fmt.Println("Production.. model.......")
		addr = "postgresql://stan@10.8.51.69:26257/db_product?sslcert=/usr/local/ickey-certs/client-stan/client.stan.crt&sslkey=/usr/local/ickey-certs/client-stan/client.stan.key&sslrootcert =/usr/local/ickey-certs/client-stan/ca.crt&sslmode=require"
	}

	db, err := gorm.Open("postgres", addr)
	if err != nil {
		fmt.Println("connect cockroach db error")
		log.Fatal(err)
	}
	db.DB().SetMaxOpenConns(1)

	return db, nil

}

func (h *Handle) doSave(s *[][]byte) error {

	begin := time.Now()
	var wg sync.WaitGroup
	for _, data := range *s {
		go func(data []byte) error {
			wg.Add(1)
			defer wg.Done()
			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				fmt.Println("#### json.Unmashal rawMessage err", err, string(data))
				return err
			}

			switch msg.Cmd {
			case "c_price":
				var p TProSellPrice
				if err := json.Unmarshal([]byte(msg.Data), &p); err != nil {
					fmt.Print("json convert data error ", err, msg.Data)
				} else {
					h.DoPriceStorage(&p)
				}
			default:
				fmt.Println("Bad command")
				return errors.New("Bad command")
			}
			return nil
		}(data)
	}
	wg.Wait()

	fmt.Printf("All save 50000 rows in cock db spend time: %v seconds.", time.Since(begin).Seconds())
	return nil
}

func (h *Handle) DoPriceStorage(p *TProSellPrice) error {
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
	err := h.db.Create(&price)
	if err != nil {
		fmt.Printf("Err busy buff####XXXXX 333333 create price err: %v  , cock price: %v    \n", err.Error.Error(), price)
		return err.Error
	}

	// select {
	// case <-ctx.Done():
	// 	//do something when finished.
	// 	return nil
	// default:
	// 	//no finish continue
	// }

	return nil
}

//Stop done.
func (h *Handle) Stop() {
	defer h.db.Close()
	h.done <- true
}

func main() {

	//os signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	l := 10000
	in := make(chan []byte, l)

	h := new(Handle)
	db, err := getconnection(true)
	if err != nil || db == nil {
		fmt.Println("connection cockroach error")
		panic(err)
	}
	h.db = db

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go h.DoProcess(ctx, in)

	// for j := 0; j < 100; j++ {
	//send to channal fo save
	for i := 0; i < l; i++ {
		fmt.Println("i=====", i)
		data := []byte(`{"cmd":"c_price","Data":{"MysqlID":87113256,"ProSellID":26612,"Sku":1003026612,"PriceType":1,"CurrencyID":1,"Price1":0,"Number1":3550,"Price2":0,"Number2":0,"Price3":0,"Number3":0,"Price4":0,"Number4":0,"Price5":0,"Number5":0,"Price6":0,"Number6":0,"Price7":0,"Number7":0,"Price8":0,"Number8":0,"Price9":0,"Number9":0,"Price10":0,"Number10":0,"Status":1,"LastUpTime":1499393237,"OpAdminID":1015,"LUptime":"2017-07-18T14:46:02.893886049+08:00"}}`)
		in <- data
	}
	// }

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)

		cancel()
		//after close for loop 3 second exist main
		select {
		case <-time.After(time.Second * 3):
			fmt.Println("receive signal termnal.")
			done <- true
		}

	}()

	// The program will wait here until it gets the
	// expected signal (as indicated by the goroutine
	// above sending a value on `done`) and then exit.
	fmt.Println("awaiting signal")
	<-done

	//send to nsq exist for loop.
	fmt.Println("exiting")
}
