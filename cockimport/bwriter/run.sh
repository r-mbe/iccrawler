#!/bin/sh

go build -o nsqriver main.go
./nsqriver  -batch=500 -concurrency=300 -debug=false
