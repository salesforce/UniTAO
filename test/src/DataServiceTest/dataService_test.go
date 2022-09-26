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

package DataServiceTest

import (
	"Data/DbConfig"
	"Data/DbIface"
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"DataService/Config"
	"DataService/DataHandler"
	"DataService/DataServer"

	"github.com/salesforce/UniTAO/lib/Schema/Record"
)

// Make sure both data service and inventory Service are running before running the test
func TestDataHander(t *testing.T) {
	configStr := `
	{
		"database": {
			"type": "dynamodb",
			"dynamodb": {
				"region": "us-west-2",
				"endpoint": "http://localhost:8000"
			}
		},
		"table": {
			"data": "DataService01"
		},
		"http": {
			"type": "http",
			"dns": "localhost",
			"port": "8002",
			"id": "DataService_01"
		},
		"inventory": {
			"url": "http://localhost:8004"
		}
	}
	`
	config := Config.Confuguration{}
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		t.Fatalf("faild to load config str. invalid format. Error:%s", err)
	}
	log.Print("config loaded")
	schemaStr := `
	{
		"region": {
			"__id": "region",
			"__type": "schema",
			"__ver": "0.0.1",
			"data": {
				"name": "region",
				"description": "geographical regions Schema",
				"properties": {
					"id": {
						"type": "string"
					},
					"description": {
						"type": "string"
					},
					"data_centers": {
						"type": "array",
						"items": {
							"type": "string",
							"contentMediaType": "inventory/data_center"
						}
					}
				}
			}
		},
		"data_center": {
			"__id": "data_center",
			"__type": "schema",
			"__ver": "0.0.1",
			"data": {
				"name": "data_center",
				"description": "geographical regions Schema",
				"properties": {
					"id": {
						"type": "string"
					},
					"description": {
						"type": "string"
					}
				}
			}
		},
		"SEA1": {
			"__id": "SEA1",
			"__type": "data_center",
			"__ver": "0.0.1",
			"data": {
				"name": "Seattle",
				"description": "Data Center in Seattle"
			}
		}
	}
	`
	schemaData := map[string]interface{}{}
	err = json.Unmarshal([]byte(schemaStr), &schemaData)
	if err != nil {
		t.Fatalf("Failed to load schema data for region. Error:%s", err)
	}
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb := MockDatabase{
			config: config,
		}
		mockDb.get = func(queryArgs map[string]interface{}) ([]map[string]interface{}, error) {
			result := []map[string]interface{}{}
			data, ok := schemaData[queryArgs[Record.DataId].(string)]
			if ok {
				result = append(result, data.(map[string]interface{}))
			}
			return result, nil
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler")
	}

	regionStr := `
	{
		"data": {
			"__id": "North_America",
			"__type": "region",
			"__ver": "1_01_01",
			"data": {
				"id": "North_America",
				"description": "North America Infrastructure",
				"data_centers": ["SEA1"]
			}
		},
		"negativeData": {
			"__id": "North_America",
			"__type": "region",
			"__ver": "1_01_01",
			"data": {
				"id": "North_America",
				"description": "North America Infrastructure",
				"data_centers": ["SEA1", "DFW4", "LAX2"]
			}
		}
	}
	`
	testData := map[string]interface{}{}
	err = json.Unmarshal([]byte(regionStr), &testData)
	if err != nil {
		t.Fatalf("failed loading test data Err:%s", err)
	}
	log.Print("get positive data for test")
	record, err := Record.LoadMap(testData["data"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to load data as record. Error:%s", err)
	}
	log.Printf("[type]=[%s],[id]=[%s]", record.Type, record.Id)
	e := handler.Validate(record)
	if e != nil {
		t.Fatalf("failed to validate positive data. Error:%s", err)
	}
	log.Printf("positive data validate passed")
	log.Print("get negative Data for test")
	record, err = Record.LoadMap(testData["negativeData"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to load negativeData as record. Error:%s", err)
	}
	log.Printf("[type]=[%s],[id]=[%s]", record.Type, record.Id)
	e = handler.Validate(record)
	if e == nil {
		t.Fatalf("failed to validate negative data.")
	}
	log.Printf("negative data validate passed")
}

func TestParseRecord(t *testing.T) {
	payload := make(map[string]interface{})
	dataType := "test"
	typeVer := "00_00_00"
	dataId := "test_01"
	record := Record.NewRecord(dataType, typeVer, dataId, payload)
	_, err := DataServer.ParseRecord([]string{}, record.Map(), "", "")
	if err != nil {
		t.Fatalf("failed to parse record. type=[%s], id=[%s]", dataType, dataId)
	}
	_, err = DataServer.ParseRecord([]string{}, record.Map(), dataType, dataId)
	if err.Status != http.StatusBadRequest {
		t.Fatalf("failed to validate post/put url path. type and id should be empty")
	}
	_, err = DataServer.ParseRecord([]string{}, record.Map(), dataType, "")
	if err.Status != http.StatusBadRequest {
		t.Fatalf("failed to parse record. type='%s', id=''", dataType)
	}
	_, err = DataServer.ParseRecord([]string{}, record.Map(), "", dataId)
	if err.Status != http.StatusBadRequest {
		t.Fatalf("failed to parse record. type='', id='%s'", dataId)
	}
	pRecord, err := DataServer.ParseRecord([]string{"true"}, record.Data, dataType, dataId)
	if err != nil {
		t.Fatalf("failed to parse record with no Reacod header. Error:%s", err)
	}
	if pRecord.Version != "0_00_00" {
		t.Fatalf("failed to create record with correct version")
	}
	_, err = DataServer.ParseRecord([]string{"true"}, record.Data, "", "")
	if err == nil {
		t.Fatalf("failed to catch missing type and id error. Error:%s", err)
	}
	if err.Status != http.StatusBadRequest {
		t.Fatalf("invalid status code for missing type and id error. expecte [%d]!=[%d]", http.StatusBadRequest, err.Status)
	}
}
