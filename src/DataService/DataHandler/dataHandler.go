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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"Data/DbConfig"
	"Data/DbIface"
	"DataService/Common"
	"DataService/Config"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/HashLock"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
	"github.com/salesforce/UniTAO/lib/Util/Template"
)

type JournalAdd func(dataType string, dataId string, before map[string]interface{}, after map[string]interface{}) *Http.HttpError

type Handler struct {
	DB         DbIface.Database
	schemaMap  map[string]*Schema.SchemaOps
	Config     Config.Confuguration
	Lock       *HashLock.HashLock
	Inventory  *DataServiceProxy
	AddJournal JournalAdd
	log        *log.Logger
}

func New(config Config.Confuguration, logger *log.Logger, connectDb func(db DbConfig.DatabaseConfig) (DbIface.Database, error)) (*Handler, *Http.HttpError) {
	if logger == nil {
		logger = log.Default()
	}
	db, err := connectDb(config.Database)
	if err != nil {
		return nil, Http.WrapError(err, "failed to connect to Database", http.StatusInternalServerError)
	}
	handler := Handler{
		schemaMap: make(map[string]*Schema.SchemaOps),
		DB:        db,
		Config:    config,
		Lock:      HashLock.NewHashLock(logger),
		log:       logger,
	}
	handler.Inventory = CreateDsProxy(&handler)
	return &handler, nil
}

func (h *Handler) Log(message string) {
	h.log.Printf("Handler: %s", message)
}

func (h *Handler) QueryDb(dataType string, dataId string) ([]map[string]interface{}, *Http.HttpError) {
	args := make(map[string]interface{})
	args[DbIface.Table] = h.Config.DataTable.Data
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
		id, version := Util.ParsePath(dataId)
		schema, err := h.LocalSchema(id, version)
		if err != nil {
			return nil, err
		}
		return schema.Record.Map(), nil
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
		errMsg := fmt.Sprintf("failed to load schema record. [type]=[%s]", dataType)
		h.Log(errMsg)
		h.Log(e.Error())
		return nil, Http.WrapError(e, errMsg, http.StatusInternalServerError)
	}
	schema, e = Schema.LoadSchemaOpsRecord(record)
	if e != nil {
		errMsg := fmt.Sprintf("failed to load Schema Record, [%s]=[%s]", Record.DataType, dataType)
		h.Log(errMsg)
		h.Log(e.Error())
		return nil, Http.WrapError(e, errMsg, http.StatusInternalServerError)
	}
	h.SetLocalSchema(dataType, schema)
	return schema, nil
}

func (h *Handler) SetLocalSchema(dataType string, schema *Schema.SchemaOps) {
	if schema == nil {
		delete(h.schemaMap, dataType)
		return
	}
	h.schemaMap[dataType] = schema
}

func (h *Handler) LocalSchema(dataType string, version string) (*Schema.SchemaOps, *Http.HttpError) {
	schema, err := h.querySchema(dataType)
	if err != nil {
		if err.Status == http.StatusNotFound {
			errMsg := fmt.Sprintf("object of type “%s” does not exist, query failed", dataType)
			h.Log(errMsg)
			return nil, Http.WrapError(err, errMsg, err.Status)
		}
		return nil, err
	}
	if schema == nil {
		errMsg := fmt.Sprintf("object of type “%s” does not exist", dataType)
		h.Log(errMsg)
		return nil, Http.WrapError(err, errMsg, err.Status)
	}
	if version == "" || schema.Schema.Version == version {
		return schema, nil
	}
	archivedType := SchemaDoc.ArchivedSchemaId(dataType, version)
	schema, err = h.querySchema(archivedType)
	if err != nil {
		h.Log(fmt.Sprintf("failed to query archive schema: %s", archivedType))
		h.Log(err.Error())
		return nil, err
	}
	if schema == nil {
		errMsg := fmt.Sprintf("object of type “%s” does not exist locally", dataType)
		h.Log(errMsg)
		return nil, Http.WrapError(err, errMsg, err.Status)
	}
	return schema, nil
}

func (h *Handler) Validate(record *Record.Record) *Http.HttpError {
	if record.Type == Record.KeyRecord {
		errMsg := fmt.Sprintf("should not validate schema of %s", Record.KeyRecord)
		h.Log(errMsg)
		return Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	schema, err := h.LocalSchema(record.Type, record.Version)
	if err != nil {
		return err
	}
	e := schema.ValidateRecord(record)
	if e != nil {
		errMsg := fmt.Sprintf("failed to validate payload against schema for type %s", record.Type)
		h.Log(errMsg)
		h.Log(e.Error())
		return Http.WrapError(e, errMsg, http.StatusBadRequest)
	}
	if record.Type != JsonKey.Schema {
		err = h.ValidateDataRefs(schema.Schema, record.Data, path.Join(record.Type, record.Id))
	} else {
		err = h.validateCmtAutoIdxOnSchema(record)
	}
	if err != nil {
		h.Log(err.Error())
		return err
	}
	return nil
}

func (h *Handler) ValidateDataRefs(doc *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string) *Http.HttpError {
	for attrName, def := range doc.Properties() {
		value, ok := data[attrName]
		if !ok {
			continue
		}
		attrDef := def.(map[string]interface{})
		attrPath := fmt.Sprintf("%s/%s", dataPath, attrName)
		ref, isRef := doc.CmtRefs[attrName]
		subDoc, isSubDoc := doc.SubDocs[attrName]
		errList := []*Http.HttpError{}
		switch attrDef[JsonKey.Type] {
		case JsonKey.String:
			if !isRef {
				continue
			}
			err := h.validateCmtRefValue(ref, value.(string), attrPath)
			if err != nil {
				errList = append(errList, err)
			}
		case JsonKey.Object:
			valueObj := value.(map[string]interface{})
			if !SchemaDoc.IsMap(attrDef) {
				return h.ValidateDataRefs(subDoc, valueObj, attrPath)
			}
			if !isRef && !isSubDoc {
				continue
			}
			for key, keyValue := range valueObj {
				keyPath := fmt.Sprintf("%s[%s]", attrPath, key)
				if isRef {
					err := h.validateCmtRefValue(ref, keyValue.(string), keyPath)
					if err != nil {
						errList = append(errList, err)
					}
					continue
				}
				if isSubDoc {
					err := h.ValidateDataRefs(subDoc, keyValue.(map[string]interface{}), keyPath)
					if err != nil {
						errList = append(errList, err)
					}
				}
			}
		case JsonKey.Array:
			valueAry := value.([]interface{})
			if !isRef && !isSubDoc {
				continue
			}
			for _, item := range valueAry {
				if isRef {
					itemPath := fmt.Sprintf("%s[%s]", attrPath, item.(string))
					err := h.validateCmtRefValue(ref, item.(string), itemPath)
					if err != nil {
						errList = append(errList, err)
					}
					continue
				}
				if isSubDoc {
					itemKey, _ := subDoc.BuildKey(item.(map[string]interface{}))
					itemPath := fmt.Sprintf("%s[%s]", attrPath, itemKey)
					err := h.ValidateDataRefs(subDoc, item.(map[string]interface{}), itemPath)
					if err != nil {
						errList = append(errList, err)
					}
				}
			}
		}
		switch len(errList) {
		case 0:
			continue
		case 1:
			return errList[0]
		}
		err := Http.NewHttpError(fmt.Sprintf("Data Reference validation failure @path=[%s]", attrPath), http.StatusBadRequest)
		for _, sub := range errList {
			err.AppendError(sub)
		}
		return err
	}
	return nil
}

func (h *Handler) validateCmtRefValue(ref *SchemaDoc.CMTDocRef, value string, dataPath string) *Http.HttpError {
	if ref.CmtType != Schema.Inventory {
		// ContentMediaType not start with inventory, we don't understand
		return nil
	}
	if ref.ContentType == JsonKey.Schema {
		return Http.NewHttpError("should not refer to schema of schema as data type", http.StatusBadRequest)
	}
	cmtRecord, err := h.Inventory.Get(ref.ContentType, value)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return Http.NewHttpError(fmt.Sprintf("reference %s:%s with value=[%s] does not exists. @path=[%s]", ref.CmtType, ref.ContentType, value, dataPath), http.StatusBadRequest)
		}
		return err
	}
	if ref.IndexTemplate != "" {
		idxTemp, ex := Template.ParseStr(ref.IndexTemplate, "{", "}")
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to parse IndexTemplate, @path=[%s]", dataPath), http.StatusInternalServerError)
		}
		attrPath, ex := idxTemp.BuildValue(cmtRecord.Data)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to build CmtKey from cmtRecord.Error:%s", ex), http.StatusInternalServerError)
		}
		indexPath := fmt.Sprintf("%s[%s]", attrPath, value)
		if indexPath != dataPath {
			return Http.NewHttpError(fmt.Sprintf("cmt Record=[%s/%s] build path=[%s] does not match data path=[%s]", cmtRecord.Type, cmtRecord.Id, indexPath, dataPath), http.StatusBadRequest)
		}
	}
	return nil
}

func (h *Handler) validateCmtAutoIdxOnSchema(record *Record.Record) *Http.HttpError {
	schemaData, _ := Json.CopyToMap(record.Data)
	schema, _ := SchemaDoc.New(schemaData)
	idxList := CmtIndex.FindAutoIndex(schema, "")
	errList := make([]*Http.HttpError, 0, len(idxList))
	for _, idx := range idxList {
		err := h.validateCmtAutoIdx(idx)
		if err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) > 0 {
		err := Http.NewHttpError(fmt.Sprintf("invalid IndexTemplate on schema: [%s]", record.Id), http.StatusBadRequest)
		for _, ex := range errList {
			err.AppendError(ex)
		}
		return err
	}
	return nil
}

func (h *Handler) validateCmtAutoIdx(idx CmtIndex.AutoIndex) *Http.HttpError {
	schemaRec, err := h.Inventory.Get(JsonKey.Schema, idx.ContentType)
	if err != nil {
		return err
	}
	targetSchema, ex := SchemaDoc.New(schemaRec.Data)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to load schema data of type=[%s]", idx.ContentType), http.StatusInternalServerError)
	}
	return idx.ValidateIndexTemplate(targetSchema)
}

func (h *Handler) Add(record *Record.Record) *Http.HttpError {
	err := h.Validate(record)
	if err != nil {
		return err
	}
	schema, _ := h.LocalSchema(record.Type, "")
	if schema.Schema.Version != record.Version {
		return Http.NewHttpError(fmt.Sprintf("invalid schema version of [%s %s] not match current schema version[%s]", record.Type, record.Version, schema.Schema.Version), http.StatusBadRequest)
	}
	idKey := fmt.Sprintf("%s/%s", record.Type, record.Id)
	h.Lock.Aquire(idKey, "HandlerAdd")
	defer h.Lock.Release(idKey, "HandlerAdd")
	h.Log(fmt.Sprintf("HandlerAdd: query exists.[%s/%s]", record.Type, record.Id))
	recordList, err := h.QueryDb(record.Type, record.Id)
	if err != nil {
		h.Log(fmt.Sprintf("HandlerAdd: query failed.[%s/%s]", record.Type, record.Id))
		return err
	}
	if len(recordList) > 0 {
		h.Log(fmt.Sprintf("HandlerAdd: already exists.[%s/%s]", record.Type, record.Id))
		if record.Type != JsonKey.Schema {
			return Http.NewHttpError(fmt.Sprintf("data [type/id]=[%s/%s] already exists", record.Type, record.Id), http.StatusConflict)
		}
		h.Log(fmt.Sprintf("HandlerAdd: upgrade schema.[%s]", record.Id))
		_, schemaVer := Util.ParseCustomPath(record.Id, JsonKey.ArchivedSchemaIdDiv)
		if schemaVer != "" {
			msg := fmt.Sprintf(`"invalid schema id=[%s], add new archived data are not supported. 
			please add new schema directly, current schema will be archived automatically."`, record.Id)
			h.Log(fmt.Sprintf("HandlerAdd: [%s]", msg))
			return Http.NewHttpError(msg, http.StatusBadRequest)
		}
		newSchema, ex := Schema.LoadSchemaOpsRecord(record)
		if ex != nil {
			return Http.WrapError(ex, "failed to load new schema record as schema", http.StatusBadRequest)
		}
		h.archiveCurrentSchema(newSchema)
	}
	return h.addData(record)
}

func CompareVersion(currentVersion string, newVersion string) (int, *Http.HttpError) {
	newVer, ex := Record.ParseVersion(newVersion)
	if ex != nil {
		return -2, Http.WrapError(ex, fmt.Sprintf("invalid new schema version[%s]", newVersion), http.StatusBadRequest)
	}
	currentVer, ex := Record.ParseVersion(currentVersion)
	if ex != nil {
		return -2, Http.WrapError(ex, fmt.Sprintf("invalid current schema version[%s]", currentVersion), http.StatusBadRequest)
	}
	verComp := Record.CompareVersion(newVer, currentVer)
	return verComp, nil
}

func (h *Handler) archiveCurrentSchema(newSchema *Schema.SchemaOps) *Http.HttpError {
	h.Log(fmt.Sprintf("HandlerAdd: archive schema [%s]", newSchema.Schema.Id))
	schema, err := h.LocalSchema(newSchema.Schema.Id, "")
	if err != nil {
		return err
	}
	verComp, err := CompareVersion(schema.Schema.Version, newSchema.Schema.Version)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("failed to validate version of schema[%s]", newSchema.Schema.Id), http.StatusBadRequest)
	}
	if verComp < 0 {
		return Http.NewHttpError(fmt.Sprintf("new schema version=[%s] is smaller than current version=[%s]", newSchema.Schema.Version, schema.Schema.Version), http.StatusBadRequest)
	}
	if verComp == 0 {
		return Http.NewHttpError(fmt.Sprintf("new schema version=[%s] is equal to current version, please provid later version to archive current one", newSchema.Schema.Version), http.StatusBadRequest)
	}
	before := Record.Record{}
	ex := Json.CopyTo(schema.Record, &before)
	if ex != nil {
		return Http.WrapError(ex, fmt.Sprintf("failed to snapshot schema [%s %s]", schema.Schema.Id, schema.Schema.Version), http.StatusInternalServerError)
	}
	schema.Record.Id = SchemaDoc.ArchivedSchemaId(schema.Schema.Id, schema.Schema.Version)
	h.Log(fmt.Sprintf("HandlerAdd: updating schema record [%s]", newSchema.Schema.Id))
	err = h.updateRecord(JsonKey.Schema, schema.Schema.Id, schema.Record)
	if err != nil {
		return err
	}
	h.Log(fmt.Sprintf("HandlerAdd: schema archived [%s]", newSchema.Schema.Id))
	h.SetLocalSchema(schema.Schema.Id, nil)
	h.SetLocalSchema(schema.Record.Id, schema)
	if h.AddJournal != nil {
		h.Log(fmt.Sprintf("HandlerAdd: Add Journal of schema archive. [%s]->[%s]", before.Id, schema.Record.Id))
		h.AddJournal(before.Type, before.Id, before.Map(), schema.Record.Map())
	}
	return nil
}

func (h *Handler) addData(record *Record.Record) *Http.HttpError {
	h.Log(fmt.Sprintf("HandlerAdd: add record [%s/%s]", record.Type, record.Id))
	e := h.DB.Create(h.Config.DataTable.Data, record.Map())
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to create record [{type}/{id}]=[%s]/%s", record.Type, record.Id), http.StatusInternalServerError)
	}
	h.Log(fmt.Sprintf("HandlerAdd: record added [%s/%s]", record.Type, record.Id))
	if h.AddJournal != nil {
		h.Log(fmt.Sprintf("HandlerAdd: add journal for new record [%s/%s]", record.Type, record.Id))
		h.AddJournal(record.Type, record.Id, nil, record.Map())
		// we should log err if failed to add Journal
	}
	return nil
}

func (h *Handler) CompareRecords(before *Record.Record, after *Record.Record) (bool, *Http.HttpError) {
	if before == nil && after == nil {
		return true, nil
	}
	if before == nil || after == nil {
		return false, nil
	}
	if before.Type != after.Type {
		return false, Http.NewHttpError(fmt.Sprintf("change of field[%s] not allowed", Record.DataType), http.StatusBadRequest)
	}

	if before.Id != after.Id {
		if before.Type != JsonKey.Schema {
			return false, Http.NewHttpError(fmt.Sprintf("[%s] update not allowed on type[%s], id can only be updatd for [%s]", Record.DataId, before.Type, JsonKey.Schema), http.StatusBadRequest)
		}
		return false, nil
	}
	if before.Version != after.Version {
		verComp, err := CompareVersion(before.Version, after.Version)
		if err != nil {
			return false, Http.WrapError(err, "failed to compare version", http.StatusBadRequest)
		}
		if verComp < 0 {
			return false, Http.NewHttpError(fmt.Sprintf("downgrade data format are not supported. version[%s] -> [%s]", before.Version, after.Version), http.StatusBadRequest)
		}
	}
	_, ex := h.LocalSchema(after.Type, after.Version)
	if ex != nil {
		return false, ex
	}
	srcStr, _ := json.MarshalIndent(before.Data, "", "    ")
	tgtStr, _ := json.MarshalIndent(after.Data, "", "    ")
	if string(srcStr) != string(tgtStr) {
		return false, nil
	}
	return true, nil
}

func (h *Handler) Set(dataType string, dataId string, record *Record.Record) *Http.HttpError {
	if _, ok := Common.InternalTypes[record.Type]; ok {
		return Http.NewHttpError(fmt.Sprintf("method[%s] on type[%s] is not allowed", http.MethodPut, record.Type), http.StatusBadRequest)
	}
	if dataType == "" {
		dataType = record.Type
	}
	if dataId == "" {
		dataId = record.Id
	}
	idKey := fmt.Sprintf("%s/%s", dataType, dataId)
	h.Lock.Aquire(idKey, "HandlerSet")
	defer h.Lock.Release(idKey, "HandlerSet")
	data, err := h.GetData(dataType, dataId)
	if err != nil && err.Status != http.StatusNotFound {
		return err
	}
	var before *Record.Record
	if data != nil {
		record, ex := Record.LoadMap(data)
		if ex != nil {
			return Http.WrapError(ex, fmt.Sprintf("failed to load data of [%s/%s] as record", record.Type, record.Id), http.StatusInternalServerError)
		}
		before = record
	}
	isSame, err := h.CompareRecords(before, record)
	if err != nil {
		return err
	}
	if !isSame {
		err = h.updateRecord(before.Type, before.Id, record)
		if err != nil {
			return err
		}
		if h.AddJournal != nil {
			h.AddJournal(record.Type, record.Id, before.Map(), record.Map())
		}
	}
	return nil
}

func (h *Handler) updateRecord(dataType string, dataId string, record *Record.Record) *Http.HttpError {
	err := h.Validate(record)
	if err != nil {
		return err
	}
	e := h.DB.Replace(h.Config.DataTable.Data, map[string]interface{}{
		Record.DataType: dataType,
		Record.DataId:   dataId,
	}, record.Map())
	if e != nil {
		return Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (h *Handler) Delete(dataType string, dataId string) *Http.HttpError {
	if dataType == JsonKey.Schema {
		err := h.deleteSchema(dataId)
		if err != nil {
			return err
		}
		delete(h.schemaMap, dataId)
	}
	_, err := h.LocalSchema(dataType, "")
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
	idKey := fmt.Sprintf("%s/%s", dataType, dataId)
	h.Lock.Aquire(idKey, "HandlerDelete")
	defer h.Lock.Release(idKey, "HandlerDelete")
	recordList, err := h.QueryDb(dataType, dataId)
	if err != nil {
		return err
	}
	if len(recordList) == 0 {
		return nil
	}
	beforeRec, e := Record.LoadMap(recordList[0])
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to load data as record.[type/id]=[%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}

	keys := make(map[string]interface{})
	keys[Record.DataType] = dataType
	keys[Record.DataId] = dataId
	e = h.DB.Delete(h.Config.DataTable.Data, keys)
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}
	if h.AddJournal != nil {
		h.AddJournal(dataType, dataId, beforeRec.Map(), nil)
	}
	return nil
}

func (h *Handler) Patch(dataType string, idPath string, headers map[string]interface{}, data interface{}) (map[string]interface{}, *Http.HttpError) {
	dataId, nextPath := Util.ParsePath(idPath)
	if dataId == "" {
		errMsg := fmt.Sprintf("invalid path=[%s/%s], expect format=[{dataType}/{dataId}/{dataPath}]", dataType, dataId)
		h.Log(errMsg)
		return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	if nextPath == "" {
		errMsg := "invalid path, no data path to drill in, expect format=[{dataType}/{dataId}/{dataPath}]"
		h.Log(errMsg)
		return nil, Http.NewHttpError(errMsg, http.StatusBadRequest)
	}
	_, err := h.LocalSchema(dataType, "")
	if err != nil {
		return nil, err
	}
	h.Log(fmt.Sprintf("Handler PATCH[%s/%s]: acquire lock", dataType, dataId))
	idKey := fmt.Sprintf("%s/%s", dataType, dataId)
	h.Lock.Aquire(idKey, "HandlerPatch")
	h.Log(fmt.Sprintf("Handler PATCH[%s/%s]: lock acquired, GetData", dataType, dataId))
	defer h.Lock.Release(idKey, "HandlerPatch")
	patchData, err := h.GetData(dataType, dataId)
	if err != nil {
		h.Log(fmt.Sprintf("cannot get data [%s/%s]", dataType, dataId))
		h.Log(err.Error())
		return nil, err
	}
	patchRecord, e := Record.LoadMap(patchData)
	if e != nil {
		errMsg := fmt.Sprintf("failed to load data [%s/%s] as record", dataType, dataId)
		h.Log(errMsg)
		h.Log(e.Error())
		return nil, Http.WrapError(e, fmt.Sprintf("failed to load data [%s/%s] as record", dataType, dataId), http.StatusInternalServerError)
	}
	patchVer, ok := headers[JsonKey.Version]
	if ok {
		h.Log(fmt.Sprintf("PATCH[%s/%s]: header path version [%s]", dataType, dataId, patchVer))
		if patchRecord.Version != patchVer {
			errMsg := fmt.Sprintf("current record:[%s/%s] version:[%s] does not match specified version:[%s]", dataType, dataId, patchRecord.Version, patchVer)
			h.Log(errMsg)
			return nil, Http.NewHttpError(errMsg, http.StatusNotModified)
		}
		h.Log("version match with header")
	}
	h.Log(fmt.Sprintf("Handler PATCH[%s/%s]: get version schema [%s]", dataType, dataId, patchRecord.Version))
	schema, err := h.LocalSchema(dataType, patchRecord.Version)
	if err != nil {
		h.Log(err.Error())
		return nil, err
	}
	before := Record.Record{}
	e = Json.CopyTo(patchRecord, &before)
	if e != nil {
		errMsg := fmt.Sprintf("failed to snapshot data [%s/%s]", dataType, dataId)
		h.Log(errMsg)
		h.Log(e.Error())
		return nil, Http.WrapError(e, errMsg, http.StatusInternalServerError)
	}
	h.Log(fmt.Sprintf("PATCH [%s/%s] @[%s]", dataType, dataId, nextPath))
	err = patchRecordByPath(schema.Schema, patchRecord, nextPath, fmt.Sprintf("%s/%s", dataType, dataId), data)
	if err != nil {
		if err.Status != http.StatusNotModified {
			h.Log(err.Error())
			return nil, err
		}
		return patchRecord.Map(), nil
	}
	verComp, err := CompareVersion(before.Version, patchRecord.Version)
	if err != nil {
		return nil, Http.WrapError(err, "failed to compare version", http.StatusBadRequest)
	}
	if verComp < 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("downgrade data format are not supported. version[%s] -> [%s]", before.Version, patchRecord.Version), http.StatusBadRequest)
	}
	err = h.updateRecord(before.Type, before.Id, patchRecord)
	if err != nil {
		h.Log(err.Error())
		return nil, err
	}
	h.Log(fmt.Sprintf("PATCH [%s/%s] complete", dataType, dataId))
	if h.AddJournal != nil {
		h.Log(fmt.Sprintf("PATCH [%s/%s] add journal", dataType, dataId))
		h.AddJournal(dataType, dataId, before.Map(), patchRecord.Map())
	} else {
		h.Log("AddJournal is not defined here.")
	}
	h.Log(fmt.Sprintf("PATCH [%s/%s] return patched record", dataType, dataId))
	return patchRecord.Map(), nil
}

func patchRecordByPath(schema *SchemaDoc.SchemaDoc, record *Record.Record, nextPath string, dataPath string, newData interface{}) *Http.HttpError {
	switch nextPath {
	case Record.DataId:
		if record.Type != JsonKey.Schema {
			return Http.NewHttpError(fmt.Sprintf("only record type of [%s], allow patch on [%s]", JsonKey.Schema, Record.DataId), http.StatusBadRequest)
		}
		archiveVer := record.Data[JsonKey.Version].(string)
		archiveType := record.Data[JsonKey.Name].(string)
		archiveId := SchemaDoc.ArchivedSchemaId(archiveType, archiveVer)
		if archiveId != newData.(string) {
			return Http.NewHttpError(fmt.Sprintf("invalid archive schema Id[%s], expect[%s]", newData, archiveId), http.StatusBadRequest)
		}
		if record.Id == newData.(string) {
			return Http.NewHttpError(fmt.Sprintf("id already updated. [%s]", record.Id), http.StatusNotModified)
		}
		record.Id = archiveId
	case Record.DataType:
		return Http.NewHttpError("Change on Record Data Type is not supported", http.StatusNotModified)
	case Record.Version:
		if record.Version == newData.(string) {
			return Http.NewHttpError(fmt.Sprintf("Type Version already updated. [%s]", record.Version), http.StatusNotModified)
		}
		record.Version = newData.(string)
	default:
		if record.Type == JsonKey.Schema {
			return Http.NewHttpError(fmt.Sprintf("Patch on [%s] not allowed", JsonKey.Schema), http.StatusBadRequest)
		}
		err := Schema.SetDataOnPath(schema, record.Data, nextPath, dataPath, newData)
		if err != nil {
			return err
		}
	}
	return nil
}
