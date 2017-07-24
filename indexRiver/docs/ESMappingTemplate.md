
# Elasticsearch Mapping Template模板创建及使用

## 打开chrome sense工具

## 需要创建的 Mapping 模板的供应商列表
```
[             "ickey",
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
              "rochester"]
```

## Must bool 嵌套优化hits命中率查询Sense语法

```
GET digikey/product/_search
{
    "from" : 0,
    "size" : 100,
    "query": {
        "function_score": {
            "query": {
                 "bool": {
                     "must": [                        
                        {
                            "bool": {
                                "should": [
                            {"regexp": {"p.keyword": ".*3M.*"}            
                            },
                            {"regexp": {"pro_sup_sno.keyword":  ".*3M.*"}
                            },
                            {"regexp": {"pro_name.keyword":  ".*3M.*"}
                            },
                            {"regexp": {"pro_maf.keyword":  ".*3M.*"}},
                            {"match": {"p":  "3M"}},
                            {"match": {"pro_sup_sno":  "3M"}},
                            {"match": {"pro_name":  "3M"}},
                            {"match": {"pro_maf":  "3M"}}                       
                          ]
                            }
                        }
                  ],
                   "filter": {
                             "term": {
                                "status": 1
                             }                   
                        }
                     
                 }
            },
           "functions": [
                {
                    "weight": 4,
                    "filter": {"term": {
                       "pro_sno_md5.keyword": "a45781e5fd9deaf6fbdce4982341b1fe"
                    }}
                },
                 {
                    "weight": 3,
                    "filter": {"term": {
                       "pro_sup_sno_md5.keyword": "81e066108faa77d4b149572293ab0718"
                    }}
                },
                 {
                    "weight": 2,
                    "filter": {"term": {
                       "pro_maf_md5.keywordkeyword": "9073eb4b768dc2669e0e18f91918bca1"
                    }}
                }
             ]
        }
    }
}

```

## 创建索引别名的目的

索引别名和零停机时间

前面提到的重新索引过程中的问题是必须更新你的应用，来使用另一个索引名。索引别名正是用来解决这个问题的！+

索引 别名 就像一个快捷方式或软连接，可以指向一个或多个索引，也可以给任何需要索引名的 API 使用。别名带给我们极大的灵活性，允许我们做到：
在一个运行的集群上无缝的从一个索引切换到另一个
给多个索引分类（例如，last_three_months）
给索引的一个子集创建 视图
我们以后会讨论更多别名的使用场景。现在我们将介绍用它们怎么在零停机时间内从旧的索引切换到新的索引。
这里有两种管理别名的途径：_alias 用于单个操作，_aliases 用于原子化多个操作。
在这一章中，我们假设你的应用采用一个叫 my_index 的索引。而事实上，my_index 是一个指向当前真实索引的别名。真实的索引名将包含一个版本号：my_index_v1, my_index_v2 等等。
开始，我们创建一个索引 my_index_v1，然后将别名 my_index 指向它	  

## 索引模板(mapping template)的作用
创建索引模板可以模糊匹配索引名字，字段mapping,并且利用公用设置

## my_index_v1 template demo 范例
```
DELETE my_index
PUT _template/my_index_v1
{
    "template": "my_index_v*",
    "settings": {
         "index": {
            "number_of_shards": "5",
            "number_of_replicas": "1"
         }
      },
    "mappings": {
        "product": {     
          "properties": { 
            "title":    { "type": "text"  }, 
            "name":     { "type": "text"  },
             "p": { 
                    "type": "keyword" , 
                    "fields": {
                       "raw": {"type": "text"}
                    }
              }, 
             "age": { "type": "integer" }  
          }
        }
      },
      "aliases": {
          "my_index": {}
      },
      "version": 1
}
```

## 获取模板过滤version
```
GET _template?filter_path=*.version
{
   "my_index_v1": {
      "version": 1
   }
}


DELETE /_template/template.es_v2

GET _template
GET _template/template.es_v2
DELETE /_template/template.es_v2

GET _template/template.es_v2

GET  my_index/_mapping

GET _aliases

GET my_index/_search
{
   "query": {
      "match_all": {}
   }
}


put my_index_v1/product/3
{
    "title": "yyyy",
    "p": "JANTX2N23333AUB",
    "age": 3
}

put my_index/product/4
{
    "title": "yyyy",
    "p": "JANTX2N-23333A-UB",
    "age": 3
}

put my_index/product/5
{
    "title": "yyyy",
    "p": "JANTX/2N23333A/UB",
    "age": 3
}

put my_index/product/6
{
    "title": "yyyy",
    "p": "JANTX/2N23333A/   UB",
    "age": 3
}

put my_index/product/9
{
    "title": "yyyy",
    "p": "UTRRRRFD",
    "age": 3
}

put my_index/product/9
{
    "title": "yyyy",
    "p": "UTRR++++++.....-----RRFD",
    "age": 3
}

put my_index/product/7
{
    "title": "yyyy",
    "p": "JANTX/2N____23333A/   UB",
    "age": 3
}


GET my_index/_search
{
   "query": {
      "match_all": {}
   }
}


GET my_index/_search
{
   "query": {
      "regexp": {
         "p": ".*RR+++++.*"
      }
   }
}

GET my_index/_search
{
   "query": {
      "regexp": {
         "p.raw": "P"
      }
   }
}
```


## 匹配keyword (not analyzed)字段 正则匹配 最优化搜索 
同时匹配 p (analyzed toLowercase char mapping filter)
```
GET ickey/_search
{
    "from" : 0,
    "size" : 100,
    "query": {
        "function_score": {
            "query": {
                "bool": {
                  "should": [
                    {"regexp": {"p.keyword": ".*standex.*"}
                    },
                    {"regexp": {"pro_sup_sno.keyword":  ".*standex.*"}
                    },
                    {"regexp": {"pro_name.keyword":  ".*Standex.*"}
                    },
                    {"regexp": {"pro_maf.keyword":  ".*Standex.*"}
                    },                    
                    {"match": {"p":  "STANDEX"}},
                    {"match": {"pro_sup_sno":  "STANDEX"}},
                    {"match": {"pro_name":  "STANDEX"}},
                    {"match": {"pro_maf":  "STANDEX"}}
                  ]
                }
            },
           "functions": [
                {
                    "weight": 4,
                    "filter": {"term": {
                       "pro_sno_md5.keyword": "9073eb4b768dc2669e0e18f91918bca1"
                    }}
                },
                 {
                    "weight": 3,
                    "filter": {"term": {
                       "pro_sup_sno_md5.keyword": "81e066108faa77d4b149572293ab0718"
                    }}
                },
                 {
                    "weight": 2,
                    "filter": {"term": {
                       "pro_maf_md5.keywordkeyword": "9073eb4b768dc2669e0e18f91918bca1"
                    }}
                }
             ]
        }
    }
}
```

## 所有以.es_v*结尾的索引的Mapping Template公用模板   本语句以版本2为例子

```
增加   则定义mapping char filter 自定义分词
PUT _template/template.es_v2
{
      "order": 0,
      "version": 2,
      "template": "*.es_v*",
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
               "lead_time_cn": {
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "bus_type": {
                  "type": "long"
               },
               "pro_sup_sno_md5": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "pro_sno_md5": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "pid": {
                  "type": "long"
               },
               "pro_maf": {
                  "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "moq": {
                  "type": "long"
               },
               "base_price": {
                  "type": "long"
               },
               "pro_sup_sno": {
                  "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "sku": {
                  "type": "long"
               },
               "package": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "lead_time_hk": {
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "pro_maf_md5": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "p": {
                  "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "@timestamp": {
                  "type": "date"
               },
               "pro_id": {
                  "type": "long"
               },
               "img_url": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "spq": {
                  "type": "long"
               },
               "base_buy_amounts": {
                  "type": "long"
               },
               "buy_price": {
                  "type": "long"
               },
               "Id": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "chrono": {
                  "ignore_above": 256,
                  "type": "keyword"
               },
               "date_code": {
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "supplier_id": {
                  "type": "long"
               },
               "currency_id": {
                  "type": "long"
               },
               "pro_desc": {
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "pro_name": {
                  "analyzer": "user_analyzer",
                  "type": "text",
                  "fields": {
                     "keyword": {
                        "ignore_above": 256,
                        "type": "keyword"
                     }
                  }
               },
               "pro_num": {
                  "type": "long"
               },
               "status": {
                  "type": "long"
               },
               "cate_id": {
                  "type": "long"
               },
			   "stock":  {
				  "type": "long"
				},
				"stocktype":  {
				  "type": "long"
				}
            }
         }
      },
      "aliases": {}
}



```
## Chip1Stop Mapping Template模板创建及使用
  其他供应商新版本模板都按照如下规则 公用模板。以.es_v* 结尾，比如

.es_v2版本的模板创建索引名称
[
    ickey.es_v2
    chip1stop.es_v2
    digikey.es_v2
    future.es_v2
    wpi.es_v2
    master.es_v2
    microchip.es_v2
    ps.es_v2
    vicor.es_v2
    avnet.es_v2
    rs.es_v2
    element14.es_v2
    element14cn.es_v2
    aipco.es_v2
    rochester.es_v2
]
 
新增版本如下
[
    .es_v3
    .es_v4
    .es_v5
    .es_v6
    ...
]   

## 其他供应商参考chip1stop建立

## 线上别名运维


## my_index查询新旧索引的别名范例
```
GET  my_index_v1/_aliases
GET  my_index_v2/_aliases

DELETE my_index_v1/_aliases/my_index
DELETE my_index_v2/_aliases/my_index

PUT my_index_v2/_aliases/my_index
```



#### 查询ickey_v*索引的别名
```
GET  ickey_v1/_aliases
GET  ickey_v2/_aliases
```

#### 删除旧索引别名
DELETE ickey_v2/_aliases/my_index
DELETE my_index_v2/_aliases/my_index

DELETE ickey
PUT ickey_v1/_aliases/ickey

# ickey索引新版本重建索引流程
比如线上运行索引ickey_v1
新版索引ickey_v2

1. 删除新版索引，如果存在的话，可以忽略但是确保新版本的索引不存在
```
DELETE ickey_v2
```
2. 需要先创建ickey_v2的mappig ---注意
  参见上面的创建mapping template ickey_v1的方法，如果字段类型没有改变，可以无需重建mappinig template
3. 导入新版数据到ickey_v2索引 参见README到索引
4. 删除 ickey_v1的别名
```
DELETE ickey_v1/_aliases/ickey
```
5. 创建新版ickey_v2的别名
```
PUT ickey_v2/_aliases/ickey
```
6. 保留最后一个版本的索引


## 总结

利用索引别名实现零停机时间
利用Mapping Template实现自动重建索引是可以修改Mapping字段类型的机会