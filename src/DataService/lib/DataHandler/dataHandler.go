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
	"UniTao/Data"
	"UniTao/Data/DbIface"
	"UniTao/DataService/lib/Config"
	"UniTao/Schema"
	"UniTao/Util"
	"fmt"
	"net/http"
	"path"
	"reflect"
)

type Handler struct {
	db        DbIface.Database
	schemaMap map[string]*Schema.Record
	config    Config.Confuguration
}

func New(config Config.Confuguration) (*Handler, error) {
	db, err := Data.ConnectDb(config.Database)
	if err != nil {
		return nil, err
	}
	handler := Handler{
		schemaMap: make(map[string]*Schema.Record),
		db:        db,
		config:    config,
	}
	return &handler, nil
}

func (h *Handler) List(dataType string) ([]string, int, error) {
	if dataType != Schema.Schema && dataType != "" {
		_, code, err := h.GetData(Schema.Schema, dataType)
		if err != nil {
			err = fmt.Errorf("failed to get schema for type=[%s], Err:%s", dataType, err)
			return nil, code, err
		}
	}
	args := make(map[string]interface{})
	args[DbIface.Table] = Schema.Data
	args[Schema.DataType] = dataType
	if args[Schema.DataType] == "" {
		args[Schema.DataType] = Schema.Schema
	}
	recordList, err := h.db.Get(args)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	result := make([]string, 0, len(recordList))
	for _, record := range recordList {
		result = append(result, record[Schema.DataId].(string))
	}
	return result, http.StatusOK, nil
}

func (h *Handler) Get(dataType string, dataId string) (map[string]interface{}, int, error) {
	if dataType == Schema.Schema {
		return h.GetData(Schema.Schema, dataId)
	}
	_, code, err := h.GetData(Schema.Schema, dataType)
	if err != nil {
		err = fmt.Errorf("failed to get schema for type=[%s], Err:%s", dataType, err)
		return nil, code, err
	}
	return h.GetData(dataType, dataId)
}

func (h *Handler) GetData(dataType string, dataId string) (map[string]interface{}, int, error) {
	args := make(map[string]interface{})
	args[DbIface.Table] = Schema.Data
	args[Schema.DataType] = dataType
	args[Schema.DataId] = dataId
	recordList, err := h.db.Get(args)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(recordList) == 0 {
		return nil, http.StatusNotFound, fmt.Errorf("failed to find [{type}/{id}]=[%s/%s]", dataType, dataId)
	}
	return recordList[0], http.StatusOK, nil
}

func (h *Handler) Lock(dataType string, dataId string) (int, error) {
	if dataType == Schema.Schema {
		return http.StatusAccepted, nil
	}
	_, code, err := h.GetData(Schema.Schema, dataType)
	if err != nil {
		err = fmt.Errorf("failed to get schema for data type=[%s], Err:%s", dataType, err)
		return code, err
	}
	return http.StatusAccepted, err
}

func (h *Handler) localSchema(dataType string) (*Schema.Record, int, error) {
	if dataType == Schema.Schema {
		return nil, http.StatusBadRequest, fmt.Errorf("should not get schema of schema")
	}
	record, ok := h.schemaMap[dataType]
	if ok {
		return record, http.StatusOK, nil
	}
	data, code, err := h.GetData(Schema.Schema, dataType)
	if err != nil {
		err = fmt.Errorf("failed to get schema for type=[%s], Err:%s", dataType, err)
		return nil, code, err
	}
	schema, err := Schema.LoadRecord(data)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to load Schema Record, [%s]=[%s], Err:\n%s", Schema.DataType, dataType, err)
	}
	h.schemaMap[dataType] = schema
	return schema, http.StatusOK, nil
}

func (h *Handler) inventoryData(dataType string, value string) (map[string]interface{}, int, error) {
	if h.config.Inv.Url == "" {
		return nil, http.StatusInternalServerError, fmt.Errorf("missing inventory URL in configuration [inventory/url]. at Data Service=[%s]", h.config.Http.Id)
	}
	dataUrl, err := Util.URLPathJoin(h.config.Inv.Url, dataType, value)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("failed to build ref data url. [url/type/id]=[%s/%s/%s]", h.config.Inv.Url, dataType, value)
	}
	data, code, err := Util.GetRestData(*dataUrl)
	if err == nil {
		mapData, ok := data.(map[string]interface{})
		if !ok {
			return nil, http.StatusBadRequest, fmt.Errorf("return data is not an object. [url]=[%s]", *dataUrl)
		}
		return mapData, code, err
	}
	return nil, code, err
}

func (h *Handler) Validate(dataType string, dataId string, data map[string]interface{}) (int, error) {
	if dataType != data[Schema.DataType].(string) {
		return http.StatusBadRequest, fmt.Errorf("path type [%s]!= payload type[%s]", dataType, data[Schema.DataType].(string))
	}
	if dataId != data[Schema.DataId].(string) {
		return http.StatusBadRequest, fmt.Errorf("path id [%s]!= payload id[%s]", dataId, data[Schema.DataId].(string))
	}
	schema, code, err := h.localSchema(dataType)
	if err != nil {
		return code, err
	}
	err = schema.Validate(data)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to validate payload against schema for type %s, Err: \n%s", dataType, err)
	}
	code, err = h.ValidateDataRefs(schema.Schema, data[Schema.RecordData].(map[string]interface{}), path.Join(dataType, dataId))
	if err != nil {
		return code, err
	}
	return code, nil
}

func (h *Handler) ValidateDataRefs(doc *Schema.Doc, data interface{}, dataPath string) (int, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return http.StatusBadRequest, fmt.Errorf("failed to convert data to map")
	}
	code, err := h.validateRefs(doc, dataMap, dataPath)
	if err != nil {
		return code, err
	}
	return h.validateSubDoc(doc, dataMap, dataPath)
}

func (h *Handler) validateRefs(doc *Schema.Doc, data map[string]interface{}, dataPath string) (int, error) {
	for _, ref := range doc.Refs {
		value, ok := data[ref.Name]
		if !ok {
			continue
		}
		refPath := path.Join(dataPath, ref.Name)
		switch reflect.TypeOf(value).Kind() {
		case reflect.String:
			hasRef, err := h.validateRefValue(ref, value.(string))
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("failed to validate value with [type]=[%s], [path]=[%s], Err:%s", ref.ContentType, refPath, err)
			}
			if !hasRef {
				return http.StatusBadRequest, fmt.Errorf("ref does not exists,[type]=[%s], [path]=[%s], [ref]=[%s]", ref.ContentType, refPath, value)
			}
		case reflect.Map:
			for key, keyVal := range value.(map[string]string) {
				keyPath := path.Join(refPath, key)
				hasRef, err := h.validateRefValue(ref, keyVal)
				if err != nil {
					return http.StatusBadRequest, fmt.Errorf("failed to validate value with [type]=[%s], [path]=[%s], Err:%s", ref.ContentType, keyPath, err)
				}
				if !hasRef {
					return http.StatusBadRequest, fmt.Errorf("ref does not exists,[type]=[%s], [path]=[%s], [ref]=[%s]", ref.ContentType, keyPath, keyVal)
				}
			}
		case reflect.Slice:
			for idx, item := range value.([]interface{}) {
				idxPath := fmt.Sprintf("%s[%d]", refPath, idx)
				hasRef, err := h.validateRefValue(ref, item.(string))
				if err != nil {
					return http.StatusBadRequest, fmt.Errorf("failed to validate value with [type]=[%s], [path]=[%s], Err:%s", ref.ContentType, idxPath, err)
				}
				if !hasRef {
					return http.StatusBadRequest, fmt.Errorf("ref does not exists,[type]=[%s], [path]=[%s], [ref]=[%s]", ref.ContentType, idxPath, item.(string))
				}
			}
		}
	}
	return http.StatusAccepted, nil
}

func (h *Handler) validateSubDoc(doc *Schema.Doc, data map[string]interface{}, dataPath string) (int, error) {
	for pPath, pDoc := range doc.SubDoc {
		pname, key := Util.ParsePath(pPath)
		subPath := path.Join(dataPath, pname)
		subData, ok := data[pname]
		if !ok {
			// property does not exists
			continue
		}
		if key == "" {
			// property is not an itemized type (slice or map)
			return h.ValidateDataRefs(pDoc, subData, subPath)
		}
		switch reflect.TypeOf(subData).Kind() {
		case reflect.Map:
			// check each key value of map
			for key, keyData := range subData.(map[string]interface{}) {
				keyPath := path.Join(subPath, key)
				code, err := h.ValidateDataRefs(pDoc, keyData, keyPath)
				if err != nil {
					return code, err
				}
			}
		case reflect.Slice:
			// check each item value of array
			for idx, idxData := range subData.([]interface{}) {
				idxPath := fmt.Sprintf("%s[%d]", subPath, idx)
				code, err := h.ValidateDataRefs(pDoc, idxData, idxPath)
				if err != nil {
					return code, err
				}
			}
		default:
			// data value does not match schema
			return http.StatusBadRequest, fmt.Errorf("data is not [%s or %s] @[path]=[%s]", reflect.Map, reflect.Slice, subPath)
		}
		// itemized data type passed
		return http.StatusAccepted, nil
	}

	return http.StatusAccepted, nil
}

func (h *Handler) validateRefValue(ref *Schema.DocRef, value string) (bool, error) {
	typePath, dataType := Util.ParsePath(ref.ContentType)
	if typePath != Schema.Inventory {
		// ContentMediaType not start with inventory, we don't understand
		return true, nil
	}
	if dataType == Schema.Schema {
		return false, fmt.Errorf("should not refer to schema of schema as data type")
	}
	_, code, err := h.localSchema(dataType)
	if err != nil && code != http.StatusNotFound {
		return false, err
	}
	if code == http.StatusNotFound {
		_, code, err := h.inventoryData(dataType, value)
		if err != nil && code != http.StatusNotFound {
			return false, fmt.Errorf("failed to query data from inventory, [type/id]=[%s/%s]", dataType, value)
		}
		if code == http.StatusNotFound {
			return false, nil
		}
		return true, nil
	}
	_, code, err = h.GetData(dataType, value)
	if err != nil {
		if code == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (h *Handler) Add(dataType string, dataId string, payload map[string]interface{}) (int, error) {
	code, err := h.Validate(dataType, dataId, payload)
	if err != nil {
		return code, err
	}
	code, err = h.Lock(dataType, dataId)
	if err != nil {
		return code, err
	}
	_, code, err = h.GetData(dataType, dataId)
	if err == nil {
		return http.StatusConflict, fmt.Errorf("data [type/id]=[%s/%s] already exists", dataType, dataId)
	}
	if code != http.StatusNotFound {
		return code, fmt.Errorf("failed to check exists for add [type/id]=[%s/%s], Err:%s", dataType, dataId, err)
	}
	err = h.db.Create(Schema.Data, payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create record [{type}/{id}]=[%s]/%s, Err:%s", dataType, dataId, err)
	}
	return http.StatusAccepted, nil
}

func (h *Handler) Set(dataType string, dataId string, payload map[string]interface{}) (int, error) {
	code, err := h.Validate(dataType, dataId, payload)
	if err != nil {
		return code, err
	}
	code, err = h.Lock(dataType, dataId)
	if err != nil {
		return code, err
	}
	err = h.db.Create(Schema.Data, payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create record [{type}/{id}]=[%s]/%s, Err:%s", dataType, dataId, err)
	}
	return http.StatusAccepted, nil
}

func (h *Handler) Delete(dataType string, dataId string) (int, error) {
	code, err := h.Lock(dataType, dataId)
	if err != nil {
		return code, err
	}
	keys := make(map[string]interface{})
	keys[Schema.DataType] = dataType
	keys[Schema.DataId] = dataId
	err = h.db.Delete(Schema.Data, keys)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete record [type/id]=[%s/%s], Err:%s", dataType, dataId, err)
	}
	return http.StatusAccepted, nil
}
