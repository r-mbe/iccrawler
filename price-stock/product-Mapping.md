
## product索引所有以*.pro_v结尾的索引都将按照一下模板创建索引的settings 和mapping

## product.pro_v2 索引模板创建Sense脚本

```
PUT _template/product.pro_v2
{
      "order": 0,
      "version": 2,
      "template": "*.pro_v*",
      "settings": {
         "index": {
            "analysis": {
               "analyzer": {
                  "user_analyzer": {
                     "char_filter": [
                        "stan_char_filter"
                     ],
                    "filter": [
                        "lowercase",
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
               "cate_id": {
                  "type": "long"
               },
               "created_time": {
                  "type": "date"
               },
               "data_sheet": {
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "img_url": {
                    "type": "keyword",
                    "ignore_above": 256                                       
               },
               "is_rohs": {
                    "type": "long"
               },
               "last_update_time": {
                  "type": "long"
               },
               "maf_id": {
                  "type": "long"
               },
               "op_admin_id": {
                  "type": "long"
               },
               "op_admin_name": {
			      "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "pro_group": {
                  "type": "long"
               },
               "pro_id": {
                  "type": "long"
               },
               "pro_name": {
			      "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "pro_sno": {
			      "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "pro_url": {
                    "type": "keyword",
                    "ignore_above": 256
               },
               "remark": {
			      "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
               },
               "status": {
                  "type": "long"
               },
               "supplier_category": {
			      "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "type": "keyword",
                        "ignore_above": 256
                     }
                  }
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