package main

import (
//	"encoding/base64"
 //   "errors"
    "flag"
    "fmt"
    "log"
    "math/rand"
//  "sync/atomic"
    "time"

    "golang.org/x/net/context"
    "golang.org/x/sync/errgroup"
    "gopkg.in/olivere/elastic.v5"

    "github.com/stanxii/indexRiver/ds"
    "github.com/stanxii/indexRiver/es"
)

func main() {
	fmt.Println("bulk delete!")

	var (
        url      = flag.String("url", "http://localhost:9200", "Elasticsearch URL")
        index    = flag.String("index", "", "Elasticsearch index name")
        typ      = flag.String("type", "", "Elasticsearch type name")
        sniff    = flag.Bool("sniff", true, "Enable or disable sniffing")
        n        = flag.Int("n", 0, "Number of documents to bulk insert")
        bulkSize = flag.Int("bulk-size", 0, "Number of documents to collect before committing")
    )
    flag.Parse()
    log.SetFlags(0)
    rand.Seed(time.Now().UnixNano())

    if *url == "" {
        log.Fatal("missing url parameter")
    }
    if *index == "" {
        log.Fatal("missing index parameter")
    }
    if *typ == "" {
        log.Fatal("missing type parameter")
    }
    if *n <= 0 {
        log.Fatal("n must be a positive number")
    }
    if *bulkSize <= 0 {
        log.Fatal("bulk-size must be a positive number")
    }

    // Create an Elasticsearch client
    client, err := elastic.NewClient(elastic.SetURL(*url), elastic.SetSniff(*sniff))
    if err != nil {
        log.Fatal(err)
    }

    // Setup a group of goroutines from the excellent errgroup package
    g, ctx := errgroup.WithContext(context.TODO())

    

    docsc := make(chan ds.DelDoc)

    // First goroutine will emit documents will delete and sent it to the second goroutine
     begin := time.Now()

    // Goroutine to create documents
    g.Go(func() error {
        defer close(docsc)

		var dels = []ds.DelDoc {
		        {Index:"warehouse", Type:"product", ID:"yPG0TITxHbRjlW8GlDBVZF4lMiKaJx5GOecRgrlior4=",},
			    {Index:"warehouse", Type:"product", ID:"5Q9ojNPdIqr1eesQVg1H3n2zXQpTZZhSvbCVqrzAwws=",},
				{Index:"warehouse", Type:"product", ID:"zb_lbAGycwbhCR2SYaDZHrUKW9ni5Tfwd9ISZ822Jl8=",},
        }

	//	var s DelDocs
	//	json.Unmarshal([]byte(str), &s)	

		for _, d := range dels {
            // Construct the document {index, type, id} d

            // Send over to 2nd goroutine, or cancel
            select {
            case docsc <- d:
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        return nil
    })



	// Second goroutine
	var total uint64
	es.BulkDelete(begin, total, bulkSize, client, g, ctx, docsc)

	   // Wait until all goroutines are finished
    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }

    // Final results
    dur := time.Since(begin).Seconds()
    sec := int(dur)
    pps := int64(float64(total) / dur)
    fmt.Printf("%10d | %6d req/s | %02d:%02d\n", total, pps, sec/60, sec%60)

	
}
