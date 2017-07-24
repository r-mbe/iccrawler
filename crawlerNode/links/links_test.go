package links

import (
	"fmt"
	"testing"
)

//TestGetListPagesNum
func TestGetListPagesNum(t *testing.T) {
	out := make(chan string)
	// done := make(chan struct{})
	u := "http://www.anglia-live.com/products/circuit-protection/fuses-breakers-holders/resettable-fuses"

	GetListPagesNum(u, out)

	for {
		select {
		case link, ok := <-out:
			if ok {
				fmt.Printf("get list pages url=:%s \n", link)
			} else {
				return
			}
		}

	}
	// done <- struct{}{}
}
