# indexRiver

# v0.0.27 features
redis 连接失败重新连接
add cate_id for 分类列表 and support full index
## add one index  supplier  id === sku
实时添加索引 and Upper Lower case search
must bool 嵌套优化hits命中率
ickey自营7-23点执行,get 10个小时内增量


## 安装glide 依赖工具
```
curl https://glide.sh/get | sh
glide --version
glide version v0.12.3
```
## go get 安装依赖包
```
go get golang.org/x/net/context
go get golang.org/x/sync/errgroup
go get gopkg.in/olivere/elastic.v5
go get github.com/elastic/beats/libbeat
go get github.com/go-sql-driver/mysql
```

## 下载编译mysqlbeat代码
##  git clone xxxx/mysqlbeat coode
glide update --no-recursive
go build main.go

##################

#How to build
go get ...

rm indexRiver
go build indexRiver.go
./indexRiver -url="http://172.18.0.2:9200" -index=aii -type=product -n=300000 -bulk-size=5000


## how to improve performance of bulk indexing..
put a redis string cmd into redis list named  [{{"es:index_river:0"}}]

## increment index ickey supplier
lpush es:index_river:0  '{"cmd": "incrindex", "index": "ickey", "type": "product"}'

## increment index chip1stop supplier
lpush es:index_river:0  '{"cmd": "incrindex", "index": "chip1stop", "type": "product"}'




## notice id is sku for ic item.
## add new one index  supplier  id === sku

```
//redis addindex
lpush es:index_river:0  '{"cmd": "addindex", "index": "ickey", "type": "product", "id": "10030029764289", "doc": {
		               "@timestamp": "2017-02-14T06:47:13.219Z",
		               "Id": "04swEgKOnyztM2w1gY3ywXjIwfzK-AaOe5AOflud2yc=",
		               "status": 1,
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

//conver to one row string to put in redis-cli>
lpush es:index_river:0  '{"cmd": "addindex", "index": "ickey", "type": "product", "id": "10030029764289", "doc": "{\"@timestamp\": \"2017-02-14T06:47:13.219Z\", \"Id\": \"04swEgKOnyztM2w1gY3ywXjIwfzK-AaOe5AOflud2yc=\", \"status\": 1, \"base_buy_amounts\": 0, \"base_price\": 0, \"bus_type\": 102, \"buy_price\": 0, \"cate_id\": 1074, \"chrono\": 1487054833, \"currency_id\": 1, \"date_code\": \"\", \"img_url\": \"/images/home/nophotodetail.jpg\", \"lead_time_cn\": \"2工作日\", \"lead_time_hk\": \"\", \"moq\": 1, \"p\": \"技术服务Part Search Lite\", \"package\": \"\", \"pid\": 47959862, \"pro_desc\": \"试用期限3个月；还税6%\", \"pro_id\": 23458402, \"pro_maf\": \"SiliconExpert\", \"pro_maf_md5\": \"c04ff8c8d8a42e4c0694172f09e7707b\", \"pro_name\": \"技术服务Part Search Lite\", \"pro_num\": 20, \"pro_sno_md5\": \"895075759dd970a16c655360a72e688b\", \"pro_sup_sno\": \"18-技术服务Part Search Lite\", \"pro_sup_sno_md5\": \"7a350d2de498eab0587b349d355327b7\", \"sku\": 10030029764289, \"spq\": 1, \"supplier_id\": 3708}"}'

/////////////////sense api
PUT ickey/product/10030029764289
{
		               "@timestamp": "2017-02-14T06:47:13.219Z",
		               "Id": "04swEgKOnyztM2w1gY3ywXjIwfzK-AaOe5AOflud2yc=",
		               "status": 1,
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
}
```

## delindex  one index row id === sku
lpush es:index_river:0  '{"cmd": "delindex", "index": "chip1stop", "type": "product", "id": "1"}'

## full index chip1stop supplier
lpush es:index_river:0  '{"cmd": "fullindex", "index": "chip1stop", "type": "product"}'


## supported suppliers
  suppliers: ["ickey",
              "chip1stop",
              "digikey",
              "future",
              "wpi",
              "master",
              "microchip",
              "ps",
              "vicor",
              "avnet",
              "rs",
              "element14",
              "element14cn",
              "aipco",
              rochester]

//str := `{"cmd": "incrindex", "index": "ickey", "type": "product"}`


## How to running for test.
go build main.go
./main -c testbeat.yml

redis-cli>lpush es:index_river:0  '{"cmd": "incrindex", "index": "chip1stop", "type": "product"}'

then check es index increment indexing.

##########################################Upgrade Index Step by Step begin ###########################
## how to upgrade new version index
 目的 ickey.es_v1 升级到ickey.es_v2的步骤  //参考ESMappingTemplate.md创建新模板

## 创建template.es_v2模板并
DELETE _template/template.es_v2
PUT _template/template.es_v2 
GET _template/template.es_v2

## 删除旧有模板
DELETE _template/template.es_v1

## 导入全量ickey.es_v2数据
修改go_ickey.sh脚本中的 -index名称 为对应版本名称: ickey.es_v2
```
#!/bin/sh

rm fullRiver
go build fullRiver.go
./fullRiver -url="http://es.search.ickey.cn:9200" -index=ickey.es_v2 -type=product -n=100000 -bulk-size=2000
```
## 执行全量脚本导入数据,导入数据的时候自动创建ickey.es_v2 的mapping.
./go_ickey.sh

## SEnse监测 ickey.es_v2的 mapping
GET ickey.es_v2/_mapping

##  此时需要删除旧的索引ickey.es_v1的 ickey 别名，并将ickey别名指向新的索引 ickey.es_v2

DELETE ickey.es_v1/_aliases/ickey

PUT ickey.es_v2/_aliases/ickey

## 再ickey别名 实际指向ickey.es_v2 导入增量ickey索引
redis-cli>lpush es:index_river:0  '{"cmd": "incrindex", "index": "ickey", "type": "product"}'

 此时所有数据已经升级完成

## 删除旧的ickey.es_v1索引数据
DELETE ickey.es_v1

#### 删除## template.es_v1模板


########################wpi 升级到索引范例 Sense用的语句
全量索引
./go_wpi.sh


DELETE master

GET master/_search
{
   "query": {
      "match_all": {}
   }
}
GET  master/_mapping



GET master.es_v2/_mapping
GET master.es_v2/_search
{
   "query": {
      "match_all": {}
   }
}

PUT  master.es_v2/_aliases/master



##########################################Upgrade Index Step by Step  end  ###########################


## How to running full index digikey
./go_digikey

## How to running for production increment index
./go_indexRiver.sh "usage start|stop|restart|rebuild"

