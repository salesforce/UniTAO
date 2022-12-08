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

package Compare

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
)

const (
	Add    = "add"
	Del    = "del"
	Modify = "modify"
)

type SchemaCompare struct {
	Schema *SchemaDoc.SchemaDoc
}

type Diff struct {
	Action   string
	DataPath string
	DataType string
	Source   interface{}
	Target   interface{}
}

func (c *SchemaCompare) CompareObj(source interface{}, target interface{}, dataPath string) []*Diff {
	nilResult := compareNil(source, target, dataPath, JsonKey.Object)
	if nilResult != nil {
		return nilResult
	}
	src := source.(map[string]interface{})
	tgt := target.(map[string]interface{})
	result := []*Diff{}
	for attr, value := range src {
		tgtValue, ok := tgt[attr]
		if !ok {
			attrChanges := c.CompareAttr(attr, value, nil, dataPath)
			result = append(result, attrChanges...)
			continue
		}
		attrChanges := c.CompareAttr(attr, value, tgtValue, dataPath)
		result = append(result, attrChanges...)
	}
	return result
}

func (c *SchemaCompare) CompareAttr(attrName string, source interface{}, target interface{}, dataPath string) []*Diff {
	attrPath := fmt.Sprintf("%s/%s", dataPath, attrName)
	attrDef, ok := c.Schema.Data[JsonKey.Properties].(map[string]interface{})[attrName].(map[string]interface{})
	if !ok {
		return compareNoSchema(source, target, attrPath)
	}
	attrType := attrDef[JsonKey.Type].(string)
	nilResult := compareNil(source, target, attrPath, attrType)
	if nilResult != nil {
		return nilResult
	}
	switch attrType {
	case JsonKey.Array:
		itemDef := attrDef[JsonKey.Items].(map[string]interface{})
		return c.compareArray(attrName, itemDef, source, target, attrPath)
	case JsonKey.Object:
		if SchemaDoc.IsMap(attrDef) {
			itemDef := attrDef[JsonKey.AdditionalProperties].(map[string]interface{})
			return c.compareMap(attrName, itemDef, source, target, attrPath)
		}
		attrSchema := c.Schema.SubDocs[attrName]
		attrCmp := SchemaCompare{Schema: attrSchema}
		return attrCmp.CompareObj(source, target, attrPath)
	default:
		if source != target {
			c := Diff{
				Action:   Modify,
				DataPath: attrPath,
				DataType: attrType,
				Source:   source,
				Target:   target,
			}
			return []*Diff{&c}
		}
		return []*Diff{}
	}
}

func (c *SchemaCompare) compareArray(attrName string, itemDef map[string]interface{}, source interface{}, target interface{}, dataPath string) []*Diff {
	srcList := source.([]interface{})
	tgtList := target.([]interface{})
	srcKeyMap := c.Schema.MapArray(attrName, srcList)
	tgtKeyMap := c.Schema.MapArray(attrName, tgtList)
	return c.compareMap(attrName, itemDef, srcKeyMap, tgtKeyMap, dataPath)
}

func (c *SchemaCompare) compareMap(attrName string, itemDef map[string]interface{}, source interface{}, target interface{}, dataPath string) []*Diff {
	srcMap := source.(map[string]interface{})
	tgtMap := target.(map[string]interface{})
	itemType := itemDef[JsonKey.Type].(string)
	var itemCmp *SchemaCompare
	if itemType == JsonKey.Object {
		itemSchema := c.Schema.SubDocs[attrName]
		itemCmp = &SchemaCompare{
			Schema: itemSchema,
		}
	}
	result := []*Diff{}
	for key, item := range srcMap {
		keyPath := fmt.Sprintf("%s[%s]", dataPath, key)
		if _, ok := tgtMap[key]; !ok {
			keyResult := compareNil(item, nil, keyPath, itemType)
			result = append(result, keyResult...)
			continue
		}
		if itemCmp != nil {
			keyResult := itemCmp.CompareObj(item, tgtMap[key], keyPath)
			result = append(result, keyResult...)
			continue
		}
		if item != tgtMap[key] {
			c := Diff{
				Action:   Modify,
				DataPath: keyPath,
				DataType: itemType,
				Source:   item,
				Target:   tgtMap[key],
			}
			result = append(result, &c)
		}
	}
	for key, item := range tgtMap {
		keyPath := fmt.Sprintf("%s[%s]", dataPath, key)
		if _, ok := srcMap[key]; !ok {
			keyResult := compareNil(nil, item, keyPath, itemType)
			result = append(result, keyResult...)
			continue
		}
	}
	return result
}

func compareNil(src interface{}, tgt interface{}, dataPath string, dataType string) []*Diff {
	if dataPath == "" {
		dataPath = "/"
	}
	if src == nil && tgt == nil {
		return []*Diff{}
	}
	if src == nil {
		c := Diff{
			Action:   Add,
			DataType: dataType,
			DataPath: dataPath,
			Target:   tgt,
		}
		return []*Diff{&c}
	}
	if tgt == nil {
		c := Diff{
			Action:   Del,
			DataType: dataType,
			DataPath: dataPath,
			Source:   src,
		}
		return []*Diff{&c}
	}
	return nil
}

func compareNoSchema(source interface{}, target interface{}, dataPath string) []*Diff {
	nilResult := compareNil(source, target, dataPath, "")
	if nilResult != nil {
		return nilResult
	}
	srcKind := reflect.TypeOf(source).Kind()
	tgtKind := reflect.TypeOf(target).Kind()
	if srcKind != tgtKind {
		c := Diff{
			Action:   Modify,
			DataType: "",
			Source:   source,
			Target:   target,
		}
		return []*Diff{&c}
	}
	switch srcKind {
	case reflect.Map:
		return compareNoSchemaMap(source, target, dataPath)
	case reflect.Slice:
		return compareNoSchemaSlice(source, target, dataPath)
	default:
		if source == target {
			c := Diff{
				Action:   Modify,
				DataType: "",
				Source:   source,
				Target:   target,
			}
			return []*Diff{&c}
		}
	}
	return []*Diff{}
}

func compareNoSchemaMap(source interface{}, target interface{}, dataPath string) []*Diff {
	diffs := []*Diff{}
	src := source.(map[string]interface{})
	tgt := target.(map[string]interface{})
	for key, value := range src {
		keyPath := fmt.Sprintf("%s/%s", dataPath, key)
		if _, ok := tgt[key]; !ok {
			keyDiffs := compareNoSchema(value, nil, keyPath)
			diffs = append(diffs, keyDiffs...)
			continue
		}
		keyDiffs := compareNoSchema(value, tgt[key], keyPath)
		diffs = append(diffs, keyDiffs...)
	}
	for key, value := range tgt {
		keyPath := fmt.Sprintf("%s/%s", dataPath, key)
		if _, ok := src[key]; !ok {
			keyDiffs := compareNoSchema(nil, value, keyPath)
			diffs = append(diffs, keyDiffs...)
		}
	}
	return diffs
}

func compareNoSchemaSlice(source interface{}, target interface{}, dataPath string) []*Diff {
	src, _ := json.MarshalIndent(source, "", "    ")
	tgt, _ := json.MarshalIndent(target, "", "    ")
	if string(src) != string(tgt) {
		c := Diff{
			Action:   Modify,
			DataType: "",
			Source:   source,
			Target:   target,
		}
		return []*Diff{&c}
	}
	return []*Diff{}
}
