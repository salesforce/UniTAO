/*
************************************************************************************************************
Copyright (c) 2022 Salesforce, Inc.
All rights reserved.

UniTAO was originally created in 2022 by Shai Herzog & Yi Huo as an
Universal No-Coding Heterogeneous Infrastructure Maintenance & Inventory system that is holistically driven by open/community-developed semantic models/schemas.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>

This copyright notice and license applies to all files in this directory or sub-directories, except when stated otherwise explicitly.
************************************************************************************************************
*/

package DataHandler

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"
	"sync"

	"Data/DbConfig"
	"Data/DbIface"
	"DataService/Config"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type JournalAdd func(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) (int, *Http.HttpError)

type Handler struct {
	DB         DbIface.Database
	schemaMap  map[string]*Schema.SchemaOps
	config     Config.Confuguration
	lock       sync.Mutex
	AddJournal JournalAdd
}

func New(config Config.Confuguration, connectDb func(db DbConfig.DatabaseConfig) (DbIface.Database, error)) (*Handler, error) {
	db, err := connectDb(config.Database)
	if err != nil {
		return nil, err
	}
	handler := Handler{
		schemaMap: make(map[string]*Schema.SchemaOps),
		DB:        db,
		config:    config,
		lock:      sync.Mutex{},
	}
	return &handler, nil
}

func (h *Handler) QueryDb(dataType string, dataId string) ([]map[string]interface{}, *Http.HttpError) {
	args := make(map[string]interface{})
	args[DbIface.Table] = h.config.DataTable.Data
	args[Record.DataType] = dataType
	if dataId != "" {
		args[Record.DataId] = dataId
	}
	recordList, err := h.DB.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	return recordList, nil
}

func (h *Handler) List(dataType string) ([]interface{}, *Http.HttpError) {
	if dataType != JsonKey.Schema && dataType != "" {
		_, err := h.GetData(JsonKey.Schema, dataType)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
		}
	}
	if dataType == "" {
		dataType = JsonKey.Schema
	}
	recordList, e := h.QueryDb(dataType, "")
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	result := make([]interface{}, 0, len(recordList))
	for _, record := range recordList {
		// not record schema is only for schema.
		if record[Record.DataId] != Record.KeyRecord {
			result = append(result, record[Record.DataId].(string))
		}
	}
	return result, nil
}

func (h *Handler) Get(dataType string, dataId string) (map[string]interface{}, *Http.HttpError) {
	if dataType == JsonKey.Schema {
		return h.GetData(JsonKey.Schema, dataId)
	}
	_, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
	}
	return h.GetData(dataType, dataId)
}

func (h *Handler) GetData(dataType string, dataId string) (map[string]interface{}, *Http.HttpError) {
	recordList, err := h.QueryDb(dataType, dataId)
	if err != nil {
		return nil, err
	}
	if len(recordList) == 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("object of type '%s' with id '%s' not found", dataType, dataId), http.StatusNotFound)
	}
	if len(recordList) > 1 {
		return nil, Http.NewHttpError(fmt.Sprintf("found [%d] record for [type/id]=[%s/%s]", len(recordList), dataType, dataId), http.StatusInternalServerError)
	}
	return recordList[0], nil
}

func (h *Handler) querySchema(dataType string) (*Schema.SchemaOps, *Http.HttpError) {
	schema, ok := h.schemaMap[dataType]
	if ok {
		return schema, nil
	}
	data, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, err
	}
	record, e := Record.LoadMap(data)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to load schema record. [type]=[%s]", dataType), http.StatusInternalServerError)
	}
	schema, e = Schema.LoadSchemaOpsRecord(record)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to load Schema Record, [%s]=[%s]", Record.DataType, dataType), http.StatusInternalServerError)
	}
	h.schemaMap[dataType] = schema
	return schema, nil
}

func (h *Handler) localSchema(dataType string, version string) (*Schema.SchemaOps, *Http.HttpError) {
	schema, err := h.querySchema(dataType)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
		}
		return nil, err
	}
	if schema == nil {
		return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
	}
	if version == "" || schema.Record.Version == version {
		return schema, nil
	}
	archivedType := SchemaDoc.ArchivedSchemaId(dataType, version)
	schema, err = h.querySchema(archivedType)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
	}
	return schema, nil
}

func (h *Handler) inventoryData(dataType string, value string) (map[string]interface{}, *Http.HttpError) {
	if h.config.Inv.Url == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("missing inventory URL in configuration [inventory/url]. at Data Service=[%s]", h.config.Http.Id), http.StatusInternalServerError)
	}
	dataUrl, err := Http.URLPathJoin(h.config.Inv.Url, dataType, value)
	if err != nil {
		return nil, Http.NewHttpError(fmt.Sprintf("failed to build ref data url. [url/type/id]=[%s/%s/%s]", h.config.Inv.Url, dataType, value), http.StatusBadRequest)
	}
	data, code, err := Http.GetRestData(*dataUrl)
	if err == nil {
		mapData, ok := data.(map[string]interface{})
		if !ok {
			return nil, Http.NewHttpError(fmt.Sprintf("return data is not an object. [url]=[%s]", *dataUrl), http.StatusBadRequest)
		}
		return mapData, nil
	}
	return nil, Http.NewHttpError(err.Error(), code)
}

func (h *Handler) Validate(record *Record.Record) *Http.HttpError {
	if record.Type == Record.KeyRecord {
		return Http.NewHttpError(fmt.Sprintf("should not validate schema of %s", Record.KeyRecord), http.StatusBadRequest)
	}
	schema, err := h.localSchema(record.Type, record.Version)
	if err != nil {
		return err
	}
	e := schema.ValidateRecord(record)
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to validate payload against schema for type %s", record.Type), http.StatusBadRequest)
	}
	if record.Type != JsonKey.Schema {
		err = h.ValidateDataRefs(schema.Schema, record.Data, path.Join(record.Type, record.Id))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) ValidateDataRefs(doc *SchemaDoc.SchemaDoc, data interface{}, dataPath string) *Http.HttpError {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return Http.NewHttpError("failed to convert data to map", http.StatusBadRequest)
	}
	err := h.validateCmtRefs(doc, dataMap, dataPath)
	if err != nil {
		return err
	}
	return h.validateSubDoc(doc, dataMap, dataPath)
}

// Validate Content Media Type Reference Values
func (h *Handler) validateCmtRefs(doc *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string) *Http.HttpError {
	errList := []*Http.HttpError{}
	for _, ref := range doc.CmtRefs {
		value, ok := data[ref.Name]
		if !ok {
			continue
		}
		refPath := path.Join(dataPath, ref.Name)
		switch reflect.TypeOf(value).Kind() {
		case reflect.String:
			hasRef, err := h.validateCmtRefValue(ref, value.(string))
			if err != nil {
				return err
			}
			if !hasRef {
				errList = append(errList, Http.NewHttpError(
					fmt.Sprintf("broken reference @[%s], '%s:%s' does not exists", refPath, ref.ContentType, value),
					http.StatusBadRequest))
			}
		case reflect.Slice:
			for idx, item := range value.([]interface{}) {
				idxPath := fmt.Sprintf("%s[%d]", refPath, idx)
				hasRef, err := h.validateCmtRefValue(ref, item.(string))
				if err != nil {
					errList = append(errList,
						Http.WrapError(err, fmt.Sprintf("broken reference @[%s], '%s:%s' value validate failed.", idxPath, ref.ContentType, item),
							http.StatusBadRequest))
					continue
				}
				if !hasRef {
					errList = append(errList, Http.NewHttpError(
						fmt.Sprintf("broken reference @[%s], '%s:%s' does not exist.", idxPath, ref.ContentType, item),
						http.StatusBadRequest))
				}
			}
		default:
			errList = append(errList, Http.NewHttpError(
				fmt.Sprintf("broken reference @[%s], 'dataType=[%s]' are not supported. only support string or array", refPath, reflect.TypeOf(value).Kind()),
				http.StatusBadRequest))
		}
	}
	if len(errList) > 0 {
		if len(errList) > 1 {
			err := Http.NewHttpError("broken references:", http.StatusBadRequest)
			for _, itemErr := range errList {
				Http.AppendError(err, itemErr)
			}
			return err
		}
		return errList[0]
	}
	return nil
}

func (h *Handler) validateCmtRefValue(ref *SchemaDoc.CMTDocRef, value string) (bool, *Http.HttpError) {
	if ref.CmtType != Schema.Inventory {
		// ContentMediaType not start with inventory, we don't understand
		return true, nil
	}
	if ref.ContentType == JsonKey.Schema {
		return false, Http.NewHttpError("should not refer to schema of schema as data type", http.StatusBadRequest)
	}
	_, err := h.localSchema(ref.ContentType, "")
	if err != nil {
		if err.Status != http.StatusNotFound {
			return false, err
		}
		_, err := h.inventoryData(ref.ContentType, value)
		if err != nil {
			if err.Status != http.StatusNotFound {
				return false, err
			}
			return false, nil
		}
		return true, nil
	}
	_, err = h.GetData(ref.ContentType, value)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (h *Handler) validateSubDoc(doc *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string) *Http.HttpError {
	for pname, pDoc := range doc.SubDocs {
		subPath := path.Join(dataPath, pname)
		subData, ok := data[pname]
		if !ok {
			// property does not exists
			continue
		}
		switch reflect.TypeOf(subData).Kind() {
		case reflect.Map:
			// Object, validate directly
			err := h.ValidateDataRefs(pDoc, subData, subPath)
			if err != nil {
				return err
			}
		case reflect.Slice:
			// check each item value of array
			for idx, idxData := range subData.([]interface{}) {
				idxPath := fmt.Sprintf("%s[%d]", subPath, idx)
				err := h.ValidateDataRefs(pDoc, idxData, idxPath)
				if err != nil {
					return err
				}
			}
		default:
			// data value does not match schema
			return Http.NewHttpError(fmt.Sprintf("data is not [%s or %s] @[path]=[%s]", reflect.Map, reflect.Slice, subPath), http.StatusBadRequest)
		}
		// itemized data type passed
		return nil
	}

	return nil
}

func (h *Handler) Add(record *Record.Record) *Http.HttpError {
	err := h.Validate(record)
	if err != nil {
		return err
	}
	if strings.Contains(record.Id, JsonKey.ArchivedSchemaIdDiv) {
		return Http.NewHttpError(fmt.Sprintf("invalid data format. type=[%s] version=[%s] is archived", record.Type, record.Version), http.StatusBadRequest)
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	recordList, err := h.QueryDb(JsonKey.Schema, record.Id)
	if err != nil {
		return err
	}
	if len(recordList) > 0 {
		if record.Type != JsonKey.Schema {
			return Http.NewHttpError(fmt.Sprintf("data [type/id]=[%s/%s] already exists", record.Type, record.Id), http.StatusConflict)
		}
		_, schemaVer := Util.ParseCustomPath(record.Id, JsonKey.ArchivedSchemaIdDiv)
		if schemaVer != "" {
			return Http.NewHttpError(fmt.Sprintf(`"invalid schema id=[%s], add new archived data are not supported. 
				please add new schema directly, current schema will be archived automatically."`, record.Id), http.StatusBadRequest)
		}
		err := h.archiveSchema(record)
		if err != nil {
			return err
		}
	}
	return h.addData(record)
}

func (h *Handler) archiveSchema(record *Record.Record) *Http.HttpError {
	schema, err := h.localSchema(record.Id, "")
	if err != nil {
		return err
	}
	newVer, ex := Record.ParseVersion(record.Version)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("invalid new schema version=[%s]", record.Version), http.StatusBadRequest)
	}
	currentVer, ex := Record.ParseVersion(schema.Record.Version)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("invalid current schema version=[%s] for current type=[%s]", record.Version, record.Id), http.StatusBadRequest)
	}
	verComp := Record.CompareVersion(newVer, currentVer)
	if verComp < 0 {
		return Http.WrapError(ex, fmt.Sprintf("new schema version=[%s] is smaller than current version=[%s]", record.Version, schema.Record.Version), http.StatusBadRequest)
	}
	if verComp == 0 {
		return Http.WrapError(ex, fmt.Sprintf("new schema version=[%s] is equal to current version, please provid later version to archive current one", record.Version), http.StatusBadRequest)
	}
	schema.Record.Id = SchemaDoc.ArchivedSchemaId(schema.Record.Id, schema.Record.Version)
	args := map[string]interface{}{
		Record.DataType: JsonKey.Schema,
		Record.DataId:   record.Id,
	}
	ex = h.DB.Replace(h.config.DataTable.Data, args, schema.Record.Map())
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to archive current schema=[%s] to [%s]", record.Id, schema.Record.Id), http.StatusInternalServerError)
	}
	return nil
}

func (h *Handler) addData(record *Record.Record) *Http.HttpError {
	e := h.DB.Create(h.config.DataTable.Data, record.Map())
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to create record [{type}/{id}]=[%s]/%s", record.Type, record.Id), http.StatusInternalServerError)
	}
	if h.AddJournal != nil {
		h.AddJournal(record.Type, record.Id, nil, record.Data)
		// we should log err if failed to add Journal
	}
	return nil
}

func (h *Handler) Set(record *Record.Record) *Http.HttpError {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.updateRecord(record)
}

func (h *Handler) updateRecord(record *Record.Record) *Http.HttpError {
	err := h.Validate(record)
	if err != nil {
		return err
	}
	recordList, err := h.QueryDb(record.Type, record.Id)
	if err != nil {
		return err
	}
	var before map[string]interface{}
	if len(recordList) > 0 {
		if record.Type == JsonKey.Schema {
			return Http.NewHttpError(fmt.Sprintf("update on existing schema=[%s] not allowed", record.Id), http.StatusBadRequest)
		}
		rec, e := Record.LoadMap(recordList[0])
		if e == nil {
			before = rec.Data
		}
	}
	e := h.DB.Create(h.config.DataTable.Data, record.Map())
	if e != nil {
		return Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	if h.AddJournal != nil {
		h.AddJournal(record.Type, record.Id, before, record.Data)
	}
	return nil
}

func (h *Handler) Delete(dataType string, dataId string) *Http.HttpError {
	if dataType == JsonKey.Schema {
		return h.deleteSchema(dataId)
	}
	_, err := h.localSchema(dataType, "")
	if err != nil {
		return err
	}
	return h.deleteData(dataType, dataId)
}

func (h *Handler) deleteSchema(dataType string) *Http.HttpError {
	schemaId, schemaVer := Util.ParseCustomPath(dataType, JsonKey.ArchivedSchemaIdDiv)
	if schemaVer != "" {
		recordList, err := h.QueryDb(schemaId, "")
		if err != nil {
			return err
		}
		for _, data := range recordList {
			rec, ex := Record.LoadMap(data)
			if ex != nil {
				return Http.WrapError(ex, fmt.Sprintf("failed to load data as Record. dataType=[%s]", schemaId), http.StatusInternalServerError)
			}
			if rec.Version == schemaVer {
				return Http.NewHttpError(fmt.Sprintf("data exists, cannot delete schema=[%s], version=[%s]", schemaId, schemaVer), http.StatusBadRequest)
			}
		}
		return h.deleteData(JsonKey.Schema, dataType)
	}
	archivedPrefix := SchemaDoc.ArchivedSchemaId(dataType, "")
	schemaList, err := h.List(JsonKey.Schema)
	if err != nil {
		return err
	}
	for _, queryType := range schemaList {
		if strings.HasPrefix(queryType.(string), archivedPrefix) {
			return Http.NewHttpError(fmt.Sprintf("there are archived schema for type=[%s], please delete all archived schema before delete type", dataType), http.StatusBadRequest)
		}
	}
	return h.deleteData(JsonKey.Schema, dataType)
}

func (h *Handler) deleteData(dataType string, dataId string) *Http.HttpError {
	h.lock.Lock()
	defer h.lock.Unlock()
	recordList, err := h.QueryDb(dataType, dataId)
	if err != nil {
		return err
	}
	if len(recordList) == 0 {
		return nil
	}
	var before map[string]interface{}
	rec, e := Record.LoadMap(recordList[0])
	if e == nil {
		before = rec.Data
	}
	keys := make(map[string]interface{})
	keys[Record.DataType] = dataType
	keys[Record.DataId] = dataId
	e = h.DB.Delete(h.config.DataTable.Data, keys)
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}
	if h.AddJournal != nil {
		h.AddJournal(dataType, dataId, before, nil)
	}
	return nil
}

func (h *Handler) Patch(dataType string, idPath string, data interface{}) (map[string]interface{}, *Http.HttpError) {
	dataId, nextPath := Util.ParsePath(idPath)
	if dataId == "" {
		return nil, Http.NewHttpError(fmt.Sprintf("invalid path=[%s/%s], expect format=[{dataType}/{dataId}/{dataPath}]", dataType, dataId), http.StatusBadRequest)
	}
	if nextPath == "" {
		return nil, Http.NewHttpError("invalid path, no data path to drill in, expect format=[{dataType}/{dataId}/{dataPath}]", http.StatusBadRequest)
	}
	_, err := h.localSchema(dataType, "")
	if err != nil {
		return nil, err
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	patchData, err := h.GetData(dataType, dataId)
	if err != nil {
		return nil, err
	}
	patchRecord, e := Record.LoadMap(patchData)
	if err != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to load data [%s/%s] as record", dataType, dataId), http.StatusInternalServerError)
	}
	schema, err := h.localSchema(dataType, patchRecord.Version)
	if err != nil {
		return nil, err
	}
	before := map[string]interface{}{}
	e = Util.ObjCopy(patchRecord.Data, &before)
	if e != nil {
		return nil, Http.WrapError(e, fmt.Sprintf("failed to snapshot data [%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}
	err = Schema.SetDataOnPath(schema.Schema, patchRecord.Data, nextPath, fmt.Sprintf("%s/%s", dataType, dataId), data)
	if err != nil {
		return nil, err
	}
	err = h.updateRecord(patchRecord)
	if err != nil {
		return nil, err
	}
	if h.AddJournal != nil {
		h.AddJournal(dataType, dataId, before, patchRecord.Data)
	}
	return patchRecord.Map(), nil
}
