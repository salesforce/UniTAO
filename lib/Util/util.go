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
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

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
	keyIdx := strings.Index(path, "[")
	if keyIdx < 1 {
		return path, "", nil
	}
	if path[len(path)-1:] != "]" {
		return path, "", nil
	}
	attrName := path[:keyIdx]
	keyStr := path[keyIdx+1 : len(path)-1]
	if keyStr == "" {
		return "", "", fmt.Errorf("invalid array path=[%s], key empty", path)
	}
	key, err := url.QueryUnescape(keyStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to unescape key=[%s], Error:%s", keyStr, err)
	}
	return attrName, key, nil
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

func CheckInvalidKeys(invalidChars []string, data map[string]interface{}) []string {
	result := make([]string, 0, len(invalidChars))
	foundChars := map[string]bool{}
	for _, c := range invalidChars {
		if _, ok := foundChars[c]; ok {
			continue
		}
		for key := range data {
			if strings.Contains(key, c) {
				foundChars[c] = true
				result = append(result, c)
				break
			}
		}
	}
	return result
}
