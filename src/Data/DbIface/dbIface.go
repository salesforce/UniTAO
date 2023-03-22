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

package DbIface

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/salesforce/UniTAO/lib/Util"
)

const (
	DbType            = "type"
	DisplayAttributes = "displayAttributes"
	Payload           = "payload"
	Table             = "table"
	PatchPath         = "patchPath"
)

type Database interface {
	Name() string
	ListTable() ([]interface{}, error)
	CreateTable(name string, data map[string]interface{}) error
	DeleteTable(name string) error
	Get(queryArgs map[string]interface{}) ([]map[string]interface{}, error)
	Create(table string, data interface{}) error
	Update(table string, keys map[string]interface{}, data interface{}) (map[string]interface{}, error)
	Replace(table string, keys map[string]interface{}, data interface{}) error
	Delete(table string, keys map[string]interface{}) error
}

// walk into data with dataPath
// return last data layer that wrapping the attrbute
// attribute path:
//
//	attr name if it is a direct attribute
//	attr name with key, if it is a map
//	attr name with idx, if it is a array
func GetDataOnPath(data map[string]interface{}, dataPath string, prevPath string) (map[string]interface{}, string, error) {
	attrPath, nextPath := Util.ParsePath(dataPath)
	if nextPath == "" {
		return data, attrPath, nil
	}
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse attrPath=[%s] @path=[%s]", attrPath, prevPath)
	}
	attrData, ok := data[attrName]
	if !ok || attrData == nil {
		return nil, "", fmt.Errorf("missing attr=[%s] @path=[%s] to drill further", attrName, prevPath)
	}
	switch reflect.TypeOf(attrData).Kind() {
	case reflect.Slice:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, "", fmt.Errorf("invalid array path=%s[%s] @path=[%s], Error:%s", attrName, key, prevPath, err)
		}
		item, ok := attrData.([]interface{})[idx].(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("array item @[%s] is not map, @path=[%s]", attrPath, prevPath)
		}
		return GetDataOnPath(item, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
	case reflect.Map:
		if key == "" {
			return GetDataOnPath(attrData.(map[string]interface{}), nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
		}
		itemData, ok := attrData.(map[string]interface{})[key]
		if !ok {
			return nil, "", fmt.Errorf("invalid key=[%s] on map attr=[%s] @path=[%s], Error:%s", key, attrName, prevPath, err)
		}
		item, ok := itemData.(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("map item @[%s] is not map, @path=[%s]", attrPath, prevPath)
		}
		return GetDataOnPath(item, nextPath, fmt.Sprintf("%s/%s", prevPath, attrPath))
	default:
		if key != "" {
			return nil, "", fmt.Errorf("invalid key=[%s], data is not array or map @path=[%s/%s]", key, prevPath, attrName)
		}
		return nil, "", fmt.Errorf("type=[%s] cannot drill in further. @path=[%s/%s]", reflect.TypeOf(attrData).Kind(), prevPath, attrPath)
	}
}

// update data on attribute.
// this function is for database that does not support patch
func SetPatchData(data map[string]interface{}, attrPath string, newData interface{}) error {
	attrName, key, err := Util.ParseArrayPath(attrPath)
	if err != nil {
		return err
	}
	if key == "" {
		if newData == nil {
			delete(data, attrName)
			return nil
		}
		data[attrName] = newData
		return nil
	}
	switch reflect.TypeOf(data[attrName]).Kind() {
	case reflect.Slice:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("invalid array item reference=%s[%s], Error:%s", attrName, key, err)
		}
		attrData := data[attrName].([]interface{})
		if newData == nil {
			if idx < 0 || idx >= len(attrData) {
				// do nothing if idx not in range
				return nil
			}
			newList := make([]interface{}, 0, len(attrData)-1)
			for newIdx, _ := range attrData {
				if newIdx != idx {
					newList = append(newList, attrData[newIdx])
				}
			}
			data[attrName] = newList
			return nil
		}
		if idx < 0 {
			data[attrName] = append([]interface{}{newData}, attrData...)
			return nil
		}
		if idx >= len(attrData) {
			data[attrName] = append(attrData, newData)
			return nil
		}
		attrData[idx] = newData
	case reflect.Map:
		attrData := data[attrName].(map[string]interface{})
		if newData == nil {
			delete(attrData, key)
			return nil
		}
		attrData[key] = newData
	default:
		return fmt.Errorf("attr=[%s] and type=[%s] is not fit to be set with Key/idx=[%s]", attrName, reflect.TypeOf(data[attrName]).Kind(), key)
	}
	return nil
}
