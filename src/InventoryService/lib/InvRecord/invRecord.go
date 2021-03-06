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

package InvRecord

import (
	"encoding/json"
	"fmt"
)

type DataServiceInfo struct {
	Id           string   `json:"__id"`
	URL          string   `json:"url"`
	TypeList     []string `json:"typeList"`
	LastSyncTime string   `json:"lastSynctime"`
}

func NewDsInfo(id string, url string) *DataServiceInfo {
	dsInfo := DataServiceInfo{
		Id:       id,
		URL:      url,
		TypeList: []string{},
	}
	return &dsInfo
}

func CreateDsInfo(payload interface{}) (*DataServiceInfo, error) {
	ds_marshalled, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload to bytes as json, Err:%s", err)
	}
	dsInfo := DataServiceInfo{}
	err = json.Unmarshal(ds_marshalled, &dsInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to load json payload as DataServiceInfo, Err:%s", err)
	}
	return &dsInfo, nil
}

func (ds *DataServiceInfo) ToIface() (interface{}, error) {
	ds_marshalled, err := json.Marshal(ds)
	if err != nil {
		err = fmt.Errorf("failed to marshal DataServiceInfo to bytes, Error:%s", err)
		return nil, err
	}
	var payload interface{}
	err = json.Unmarshal(ds_marshalled, &payload)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal bytes to interface{}, Error:%s", err)
		return nil, err
	}
	return payload, nil
}
