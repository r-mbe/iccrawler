#!/bin/sh

rm full-river
go build -o fullStock full-river.go

#for test
./fullStock  -fromdb=t_pro_sell_stock -todb=t_pro_sell_stock
