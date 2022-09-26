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

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	LatestVer = "0.0.1"
)

type DataServiceInfo struct {
	Id           string   `json:"dsId"`
	URL          []string `json:"url"`
	LastSyncTime string   `json:"lastSynctime"`
	goodUrl      string
}

func NewDsInfo(id string, url string) *Record.Record {
	dsInfo := DataServiceInfo{
		Id:  id,
		URL: []string{url},
	}
	dsMap, _ := Util.StructToMap(dsInfo)
	record := Record.NewRecord(Schema.Inventory, LatestVer, id, dsMap)
	return record
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

func (ds *DataServiceInfo) GetUrl() (string, error) {
	if ds.goodUrl == "" || !Http.SiteReachable(ds.goodUrl) {
		ds.goodUrl = ""
		for _, url := range ds.URL {
			if Http.SiteReachable(url) {
				ds.goodUrl = url
			}
		}
		if ds.goodUrl == "" {
			return "", fmt.Errorf("no good url is reachable for DS=[%s]", ds.Id)
		}
	}

	return ds.goodUrl, nil
}
