#!/bin/sh

killall testRiver
rm testRiver
go build main.go 
mv main testRiver
./testRiver -c ./testbeat.yml

