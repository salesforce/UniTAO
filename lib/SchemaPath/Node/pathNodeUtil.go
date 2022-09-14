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

package Node

import (
	"fmt"
	"net/http"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Error"
	"github.com/salesforce/UniTAO/lib/SchemaPath/PathCmd"
)

func GetNodeValue(node *PathNode, dataFromParent interface{}) (interface{}, *Error.SchemaPathErr) {
	var data map[string]interface{}
	if node.DataType != "" {
		if node.Prev == nil {
			// root record, get record data directly.
			recordData, err := node.GetRecordData()
			if err != nil {
				return nil, err
			}
			data = recordData.(map[string]interface{})
		} else {
			// CMT node. get dataId from parent
			dataId, ok := dataFromParent.(string)
			if !ok {
				return nil, &Error.SchemaPathErr{
					Code:    http.StatusInternalServerError,
					PathErr: fmt.Errorf("invalid CMT Ref value. it is not a string"),
				}
			}
			node.DataId = dataId
			recordData, err := node.GetRecordData()
			if err != nil {
				return nil, err
			}
			data = recordData.(map[string]interface{})
		}
	} else {
		parentObj, ok := dataFromParent.(map[string]interface{})
		if !ok {
			return nil, &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: fmt.Errorf("node data should always be object. @path=[%s]", node.FullPath()),
			}
		}
		data = parentObj
	}
	if node.AttrName == "" {
		//no further path
		return data, nil
	}
	attrValue, ok := data[node.AttrName]
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusNotFound,
			PathErr: fmt.Errorf("attr=[%s] does not exists @path=[%s]", node.AttrName, node.FullPath()),
		}
	}
	if attrValue == nil {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusNotFound,
			PathErr: fmt.Errorf("attr=[%s] is nil @path=[%s]", node.AttrName, node.FullPath()),
		}
	}
	if node.Idx == "" || node.Idx == PathCmd.ALL {
		return attrValue, nil
	}
	if node.AttrDef[JsonKey.Type] == JsonKey.Array {
		valueList, ok := attrValue.([]interface{})
		if !ok {
			return nil, &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: fmt.Errorf("invalid data, attr=[%s] is not a list. path=[%s]", node.AttrName, node.FullPath()),
			}
		}
		return GetNodeListIdx(node, valueList, node.Idx)
	}
	valueMap, ok := attrValue.(map[string]interface{})
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusInternalServerError,
			PathErr: fmt.Errorf("invalid data, attr=[%s] is not a map. path=[%s]", node.AttrName, node.FullPath()),
		}
	}
	return GetNodeMapKey(node, valueMap, node.Idx)
}

func GetNodeListIdx(node *PathNode, value []interface{}, key string) (interface{}, *Error.SchemaPathErr) {
	itemType := node.AttrDef[JsonKey.Items].(map[string]interface{})[JsonKey.Type].(string)
	for idx, item := range value {
		if itemType == JsonKey.String {
			if item == key {
				return item, nil
			}
		}
		itemKey, err := node.Next.Schema.BuildKey(item.(map[string]interface{}))
		if err != nil {
			return nil, &Error.SchemaPathErr{
				Code:    http.StatusInternalServerError,
				PathErr: fmt.Errorf("data error, failed to get key item #%d, @path=[%s] Error:%s", idx, err, node.FullPath()),
			}
		}
		if itemKey == key {
			return item, nil
		}
	}
	return nil, &Error.SchemaPathErr{
		Code:    http.StatusNotFound,
		PathErr: fmt.Errorf("key %s[%s] does not exists at path=[%s]", node.AttrName, key, node.FullPath()),
	}
}

func GetNodeMapKey(node *PathNode, valueMap map[string]interface{}, key string) (interface{}, *Error.SchemaPathErr) {
	idxValue, ok := valueMap[key]
	if !ok {
		return nil, &Error.SchemaPathErr{
			Code:    http.StatusNotFound,
			PathErr: fmt.Errorf("key %s[%s] does not exists at path=[%s]", node.AttrName, node.Idx, node.FullPath()),
		}
	}
	return idxValue, nil
}
