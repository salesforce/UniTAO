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

package RefRecord

import (
	"encoding/json"
	"fmt"
	"net/http"

	"InventoryService/InvRecord"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

const (
	Referral  = "referral"
	LatestVer = "0.0.1"
)

type ReferralData struct {
	DataType  string                 `json:"DataType"`
	SchemaVer string                 `json:"SchemaVersion"`
	DsId      string                 `json:"DataServiceId"`
	AuthUrl   string                 `json:"AuthUrl"`
	AuthType  string                 `json:"AuthType"`
	Schema    map[string]interface{} `json:"Schema"`
	DsInfo    *InvRecord.DataServiceInfo
}

func LoadMap(data map[string]interface{}) (*ReferralData, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal data for ReferralRecord. Error:%s", err)
	}
	record := ReferralData{}
	err = json.Unmarshal(raw, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map data for ReferralRecord. Error:%s", err)
	}
	return &record, nil
}

func (r *ReferralData) GetSchema() *Http.HttpError {
	if r.DsInfo == nil {
		return Http.NewHttpError(fmt.Sprintf("failed to load DsInfo for type=[%s]", r.DataType), http.StatusInternalServerError)

	}
	dsUrl, err := r.DsInfo.GetUrl()
	if err != nil {
		return Http.NewHttpError(fmt.Sprintf("no good url to DS=[%s], error:%s", r.DsInfo.Id, err), http.StatusInternalServerError)
	}
	schemaUrl := fmt.Sprintf("%s/%s/%s", dsUrl, JsonKey.Schema, r.DataType)
	schemaData, code, err := Http.GetRestData(schemaUrl)
	if err != nil {
		return Http.NewHttpError(err.Error(), code)
	}
	schema, ok := schemaData.(map[string]interface{})
	if !ok {
		return Http.NewHttpError(fmt.Sprintf("failed to parse schema record. from path=[%s]", schemaUrl), http.StatusInternalServerError)
	}
	r.Schema = schema
	return nil
}

func (r *ReferralData) GetRecord() *Record.Record {
	rMap, _ := Util.StructToMap(r)
	return Record.NewRecord(Referral, LatestVer, r.DataType, rMap)
}
