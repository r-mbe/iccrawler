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

        //"github.com/elastic/beats/libbeat/common"

	"github.com/stanxii/indexRiver/ds"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"gopkg.in/olivere/elastic.v5"
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

//func  (bd *DelBulkClient) bulkDelete(total *uint64, client *elastic.Client, ctx *context.Context, docsc <-chan ds.DelDoc) error {
func BulkDelete(begin time.Time, total uint64, bulkSize *int, client *elastic.Client, g *errgroup.Group, ctx context.Context, docsc <-chan ds.DelDoc) error {
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

		// Enqueue the document
		//bulk.Add(elastic.NewBulkDeleteRequest().Index(*index).Type(*typ).Id(d.ID)
		bulk.Add(elastic.NewBulkDeleteRequest().Index(doc.Index).Type(doc.Type).Id(doc.ID))
		if bulk.NumberOfActions() >= *bulkSize {
			// Commit
			res, err := bulk.Do(ctx)
			if err != nil {
				fmt.Errorf("Bulk Do stan error: %s\n", err)
				return err
			}
			if res.Errors {
				// Look up the failed documents with res.Failed(), and e.g. recommit
				return errors.New("bulk commit failed")
			}
			// "bulk" is reset after Do, so you can reuse it
			// Document with Id="1" should not exist
			exists, err := client.Exists().Index(doc.Index).Type(doc.Type).Id("1").Do(context.TODO())
			if err != nil {
				fmt.Errorf("Bulk Do stan error index  why exist?: %s\n", err)
			}
			if exists {
				fmt.Errorf("expected exists %v; got %v", false, exists)
			}

			// Document with Id="2" should exist
			exists, err = client.Exists().Index(doc.Index).Type(doc.Type).Id("2").Do(context.TODO())
			if err != nil {
				fmt.Errorf("index expected exists but  why exist?: %s\n", err)
			}
			if !exists {
				fmt.Errorf("expected exists %v; got %v", true, exists)
			}
			//finished check result if delete ok
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
