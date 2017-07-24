package mysqldb

import (
    "fmt"
    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql" // import your used driver
    "strings"
    "sync"
    "time"
)


func init() {
  
    orm.RegisterDataBase("default", "mysql", "@tcp(192.168.100.3306)/guttv_vod?charset=utf8", 30)

    orm.RegisterModelWithPrefix("t_", new(Series), new(Product), new(ServiceGroup))
    orm.RunSyncdb("default", false, false)
}

func init() {
	//db, _ = sql.Open("mysql", "root:ickey_2016@tcp(10.8.11.225:3306)/db_product")
        //slave
	//db, _ = sql.Open("mysql", "dbproduct_slave:RdEa211e8HPnK2zM@tcp(mysqldb.product.r.01.ickey.cn:3306)/db_product")
        //master
	db, _ = sql.Open("mysql", "dbproduct_master:RdEa211e8HPnK2zM@tcp(mysqldb.product.rw.01.ickey.cn:3306)/db_product")
}

type Mysql struct {
    sql   string
    total int64
    lock  *sync.Mutex
}

func NewDB() (db *Mysql) {
    db = new(Mysql)
    db.New()
    return db
}

func (this *Mysql) New() {
    //this.sql = "SELECT s.*, p.code ProductCode, p.name pName  FROM guttv_vod.t_series s inner join guttv_vod.t_product p on p.itemcode=s.code  and p.isdelete=0 limit ?,?"
    this.sql = "SELECT s.*, p.code ProductCode, p.name pName  FROM guttv_vod.t_series s , guttv_vod.t_product p where p.itemcode=s.code  and p.isdelete=0 limit ?,?"
    this.total = 0
    this.lock = &sync.Mutex{}
}