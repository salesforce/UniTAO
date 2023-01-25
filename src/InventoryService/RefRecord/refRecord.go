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
	"log"
	"net/http"

	"InventoryService/InvRecord"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util/Http"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

const (
	Referral     = "referral"
	LatestVer    = "0.0.1"
	SchemaRecord = `{
		"__id": "referral",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "referral",
			"description": "referral record schema",
			"version": "0.0.1",
			"properties": {
				"DataType": {
					"type": "string"
				},
				"DataServiceId": {
					"type": "string"
				},
				"AuthUrl": {
					"type": "string"
				},
				"AuthType": {
					"type": "string"
				}
			}
		}
	}`
)

type ReferralData struct {
	DataType string `json:"DataType"`
	DsId     string `json:"DataServiceId"`
	AuthUrl  string `json:"AuthUrl"`
	AuthType string `json:"AuthType"`
	Schema   map[string]interface{}
	DsInfo   *InvRecord.DataServiceInfo
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

func (r *ReferralData) GetSchema(dataType string, logger *log.Logger) (*Record.Record, *Http.HttpError) {
	if logger == nil {
		logger = log.Default()
	}
	if r.DsInfo == nil {
		msg := fmt.Sprintf("failed to load DsInfo for type=[%s]", dataType)
		logger.Print(msg)
		return nil, Http.NewHttpError(msg, http.StatusInternalServerError)

	}
	dsUrl, err := r.DsInfo.GetUrl()
	if err != nil {
		msg := fmt.Sprintf("no good url to DS=[%s], error:%s", r.DsInfo.Id, err)
		logger.Print(msg)
		return nil, Http.NewHttpError(msg, http.StatusInternalServerError)
	}
	schemaUrl := fmt.Sprintf("%s/%s/%s", dsUrl, JsonKey.Schema, dataType)
	schemaData, code, err := Http.GetRestData(schemaUrl)
	if err != nil {
		logger.Print(err.Error())
		return nil, Http.NewHttpError(err.Error(), code)
	}
	schema, ok := schemaData.(map[string]interface{})
	if !ok {
		msg := fmt.Sprintf("failed to parse schema record. from path=[%s]", schemaUrl)
		logger.Print(msg)
		return nil, Http.NewHttpError(msg, http.StatusInternalServerError)
	}
	schemaRecord, err := Record.LoadMap(schema)
	if err != nil {
		msg := "schema from dataservice is not in Record format."
		logger.Printf("%s, Error: %s", msg, err)
		return nil, Http.WrapError(err, msg, http.StatusInternalServerError)
	}
	return schemaRecord, nil
}

func (r *ReferralData) GetRecord() *Record.Record {
	rMap, _ := Json.CopyToMap(r)
	return Record.NewRecord(Referral, LatestVer, r.DataType, rMap)
}
