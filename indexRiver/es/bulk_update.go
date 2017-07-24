//     bulk_insert -index=warehouse -type=product -n=100000 -bulk-size=1000
//stan@ubuntuos:~/gowork/src/github.com/stanxii/allindex$ ./bulk_insert -index=warehouse -type=product -n=5000000 -bulk-size=5000
//   5000000 |   6585 req/s | 12:39
//
//
package es

import (
	//	"encoding/base64"
	//	"encoding/json"
	"errors"
	//	"flag"
	"fmt"
	//	"log"
	//	"math/rand"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/olivere/elastic.v5"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	//"github.com/stanxii/indexRiver/ds"
)

//type DelBulkClient struct {
//}

/*
	suppliers := map[string]int{
		"advancedmp": 29,
		"aii": 6,
        "aipco": 27,
		"avnet": 14,
		"bristol": 16,
        "chip1stop" :1,
        "digikey": 2,
        "element14": 26,
		"element14cn": 34,
        "future": 3,
        "hdi": 3282,
        "ickey": 30,
        "microchip": 10,
        "ps": 11,
        "rutronik": 12,
        "vicor": 13,
        "wpi": 4,
        "rochester": 33,
        "master": 7,
	}
*/

func BulkUpdate(begin time.Time, total uint64, bulkSize *int, client *elastic.Client, g *errgroup.Group, ctx context.Context, docsc <-chan common.MapStr) error {
	// Second goroutine will consume the documents sent from the first and bulk insert into ES
	// g.Go(func() error {
	bulk := client.Bulk()
	for doc := range docsc {
		// Simple progress
		current := atomic.AddUint64(&total, 1)
		dur := time.Since(begin).Seconds()
		sec := int(dur)
		pps := int64(float64(current) / dur)
		fmt.Printf("%10d | %6d req/s | %02d:%02d\r", current, pps, sec/60, sec%60)

		//convert string d to json
		//var doc ds.DelDoc
		//json.Unmarshal([]byte(d), &doc)
		d, err := doc.GetValue("doc")
		if err != nil {
			logp.Err("Get #%v error doc content from with index type id mapstr:", err)
			return err
		}

		//for do not except type assertion panic , so tpye judge first!!
        indexV, ok := doc["Index"].(string)
        if !ok {
			logp.Err("Get #%v error doc content from with index type id mapstr:", err)
			return err
        }
        fmt.Printf("type assert index= %s", indexV)
		fmt.Printf("json: index:%s, type:%s, id:%s,   doc: %s", doc["Index"].(string), doc["Type"].(string), doc["ID"].(string), d)

		// Enqueue the document notice why &d === d the same useful
		//bulk.Add(elastic.NewBulkUpdateRequest().Index(doc["Index"].(string)).Type(doc["Type"].(string)).Id(doc["ID"].(string)).Doc(&d))
//will panic doc: {"xxx":"this is doc 1"}panic: interface conversion: interface {} is common.MapStr, not map[string]interface {}
//		bulk.Add(elastic.NewBulkUpdateRequest().Index(doc["Index"].(string)).Type(doc["Type"].(string)).Id(doc["ID"].(string)).Doc(d.(map[string]interface{})))

		bulk.Add(elastic.NewBulkUpdateRequest().Index(doc["Index"].(string)).Type(doc["Type"].(string)).Id(doc["ID"].(string)).Doc(d.(common.MapStr)))
		if bulk.NumberOfActions() >= *bulkSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				fmt.Errorf("Bulk Do stan error: %s\n", err)
				return err
			}
			if res.Errors {
				// Look up the failed documents with res.Failed(), and e.g. recommit
				fmt.Errorf("Bulk update error. whaer?\n")
				fmt.Printf("res:: %v ", res)
				return errors.New("bulk commit failed")
			}
			//finished update ok
		}

		select {
		default:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Commit the final batch before exiting
	if bulk.NumberOfActions() > 0 {
		_, err := bulk.Do(ctx)
		if err != nil {
			return err
		}
	}
	//   return nil
	//})

	return nil

}
