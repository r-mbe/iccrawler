#!/bin/sh

rm fullRiver
go build fullRiver.go

#for test
#./fullRiver -url="http://10.8.15.9:9200" -index=element14cn -type=product -n=100000 -bulk-size=2000

# ./fullRiver -url="http://10.8.51.121:9200" -index=element14cn.es_v2 -type=product -n=100000 -bulk-size=2000

#for product
./fullRiver -url="http://es.search.ickey.cn:9200" -index=element14cn.es_v2 -type=product -n=100000 -bulk-size=2000

