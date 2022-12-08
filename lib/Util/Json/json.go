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

package Json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

func LoadJsonFile(filePath string) (interface{}, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		newErr := fmt.Errorf("failed to open JSON file: [%s]", filePath)
		return nil, newErr
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var data interface{}
	json.Unmarshal([]byte(byteValue), &data)
	return data, nil
}

func LoadJSONMap(filePath string) (map[string]interface{}, error) {
	data, err := LoadJsonFile(filePath)
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(data).Kind() != reflect.Map {
		return nil, fmt.Errorf("data is not map, [path]=[%s], Error:%s", filePath, err)
	}
	return data.(map[string]interface{}), nil
}

func LoadJSONList(filePath string) ([]interface{}, error) {
	data, err := LoadJsonFile(filePath)
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return nil, fmt.Errorf("data is not Slice, [path]=[%s], Error:%s", filePath, err)
	}
	return data.([]interface{}), nil
}

func StructToMap(sData interface{}) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := CopyTo(sData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Copy(data interface{}) (interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Util.JsonCopy failed to Marshal data, Error: %s", err)
	}
	var result interface{}
	err = json.Unmarshal(dataBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("Util.JsonCopy failed to UnMarshal data, Error: %s", err)
	}
	return result, nil
}

func CopyTo(src interface{}, targetAddr interface{}) error {
	dataBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("Util.JsonCopy failed to Marshal data, Error: %s", err)
	}
	err = json.Unmarshal(dataBytes, targetAddr)
	if err != nil {
		return fmt.Errorf("Util.JsonCopy failed to UnMarshal data, Error: %s", err)
	}
	return nil
}
