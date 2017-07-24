package seed

import (
	"fmt"
	"log"
	"testing"
)

func TestRootLinksGet(t *testing.T) {
	u := "http://www.anglia-live.com/products/cases-enclosures/enclosure-accessories/accessories/1193644001_din-clip-metal-100x8x51-clr"
	seed := &Seed{}
	urls, err := seed.RootLinksGet(u)
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range urls {
		fmt.Printf("a[%d] = %s\n", i, v)
		t.Logf("xxxx t.log===%s", v)
	}

}
