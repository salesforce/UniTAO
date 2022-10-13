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

package Schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	Inventory = "inventory"
)

type SchemaOps struct {
	Record     *Record.Record
	SchemaData map[string]interface{}
	Schema     *SchemaDoc.SchemaDoc
	Meta       *jsonschema.Schema
}

func LoadSchemaOpsRecord(record *Record.Record) (*SchemaOps, error) {
	recWithSchema := SchemaOps{
		Record: record,
	}
	err := recWithSchema.init()
	if err != nil {
		return nil, fmt.Errorf("failed while init SchemaOps. Error:%s", err)
	}
	return &recWithSchema, nil
}

func LoadSchemaOpsData(dataType string, typeVer string, data map[string]interface{}) (*SchemaOps, error) {
	dataId, ok := data[JsonKey.Name]
	if !ok {
		return nil, fmt.Errorf("missing required field=[%s] from data", JsonKey.Name)
	}
	record := Record.NewRecord(dataType, typeVer, dataId.(string), data)
	return LoadSchemaOpsRecord(record)
}

func (schema *SchemaOps) init() error {
	if schema.Record.Type != JsonKey.Schema {
		return fmt.Errorf("schema record has wrong [%s], [%s]!=[%s]", Record.DataType, schema.Record.Id, JsonKey.Schema)
	}
	schemaData, err := Util.JsonCopy(schema.Record.Data)
	if err != nil {
		return fmt.Errorf("copy schema.Record.Data failed. Error: %s", err)
	}
	doc, err := SchemaDoc.New(schemaData.(map[string]interface{}), schema.Record.Id, nil)
	if err != nil {
		return fmt.Errorf("failed to create Schema Doc, err: %s", err)
	}
	schema.Schema = doc
	schemaBytes, err := json.MarshalIndent(doc.Data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to MarshalIndent value [field]=[data], Err:%s", err)
	}
	meta, err := jsonschema.CompileString(schema.Record.Id, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to compile schema, [%s]=[%s] Err:%s", Record.DataId, schema.Record.Id, err)
	}
	schema.Meta = meta
	return nil
}

func (schema *SchemaOps) ValidateRecord(record *Record.Record) error {
	if schema.Record.Id != record.Type {
		return fmt.Errorf("schema id and payload data type does not match, [%s]!=[%s]", schema.Record.Id, record.Type)
	}
	return schema.ValidateData(record.Data)
}

func (schema *SchemaOps) ValidateData(data map[string]interface{}) error {
	err := schema.Meta.Validate(data)
	if err != nil {
		return fmt.Errorf("schema validation failed. Error:\n%s", err)
	}
	return nil
}

func GetDataOnPath(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string, prevPath string) (*SchemaDoc.SchemaDoc, map[string]interface{}, string, *Http.HttpError) {
	attrPath, nextPath := Util.ParsePath(dataPath)
	if nextPath == "" {
		return schema, data, attrPath, nil
	}
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return nil, nil, "", Http.NewHttpError(fmt.Sprintf("failed to parse attrPath=[%s] @path=[%s]", attrPath, prevPath), http.StatusBadRequest)
	}
	attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return nil, nil, "", Http.NewHttpError(fmt.Sprintf("invalid path does not exists. @[%s/%s]", prevPath, attrName), http.StatusBadRequest)
	}
	attrData, ok := data[attrName]
	if !ok {
		return nil, nil, "", Http.NewHttpError(fmt.Sprintf("path=[%s/%s] does not exists", prevPath, attrPath), http.StatusBadRequest)
	}
	switch attrDef[JsonKey.Type].(string) {
	case JsonKey.Array:
		if key == "" {
			return nil, nil, "", Http.NewHttpError(fmt.Sprintf("need to specify key to drill in array. @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
		}
		attrAry := attrData.([]interface{})
		itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
		if itemType != JsonKey.Object {
			//compare key with generated key from sub object
			return nil, nil, "", Http.NewHttpError(fmt.Sprintf("array of [%s] cannot drill further @path=[%s]", itemType, prevPath), http.StatusBadRequest)
		}
		subDoc := schema.SubDocs[attrName]
		for idx := range attrAry {
			item := attrAry[idx].(map[string]interface{})
			itemKey, err := subDoc.BuildKey(item)
			if err != nil {
				return nil, nil, "", Http.NewHttpError(fmt.Sprintf("not able to compare key with item=[%d] @path=[%s/%s]", idx, prevPath, attrName), http.StatusInternalServerError)
			}
			if itemKey == key {
				return GetDataOnPath(subDoc, item, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
			}
		}
		return nil, nil, "", Http.NewHttpError(fmt.Sprintf("item not found in array @path=[%s/%s]", prevPath, attrPath), http.StatusNotFound)
	case JsonKey.Object:
		mapData := attrData.(map[string]interface{})
		subDoc := schema.SubDocs[attrName]
		if !SchemaDoc.IsMap(attrDef) {
			if key != "" {
				return nil, nil, "", Http.NewHttpError(fmt.Sprintf("data is not map @path=[%s/%s] to drill in with key=[%s]", prevPath, attrName, key), http.StatusBadRequest)
			}
			return GetDataOnPath(subDoc, mapData, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
		}
		if key == "" {
			return nil, nil, "", Http.NewHttpError(fmt.Sprintf("need to specify key to drill in map. @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
		}
		keyData, ok := mapData[key].(map[string]interface{})
		if !ok {
			return nil, nil, "", Http.NewHttpError(fmt.Sprintf("item not found in map @path=[%s/%s]", prevPath, attrPath), http.StatusNotFound)
		}
		return GetDataOnPath(subDoc, keyData, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
	default:
		return nil, nil, "", Http.NewHttpError(fmt.Sprintf("invalid path, type=[%s] cannot walk in. @path=[%s/%s]", attrDef[JsonKey.Type].(string), prevPath, attrName), http.StatusBadRequest)
	}
}

func SetAttrData(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrPath string, prevPath string, newData interface{}) *Http.HttpError {
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return Http.NewHttpError(fmt.Sprintf("invalid attrPath=[%s] @[%s]", attrPath, prevPath), http.StatusBadRequest)
	}
	attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("attr=[%s] not defined at path=[%s]", attrName, prevPath), http.StatusBadRequest)
	}
	if key == "" {
		data[attrName] = newData
		return nil
	}
	switch attrDef[JsonKey.Type].(string) {
	case JsonKey.Array:
		itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
		if itemType != JsonKey.Object {
			if SchemaDoc.IsCmtRef(attrDef[JsonKey.Items].(map[string]interface{})) {
				return SetArrayCmt(data, attrName, key, newData)
			}
			setIdx, err := strconv.Atoi(key)
			if err != nil {
				return Http.WrapError(err, fmt.Sprintf("invalid key=[%s] to delete. cannot parse to int", key), http.StatusBadRequest)
			}
			return SetArraySimple(data, attrName, setIdx, newData)
		}
		return SetArrayObject(schema, data, attrName, key, newData)
	case JsonKey.Object:
		if !SchemaDoc.IsMap(attrDef) {
			return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not empty. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
		}
		data[attrName].(map[string]interface{})[key] = newData
		return nil
	default:
		return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not empty. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
	}
}

func SetArrayCmt(data map[string]interface{}, attrName string, key string, newData interface{}) *Http.HttpError {
	newRef := newData.(string)
	dataList := data[attrName].([]interface{})
	refHash := Util.IdxList(dataList)
	keyIdx, keyExists := refHash[key]
	refIdx, refExists := refHash[newRef]
	if keyExists {
		if refExists {
			if refIdx != keyIdx {
				return Http.NewHttpError(fmt.Sprintf("invalid operation, ref=[%s] already exists.", newRef), http.StatusBadRequest)
			}
			return nil
		}
		dataList[keyIdx] = newRef
		return nil
	}
	if refExists {
		return nil
	}
	if newData != nil {
		data[attrName] = append(dataList, newRef)
	}
	return nil
}

func SetArraySimple(data map[string]interface{}, attrName string, idx int, newData interface{}) *Http.HttpError {
	dataList := data[attrName].([]interface{})
	if newData == nil {
		// to delete on index
		if idx < 0 || idx >= len(dataList)-1 {
			return Http.NewHttpError(fmt.Sprintf("invalid idx=[%d], expect between (0 - %d)", idx, len(dataList)-1), http.StatusBadRequest)
		}
		data[attrName] = append(dataList[:idx], dataList[idx+1:]...)
	} else {
		switch i := idx; {
		case i < 0:
			data[attrName] = append([]interface{}{newData}, dataList...)
		case i > len(dataList)-1:
			data[attrName] = append(dataList, newData)
		default:
			dataList[idx] = newData
		}
	}
	return nil
}

func SetArrayObject(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrName string, key string, newData interface{}) *Http.HttpError {
	subSchema := schema.SubDocs[attrName]
	dataList := data[attrName].([]interface{})
	if newData != nil {
		newKey, err := subSchema.BuildKey(newData.(map[string]interface{}))
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to get key from request data. type=[%s]", schema.Id), http.StatusBadRequest)
		}
		if newKey != key {
			return Http.NewHttpError(fmt.Sprintf("path not match object key=[%s]", newKey), http.StatusBadRequest)
		}
	}
	for idx := range dataList {
		item := dataList[idx].(map[string]interface{})
		itemKey, err := subSchema.BuildKey(item)
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("data err idx=[%d]", idx), http.StatusInternalServerError)
		}
		if itemKey == key {
			if newData == nil {
				data[attrName] = append(dataList[:idx], dataList[idx+1:]...)
			} else {
				dataList[idx] = newData
			}
			return nil
		}
	}
	if newData != nil {
		data[attrName] = append(dataList, newData)
	}
	return nil
}
