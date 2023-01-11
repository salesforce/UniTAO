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

package Record

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	Data      = "data"
	DataId    = "__id"
	DataType  = "__type"
	KeyRecord = "record"
	NotRecord = "No-Record-Framework"
	Version   = "__ver"
	Schema    = `{
		"__id": "record",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "record",
			"version": "0.0.1",
			"description": "schema of data record",
			"properties": {
				"__id": {
					"type": "string"
				},
				"__type": {
					"type": "string"
				},
				"__ver": {
					"type": "string"         
				},
				"data": {
					"type": "object"
				}
			}
		}            
	}`
)

type Record struct {
	Id      string                 `json:"__id"`
	Type    string                 `json:"__type"`
	Version string                 `json:"__ver"`
	Data    map[string]interface{} `json:"data"`
}

func IsRecord(data map[string]interface{}) bool {
	record, _ := LoadStr(Schema)
	doc, _ := SchemaDoc.New(record.Data)
	schemaBytes, _ := json.MarshalIndent(doc.Data, "", "    ")
	meta, _ := jsonschema.CompileString("Record", string(schemaBytes))
	err := meta.Validate(data)
	return err == nil
}

func NewRecord(dataType string, typeVersion string, dataId string, data map[string]interface{}) *Record {
	record := Record{
		Id:      dataId,
		Type:    dataType,
		Version: typeVersion,
		Data:    data,
	}
	return &record
}

func ParseVersion(version string) ([]int, error) {
	verList := []int{}
	idx := 0
	if version == "" {
		return nil, fmt.Errorf("empty version")
	}
	for version != "" {
		idx += 1
		ver, nextVer := Util.ParseCustomPath(version, ".")
		verInt, err := strconv.Atoi(ver)
		if err != nil {
			return nil, fmt.Errorf("failed to convert part[%d]=[%s] to int", idx, ver)
		}
		if verInt < 0 {
			return nil, fmt.Errorf("invalid version, negative part[%d]=[%s]", idx, ver)
		}
		verList = append(verList, verInt)
		version = nextVer
	}
	if len(verList) < 3 {
		return nil, fmt.Errorf("invalid version format. at least 3 number. (xxx.xxx.xxx)")
	}
	return verList, nil
}

func CompareVersion(ver1List []int, ver2List []int) int {
	var shortList []int
	var longList []int
	v1Short := true
	if len(ver1List) <= len(ver2List) {
		shortList = ver1List
		longList = ver2List
	} else {
		v1Short = false
		shortList = ver2List
		longList = ver1List
	}
	gap := len(longList) - len(shortList)
	for idx := range longList {
		shortVal := 0
		if idx >= gap {
			shortVal = shortList[idx-gap]
		}
		if longList[idx] == shortVal {
			continue
		}
		if longList[idx] > shortVal {
			if v1Short {
				return -1
			}
			return 1
		}
		if v1Short {
			return 1
		}
		return -1
	}
	return 0
}

func LoadMap(data map[string]interface{}) (*Record, error) {
	if data == nil {
		return nil, nil
	}
	if !IsRecord(data) {
		return nil, fmt.Errorf("data is not a record")
	}
	recordBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal doc to string, Err:%s", err)
	}
	return LoadStr(string(recordBytes))
}

func LoadStr(dataStr string) (*Record, error) {
	record := Record{}
	err := json.Unmarshal([]byte(dataStr), &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data to Record. Error:%s", err)
	}
	return &record, nil
}

func (rec *Record) Raw() *string {
	rawbytes, _ := json.MarshalIndent(rec, "", "    ")
	rawStr := string(rawbytes)
	return &rawStr
}

func (rec *Record) RawData() *string {
	rawbytes, _ := json.MarshalIndent(rec.Data, "", "    ")
	rawStr := string(rawbytes)
	return &rawStr
}

func (rec *Record) Map() map[string]interface{} {
	data := make(map[string]interface{})
	json.Unmarshal([]byte(*rec.Raw()), &data)
	return data
}
