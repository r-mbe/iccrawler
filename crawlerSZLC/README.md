# crawler szlcsc

## v0.0.2
crawler with proxy checker micro service.


## build

go build -o go-crawler-szlc main.go

nohup ./go-crawler-szlc >/dev/null 2>&1 &

# Test a speciy function

go test -v  -test.run TestGetPageListDetail
