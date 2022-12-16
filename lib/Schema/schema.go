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
	for _, c := range JsonKey.InvalidKeyChars {
		if strings.Contains(record.Id, c) {
			return fmt.Errorf("invalid char found in record id:[%s], please make sure not include following chars:\n%s", record.Id, JsonKey.InvalidKeyChars)
		}
	}
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
		schemaData, _ := Json.CopyToMap(record.Data)
		s, err := SchemaDoc.New(schemaData)
		if err != nil {
			return err
		}
		autoIdxList := CmtIndex.FindAutoIndex(s, "")
		errList := make([]string, 0, len(autoIdxList))
		for _, autoIdx := range autoIdxList {
			err := CmtIndex.ValidateIndexTemplate(record.Id, autoIdx)
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
			invalidChars := Util.CheckInvalidKeys(JsonKey.InvalidKeyChars, value)
			if len(invalidChars) > 0 {
				return fmt.Errorf("in map %s, found invalid chars %s. @path=[%s]", attr, invalidChars, dataPath)
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
		if newData == nil {
			return delAttrData(schema, data, attrPath, prevPath)
		}
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

func delAttrData(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrPath string, prevPath string) *Http.HttpError {
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return Http.NewHttpError(fmt.Sprintf("invalid attrPath=[%s] @[%s]", attrPath, prevPath), http.StatusBadRequest)
	}
	attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("attr=[%s] not defined at path=[%s]", attrName, prevPath), http.StatusBadRequest)
	}
	if key == "" {
		if _, ok := data[attrName]; !ok {
			return Http.NewHttpError(fmt.Sprintf("attr[%s] already deleted @path=[%s]", attrName, prevPath), http.StatusNotModified)
		}
		delete(data, attrName)
		return nil
	}
	switch attrDef[JsonKey.Type].(string) {
	case JsonKey.Array:
		ex := delAttrAry(schema, data, attrName, attrDef[JsonKey.Items].(map[string]interface{}), key)
		if ex != nil {
			ex.Context = append(ex.Context, fmt.Sprintf("path=[%s/%s]", prevPath, attrName))
			return ex
		}
	case JsonKey.Object:
		mapData := data[attrName].(map[string]interface{})
		if !SchemaDoc.IsMap(attrDef) {
			return Http.NewHttpError(fmt.Sprintf("invalid path, attr[%s] is not a map. @path=[%s]", attrName, prevPath), http.StatusBadRequest)
		}
		if _, ok := mapData[key]; !ok {
			return Http.NewHttpError(fmt.Sprintf("key[%s] does not exists to delete", key), http.StatusNotModified)
		}
		delete(mapData, key)
	default:
		return Http.NewHttpError(fmt.Sprintf("invalid path attr[%s] is a [%s] which has no key. @path=[%s]", attrName, attrDef[JsonKey.Type], prevPath), http.StatusBadRequest)
	}
	return nil
}

func delAttrAry(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrName string, itemDef map[string]interface{}, key string) *Http.HttpError {
	attrData, ok := data[attrName]
	if !ok {
		return nil
	}
	arrayData := attrData.([]interface{})
	itemType := itemDef[JsonKey.Type].(string)
	if itemType != JsonKey.Object {
		isIdx := true
		setIdx, err := strconv.Atoi(key)
		if err != nil {
			isIdx = false
		}
		if isIdx {
			newAry := Util.ListDel(arrayData, setIdx)
			if newAry == nil {
				return Http.NewHttpError(fmt.Sprintf("idx[%d] does not exists", setIdx), http.StatusNotModified)
			}
			data[attrName] = newAry
			return nil
		}
		newArray := []interface{}{}
		foundKey := false
		for _, item := range arrayData {
			if item.(string) != key {
				newArray = append(newArray, item.(string))
				continue
			}
			foundKey = true
		}
		if !foundKey {
			return Http.NewHttpError(fmt.Sprintf("key[%s] does not exists to delete", key), http.StatusNotModified)
		}
		data[attrName] = newArray
		return nil
	}
	subDoc := schema.SubDocs[attrName]
	newArray := []interface{}{}
	foundKey := false
	for idx, item := range arrayData {
		itemKey, err := subDoc.BuildKey(item.(map[string]interface{}))
		if err != nil {
			return Http.WrapError(err, fmt.Sprintf("failed to get key @[%d]", idx), http.StatusInternalServerError)
		}
		if itemKey != key {
			newArray = append(newArray, item)
			continue
		}
		foundKey = true
	}
	if !foundKey {
		return Http.NewHttpError(fmt.Sprintf("object of [%s] does not exists to delete", key), http.StatusNotModified)
	}
	data[attrName] = newArray
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
	switch attrDef[JsonKey.Type].(string) {
	case JsonKey.Array:
		ex := setAttrAry(schema, data, attrName, attrDef[JsonKey.Items].(map[string]interface{}), key, newData)
		if ex != nil {
			ex.Context = append(ex.Context, fmt.Sprintf("path @[%s/%s]", prevPath, attrName))
			return ex
		}
	case JsonKey.Object:
		if !SchemaDoc.IsMap(attrDef) {
			if key != "" {
				return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not supported. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
			}
			data[attrName] = newData
		} else {
			itemType := attrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
			attrData := data[attrName].(map[string]interface{})
			if key == "" {
				newKey, ex := getHashKey(schema, itemType, newData, attrName)
				if ex != nil {
					ex.Context = append(ex.Context, fmt.Sprintf("path @[%s/%s]", prevPath, attrName))
				}
				key = newKey
			}
			attrData[key] = newData
		}
	default:
		if key != "" {
			return Http.NewHttpError(fmt.Sprintf("invalid path, attr=[%s], type=[%s] key=[%s] is not empty. path=[%s]", attrName, attrDef[JsonKey.Type].(string), key, prevPath), http.StatusBadRequest)
		}
		data[attrName] = newData
	}
	return nil
}

func setAttrAry(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, attrName string, itemDef map[string]interface{}, key string, newData interface{}) *Http.HttpError {
	if reflect.TypeOf(newData).Kind() == reflect.Slice {
		if key != "" {
			return Http.NewHttpError(fmt.Sprintf("invalid path. key[%s] expect to be empty", key), http.StatusBadRequest)
		}
		data[attrName] = newData
		return nil
	}
	attrData, ok := data[attrName].([]interface{})
	if !ok {
		data[attrName] = []interface{}{
			newData,
		}
		return nil
	}
	itemType := itemDef[JsonKey.Type].(string)
	if key != "" {
		if itemType != JsonKey.Object {
			idx, ex := strconv.Atoi(key)
			if ex == nil {
				switch {
				case idx < 0:
					data[attrName] = append([]interface{}{newData}, attrData...)
				case idx < len(attrData):
					attrData[idx] = newData
				default:
					data[attrName] = append(attrData, newData)
				}
				return nil
			} else {
				if itemType != JsonKey.String {
					return Http.NewHttpError(fmt.Sprintf("item type=[%s] does not support idx:[%s]", itemType, key), http.StatusBadRequest)
				}
			}
		}
	}
	newKey, ex := getHashKey(schema, itemType, newData, attrName)
	if ex != nil {
		return ex
	}
	if key == "" {
		key = newKey
	}
	newArray := []interface{}{}
	foundKey := false
	for idx, item := range attrData {
		var itemKey string
		if itemType == JsonKey.Object {
			k, err := schema.SubDocs[attrName].BuildKey(item.(map[string]interface{}))
			if err != nil {
				return Http.WrapError(err, fmt.Sprintf("failed to get key @idx=[%d]", idx), http.StatusInternalServerError)
			}
			itemKey = k
		} else {
			itemKey = item.(string)
		}
		if itemKey == key {
			newArray = append(newArray, newData)
			foundKey = true
			continue
		}
		newArray = append(newArray, item)
	}
	if !foundKey {
		if key != newKey {
			return Http.NewHttpError(fmt.Sprintf("key=[%s] does not exists in list to modify", key), http.StatusNotFound)
		}
		newArray = append(newArray, newData)
	}
	data[attrName] = newArray
	return nil
}

func getHashKey(schema *SchemaDoc.SchemaDoc, itemType string, newData interface{}, attrName string) (string, *Http.HttpError) {
	switch itemType {
	case JsonKey.Object:
		subDoc := schema.SubDocs[attrName]
		subKey, err := subDoc.BuildKey(newData.(map[string]interface{}))
		if err != nil {
			return "", Http.WrapError(err, "failed to build key from given data.", http.StatusBadRequest)
		}
		return subKey, nil
	case JsonKey.String:
		return newData.(string), nil
	default:
		return "", Http.NewHttpError(fmt.Sprintf("item type=[%s] does not support, only [%s, %s] support build key", itemType, JsonKey.String, JsonKey.Map), http.StatusBadRequest)
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
		// key already deleted
		return Http.NewHttpError(fmt.Sprintf("key=[%s] already deleted", key), http.StatusNotModified)
	}
	if key != newData.(string) {
		// key to repoint does not exists, return not found
		return Http.NewHttpError(fmt.Sprintf("key=[%s] does not exists", key), http.StatusNotFound)
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
