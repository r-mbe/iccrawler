#!/bin/sh

rm full-river
go build full-river.go

#for test
#./full-river -url="http://10.8.15.9:9200" -index=keywords.key_v2 -type=product -n=100000 -bulk-size=2000


#for product
./full-river -url="http://es.search.ickey.cn:9200" -index=keywords.key_v2 -type=product -n=100000 -bulk-size=2000

