package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"techtoolkit.ickey.cn/indexRiver/beater"
)

func main() {
	err := beat.Run("mysqlbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
