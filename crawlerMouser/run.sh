#!/bin/sh
rm ./go-crawler-mouser
go build -o  go-crawler-mouser crawler.go

# test
#./go-crawler-mouser


# product  for nohup.out too large more then 40G after running one night.
# nohup  nohuprooll.sh  2>%1  |  ./logger  go-crawler-mouser  &


# nohup ./go-crawler-mouser >/dev/null 2>&1 &


nohup ./go-crawler-mouser &
