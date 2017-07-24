package river

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/siddontang/go-mysql/schema"
	"techtoolkit.ickey.cn/cockimport/cockroach"
)

const (
	syncInsertDoc = iota
	syncDeleteDoc
	syncUpdateDoc
)

const (
	fieldTypeList = "list"
)

type posSaver struct {
	pos   mysql.Position
	force bool
}

type eventHandler struct {
	r *River
}

func (h *eventHandler) OnRotate(e *replication.RotateEvent) error {
	pos := mysql.Position{
		string(e.NextLogName),
		uint32(e.Position),
	}

	h.r.syncCh <- posSaver{pos, true}

	return h.r.ctx.Err()
}

func (h *eventHandler) OnDDL(nextPos mysql.Position, _ *replication.QueryEvent) error {
	h.r.syncCh <- posSaver{nextPos, true}
	return h.r.ctx.Err()
}

func (h *eventHandler) OnXID(nextPos mysql.Position) error {
	h.r.syncCh <- posSaver{nextPos, false}
	return h.r.ctx.Err()
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	rule, ok := h.r.rules[ruleKey(e.Table.Schema, e.Table.Name)]
	if !ok {
		return nil
	}

	var reqs []*cockroach.BulkRequest
	var err error
	switch e.Action {
	case canal.InsertAction:
		fmt.Printf("####XXXX#### canal.InsertAction: table: %v\n", e.Table.Name)
		if strings.EqualFold(e.Table.Name, "t_pro_sell_stock") || strings.EqualFold(e.Table.Name, "t_pro_sell_price") {
			reqs, err = h.r.makeInsertRequest(rule, e.Rows)
		} else {
			seelog := *h.r.cock.Seelog
			seelog.Errorf("action canal.InsertAction: %v", err.Error())
			return nil
		}
	case canal.DeleteAction:
		fmt.Printf("####XXXX#### canal.DeleteAction: table: %v\n", e.Table.Name)
		if strings.EqualFold(e.Table.Name, "t_pro_sell_stock") || strings.EqualFold(e.Table.Name, "t_pro_sell_price") {
			reqs, err = h.r.makeDeleteRequest(rule, e.Rows)
		} else {
			seelog := *h.r.cock.Seelog
			seelog.Errorf("action.DeleteAction un support table %v", err.Error())
			return nil
		}
	case canal.UpdateAction:
		fmt.Printf("####XXXX#### canal.UpdateAction: table: %v\n", e.Table.Name)
		if strings.EqualFold(e.Table.Name, "t_pro_sell_stock") {
			reqs, err = h.r.makeStockUpdateRequest(rule, e.Rows)
			if err != nil {
				fmt.Printf("makeStockUpdateRequest err canal.UpdateAction: table: %v\n", e.Table.Name)
				seelog := *h.r.cock.Seelog
				seelog.Errorf("makeStockUpdateRequest: %v", err.Error())
				return errors.Errorf("make %s makeStockUpdateRequest  err %v", e.Action, err)
			}
		} else if strings.EqualFold(e.Table.Name, "t_pro_sell_price") {
			reqs, err = h.r.makePriceUpdateRequest(rule, e.Rows)
			if err != nil {
				seelog := *h.r.cock.Seelog
				seelog.Errorf("makePriceUpdateRequest: %v", err.Error())
				return errors.Errorf("make %s makePriceUpdateRequest  err %v", e.Action, err)
			}
		} else {
			seelog := *h.r.cock.Seelog
			seelog.Errorf("un support tablename makePriceUpdateRequest: %v", e.Table.Name)
			return nil
		}
	default:
		err = errors.Errorf("invalid rows action %s", e.Action)
	}

	if err != nil {
		h.r.cancel()
		return errors.Errorf("make %s ES request err %v, close sync", e.Action, err)
	}

	h.r.syncCh <- reqs

	return h.r.ctx.Err()
}

func (h *eventHandler) String() string {
	return "ESRiverEventHandler"
}

func (r *River) syncLoop() {
	bulkSize := r.c.BulkSize
	if bulkSize == 0 {
		bulkSize = 128
	}

	interval := r.c.FlushBulkTime.Duration
	if interval == 0 {
		interval = 200 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer r.wg.Done()

	lastSavedTime := time.Now()
	reqs := make([]*cockroach.BulkRequest, 0, 1024)

	var pos mysql.Position

	for {
		needFlush := false
		needSavePos := false

		select {
		case v := <-r.syncCh:
			switch v := v.(type) {
			case posSaver:
				now := time.Now()
				if v.force || now.Sub(lastSavedTime) > 3*time.Second {
					lastSavedTime = now
					needFlush = true
					needSavePos = true
					pos = v.pos
				}
			case []*cockroach.BulkRequest:
				reqs = append(reqs, v...)
				needFlush = len(reqs) >= bulkSize
			}
		case <-ticker.C:
			needFlush = true
		case <-r.ctx.Done():
			return
		}

		if needFlush {
			// TODO: retry some times?
			if err := r.doBulk(reqs); err != nil {
				log.Errorf("do ES bulk err %v, close sync", err)
				r.cancel()
				return
			}
			reqs = reqs[0:0]
		}

		if needSavePos {
			if err := r.master.Save(pos); err != nil {
				log.Errorf("save sync position %s err %v, close sync", pos, err)
				r.cancel()
				return
			}
		}
	}
}

// for insert and delete
func (r *River) makeRequest(rule *Rule, action string, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	reqs := make([]*cockroach.BulkRequest, 0, len(rows))

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

		req := &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: id, Parent: parentID}

		if action == canal.DeleteAction {
			req.Action = cockroach.ActionDelete
			r.st.DeleteNum.Add(1)
		} else {
			r.makeInsertReqData(req, rule, values)
			r.st.InsertNum.Add(1)
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (r *River) makeInsertRequest(rule *Rule, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	return r.makeRequest(rule, canal.InsertAction, rows)
}

func (r *River) makeDeleteRequest(rule *Rule, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	return r.makeRequest(rule, canal.DeleteAction, rows)
}

func (r *River) makeUpdateRequest(rule *Rule, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*cockroach.BulkRequest, 0, len(rows))

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

		req := &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = cockroach.ActionDelete
			reqs = append(reqs, req)

			req = &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
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

func (r *River) makeStockUpdateRequest(rule *Rule, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*cockroach.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {

		ibeforeID, err := r.getStanDocID(rule, rows[i])
		beforeID := string(ibeforeID)

		if err != nil {
			return nil, errors.Trace(err)
		}

		iafterID, err := r.getStanDocID(rule, rows[i+1])
		afterID := string(iafterID)
		fmt.Printf("i=%d  beforeID=%v afterID=%v \n", i, beforeID, afterID)
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

		req := &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = cockroach.ActionDelete
			reqs = append(reqs, req)

			req = &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
			r.makeInsertReqData(req, rule, rows[i+1])

			r.st.DeleteNum.Add(1)
			r.st.InsertNum.Add(1)
		} else {
			r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			r.st.UpdateNum.Add(1)
			for k, _ := range req.Data {
				fmt.Println("mapppppp stock k = ppppppppppppppppppppppppppppppppppppppp%v \n", k)
				//delete(req.Data, k)
			}

			req.Data["sku"] = rows[i][0].(int64)
			//customer print column data
			fmt.Printf("## XXX req.Data = %v\n", req.Data)

		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (r *River) makePriceUpdateRequest(rule *Rule, rows [][]interface{}) ([]*cockroach.BulkRequest, error) {
	if len(rows)%2 != 0 {
		return nil, errors.Errorf("invalid update rows event, must have 2x rows, but %d", len(rows))
	}

	reqs := make([]*cockroach.BulkRequest, 0, len(rows))

	for i := 0; i < len(rows); i += 2 {
		beforeID, err := r.getStanDocID(rule, rows[i])
		if err != nil {
			return nil, errors.Trace(err)
		}

		afterID, err := r.getStanDocID(rule, rows[i+1])

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

		req := &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: beforeID, Parent: beforeParentID}

		if beforeID != afterID || beforeParentID != afterParentID {
			req.Action = cockroach.ActionDelete
			reqs = append(reqs, req)

			req = &cockroach.BulkRequest{Index: rule.Index, Type: rule.Type, ID: afterID, Parent: afterParentID}
			r.makeInsertReqData(req, rule, rows[i+1])

			r.st.DeleteNum.Add(1)
			r.st.InsertNum.Add(1)
		} else {
			r.makeUpdateReqData(req, rule, rows[i], rows[i+1])
			r.st.UpdateNum.Add(1)

			for k, _ := range req.Data {
				fmt.Println("mapppppp stock k = ppppppppppppppppppppppppppppppppppppppp%v \n", k)
				//delete(req.Data, k)
			}



			req.Data["sku"] = rows[i][0].(int64)
			//customer print column data
			fmt.Printf("## XXX req.Data = %v\n", req.Data)

		}

		reqs = append(reqs, req)
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
	case schema.TYPE_JSON:
		var f interface{}
		var err error
		switch v := value.(type) {
		case string:
			err = json.Unmarshal([]byte(v), &f)
		case []byte:
			err = json.Unmarshal(v, &f)
		}
		if err == nil && f != nil {
			return f
		}
	}

	return value
}

func (r *River) getFieldParts(k string, v string) (string, string, string) {
	composedField := strings.Split(v, ",")

	mysql := k
	cockroach := composedField[0]
	fieldType := ""

	if 0 == len(cockroach) {
		cockroach = mysql
	}
	if 2 == len(composedField) {
		fieldType = composedField[1]
	}

	return mysql, cockroach, fieldType
}

func (r *River) makeInsertReqData(req *cockroach.BulkRequest, rule *Rule, values []interface{}) {
	req.Data = make(map[string]interface{}, len(values))
	req.Action = cockroach.ActionIndex

	for i, c := range rule.TableInfo.Columns {
		if !rule.CheckFilter(c.Name) {
			continue
		}
		mapped := false
		for k, v := range rule.FieldMapping {
			mysql, elastic, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				v := r.makeReqColumnData(&c, values[i])
				if fieldType == fieldTypeList {
					if str, ok := v.(string); ok {
						req.Data[elastic] = strings.Split(str, ",")
					} else {
						req.Data[elastic] = v
					}
				} else {
					req.Data[elastic] = v
				}
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, values[i])
		}
	}
}

func (r *River) makeUpdateReqData(req *cockroach.BulkRequest, rule *Rule,
	beforeValues []interface{}, afterValues []interface{}) {
	req.Data = make(map[string]interface{}, len(beforeValues))

	// maybe dangerous if something wrong delete before?
	req.Action = cockroach.ActionUpdate

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
			mysql, elastic, fieldType := r.getFieldParts(k, v)
			if mysql == c.Name {
				mapped = true
				// has custom field mapping
				v := r.makeReqColumnData(&c, afterValues[i])
				str, ok := v.(string)
				if ok == false {
					req.Data[c.Name] = v
				} else {
					if fieldType == fieldTypeList {
						req.Data[elastic] = strings.Split(str, ",")
					} else {
						req.Data[elastic] = str
					}
				}
			}
		}
		if mapped == false {
			req.Data[c.Name] = r.makeReqColumnData(&c, afterValues[i])
		}

	}
}

// If id in toml file is none, get primary keys in one row and format them into a string, and PK must not be nil
// Else get the ID's column in one row and format them into a string
func (r *River) getDocID(rule *Rule, row []interface{}) (string, error) {
	var (
		ids []interface{}
		err error
	)
	if rule.ID == nil {
		ids, err = canal.GetPKValues(rule.TableInfo, row)
		if err != nil {
			return "", err
		}
	} else {
		ids = make([]interface{}, 0, len(rule.ID))
		for _, column := range rule.ID {
			// fmt.Printf("GetColumnValue: %v  rule=%v tableinfo: %v  \n", rule.ID, rule, rule.TableInfo)
			value, err := canal.GetColumnValue(rule.TableInfo, column, row)
			//fmt.Printf("####pk===%v \n", value)
			if err != nil {
				return "", err
			}
			ids = append(ids, value)
		}
	}

	var buf bytes.Buffer

	sep := ""
	for i, value := range ids {
		if value == nil {
			return "", errors.Errorf("The %ds id or PK value is nil", i)
		}

		buf.WriteString(fmt.Sprintf("%s%v", sep, value))
		sep = ":"
	}

	return buf.String(), nil
}

getPriceDocID

func (r *River) getStanDocID(rule *Rule, row []interface{}) (int64, error) {
	var (
		ids []interface{}
		err error
	)
	var value interface{}
	if rule.ID == nil {
		ids, err = canal.GetPKValues(rule.TableInfo, row)
		if err != nil {
			return 0, err
		}
		log.Info("get stan doc ids = %v\n", ids)
	} else {
		for _, column := range rule.ID {
			//fmt.Printf("GetColumnValue: %v  rule=%v tableinfo: %v  \n", rule.ID, rule, rule.TableInfo)
			var err error
			value, err = canal.GetColumnValue(rule.TableInfo, column, row)
			fmt.Printf("####pk===%v %T\n", value, value)
			break
			if err != nil {
				//fmt.Printf("####pk err ===%v \n", err)
				return 0, err
			}
		}
	}

	return value.(int64), err

}

func (r *River) getPriceDocID(rule *Rule, row []interface{}) (int64, error) {
	var (
		ids []interface{}
		err error
	)
	var value interface{}
	if rule.ID == nil {
		ids, err = canal.GetPKValues(rule.TableInfo, row)
		if err != nil {
			return 0, err
		}
		log.Info("get stan doc ids = %v\n", ids)
	} else {
		for _, column := range rule.ID {
			//fmt.Printf("GetColumnValue: %v  rule=%v tableinfo: %v  \n", rule.ID, rule, rule.TableInfo)
			var err error
			value, err = canal.GetColumnValue(rule.TableInfo, column, row)
			fmt.Printf("####pk===%v %T\n", value, value)
			break
			if err != nil {
				//fmt.Printf("####pk err ===%v \n", err)
				return 0, err
			}
		}
	}

	return value.(int64), err

}

func (r *River) getParentID(rule *Rule, row []interface{}, columnName string) (string, error) {
	index := rule.TableInfo.FindColumn(columnName)
	if index < 0 {
		return "", errors.Errorf("parent id not found %s(%s)", rule.TableInfo.Name, columnName)
	}

	return fmt.Sprint(row[index]), nil
}

func (r *River) doBulk(reqs []*cockroach.BulkRequest) error {
	if len(reqs) == 0 {
		return nil
	}

	if resp, err := r.cock.Bulk(reqs); err != nil {
		log.Errorf("sync docs err %v after binlog %s", err, r.canal.SyncedPosition())
		return errors.Trace(err)
	} else if resp.Code/100 == 2 || resp.Errors {
		for i := 0; i < len(resp.Items); i++ {
			for action, item := range resp.Items[i] {
				if len(item.Error) > 0 {
					seelog := *r.cock.Seelog
					// log.Errorf("%s index: %s, type: %s, id: %s, status: %d, error: %s",
					seelog.Errorf("%s index: %s, type: %s, id: %s, status: %d, error: %s",
						action, item.Index, item.Type, item.ID, item.Status, item.Error)
				}
			}
		}
	}

	return nil
}
