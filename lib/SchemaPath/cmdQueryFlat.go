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

package SchemaPath

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
	"github.com/salesforce/UniTAO/lib/Util"
)

type CmdQueryFlat struct {
	p *Node.PathNode
}

func (c *CmdQueryFlat) Name() string {
	return PathCmd.CmdFlat
}

func (c *CmdQueryFlat) WalkValue() (interface{}, *Error.SchemaPathErr) {
	/*
		TODO:
		1, get schema of current path.
		2, get value of current path
		3, base on current schema to define how to get flat value of current
	*/
	cmdValue := CmdQueryValue{
		p: c.p,
	}
	nodeValue, err := cmdValue.WalkValue()
	if err != nil {
		return nil, err
	}
	validateErr := ValidateFlatValue(nodeValue)
	if validateErr == nil {
		// the return value is already a flat value
		if reflect.TypeOf(nodeValue).Kind() == reflect.Slice {
			dedupeList, nErr := Util.DeDupeList(nodeValue.([]interface{}))
			if nErr != nil {
				return nil, &Error.SchemaPathErr{
					Code:    http.StatusInternalServerError,
					PathErr: fmt.Errorf("failed to dedupe result. Error:%s", nErr),
				}
			}
			return dedupeList, nil
		}
		return nodeValue, nil
	}
	if reflect.TypeOf(nodeValue).Kind() == reflect.Slice {
		valueList, err := FlatNodeArray(c.p, nodeValue.([]interface{}))
		if err != nil {
			return nil, err
		}
		dedupeList, nErr := Util.DeDupeList(valueList.([]interface{}))
		if nErr != nil {
			return nil, &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: fmt.Errorf("failed to dedupe result. Error:%s", nErr),
			}
		}
		return dedupeList, nil
	}
	return FlatNodeMap(c.p, nodeValue.(map[string]interface{}))
}

func FlatMergeEmbedArray(arrayValue []interface{}) ([]interface{}, error) {
	if reflect.TypeOf(arrayValue[0]).Kind() != reflect.Slice {
		return arrayValue, nil
	}
	resultAry := []interface{}{}
	for idx, item := range arrayValue {
		itemAry, ok := item.([]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to convert item @[%d] to []interface{}", idx)
		}
		embedAry, err := FlatMergeEmbedArray(itemAry)
		if err != nil {
			return nil, fmt.Errorf("failed to Merge Embeded Array @[%d], Error:%s", idx, err)
		}
		resultAry = append(resultAry, embedAry...)
	}
	return resultAry, nil
}

func FlatNodeArray(node *Node.PathNode, nodeValue []interface{}) (interface{}, *Error.SchemaPathErr) {
	valueAry, err := FlatMergeEmbedArray(nodeValue)
	if err != nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: err,
		}
	}
	err = ValidateFlatArray(valueAry)
	if err == nil {
		return valueAry, nil
	}
	// array is not simple, item is map
	for node.Next != nil {
		node = node.Next
	}
	if node.AttrName == "" {
		resultAry, err := FlatSchemaArray(node.Schema, valueAry)
		if err != nil {
			return nil, &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: err,
			}
		}
		return resultAry, nil
	}
	itemSchema, ok := node.Schema.SubDocs[node.AttrName]
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("missing subDoc for attr=[%s] @path=[%s]", node.AttrName, node.FullPath()),
		}
	}
	resultAry, err := FlatSchemaArray(itemSchema, valueAry)
	if err != nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: err,
		}
	}
	return resultAry, nil
}

func FlatNodeMap(node *Node.PathNode, nodeValue map[string]interface{}) (interface{}, *Error.SchemaPathErr) {
	for node.Next != nil {
		node = node.Next
	}
	result := map[string]interface{}{}
	for attr, attrValue := range nodeValue {
		attrDef, ok := node.Schema.Data[JsonKey.Properties].(map[string]interface{})[attr]
		if !ok {
			simpleValue, err := FlatSimpleValue(attrValue)
			if err != nil {
				return nil, &Error.SchemaPathErr{
					Code:    http.StatusBadRequest,
					PathErr: fmt.Errorf("attr=[%s] not defined in schema, cannot convert its value to simple value. Error:%s", attr, err),
				}
			}
			result[attr] = simpleValue
			continue
		}
		switch attrDef.(map[string]interface{})[JsonKey.Type].(string) {
		case JsonKey.Array:
			itemDef := attrDef.(map[string]interface{})[JsonKey.Items].(map[string]interface{})
			if itemDef[JsonKey.Type].(string) == JsonKey.Object {
				itemSchema, ok := node.Schema.SubDocs[attr]
				if !ok {
					return nil, &Error.SchemaPathErr{
						Code:    http.StatusInternalServerError,
						PathErr: fmt.Errorf("missing subDoc for attr=[%s] @path=[%s]", attr, node.FullPath()),
					}
				}
				attrAryValue, err := FlatSchemaArray(itemSchema, attrValue.([]interface{}))
				if err != nil {
					return nil, &Error.SchemaPathErr{
						Code:    http.StatusInternalServerError,
						PathErr: fmt.Errorf("failed to flat array with schema, attr=[%s], path=[%s], Error:%s", attr, node.FullPath(), err),
					}
				}
				result[attr] = attrAryValue
			} else {
				result[attr] = attrValue.([]interface{})
			}
		case JsonKey.Object:
			attrMapValue := attrValue.(map[string]interface{})
			attrAryValue := make([]interface{}, 0, len(attrMapValue))
			for key := range attrMapValue {
				attrAryValue = append(attrAryValue, key)
			}
			result[attr] = attrAryValue
		default:
			result[attr] = attrValue
		}
	}
	return result, nil
}

func FlatSchemaArray(schema *SchemaDoc.SchemaDoc, valueAry []interface{}) ([]interface{}, error) {
	resultAry := make([]interface{}, 0, len(valueAry))
	for _, item := range valueAry {
		// the leaf is an object.
		itemMap := item.(map[string]interface{})
		itemKey, err := schema.BuildKey(itemMap)
		if err != nil {
			return nil, err
		}
		resultAry = append(resultAry, itemKey)
	}
	return resultAry, nil
}

func FlatSimpleValue(value interface{}) (interface{}, error) {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		err := ValidateFlatArray(value.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("is not an array of simple value")
		}
		return value, nil
	case reflect.Map:
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("is not an map[string]interface{}")
		}
		result := make([]interface{}, 0, len(valueMap))
		for key, _ := range valueMap {
			result = append(result, key)
		}
		return result, nil
	default:
		return value, nil
	}
}

func ValidateFlatValue(value interface{}) error {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		return ValidateFlatArray(value.([]interface{}))
	case reflect.Map:
		return ValidateFlatMap(value.(map[string]interface{}))
	default:
		return nil
	}
}

func ValidateFlatArray(value []interface{}) error {
	for idx, item := range value {
		valueType := reflect.TypeOf(item).Kind()
		if valueType == reflect.Slice || valueType == reflect.Map {
			return fmt.Errorf("invalid flat array item @idx=[%d]. item type=[%s] should only ne simple type", idx, valueType)
		}
	}
	return nil
}

func ValidateFlatMap(value map[string]interface{}) error {
	for key, item := range value {
		switch reflect.TypeOf(item).Kind() {
		case reflect.Map:
			return fmt.Errorf("invalid get flat value on attr=[%s], type=[%s], expected=[%s]", key, reflect.Map, reflect.Slice)
		case reflect.Slice:
			ifaceAry := item.([]interface{})
			err := ValidateFlatArray(ifaceAry)
			if err != nil {
				return fmt.Errorf("attr=[%s] is not a simple array. Error:%s", key, err)
			}
		}
	}
	return nil
}
