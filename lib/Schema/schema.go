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
	"reflect"
	"strconv"
	"strings"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
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
	schemaData, err := Json.Copy(schema.Record.Data)
	if err != nil {
		return fmt.Errorf("copy schema.Record.Data failed. Error: %s", err)
	}
	doc, err := SchemaDoc.New(schemaData.(map[string]interface{}))
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
	if schema.Schema.Id != record.Type {
		return fmt.Errorf("schema id and payload data type does not match, [%s]!=[%s]", schema.Record.Id, record.Type)
	}
	if schema.Schema.Version != record.Version {
		return fmt.Errorf("schema version=[%s] does not match record schema version=[%s]", schema.Schema.Version, record.Version)
	}
	_, err := Record.ParseVersion(record.Version)
	if err != nil {
		return err
	}
	if record.Type == JsonKey.Schema {
		for _, char := range JsonKey.InvalidTypeChars {
			if strings.Contains(record.Id, char) {
				return fmt.Errorf("data type name [%s] contain illigal key [%s]", record.Type, char)
			}
		}
		s, err := SchemaDoc.New(record.Data)
		if err != nil {
			return err
		}
		autoIdxList := CmtIndex.FindAutoIndex(s, "")
		errList := make([]string, 0, len(autoIdxList))
		for _, autoIdx := range autoIdxList {
			err := CmtIndex.ValidateIndexTemplate(autoIdx)
			if err != nil {
				errList = append(errList, err.Error())
			}
		}
		if len(errList) > 0 {
			errMsg, _ := json.MarshalIndent(errList, "", "     ")
			return fmt.Errorf(string(errMsg))
		}
	} else {
		if strings.Contains(record.Type, JsonKey.ArchivedSchemaIdDiv) {
			return fmt.Errorf("cannot add data with archived dataType=[%s]", record.Type)
		}
	}
	err = schema.Meta.Validate(record.Data)
	if err != nil {
		return fmt.Errorf("schema validation failed. Error:\n%s", err)
	}
	if len(schema.Schema.KeyTemplate.Vars) > 0 {
		dataId, err := schema.Schema.BuildKey(record.Data)
		if err != nil {
			return fmt.Errorf("failed to build Key from data. template=[%s], Error:%s", schema.Schema.KeyTemplate.Template, err)
		}
		if dataId != record.Id {
			return fmt.Errorf("invalid record Id=[%s] does not match key tempate=[%s] value=[%s]", record.Id, schema.Schema.KeyTemplate.Template, dataId)
		}
	}
	err = ValidateSchemaKeys(schema.Schema, record.Data, "")
	if err != nil {
		return err
	}
	return nil
}

func ValidateSchemaKeys(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string) error {
	properties := schema.Data[JsonKey.Properties].(map[string]interface{})
	for attr := range properties {
		attrDef := properties[attr].(map[string]interface{})
		attrType := attrDef[JsonKey.Type].(string)
		switch attrType {
		case JsonKey.Array:
			valueList, ok := data[attr].([]interface{})
			if !ok || len(valueList) == 0 {
				continue
			}
			itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
			switch itemType {
			case JsonKey.String:
				if cmtRef, ok := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.ContentMediaType].(string); ok {
					cmtMap, err := Util.CountListIdx(valueList)
					if err != nil {
						return err
					}
					for key, count := range cmtMap {
						if count > 1 {
							return fmt.Errorf("duplicate key=[%s] on [%s]=[%s] path=[%s/%s]", key, JsonKey.ContentMediaType, cmtRef, dataPath, attr)
						}
					}
				}
			case JsonKey.Object:
				itemDoc := schema.SubDocs[attr]
				itemKeyMap := map[string]int{}
				for idx, item := range valueList {
					itemPath := fmt.Sprintf("%s/%s[%d]", dataPath, attr, idx)
					itemData := item.(map[string]interface{})
					err := ValidateSchemaKeys(itemDoc, itemData, itemPath)
					if err != nil {
						return err
					}
					itemKey, err := itemDoc.BuildKey(itemData)
					if err != nil {
						return fmt.Errorf("failed to build key @path [%s]", itemPath)
					}
					if _, ok := itemKeyMap[itemKey]; !ok {
						itemKeyMap[itemKey] = idx
					} else {
						return fmt.Errorf("duplicate item key=[%s] @path=[%s] & [%d]", itemKey, itemPath, itemKeyMap[itemKey])
					}
				}
			}
		case JsonKey.Object:
			value, ok := data[attr].(map[string]interface{})
			if !ok || len(value) == 0 {
				continue
			}
			nextDoc := schema.SubDocs[attr]
			if SchemaDoc.IsMap(attrDef) {
				itemType := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})[JsonKey.Type].(string)
				switch itemType {
				case JsonKey.String:
					if _, ok := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})[JsonKey.ContentMediaType]; ok {
						for key, cmt := range value {
							itemPath := fmt.Sprintf("%s/%s[%s]", dataPath, attr, key)
							if key != cmt.(string) {
								return fmt.Errorf("for hash of %s, key[%s] not match value [%s] @[%s]", JsonKey.ContentMediaType, key, cmt.(string), itemPath)
							}
						}
					}
				case JsonKey.Object:
					for key, item := range value {
						itemPath := fmt.Sprintf("%s/%s[%s]", dataPath, attr, key)
						itemObj := item.(map[string]interface{})
						err := ValidateSchemaKeys(nextDoc, itemObj, itemPath)
						if err != nil {
							return err
						}
						if len(nextDoc.KeyTemplate.Vars) > 0 {
							// only validate map key if item doc has key defined
							itemKey, err := nextDoc.BuildKey(itemObj)
							if err != nil {
								return fmt.Errorf("failed to generate item key @[%s], Error:%s", itemPath, err)
							}
							if key != itemKey {
								return fmt.Errorf("hash key[%s] not match object key[%s] @[%s]", key, itemKey, itemPath)
							}
						}
					}
				}
				continue
			}
			err := ValidateSchemaKeys(nextDoc, value, fmt.Sprintf("%s/%s", dataPath, attr))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SetDataOnPath(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, dataPath string, prevPath string, newData interface{}) *Http.HttpError {
	attrPath, nextPath := Util.ParsePath(dataPath)
	if nextPath == "" {
		return setAttrData(schema, data, attrPath, prevPath, newData)
	}
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return Http.NewHttpError(fmt.Sprintf("failed to parse attrPath=[%s] @path=[%s]", attrPath, prevPath), http.StatusBadRequest)
	}
	attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("invalid path does not exists. @[%s/%s]", prevPath, attrName), http.StatusBadRequest)
	}
	attrData, ok := data[attrName]
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("path=[%s/%s] does not exists", prevPath, attrPath), http.StatusBadRequest)
	}
	switch attrDef[JsonKey.Type].(string) {
	case JsonKey.Array:
		if key == "" {
			return Http.NewHttpError(fmt.Sprintf("need to specify key to drill in array. @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
		}
		attrAry := attrData.([]interface{})
		itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
		if itemType != JsonKey.Object {
			//compare key with generated key from sub object
			return Http.NewHttpError(fmt.Sprintf("array of [%s] cannot drill further @path=[%s]", itemType, prevPath), http.StatusBadRequest)
		}
		subDoc := schema.SubDocs[attrName]
		for idx := range attrAry {
			item := attrAry[idx].(map[string]interface{})
			itemKey, err := subDoc.BuildKey(item)
			if err != nil {
				return Http.NewHttpError(fmt.Sprintf("not able to compare key with item=[%d] @path=[%s/%s]", idx, prevPath, attrName), http.StatusInternalServerError)
			}
			if itemKey == key {
				return SetDataOnPath(subDoc, item, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath), newData)
			}
		}
		return Http.NewHttpError(fmt.Sprintf("item not found in array @path=[%s/%s]", prevPath, attrPath), http.StatusNotFound)
	case JsonKey.Object:
		mapData := attrData.(map[string]interface{})
		subDoc := schema.SubDocs[attrName]
		if !SchemaDoc.IsMap(attrDef) {
			if key != "" {
				return Http.NewHttpError(fmt.Sprintf("data is not map @path=[%s/%s] to drill in with key=[%s]", prevPath, attrName, key), http.StatusBadRequest)
			}
			return SetDataOnPath(subDoc, mapData, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath), newData)
		}
		if key == "" {
			return Http.NewHttpError(fmt.Sprintf("need to specify key to drill in map. @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
		}
		keyData, ok := mapData[key].(map[string]interface{})
		if !ok {
			return Http.NewHttpError(fmt.Sprintf("item not found in map @path=[%s/%s]", prevPath, attrPath), http.StatusNotFound)
		}
		err := SetDataOnPath(subDoc, keyData, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath), newData)
		if err != nil {
			return err
		}
		if len(subDoc.KeyTemplate.Vars) > 0 {
			// only update key when key is defined @item Do
			newKey, e := subDoc.BuildKey(keyData)
			if e != nil {
				return Http.WrapError(e, fmt.Sprintf("new data cannot generate key. @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
			}
			if newKey != key {
				delete(mapData, key)
				mapData[newKey] = keyData
			}
		}
	default:
		return Http.NewHttpError(fmt.Sprintf("invalid path, type=[%s] cannot walk in. @path=[%s/%s]", attrDef[JsonKey.Type].(string), prevPath, attrName), http.StatusBadRequest)
	}
	return nil
}

func setAttrData(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrPath string, prevPath string, newData interface{}) *Http.HttpError {
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
			isIdx := true
			setIdx, err := strconv.Atoi(key)
			if err != nil {
				isIdx = false
			}
			if !isIdx {
				if itemType == JsonKey.String {
					return setStrArrayItem(data, attrName, key, newData)
				}
				return Http.WrapError(err, fmt.Sprintf("invalid key=[%s] to set for simple array. cannot parse to int", key), http.StatusBadRequest)
			}
			return setArraySimple(data, attrName, setIdx, newData)
		}
		return setArrayObject(schema, data, attrName, key, newData)
	case JsonKey.Object:
		if !SchemaDoc.IsMap(attrDef) {
			return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not empty. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
		}
		itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
		attrData := data[attrName].(map[string]interface{})
		switch itemType {
		case JsonKey.Object:
			itemDoc := schema.SubDocs[attrName]
			newKey, err := itemDoc.BuildKey(newData.(map[string]interface{}))
			if err != nil {
				return Http.WrapError(err, fmt.Sprintf("failed to get key from newData, @path=[%s/%s]", prevPath, attrPath), http.StatusBadRequest)
			}
			if newKey != key {
				delete(attrData, key)
			}
			attrData[newKey] = newData
		case JsonKey.String:
			if newData == nil {
				delete(attrData, key)
				return nil
			}
			strVal, ok := newData.(string)
			if !ok {
				return Http.NewHttpError(fmt.Sprintf("new Cmt Value is not a string. type=[%s] is not [string]", reflect.TypeOf(newData).Kind()), http.StatusBadRequest)
			}
			// for map, we don't need to check duplication, as the key is redfine its purpose even duplicated
			attrData[key] = strVal
		default:
			attrData[key] = newData
		}
		return nil
	default:
		return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not empty. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
	}
}

func setStrArrayItem(data map[string]interface{}, attrName string, key string, newData interface{}) *Http.HttpError {
	dataList := data[attrName].([]interface{})
	refHash := Util.IdxList(dataList)
	keyIdx, keyExists := refHash[key]
	refIdx := -1
	refExists := false
	if newData != nil {
		idx, ok := refHash[newData.(string)]
		refIdx = idx
		refExists = ok
	}
	if keyExists {
		if newData == nil {
			data[attrName] = append(dataList[:keyIdx], dataList[keyIdx+1:]...)
			return nil
		}
		if refExists {
			if refIdx != keyIdx {
				return Http.NewHttpError(fmt.Sprintf("invalid operation, ref=[%s] already exists.", newData.(string)), http.StatusBadRequest)
			}
			return Http.NewHttpError(fmt.Sprintf("key=[%s] already exists", key), http.StatusNotModified)
		}
		dataList[keyIdx] = newData.(string)
		return nil
	}
	if newData == nil {
		return Http.NewHttpError(fmt.Sprintf("key=[%s] does not exists", key), http.StatusNotModified)
	}
	if refExists {
		return Http.NewHttpError(fmt.Sprintf("key ref=[%s] already exists", newData.(string)), http.StatusNotModified)
	}
	data[attrName] = append(dataList, newData.(string))
	return nil
}

func setArraySimple(data map[string]interface{}, attrName string, idx int, newData interface{}) *Http.HttpError {
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

func setArrayObject(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrName string, key string, newData interface{}) *Http.HttpError {
	subSchema := schema.SubDocs[attrName]
	dataList := data[attrName].([]interface{})
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
		return nil
	}
	return Http.NewHttpError("key=[%s] not found. no change", http.StatusNotModified)
}
