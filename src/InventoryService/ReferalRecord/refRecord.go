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

package ReferalRecord

import (
	"encoding/json"
	"fmt"
	"net/http"

	"InventoryService/InvRecord"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Util"
)

type ReferalRecord struct {
	Id        string                 `json:"__id"`
	DataType  string                 `json:"__type"`
	SchemaVer string                 `json:"__ver"`
	DsId      string                 `json:"DataServiceId"`
	DsUrl     string                 `json:"DataServcieUrl"`
	AuthUrl   string                 `json:"AuthUrl"`
	AuthType  string                 `json:"AuthType"`
	Schema    map[string]interface{} `json:"Schema"`
	DsInfo    *InvRecord.DataServiceInfo
}

func LoadMap(data map[string]interface{}) (*ReferalRecord, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal data for ReferalRecord. Error:%s", err)
	}
	record := ReferalRecord{}
	err = json.Unmarshal(raw, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map data for ReferalRecord. Error:%s", err)
	}
	return &record, nil
}

func (r *ReferalRecord) GetSchema() (int, error) {
	schemaUrl := fmt.Sprintf("%s/%s/%s", r.DsUrl, JsonKey.Schema, r.DataType)
	schemaData, code, err := Util.GetRestData(schemaUrl)
	if err != nil {
		return code, err
	}
	schema, ok := schemaData.(map[string]interface{})
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("failed to parse schema record. from path=[%s]", schemaUrl)
	}
	r.Schema = schema
	return http.StatusOK, nil
}

func (r *ReferalRecord) SetDsInfo(ds *InvRecord.DataServiceInfo) error {
	r.DsInfo = ds
	dsUrl, err := ds.GetUrl()
	if err != nil {
		return err
	}
	r.DsUrl = dsUrl
	return nil
}

func (r *ReferalRecord) ToMap() (map[string]interface{}, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
