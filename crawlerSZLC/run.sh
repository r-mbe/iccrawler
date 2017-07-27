#!/bin/sh
rm ./go-crawler-szlc
go build -o  go-crawler-szlc crawler.go

# test
#./go-crawler-szlc


# product  for nohup.out too large more then 40G after running one night.
# nohup  nohuprooll.sh  2>%1  |  ./logger  go-crawler-szlc  &


nohup ./go-crawler-szlc >/dev/null 2>&1 &
