package beater

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/logp"
)

func (bt *Mysqlbeat) addIndexHandler(b *beat.Beat, cmd *RedisCmd) error {
	fmt.Println("now call addIndexHandler function")
	//fmt.Println(cmd)

	index := strings.ToLower(cmd.Index)
	typ := strings.ToLower(cmd.Type)
	doc := cmd.Doc

	fmt.Printf("cmd:%v, type:%v, id:%v,  doc: %v  %T", index, typ, cmd.ID, doc, doc)
	_, err := bt.esClient.Index().Index(index).Type(typ).Id(cmd.ID).BodyString(doc).Do(context.TODO())

	if err != nil {
		fmt.Printf("Index one row data error cmd=%v \n", cmd)
		logp.Info("Index one row data error cmd=%v", cmd)
	}

	//flush to file
	_, err = bt.esClient.Flush().Index(index).Do(context.TODO())
	if err != nil {
		fmt.Printf("Add index flush errcmd%v \n", cmd)
		logp.Info("Add index flush err error cmd %v", cmd)
	}

	return nil
}
