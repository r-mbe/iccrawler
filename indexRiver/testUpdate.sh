#!/bin/sh

rm testUpdate
go build testUpdate.go
./testUpdate -url="http://172.18.0.3:9200" -index=aii -type=product -n=100000 -bulk-size=1000

