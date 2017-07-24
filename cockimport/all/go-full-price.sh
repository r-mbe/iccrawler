#!/bin/sh

rm full-river
go build -o fullPrice full-river.go

#for test
./fullPrice  -fromdb=t_pro_sell_price -todb=t_pro_sell_price  -w=1000 -q=50000


#for product
#./full-river -url="http://es.search.ickey.cn:9200" -index=t_pro_sell_stock.pro_v2 -type=product -n=100000 -bulk-size=2000
