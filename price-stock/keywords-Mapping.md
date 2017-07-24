
## product索引所有以*.pro_v结尾的索引都将按照一下模板创建索引的settings 和mapping

## product.pro_v2 索引模板创建Sense脚本

```
PUT _template/template.key_v2
{
      "order": 0,
      "version": 2,
      "template": "*.key_v*",
      "settings": {
         "index": {
            "analysis": {
               "analyzer": {
                  "user_analyzer": {
                     "char_filter": [
                        "stan_char_filter"
                     ],
                    "filter": [
                        "uppercase",
                        "asciifolding"
                    ],
                    "tokenizer": "standard"
                  }
               },
               "char_filter": {
                  "stan_char_filter": {
                     "type": "mapping",
                     "mappings": [
                        "！ => !",
                        "＂ => \"",
                        "＃ => #",
                        "＄ => $",
                        "％ => %",
                        "＆ => &",
                        "·  => ' '",
                        "（ => (",
                        "） => )",
                        "＊ => *",
                        "＋ => +",
                        "， => ,",
                        "－ => -",
                        "． => .",
                        "／ => /",
                        "： => :",
                        "； => ;",
                        "＜ => <",
                        "＝ => =",
                        "＞ => >",
                        "？ => ?",
                        "＠ => @",
                        "［ => [",
                        "＼ => \"",
                        "］ => ]",
                        "＾ => ^",
                        "＿ => _",
                        "＇ => \"",
                        "｛ => {",
                        "｜ => |",
                        "｝ => }",
                        "～ => ~"
                     ]
                  }
               }
            },
            "number_of_shards": "5",
            "number_of_replicas": "1"
         }
      },
      "mappings": {
	      "product": {
            "properties": {
               "@timestamp": {
                  "type": "date"
               },
               "id": {
                  "type": "long"
               },
               "created_time": {
                  "type": "date"
               },
               "keywords": {
                  "type": "text",
                  "fields": {
                     "keywords": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "uid": {
                  "type": "long"
               },
               "ip": {
                    "type": "keyword",
                    "ignore_above": 256                                       
               },
               "fromid": {
                    "type": "long"
               },
               "guestid": {
                  "type": "long"
               },
               "status": {
                  "type": "long"
               },
               "source_type": {
                  "type": "long"
               },               
               "source_host": {
                    "type": "keyword",
                    "ignore_above": 256
               },               
               "source_referrer": {
                    "type": "keyword",
                    "ignore_above": 256
               }
      
            }
         }
      },
      "aliases": {}
}

```

## 删除模板
DELETE _template/product.pro_v2

## 查询模板
GET _template/product.pro_v2


## 删除旧索引
DELETE 

DELETE product

## 创建索引别名

PUT  product.es_v2/_aliases/product


## 每次升级索引模板顺序

删除旧索引
删除别名
删除旧模板
创建新模板
创建新索引
创建别名


## 对列入队
lpush es:index_river:0  '{"cmd": "incrindex", "index": "ickey", "type": "product"}'

lpush es:index_river:0  '{"cmd": "incrindex", "index": "chip1stop", "type": "product"}'

lpush es:index_river:0  '{"cmd": "incrindex", "index": "digikey", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "future", "type": "product"}'


lpush es:index_river:0  '{"cmd": "incrindex", "index": "wpi", "type": "product"}'

lpush es:index_river:0  '{"cmd": "incrindex", "index": "master", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "microchip", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "ps", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "vicor", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "avnet", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "rs", "type": "product"}'

lpush es:index_river:0  '{"cmd": "incrindex", "index": "element14", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "element14cn", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "aipco", "type": "product"}'
lpush es:index_river:0  '{"cmd": "incrindex", "index": "rochester", "type": "product"}'