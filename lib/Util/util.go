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

package Util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type HttpConfig struct {
	HttpType  string                 `json:"type"`
	DnsName   string                 `json:"dns"`
	Port      string                 `json:"port"`
	Id        string                 `json:"id"`
	HeaderCfg map[string]interface{} `json:"headers"`
}

func ParsePath(path string) (string, string) {
	return ParseCustomPath(path, "/")
}

func ParseCustomPath(path string, div string) (string, string) {
	queryPath := path
	for strings.HasPrefix(queryPath, div) {
		queryPath = strings.TrimPrefix(queryPath, div)
	}
	for strings.HasSuffix(queryPath, div) {
		queryPath = strings.TrimSuffix(queryPath, div)
	}
	devIdx := strings.Index(queryPath, div)
	if devIdx < 0 {
		return queryPath, ""
	}
	currentPath := queryPath[:devIdx]
	nextPath := queryPath[devIdx+len(div):]
	return currentPath, nextPath
}

func ParseArrayPath(path string) (string, string, error) {
	if path[len(path)-1:] != "]" {
		return path, "", nil
	}
	keyIdx := strings.Index(path, "[")
	if keyIdx < 1 {
		return path, "", nil
	}
	attrName := path[:keyIdx]
	key := path[keyIdx+1 : len(path)-1]
	if key == "" {
		return "", "", fmt.Errorf("invalid array path=[%s], key empty", path)
	}
	return attrName, key, nil
}

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

func IdxList(searchAry []interface{}) map[interface{}]int {
	hash := map[interface{}]int{}
	for idx, item := range searchAry {
		hash[item] = idx
	}
	return hash
}

func CountListIdx(itemList []interface{}) (map[interface{}]int, error) {
	cMap := map[interface{}]int{}
	if len(itemList) == 0 {
		return cMap, nil
	}
	itemType := reflect.TypeOf(itemList[0]).Kind()
	if itemType == reflect.Slice || itemType == reflect.Map {
		return nil, fmt.Errorf("cannot count list of type=[%s]", itemType)
	}
	for idx, item := range itemList {
		thisType := reflect.TypeOf(item).Kind()
		if thisType != itemType {
			return nil, fmt.Errorf("inconsist data type [%s]!=[%s] @%d", itemType, thisType, idx)
		}
		if _, found := cMap[item]; !found {
			cMap[item] = 1
		} else {
			cMap[item] += 1
		}
	}
	return cMap, nil
}

func DeDupeList(itemList []interface{}) ([]interface{}, error) {
	searchMap, err := CountListIdx(itemList)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, len(searchMap))
	for item := range searchMap {
		result = append(result, item)
	}
	return result, nil
}

func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func StructToMap(sData interface{}) (map[string]interface{}, error) {
	sDataBytes, err := json.Marshal(sData)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal data. Error:%s", err)
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(sDataBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal sData to map[string]interface{}, Error:%s", err)
	}
	return data, nil
}

func JsonCopy(data interface{}) (interface{}, error) {
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

func ObjCopy(src interface{}, targetAddr interface{}) error {
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

func PrefixStrLst(strList []string, prefix string) {
	for idx := range strList {
		strList[idx] = fmt.Sprintf("%s%s", prefix, strList[idx])
	}
}

func RootDir() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir, err := filepath.Abs(fmt.Sprintf("%s/../../", filepath.Dir(filename)))
	if err != nil {
		return "", err
	}
	return dir, nil
}
