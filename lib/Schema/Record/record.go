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
)

const (
	DataId    = "__id"
	DataType  = "__type"
	KeyRecord = "record"
	NotRecord = "No-Record-Framework"
	Version   = "__ver"
)

type Record struct {
	Id      string                 `json:"__id"`
	Type    string                 `json:"__type"`
	Version string                 `json:"__ver"`
	Data    map[string]interface{} `json:"data"`
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

func LoadMap(data map[string]interface{}) (*Record, error) {
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
