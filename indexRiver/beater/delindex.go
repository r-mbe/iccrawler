package beater

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
)

func (bt *Mysqlbeat) delIndexHandler(b *beat.Beat, cmd *RedisCmd) error {

	// Delete an index.
	index := strings.ToLower(cmd.Index)
	typ := strings.ToLower(cmd.Type)
	res, err := bt.esClient.Delete().Index(index).Type(typ).Id(cmd.ID).Do(context.TODO())

	if err != nil {
		fmt.Printf("delete  one data error cmd=%v \n", cmd)
		logp.Info("delete one row data error cmd=%v", cmd)
	}

	if res.Found != true {
		fmt.Printf("delete  one row res found error cmd=%v \n", cmd)
		logp.Info("delete  one row res found error cmd=%v", cmd)
	}

	_, err = bt.esClient.Flush().Index(index).Do(context.TODO())
	if err != nil {
		fmt.Printf("Delete one row index flush error cmd=%v \n", cmd)
		logp.Info("Delete one row index flush error cmd=%v", cmd)
	}

	return nil
}
