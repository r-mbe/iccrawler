#!/bin/sh

rm full-river
go build full-river.go

#for test
./fullRiver -url="http://10.8.167.9:9200" -index=product.pro_v2 -type=product -n=100000 -bulk-size=2000


#for product
#./full-river -url="http://es.search.ickey.cn:9200" -index=product.pro_v2 -type=product -n=100000 -bulk-size=2000
