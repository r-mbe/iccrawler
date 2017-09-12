package seed

import "techtoolkit.ickey.cn/crawlerPassive/request"

//Seed seed struct
type Seed struct {
}

//RootLinksGet find root seed urs []string seed.
func (s *Seed) RootLinksGet(url string) ([]string, error) {

	return request.GetCatListFromNodeJS(url)
}

//GetPagesFromNodeJS seed get pages
func (s *Seed) GetPagesFromNodeJS(ur string) (string, int, string, error) {
	return request.GetPagesFromNodeJS(ur)
}

//GetPageListDetail seed get pages
func (s *Seed) GetPageListDetail(ur string) ([]request.PartNumber, error) {
	return request.GetPageListDetail(ur)
}
