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
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
)

type HttpConfig struct {
	HttpType string `json:"type"`
	DnsName  string `json:"dns"`
	Port     string `json:"port"`
	Id       string `json:"id"`
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

func ResponseJson(w http.ResponseWriter, data interface{}, status int) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func LoadJSONMap(filePath string) (map[string]interface{}, error) {
	var sample map[string]interface{}
	data, err := LoadJSON(filePath, reflect.TypeOf(sample))
	if err != nil {
		return nil, err
	}
	return data.(map[string]interface{}), nil
}

func LoadJSONList(filePath string) ([]interface{}, error) {
	var sample []interface{}
	data, err := LoadJSON(filePath, reflect.TypeOf(sample))
	if err != nil {
		return nil, err
	}
	return data.([]interface{}), nil
}

func LoadJSON(filePath string, expectedType reflect.Type) (interface{}, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		newErr := fmt.Errorf("failed to open JSON file: [%s]", filePath)
		return nil, newErr
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var data interface{}
	json.Unmarshal([]byte(byteValue), &data)
	dataType := reflect.TypeOf(data)
	if expectedType != dataType {
		return nil, fmt.Errorf("expect data type [%s]!=[%s] from file=[%s]", expectedType, dataType, filePath)
	}
	return data, nil
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

func GetRestData(url string) (interface{}, int, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("failed to get response, [url]=[%s], Err:%s", url, err)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to read data from response.Body. [url]=[%s], Err:%s", url, err)
		}
		var payload interface{}
		err = json.Unmarshal(responseData, &payload)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to parse response.[url]=[%s], Err:%s", url, err)
		}
		return payload, response.StatusCode, nil
	}
	return nil, response.StatusCode, fmt.Errorf("invalid response from url=[%s], Err:%s", url, string(responseData))
}

func SearchStrList(searchAry []string, value string) bool {
	for _, item := range searchAry {
		if item == value {
			return true
		}
	}
	return false
}

func URLPathJoin(sUrl string, sPath ...string) (*string, error) {
	u, err := url.Parse(sUrl)
	if err != nil {
		return nil, err
	}
	pathList := []string{u.Path}
	pathList = append(pathList, sPath...)
	u.Path = path.Join(pathList...)
	jUrl := u.String()
	return &jUrl, nil

}
