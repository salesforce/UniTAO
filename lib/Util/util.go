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
	"net/http"
	"os"
	"reflect"
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
	queryPath := path
	if len(queryPath) > 0 && queryPath[0:1] == "/" {
		queryPath = queryPath[1:]
	}
	//log.Printf("path[0:1]:%s, queryPath:%s", path[0:1], queryPath)
	devIdx := strings.Index(queryPath, "/")
	if devIdx < 0 {
		return queryPath, ""
	}
	currentPath := queryPath[0:devIdx]
	nextPath := queryPath[devIdx+1:]
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

func LoadJSONPayload(r *http.Request, payload map[string]interface{}) (int, error) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		newErr := fmt.Errorf("failed to read %s body. %s", r.Method, err)
		return http.StatusInternalServerError, newErr
	}
	err = json.Unmarshal([]byte(reqBody), &payload)
	if err != nil {
		newErr := fmt.Errorf("failed to load payload as json, %s", err)
		return http.StatusBadRequest, newErr
	}
	return 0, nil
}

func SearchStrList(searchAry []string, value string) bool {
	for _, item := range searchAry {
		if item == value {
			return true
		}
	}
	return false
}

func DeDupeList(itemList []interface{}) ([]interface{}, error) {
	if len(itemList) == 0 {
		return itemList, nil
	}
	itemType := reflect.TypeOf(itemList[0]).Kind()
	if itemType == reflect.Slice || itemType == reflect.Map {
		return nil, fmt.Errorf("cannot dedupe list of type=[%s]", itemType)
	}
	searchMap := map[interface{}]int{}
	result := make([]interface{}, 0, len(itemList))
	for idx, item := range itemList {
		thisType := reflect.TypeOf(item).Kind()
		if thisType != itemType {
			return nil, fmt.Errorf("inconsist data type [%s]!=[%s] @%d", itemType, thisType, idx)
		}
		if _, found := searchMap[item]; !found {
			searchMap[item] = 1
			result = append(result, item)
		}
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

func ObjCopy(src interface{}, target interface{}) error {
	dataBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("Util.JsonCopy failed to Marshal data, Error: %s", err)
	}
	err = json.Unmarshal(dataBytes, target)
	if err != nil {
		return fmt.Errorf("Util.JsonCopy failed to UnMarshal data, Error: %s", err)
	}
	return nil
}

func PrefixStrLst(strList []string, prefix string) []string {
	newList := make([]string, 0, len(strList))
	for _, line := range strList {
		newList = append(newList, fmt.Sprintf("%s%s", prefix, line))
	}
	return newList
}
