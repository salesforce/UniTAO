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

package CmtIndex

import (
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Template"
)

type AutoIndex struct {
	AttrPath      string
	ContentType   string
	IndexTemplate string
}

// validate all variable required by indexTemplate are defined in target schema root
func (idx *AutoIndex) ValidateIndexTemplate(schema *SchemaDoc.SchemaDoc) *Http.HttpError {
	temp, err := Template.ParseStr(idx.IndexTemplate, "{", "}")
	if err != nil {
		return Http.WrapError(err, fmt.Sprintf("IndexTemplate[%s]: template parse failed", idx.IndexTemplate), http.StatusBadRequest)
	}
	errList := []string{}
	for _, iAttrName := range temp.Vars {
		attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[iAttrName]
		if !ok {
			errList = append(errList, fmt.Sprintf("attr[%s] not defined in [%s/%s]", iAttrName, schema.Id, schema.Version))
			continue
		}
		attrType := attrDef.(map[string]interface{})[JsonKey.Type].(string)
		if attrType != JsonKey.String {
			errList = append(errList, fmt.Sprintf("attr[%s] type[%s] is not [%s]", iAttrName, attrType, JsonKey.String))
		}
	}
	if len(errList) > 0 {
		ex := Http.NewHttpError(fmt.Sprintf("IndexTemplate[%s]: variable validation failed", idx.IndexTemplate), http.StatusBadRequest)
		ex.Context = append(ex.Context, errList...)
		return ex
	}
	return nil
}

func (idx *AutoIndex) ExplorerIdxPath(schema *SchemaDoc.SchemaDoc, record *Record.Record) map[string]map[string]interface{} {
	pathMap := explorerPath(schema, record.Data, "", idx.AttrPath)
	resultMap := map[string]map[string]interface{}{}
	for idxPath, valueMap := range pathMap {
		typePath := fmt.Sprintf("%s/%s%s", record.Type, record.Id, idxPath)
		resultMap[typePath] = valueMap
	}
	return resultMap
}

func explorerPath(schema *SchemaDoc.SchemaDoc, data map[string]interface{}, prevPath string, nextPath string) map[string]map[string]interface{} {
	attr, nextPath := Util.ParsePath(nextPath)
	attrName, _, ex := Util.ParseArrayPath(attr)
	if ex != nil {
		return nil
	}
	attrDef, ok := schema.Data[JsonKey.Properties].(map[string]interface{})[attrName]
	if !ok {
		return nil
	}
	attrType := attrDef.(map[string]interface{})[JsonKey.Type].(string)
	if attrType != JsonKey.Array && attrType != JsonKey.Object {
		return nil
	}
	nextData, ok := data[attrName]
	if !ok || nextData == nil {
		return nil
	}
	switch attrType {
	case JsonKey.Array:
		return explorerArrayPath(schema, attrName, nextData.([]interface{}), prevPath, nextPath)
	case JsonKey.Object:
		if !SchemaDoc.IsMap(attrDef.(map[string]interface{})) {
			nextSchema := schema.SubDocs[attrName]
			return explorerPath(nextSchema, nextData.(map[string]interface{}), fmt.Sprintf("%s/%s", prevPath, attrName), nextPath)
		}
		return explorerMapPath(schema, attrName, nextData.(map[string]interface{}), prevPath, nextPath)
	}
	return nil
}

func explorerArrayPath(schema *SchemaDoc.SchemaDoc, attrName string, arrayData []interface{}, prevPath string, nextPath string) map[string]map[string]interface{} {
	if nextPath == "" {
		return getArrayValue(prevPath, attrName, arrayData)
	}
	if len(arrayData) == 0 {
		return nil
	}
	return getArrayIdxPath(schema, attrName, arrayData, prevPath, nextPath)
}

func getArrayValue(prevPath string, attrName string, arrayData []interface{}) map[string]map[string]interface{} {
	valueMap := map[string]interface{}{}
	for idx, value := range arrayData {
		if _, ok := valueMap[value.(string)]; !ok {
			valueMap[value.(string)] = idx
		}
	}
	pathMap := map[string]map[string]interface{}{
		fmt.Sprintf("%s/%s", prevPath, attrName): valueMap,
	}
	return pathMap
}

func getArrayIdxPath(schema *SchemaDoc.SchemaDoc, attrName string, arrayData []interface{}, prevPath string, nextPath string) map[string]map[string]interface{} {
	pathMap := map[string]map[string]interface{}{}
	itemSchema := schema.SubDocs[attrName]
	for _, item := range arrayData {
		itemData := item.(map[string]interface{})
		key, ex := itemSchema.BuildKey(itemData)
		if ex != nil {
			continue
		}
		keyPath := fmt.Sprintf("%s/%s[%s]", prevPath, attrName, key)
		keyPathMap := explorerPath(itemSchema, itemData, keyPath, nextPath)
		for idxPath, valueMap := range keyPathMap {
			pathMap[idxPath] = valueMap
		}
	}
	return pathMap
}

func explorerMapPath(schema *SchemaDoc.SchemaDoc, attrName string, mapData map[string]interface{}, prevPath string, nextPath string) map[string]map[string]interface{} {
	if nextPath == "" {
		return getMapValue(prevPath, attrName, mapData)
	}
	if len(mapData) == 0 {
		return nil
	}
	return getMapIdxPath(schema, attrName, mapData, prevPath, nextPath)
}

func getMapValue(prevPath string, attrName string, mapData map[string]interface{}) map[string]map[string]interface{} {
	valueMap := map[string]interface{}{}
	for key, value := range mapData {
		if _, ok := valueMap[value.(string)]; !ok {
			valueMap[value.(string)] = key
		}
	}
	pathMap := map[string]map[string]interface{}{
		fmt.Sprintf("%s/%s", prevPath, attrName): valueMap,
	}
	return pathMap
}

func getMapIdxPath(schema *SchemaDoc.SchemaDoc, attrName string, mapData map[string]interface{}, prevPath string, nextPath string) map[string]map[string]interface{} {
	pathMap := map[string]map[string]interface{}{}
	itemSchema := schema.SubDocs[attrName]
	for key, keyData := range mapData {
		keyPath := fmt.Sprintf("%s/%s[%s]", prevPath, attrName, key)
		keyPathMap := explorerPath(itemSchema, keyData.(map[string]interface{}), keyPath, nextPath)
		for idxPath, valueMap := range keyPathMap {
			pathMap[idxPath] = valueMap
		}
	}
	return pathMap
}
