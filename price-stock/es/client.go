package es

import (
	"context"
	"encoding/json"

	mainlog "github.com/cihub/seelog"
	"github.com/juju/errors"
	"github.com/ngaut/log"

	"gopkg.in/olivere/elastic.v5"
)

// Although there are many Elasticsearch clients with Go, I still want to implement one by myself.
// Because we only need some very simple usages.
type Client struct {
	esc    *elastic.Client
	Seelog *mainlog.LoggerInterface
}

//+ New es client wrapper
func NewClient(esServer string) (*Client, error) {
	c := new(Client)

	log.Infof("es server string =%v", esServer)
	esc, err := elastic.NewClient(elastic.SetURL(esServer), elastic.SetSniff(true))
	if err != nil || nil == esc {
		errors.Errorf("es.client package connection es...... error: %v", err)
		return nil, err
	}

	//init seelog
	Seelog, _ := mainlog.LoggerFromConfigAsFile("etc/seelog-price-stock.xml")

	// seelog.Info("需要输入的日志")

	c.esc = esc
	c.Seelog = &Seelog

	return c, nil
}

//-

type ResponseItem struct {
	ID      string                 `json:"_id"`
	Index   string                 `json:"_index"`
	Type    string                 `json:"_type"`
	Version int                    `json:"_version"`
	Found   bool                   `json:"found"`
	Source  map[string]interface{} `json:"_source"`
}

type Response struct {
	Code int
	ResponseItem
}

// See http://www.elasticsearch.org/guide/en/elasticsearch/guide/current/bulk.html
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionIndex  = "index"
)

type BulkRequest struct {
	Action string
	Index  string
	Type   string
	ID     string
	Parent string

	Data map[string]interface{}
}

//stock sku == es index id
func (c *Client) PrepareIndex(reqs []*BulkRequest, indexs []string) ([]*BulkRequest, error) {
	var res []*BulkRequest
	for _, r := range reqs {

		//filter format sku need len >= 14  in production.
		if len(r.ID) < 14 || len(r.Data) <= 0 {

			//for test db sku not alwarys len=14
			//log.Info("sku len <=14 or map len <=0 ignore it!! %v", r.Data)
			continue
		}

		for _, index := range indexs {

			exists, err := c.esc.Exists().Index(index).Type(r.Type).Id(r.ID).Do(context.TODO())
			// exists, err := c.esc.Exists().Index(index).Type(r.Type).Id("100303286580").Do(context.TODO())
			if err != nil {
				//log.Info("Err client.go Parperent bulk Req=%v Elastics exists check id error:%v", r, err)
				//return is, err
				continue
			}

			if exists {
				//exists
				//log.Info("exists......prepare Will bulk index=%v  data=%v", index, r.Data)
				//we finally find sku es id return immedially
				r.Index = index
				res = append(res, r)
				return res, nil
			}
		}
	}

	return nil, nil
}

//t_pro_sell id === ESindex's pid
func (c *Client) PrepareOfflineIndex(reqs []*BulkRequest, indexs []string) ([]*BulkRequest, error) {
	var res []*BulkRequest
	for _, r := range reqs {

		//filter format sku need len >= 14  in production.

		if len(r.ID) < 0 || len(r.Data) <= 0 {

			//for test db sku not alwarys len=14
			//log.Info("sku len <=14 or map len <=0 ignore it!! %v", r.Data)
			continue
		}

		for _, index := range indexs {

			termQuery := elastic.NewTermQuery("pid", r.ID)
			searchResult, err := c.esc.Search().
				Index(index).
				Query(termQuery).
				Do(context.TODO())

			if nil != err {
				//this index no pid data.
				continue
			}

			if searchResult.Hits.TotalHits > 0 {
				//we found the first index include the pid not need others index.
				// Iterate through results
				for _, hit := range searchResult.Hits.Hits {
					// hit.Index contains the name of the index

					r.Index = hit.Index
					r.Type = hit.Type
					//get index and id by pid
					r.ID = hit.Id
					res = append(res, r)
					return res, nil
				}
			}

		}
	}

	return nil, nil
}

func (c *Client) exists(r *BulkRequest, indexs []string) ([]string, error) {

	var is []string

	for _, index := range indexs {
		exists, err := c.esc.Exists().Index(index).Type(r.Type).Id(r.ID).Do(context.TODO())
		if err != nil {
			log.Info("XXX client.go Elastics exists check id error:%v", err)
			return is, err
		}

		if exists {
			//exists
			is = append(is, index)
		}
	}

	return is, nil
}

//!+
func (c *Client) translateData(r *BulkRequest) (map[string]interface{}, error) {
	doc := make(map[string]interface{})

	data, err := json.Marshal(r.Data)
	if err != nil {
		return nil, errors.Trace(err)
	}

	log.Info("XXX Update translateData index=%v, type=%v id=%v doc=%v", r.Index, r.Type, r.ID, string(data))

	log.Info("translate old data=%v", string(data))
	log.Info("translate old data=%v", r.Data)

	doc["sku"] = r.Data["sku"]

	log.Info("sku = %v stocknum =%v  type=%T", r.Data["sku"], r.Data["stock_num"], r.Data["stock_num"])

	stock := r.Data["stock_num"].(int32) + r.Data["virtual_num"].(int32) - r.Data["frozen_num"].(int32)

	if stock < 0 {
		log.Info("stock < 0 stocknum:%v  %T  ,  virtual_num=%v,  frozen_num=%v", r.Data["stock_num"], r.Data["stock_num"], r.Data["virtual_num"], r.Data["frozen_num"])
		stock = 0
	}

	log.Info("after compute stock=%d", stock)
	doc["stock"] = stock

	return doc, nil
}

//!-

//need translate bulk e.RowEvent data to es special data.
func (c *Client) bulk(bulk *elastic.BulkService, r *BulkRequest) error {

	if len(r.Index) <= 0 || len(r.Type) <= 0 {
		errors.Errorf("r.Index or Type len <= 0.")
		return nil
	}

	switch r.Action {
	case ActionDelete:
		//nothing to do
		//log.Info("Delete now will add update bulk request index=%v, type=%v id=%v ", r.Index, r.Type, r.ID)
		bulk.Add(elastic.NewBulkDeleteRequest().Index(r.Index).Type(r.Type).Id(r.ID))
	case ActionUpdate:
		// doc := map[string]interface{}{
		// 	"doc": r.Data,
		// }

		data, err := json.Marshal(r.Data)
		if err != nil {
			return errors.Trace(err)
		}

		//translate row data to update stock struct
		// doc, err := c.translateData(r)
		// if err != nil {
		// 	log.Errorf("translateData error old=%v, err=%v", r.Data, err)
		// 	return nil
		// }

		//log.Info("XXX Update now will add update bulk request index=%v, type=%v id=%v doc=%v", r.Index, r.Type, r.ID, string(doc))
		//es update doc must is map[string]interface{}{"doc":{}}  not a string

		//before prepareIndex will compute correct index
		//log.Info("Now== add bulk update req into index=%v, type=%v, id=%v, r.Data=%v string(r.data)=%v", r.Index, r.Type, r.ID, r.Data, string(data))

		seelog := *c.Seelog
		seelog.Info("Now== add bulk update req into index=%v, type=%v, id=%v, r.Data=%v string(r.data)=%v", r.Index, r.Type, r.ID, r.Data, string(data))

		bulk.Add(elastic.NewBulkUpdateRequest().Index(r.Index).Type(r.Type).Id(r.ID).Doc(r.Data))

	default:
		// doc := map[string]interface{}{
		// 	"doc": r.Data,
		// }
		// data, err := json.Marshal(doc)
		// if err != nil {
		// 	return errors.Trace(err)
		// }
		//for create and index
		//log.Info("XXX Insert now will add update bulk request index=%v, type=%v id=%v doc=%v", r.Index, r.Type, r.ID, string(data))
		bulk.Add(elastic.NewBulkIndexRequest().Index(r.Index).Type(r.Type).Id(r.ID).Doc(r.Data))

	}

	return nil
}

func (c *Client) Do(bulk *elastic.BulkService) (*elastic.BulkResponse, error) {
	var res *elastic.BulkResponse
	var err error
	res, err = bulk.Do(context.TODO())
	if err != nil {
		return nil, err
	}

	if nil == res {
		log.Infof("Now Bul res = nil res=%v 1\n", err, res)
		return nil, nil
	}

	return res, nil
}

func (c *Client) DoBulk(bulk *elastic.BulkService, items []*BulkRequest) (*elastic.BulkResponse, error) {

	for _, item := range items {
		//log.Infof("Now call Client Itera itera loop loop _____________ item=%v step 1\n", item)
		if err := c.bulk(bulk, item); err != nil {
			return nil, errors.Trace(err)
		}
	}

	resp, err := c.Do(bulk)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if bulk.NumberOfActions() != 0 {
		log.Info("expect bulk.NumberOfActions=%d got %d\n", 0, bulk.NumberOfActions())
	}

	//log.Info("after check expect bulk.NumberOfActions=%d got %d\n", 0, bulk.NumberOfActions())

	return resp, err
}

// main es Bulk api
// only support parent in 'Bulk' related apis
func (c *Client) Bulk(items []*BulkRequest) (*elastic.BulkResponse, error) {

	bulk := c.esc.Bulk()
	return c.DoBulk(bulk, items)
}
