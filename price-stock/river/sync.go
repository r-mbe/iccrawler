package river

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/juju/errors"
	"github.com/ngaut/log"

	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/schema"

	"techtoolkit.ickey.cn/price-stock/es"
)

const (
	syncInsertDoc = iota
	syncDeleteDoc
	syncUpdateDoc
)

const (
	fieldTypeList = "list"
)

type rowsEventHandler struct {
	r *River
}

func (h *rowsEventHandler) Do(e *canal.RowsEvent) error {
	rule, ok := h.r.rules[ruleKey(e.Table.Schema, e.Table.Name)]
	if !ok {
		return nil
	}

	var reqs []*es.BulkRequest
	var err error
	switch e.Action {
	case canal.InsertAction:
		//reqs, err = h.r.makeInsertRequest(rule, e.Rows)
		//only stock update action.

		if strings.EqualFold(e.Table.Name, "t_product") {
			reqs, err = h.r.makeInsertRequest(rule, e.Rows)
		} else {
			logp.Info("action canal.InsertAction... Un support table : %v.", e.Table.Name)
			return nil
		}

	case canal.DeleteAction:
		//reqs, err = h.r.makeInsertRequest(rule, e.Rows)
		//only stock update action.
		// reqs, err = h.r.makeDeleteRequest(rule, e.Rows)

		if strings.EqualFold(e.Table.Name, "t_product") {
			reqs, err = h.r.makeDeleteRequest(rule, e.Rows)
		} else {
			logp.Info("action canal.DeleteAction... Un support table: %v.", e.Table.Name)
			return nil
		}
	case canal.UpdateAction:
		//add stan for different table different stragait. only do our need table.
		if strings.EqualFold(e.Table.Name, "t_product") {
			reqs, err = h.r.makeUpdateRequest(rule, e.Rows)
		} else if strings.EqualFold(e.Table.Name, "t_pro_sell_stock") {
			reqs, err = h.r.makeStockUpdateRequest(rule, e.Rows)

			if reqs, err = h.r.es.PrepareIndex(reqs, h.r.c.EsIndexs); err != nil {
				log.Errorf("do prepare ES bulks new indexs err %v, stop", err)
				return errors.Errorf("make %s  do prepare ES bulks new indexs err %v", e.Action, err)
			}
		} else if strings.EqualFold(e.Table.Name, "t_pro_sell") {
			reqs, err = h.r.makeTProSellUpdateRequest(rule, e.Rows)

			if reqs, err = h.r.es.PrepareOfflineIndex(reqs, h.r.c.EsIndexs); err != nil {
				log.Errorf("do prepare ES bulks new indexs err %v, stop", err)
				return errors.Errorf("make %s  do prepare ES bulks new indexs err %v", e.Action, err)
			}
		} else {
			log.Info("un support tablename=%v", e.Table.Name)
			return nil
		}

	default:
		return errors.Errorf("invalid rows action %s", e.Action)
	}

	if err != nil {
		return errors.Errorf("make %s ES request err %v", e.Action, err)
	}

	if 0 == len(reqs) {
		return nil
	}

	//filter id==sku not format 14len

	if err := h.r.doBulk(reqs); err != nil {
		log.Errorf("do ES bulks err %v, stop", err)
		return canal.ErrHandleInterrupted
	}

	return nil
}

func (h *rowsEventHandler) String() string {
	return "ESRiverRowsEventHandler"
}

// for insert and delete
func (r *River) makeRequest(rule *Rule, action string, rows [][]interface{}) ([]*es.BulkRequest, error) {
	reqs := make([]*es.BulkRequest, 0, len(rows))

	for _, values := range rows {
		id, err := r.getDocID(rule, values)
		if err != nil {
			return nil, errors.Trace(err)
		}

		parentID := ""
		if len(rule.Parent) > 0 {
			if parentID, err = r.getParentID(rule, values, rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: id, Parent: parentID}

		if action == canal.DeleteAction {
			req.Action = es.ActionDelete
			r.st.DeleteNum.Add(1)
		} else {
			r.makeInsertReqData(req, rule, values)
			r.st.InsertNum.Add(1)
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (r *River) makeInsertRequest(rule *Rule, rows [][]interface{}) ([]*es.BulkRequest, error) {
	return r.makeRequest(rule, canal.InsertAction, rows)
}

func (r *River) makeDeleteRequest(rule *Rule, rows [][]interface{}) ([]*es.BulkRequest, error) {
	return r.makeRequest(rule, canal.DeleteAction, rows)
}

func (r *River) makeUpdateRequest(rule *Rule, rows [][]interface{}) ([]*es.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*es.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {
		beforeID, err := r.getDocID(rule, rows[i])
		if err != nil {
			return nil, errors.Trace(err)
		}

		afterID, err := r.getDocID(rule, rows[i+1])

		if err != nil {
			return nil, errors.Trace(err)
		}

		beforeParentID, afterParentID := "", ""
		if len(rule.Parent) > 0 {
			if beforeParentID, err = r.getParentID(rule, rows[i], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
			if afterParentID, err = r.getParentID(rule, rows[i+1], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = es.ActionDelete
			reqs = append(reqs, req)

			req = &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
			r.makeInsertReqData(req, rule, rows[i+1])

			r.st.DeleteNum.Add(1)
			r.st.InsertNum.Add(1)
		} else {
			r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			r.st.UpdateNum.Add(1)
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (r *River) makeStockUpdateRequest(rule *Rule, rows [][]interface{}) ([]*es.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*es.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {
		beforeID, err := r.getDocID(rule, rows[i])
		if err != nil {
			return nil, errors.Trace(err)
		}

		afterID, err := r.getDocID(rule, rows[i+1])

		if err != nil {
			return nil, errors.Trace(err)
		}

		beforeParentID, afterParentID := "", ""
		if len(rule.Parent) > 0 {
			if beforeParentID, err = r.getParentID(rule, rows[i], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
			if afterParentID, err = r.getParentID(rule, rows[i+1], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = es.ActionDelete
			reqs = append(reqs, req)

			req = &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
			r.makeInsertReqData(req, rule, rows[i+1])

			r.st.DeleteNum.Add(1)
			r.st.InsertNum.Add(1)
		} else {
			r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			r.st.UpdateNum.Add(1)

			var stockNum int32
			if _, ok := req.Data["stock_num"]; ok {
				t := req.Data["stock_num"].(int32)
				stockNum = t
			} else {
				t := rows[i][1].(int32)
				stockNum = t
			}

			var frozenNum int32
			if _, ok := req.Data["frozen_num"]; ok {
				t := req.Data["frozen_num"].(int32)
				frozenNum = t
			} else {
				t := rows[i][2].(int32)
				frozenNum = t
			}

			var virtualNum int32
			if _, ok := req.Data["virtual_num"]; ok {
				t := req.Data["virtual_num"].(int32)
				virtualNum = t
			} else {
				t := rows[i][3].(int32)
				virtualNum = t
			}

			// var stockType int64
			// if _, ok := req.Data["stock_type"]; ok {
			// 	stockType = (int64)(req.Data["stock_type"].(int))
			// } else {
			// 	stockType = row[i][4]
			// }

			//clear all key in req.Data
			for k, _ := range req.Data {
				delete(req.Data, k)
			}

			//put new req Data
			req.Data["stock"] = stockNum + virtualNum - frozenNum
			//req.Data["stock_type"] = stockType

			// log.Info("==============XXXXXXXXXXX====index=%v id=%v =====make new req Data %v", req.Index, req.ID, req.Data)

		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

//+! only t_pro_sell status changed from 1 to 0--> offline need river to es  other use the common river update all data.
func (r *River) makeTProSellUpdateRequest(rule *Rule, rows [][]interface{}) ([]*es.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*es.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {
		beforeID, err := r.getDocID(rule, rows[i])
		if err != nil {
			return nil, errors.Trace(err)
		}

		afterID, err := r.getDocID(rule, rows[i+1])

		if err != nil {
			return nil, errors.Trace(err)
		}

		beforeParentID, afterParentID := "", ""
		if len(rule.Parent) > 0 {
			if beforeParentID, err = r.getParentID(rule, rows[i], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
			if afterParentID, err = r.getParentID(rule, rows[i+1], rule.Parent); err != nil {
				return nil, errors.Trace(err)
			}
		}

		req := &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = es.ActionDelete
			reqs = append(reqs, req)

			req = &es.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
			r.makeInsertReqData(req, rule, rows[i+1])

			r.st.DeleteNum.Add(1)
			r.st.InsertNum.Add(1)
		} else {
			r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			r.st.UpdateNum.Add(1)
		}

		if _, ok := req.Data["status"]; ok {
			if 0 == req.Data["status"].(int8) {
				// log.Info("==============XXXXXXXXXXX====index=%v id=%v =====make new req Data %v", req.Index, req.ID, req.Data)
				reqs = append(reqs, req)
			}
		}

	}

	return reqs, nil
}

func (r *River) makeReqColumnData(col *schema.TableColumn, value interface{}) interface{} {
	switch col.Type {
	case schema.TYPE_ENUM:
		switch value := value.(type) {
		case int64:
			// for binlog, ENUM may be int64, but for dump, enum is string
			eNum := value - 1
			if eNum < 0 || eNum >= int64(len(col.EnumValues)) {
				// we insert invalid enum value before, so return empty
				log.Warnf("invalid binlog enum index %d, for enum %v", eNum, col.EnumValues)
				return ""
			}

			return col.EnumValues[eNum]
		}
	case schema.TYPE_SET:
		switch value := value.(type) {
		case int64:
			// for binlog, SET may be int64, but for dump, SET is string
			bitmask := value
			sets := make([]string, 0, len(col.SetValues))
			for i, s := range col.SetValues {
				if bitmask&int64(1<<uint(i)) > 0 {
					sets = append(sets, s)
				}
			}
			return strings.Join(sets, ",")
		}
	case schema.TYPE_BIT:
		switch value := value.(type) {
		case string:
			// for binlog, BIT is int64, but for dump, BIT is string
			// for dump 0x01 is for 1, \0 is for 0
			if value == "\x01" {
				return int64(1)
			}

			return int64(0)
		}
	case schema.TYPE_STRING:
		switch value := value.(type) {
		case []byte:
			return string(value[:])
		}
	}

	return value
}

func (r *River) getFieldParts(k string, v string) (string, string, string) {
	composedField := strings.Split(v, ",")

	mysql := k
	es := composedField[0]
	fieldType := ""

	if 0 == len(es) {
		es = mysql
	}
	if 2 == len(composedField) {
		fieldType = composedField[1]
	}

	return mysql, es, fieldType
}

func (r *River) makeInsertReqData(req *es.BulkRequest, rule *Rule, values []interface{}) {
	req.Data = make(map[string]interface{}, len(values))
	req.Action = es.ActionIndex

	for i, c := range rule.TableInfo.Columns {
		if !rule.CheckFilter(c.Name) {
			continue
		}
		mapped := false
		for k, v := range rule.FieldMapping {
			mysql, es, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				v := r.makeReqColumnData(&c, values[i])
				if fieldType == fieldTypeList {
					if str, ok := v.(string); ok {
						req.Data[es] = strings.Split(str, ",")
					} else {
						req.Data[es] = v
					}
				} else {
					req.Data[es] = v
				}
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, values[i])
		}
	}
}

func (r *River) makeUpdateReqData(req *es.BulkRequest, rule *Rule,
	beforeValues []interface{}, afterValues []interface{}) {
	req.Data = make(map[string]interface{}, len(beforeValues))

	// maybe dangerous if something wrong delete before?
	req.Action = es.ActionUpdate

	for i, c := range rule.TableInfo.Columns {
		mapped := false
		if !rule.CheckFilter(c.Name) {
			continue
		}
		if reflect.DeepEqual(beforeValues[i], afterValues[i]) {
			//nothing changed
			continue
		}
		for k, v := range rule.FieldMapping {
			mysql, es, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				// has custom field mapping
				v := r.makeReqColumnData(&c, afterValues[i])
				str, ok := v.(string)
				if ok == false {
					req.Data[c.Name] = v
				} else {
					if fieldType == fieldTypeList {
						req.Data[es] = strings.Split(str, ",")
					} else {
						req.Data[es] = str
					}
				}
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, afterValues[i])
		}

	}
}

// Get primary keys in one row and format them into a string
// PK must not be nil
func (r *River) getDocID(rule *Rule, row []interface{}) (string, error) {
	pks, err := canal.GetPKValues(rule.TableInfo, row)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	sep := ""
	for i, value := range pks {
		if value == nil {
			return "", errors.Errorf("The %ds PK value is nil", i)
		}

		buf.WriteString(fmt.Sprintf("%s%v", sep, value))
		sep = ":"
	}

	return buf.String(), nil
}

func (r *River) getParentID(rule *Rule, row []interface{}, columnName string) (string, error) {
	index := rule.TableInfo.FindColumn(columnName)
	if index < 0 {
		return "", errors.Errorf("parent id not found %s(%s)", rule.TableInfo.Name, columnName)
	}

	return fmt.Sprint(row[index]), nil
}

func (r *River) doBulk(reqs []*es.BulkRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	if _, err := r.es.Bulk(reqs); err != nil {
		log.Errorf("sync docs err %v after binlog %s", err, r.canal.SyncedPosition())
		return errors.Trace(err)
	}

	//else if nil != resp {
	//	log.Info("bulk response %v ", resp)
	//}

	return nil
}
