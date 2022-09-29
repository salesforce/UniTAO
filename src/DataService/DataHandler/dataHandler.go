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

	"Data/DbConfig"
	"Data/DbIface"
	"DataService/Config"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

type Handler struct {
	db        DbIface.Database
	schemaMap map[string]*Schema.SchemaOps
	config    Config.Confuguration
}

func New(config Config.Confuguration, connectDb func(db DbConfig.DatabaseConfig) (DbIface.Database, error)) (*Handler, error) {
	db, err := connectDb(config.Database)
	if err != nil {
		return nil, err
	}
	handler := Handler{
		schemaMap: make(map[string]*Schema.SchemaOps),
		db:        db,
		config:    config,
	}
	return &handler, nil
}

func (h *Handler) List(dataType string) ([]string, *Http.HttpError) {
	if dataType != JsonKey.Schema && dataType != "" {
		_, err := h.GetData(JsonKey.Schema, dataType)
		if err != nil {
			return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
		}
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = h.config.DataTable.Data
	args[Record.DataType] = dataType
	if args[Record.DataType] == "" {
		args[Record.DataType] = JsonKey.Schema
	}
	recordList, e := h.db.Get(args)
	if e != nil {
		return nil, Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	result := make([]string, 0, len(recordList))
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
	args := make(map[string]interface{})
	args[DbIface.Table] = h.config.DataTable.Data
	args[Record.DataType] = dataType
	args[Record.DataId] = dataId
	recordList, err := h.db.Get(args)
	if err != nil {
		return nil, Http.NewHttpError(err.Error(), http.StatusInternalServerError)
	}
	if len(recordList) == 0 {
		return nil, Http.NewHttpError(fmt.Sprintf("object of type '%s' with id '%s' not found", dataType, dataId), http.StatusNotFound)
	}
	return recordList[0], nil
}

func (h *Handler) Lock(dataType string, dataId string) *Http.HttpError {
	if dataType == JsonKey.Schema {
		return nil
	}
	_, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
	}
	return nil
}

func (h *Handler) localSchema(dataType string) (*Schema.SchemaOps, *Http.HttpError) {
	schema, ok := h.schemaMap[dataType]
	if ok {
		return schema, nil
	}
	data, err := h.GetData(JsonKey.Schema, dataType)
	if err != nil {
		return nil, Http.WrapError(err, fmt.Sprintf("object of type “%s” does not exist", dataType), err.Status)
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
	schema, err := h.localSchema(record.Type)
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
	_, err := h.localSchema(ref.ContentType)
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
	err = h.Lock(record.Type, record.Id)
	if err != nil {
		return err
	}
	_, err = h.GetData(record.Type, record.Id)
	if err == nil {
		return Http.NewHttpError(fmt.Sprintf("data [type/id]=[%s/%s] already exists", record.Type, record.Id), http.StatusConflict)
	}
	if err.Status != http.StatusNotFound {
		return Http.WrapError(err, fmt.Sprintf("failed to check exists for add [type/id]=[%s/%s]", record.Type, record.Id), err.Status)
	}
	e := h.db.Create(h.config.DataTable.Data, record.Map())
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to create record [{type}/{id}]=[%s]/%s", record.Type, record.Id), http.StatusInternalServerError)
	}
	return nil
}

func (h *Handler) Set(record *Record.Record) *Http.HttpError {
	err := h.Validate(record)
	if err != nil {
		return err
	}
	err = h.Lock(record.Type, record.Id)
	if err != nil {
		return err
	}
	e := h.db.Create(h.config.DataTable.Data, record.Map())
	if e != nil {
		return Http.NewHttpError(e.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (h *Handler) Delete(dataType string, dataId string) *Http.HttpError {
	err := h.Lock(dataType, dataId)
	if err != nil {
		return err
	}
	keys := make(map[string]interface{})
	keys[Record.DataType] = dataType
	keys[Record.DataId] = dataId
	e := h.db.Delete(h.config.DataTable.Data, keys)
	if e != nil {
		return Http.WrapError(e, fmt.Sprintf("failed to delete record [type/id]=[%s/%s]", dataType, dataId), http.StatusInternalServerError)
	}
	return nil
}
