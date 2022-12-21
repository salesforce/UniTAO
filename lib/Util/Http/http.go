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

var UpdateMethods = map[string]bool{
	http.MethodPost:  true,
	http.MethodPatch: true,
	http.MethodPut:   true,
}

var SuccessCodes = map[int]bool{
	http.StatusAccepted: true,
	http.StatusOK:       true,
	http.StatusCreated:  true,
}

type Config struct {
	HttpType  string                 `json:"type"`
	DnsName   string                 `json:"dns"`
	Port      string                 `json:"port"`
	Id        string                 `json:"id"`
	HeaderCfg map[string]interface{} `json:"headers"`
}

func LoadRequest(r *http.Request) (interface{}, *HttpError) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, WrapError(err, "failed to read body from request", http.StatusBadRequest)
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(reqBody, &data)
	if err != nil {
		strData := string(reqBody)
		if strData == "" {
			return nil, nil
		}
		return strData, nil
	}
	return data, nil
}

func ResponseJson(w http.ResponseWriter, data interface{}, status int, httpCfg Config) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonStr := fmt.Sprintf("%s\n", string(jsonData))
	w.Header().Set("Content-Type", "application/json")
	Response(w, []byte(jsonStr), status, httpCfg)
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

func ResponseErr(w http.ResponseWriter, err error, code int, httpCfg Config) {
	if !IsHttpError(err) {
		err = NewHttpError(err.Error(), code)
	} else {
		code = err.(*HttpError).Status
	}
	ResponseJson(w, err, code, httpCfg)
}

func GetRestData(url string) (interface{}, int, error) {
	client := &http.Client{}
	response, err := client.Get(url)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to get response, [url]=[%s], Err:%s", url, err)
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

func SubmitPayload(dataUrl string, method string, headers map[string]interface{}, payload interface{}) (*http.Response, int, error) {
	if _, ok := UpdateMethods[method]; !ok {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid method=[%s], not a update method", method)
	}
	client := &http.Client{}
	var req *http.Request
	if payload == nil {
		if method != http.MethodGet && method != http.MethodDelete && method != http.MethodPatch {
			return nil, http.StatusBadRequest, fmt.Errorf("invalid method=[%s], only [%s, %s] allow payload=nil", method, http.MethodPatch, http.MethodDelete)
		}
		r, err := http.NewRequest(method, dataUrl, nil)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create request. Error:%s", err)
		}
		req = r
	} else {
		payloadType := reflect.TypeOf(payload).Kind()
		var json_data []byte
		if payloadType == reflect.Slice || payloadType == reflect.Map {
			data, err := json.MarshalIndent(payload, "", "    ")
			if err != nil {
				return nil, http.StatusBadRequest, fmt.Errorf("failed to marshal payload. Error: %s", err)
			}
			json_data = data
		} else {
			json_data = []byte(payload.(string))
		}
		r, err := http.NewRequest(method, dataUrl, bytes.NewBuffer(json_data))
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to create request. Error:%s", err)
		}
		req = r
	}
	defaultHeaders := map[string]interface{}{
		"Content-Type": "application/json",
	}
	for hKey, hValue := range headers {
		code, err := AddHeaders(req, hKey, hValue)
		if err != nil {
			return nil, code, err
		}
		delete(defaultHeaders, hKey)
	}
	for hKey, hValue := range defaultHeaders {
		code, err := AddHeaders(req, hKey, hValue)
		if err != nil {
			return nil, code, err
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, http.StatusInternalServerError, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, resp.StatusCode, fmt.Errorf("failed to [%s] data to [url]=[%s], Code: %d", method, dataUrl, resp.StatusCode)
	}
	return resp, resp.StatusCode, nil
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

func ParseHeaders(r *http.Request) map[string]interface{} {
	headers := map[string]interface{}{}
	for key, valList := range r.Header {
		if len(valList) == 0 {
			continue
		}
		if len(valList) > 1 {
			headers[key] = valList
		}
		headers[key] = valList[0]
	}
	return headers
}

func AddHeaders(r *http.Request, key string, value interface{}) (int, error) {
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		valList, ok := value.([]string)
		if !ok {
			return http.StatusBadRequest, fmt.Errorf("invalid header value. key=[%s], should be string or []string", key)
		}
		for _, valStr := range valList {
			r.Header.Add(key, valStr)
		}
	} else {
		hValStr, ok := value.(string)
		if !ok {
			return http.StatusBadRequest, fmt.Errorf("invalid header value. key=[%s], should be string or []string", key)
		}
		r.Header.Add(key, hValStr)
	}
	return http.StatusOK, nil
}
