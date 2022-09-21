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

package Http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"time"
)

type Config struct {
	HttpType  string                 `json:"type"`
	DnsName   string                 `json:"dns"`
	Port      string                 `json:"port"`
	Id        string                 `json:"id"`
	HeaderCfg map[string]interface{} `json:"headers"`
}

func ResponseJson(w http.ResponseWriter, data interface{}, status int, httpCfg Config) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData = append(jsonData, "\n"...)
	w.Header().Set("Content-Type", "application/json")
	Response(w, jsonData, status, httpCfg)
}

func ResponseText(w http.ResponseWriter, txt []byte, status int, httpCfg Config) {
	w.Header().Set("Content-Type", "text/plain")
	Response(w, txt, status, httpCfg)
}

func Response(w http.ResponseWriter, txt []byte, status int, httpCfg Config) {
	for key, value := range httpCfg.HeaderCfg {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			for _, item := range value.([]interface{}) {
				w.Header().Set(key, item.(string))
			}
		default:
			w.Header().Set(key, value.(string))
		}

	}
	w.WriteHeader(status)
	w.Write(txt)
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

func PostRestData(dataUrl string, payload interface{}) (int, error) {
	json_data, err := json.Marshal(payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to marshal payload. Error: %s", err)
	}
	resp, err := http.Post(dataUrl, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to post data to [url]=[%s], Error:%s", dataUrl, err)
	}
	return http.StatusAccepted, nil
}

func SiteReachable(url string) bool {
	timeout := 1 * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Get(url)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	return true
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
