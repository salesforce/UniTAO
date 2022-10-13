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
func TestDataHandler(t *testing.T) {
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
		"schema": {
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
			}
		},
		"data_center": {
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
	}
	`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, schemaStr)
		if err != nil {
			return nil, err
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

func TestDataHandlerPatchAttr(t *testing.T) {
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
	dataStr := `{
		"schema": {
			"test": {
				"__id": "test",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"properties": {
						"attr1": {
							"type": "string"
						},
						"attr2": {
							"type": "integer"
						},
						"attr3": {
							"type": "object",
							"$ref": "#/definitions/subTest"
						}
					},
					"definitions": {
						"subTest": {
							"name": "subTest",
							"properties": {
								"subAttr1": {
									"type": "string"
								}
							}
						}
					}
				}
			}
		},
		"test": {
			"test01": {
				"__id": "test01",
				"__type": "test",
				"__ver": "0.0.1",
				"data": {
					"attr1": "test",
					"attr2": 1,
					"attr3": {
						"subAttr1": "test"
					}
				}				
			}
		}
	}`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dataStr)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler")
	}
	handler.Patch("test", "test01/attr1", "ok")
	data, e := handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].(string) != "ok" {
		t.Fatalf("patch failed")
	}
	handler.Patch("test", "test01/attr2", 2)
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if data[Record.Data].(map[string]interface{})["attr2"].(float64) != 2 {
		t.Fatalf("patch failed")
	}
	_, e = handler.Patch("test", "test01/attr2", "test")
	if e == nil {
		t.Fatalf("failed to catch the wrong format")
	}
	_, e = handler.Patch("test", "test01/attr3/subAttr1", "ok")
	if e != nil {
		t.Fatalf("failed to patch next level of attr. Err: %s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].(map[string]interface{})["subAttr1"].(string) != "ok" {
		t.Fatalf("patch failed")
	}
	pathData := map[string]interface{}{
		"subAttr1": "okAgain",
	}
	_, e = handler.Patch("test", "test01/attr3", pathData)
	if e != nil {
		t.Fatalf("failed to patch next level of attr. Err: %s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].(map[string]interface{})["subAttr1"].(string) != "okAgain" {
		t.Fatalf("patch failed")
	}
}

func TestDataHandlerPatchArrayObj(t *testing.T) {
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
	dataStr := `{
		"schema": {
			"test": {
				"__id": "test",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"properties": {
						"attr1": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/subTest"
							}
						},
						"attr2": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"attr3": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"attr4": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/testCmt"
							}
						}
					},
					"definitions": {
						"subTest": {
							"name": "subTest",
							"key": "{subKey}",
							"properties": {
								"subKey": {
									"type": "string"
								},
								"subAttr1": {
									"type": "string"
								}
							}
						}
					}
				}
			},
			"testCmt": {
				"__id": "testCmt",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "testCmt",
					"key": "{cmtKey}",
					"properties": {
						"cmtKey": {
							"type": "string"
						},
						"cmtValue": {
							"type": "string"
						}
					}
				}
			}
		},
		"test": {
			"test01": {
				"__id": "test01",
				"__type": "test",
				"__ver": "0.0.1",
				"data": {
					"attr1": [],
					"attr2": [],
					"attr3": [],
					"attr4": []
				}				
			}
		},
		"testCmt": {
			"testCmt01": {
				"__id": "testCmt01",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt01",
					"cmtValue": "testValue"
				}
			},
			"testCmt02": {
				"__id": "testCmt02",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt02",
					"cmtValue": "testValue"
				}
			}
		}
	}`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dataStr)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler, Error:%s", err)
	}
	subData := map[string]interface{}{
		"subKey":   "attr1K1",
		"subAttr1": "ok",
	}
	_, e := handler.Patch("test", "test01/attr1[attr1K1]", subData)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e := handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr1"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].([]interface{})[0].(map[string]interface{})["subAttr1"].(string) != "ok" {
		t.Fatalf("failed to patch new data into array")
	}
	subData["subAttr1"] = "ok1"
	_, e = handler.Patch("test", "test01/attr1[attr1K1]", subData)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr1"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].([]interface{})[0].(map[string]interface{})["subAttr1"].(string) != "ok1" {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr1[attr1K1]/subAttr1", "ok")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr1"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].([]interface{})[0].(map[string]interface{})["subAttr1"].(string) != "ok" {
		t.Fatalf("failed to patch new data into array")
	}
	subData["subKey"] = "attr1K2"
	subData["subAttr1"] = "ok2"
	_, e = handler.Patch("test", "test01/attr1[attr1K2]", subData)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr1"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].([]interface{})[1].(map[string]interface{})["subAttr1"].(string) != "ok2" {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr1[attr1K1]", nil)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr1"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr1"].([]interface{})[0].(map[string]interface{})["subAttr1"].(string) != "ok2" {
		t.Fatalf("failed to patch new data into array")
	}
}

func TestDataHandlerPatchArraySimpleStr(t *testing.T) {
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
	dataStr := `{
		"schema": {
			"test": {
				"__id": "test",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"properties": {
						"attr1": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/subTest"
							}
						},
						"attr2": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"attr3": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"attr4": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/testCmt"
							}
						}
					},
					"definitions": {
						"subTest": {
							"name": "subTest",
							"key": "{subKey}",
							"properties": {
								"subKey": {
									"type": "string"
								},
								"subAttr1": {
									"type": "string"
								}
							}
						}
					}
				}
			},
			"testCmt": {
				"__id": "testCmt",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "testCmt",
					"key": "{cmtKey}",
					"properties": {
						"cmtKey": {
							"type": "string"
						},
						"cmtValue": {
							"type": "string"
						}
					}
				}
			}
		},
		"test": {
			"test01": {
				"__id": "test01",
				"__type": "test",
				"__ver": "0.0.1",
				"data": {
					"attr1": [],
					"attr2": [],
					"attr3": [],
					"attr4": []
				}				
			}
		},
		"testCmt": {
			"testCmt01": {
				"__id": "testCmt01",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt01",
					"cmtValue": "testValue"
				}
			},
			"testCmt02": {
				"__id": "testCmt02",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt02",
					"cmtValue": "testValue"
				}
			}
		}
	}`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dataStr)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler, Error:%s", err)
	}
	_, e := handler.Patch("test", "test01/attr2[-1]", "test")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e := handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr2"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr2"].([]interface{})[0].(string) != "test" {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr2[0]", "ok")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr2"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr2"].([]interface{})[0].(string) != "ok" {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr2[-1]", "test")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr2"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr2"].([]interface{})[0].(string) != "test" {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr2[100]", "test01")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr2"].([]interface{})) != 3 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr2"].([]interface{})[2].(string) != "test01" {
		t.Fatalf("failed to patch new data into array")
	}
}

func TestDataHandlerPatchArrayInt(t *testing.T) {
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
	dataStr := `{
		"schema": {
			"test": {
				"__id": "test",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"properties": {
						"attr1": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/subTest"
							}
						},
						"attr2": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"attr3": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"attr4": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/testCmt"
							}
						}
					},
					"definitions": {
						"subTest": {
							"name": "subTest",
							"key": "{subKey}",
							"properties": {
								"subKey": {
									"type": "string"
								},
								"subAttr1": {
									"type": "string"
								}
							}
						}
					}
				}
			},
			"testCmt": {
				"__id": "testCmt",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "testCmt",
					"key": "{cmtKey}",
					"properties": {
						"cmtKey": {
							"type": "string"
						},
						"cmtValue": {
							"type": "string"
						}
					}
				}
			}
		},
		"test": {
			"test01": {
				"__id": "test01",
				"__type": "test",
				"__ver": "0.0.1",
				"data": {
					"attr1": [],
					"attr2": [],
					"attr3": [],
					"attr4": []
				}				
			}
		},
		"testCmt": {
			"testCmt01": {
				"__id": "testCmt01",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt01",
					"cmtValue": "testValue"
				}
			},
			"testCmt02": {
				"__id": "testCmt02",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt02",
					"cmtValue": "testValue"
				}
			}
		}
	}`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dataStr)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler, Error:%s", err)
	}
	_, e := handler.Patch("test", "test01/attr3[-1]", 0)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e := handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr3"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[0].(float64) != 0 {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr3[-1]", -1)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr3"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[0].(float64) != -1 {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr3[100]", 1)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr3"].([]interface{})) != 3 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[2].(float64) != 1 {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr3[1]", nil)
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr3"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[0].(float64) != -1 {
		t.Fatalf("failed to patch new data into array")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[1].(float64) != 1 {
		t.Fatalf("failed to patch new data into array")
	}
	_, e = handler.Patch("test", "test01/attr4[testCmt01]", "testCmt01")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr3"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[0].(float64) != -1 {
		t.Fatalf("failed to patch new data into array")
	}
	if data[Record.Data].(map[string]interface{})["attr3"].([]interface{})[1].(float64) != 1 {
		t.Fatalf("failed to patch new data into array")
	}
}

func TestDataHandlerPatchArrayCmt(t *testing.T) {
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
	dataStr := `{
		"schema": {
			"test": {
				"__id": "test",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"properties": {
						"attr1": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/subTest"
							}
						},
						"attr2": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"attr3": {
							"type": "array",
							"items": {
								"type": "integer"
							}
						},
						"attr4": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/testCmt"
							}
						}
					},
					"definitions": {
						"subTest": {
							"name": "subTest",
							"key": "{subKey}",
							"properties": {
								"subKey": {
									"type": "string"
								},
								"subAttr1": {
									"type": "string"
								}
							}
						}
					}
				}
			},
			"testCmt": {
				"__id": "testCmt",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "testCmt",
					"key": "{cmtKey}",
					"properties": {
						"cmtKey": {
							"type": "string"
						},
						"cmtValue": {
							"type": "string"
						}
					}
				}
			}
		},
		"test": {
			"test01": {
				"__id": "test01",
				"__type": "test",
				"__ver": "0.0.1",
				"data": {
					"attr1": [],
					"attr2": [],
					"attr3": [],
					"attr4": []
				}				
			}
		},
		"testCmt": {
			"testCmt01": {
				"__id": "testCmt01",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt01",
					"cmtValue": "testValue"
				}
			},
			"testCmt02": {
				"__id": "testCmt02",
				"__type": "testCmt",
				"__ver": "0.0.1",
				"data": {
					"cmtKey": "testCmt02",
					"cmtValue": "testValue"
				}
			}
		}
	}`
	connectDb := func(config DbConfig.DatabaseConfig) (DbIface.Database, error) {
		mockDb, err := NewMockDb(config, dataStr)
		if err != nil {
			return nil, err
		}
		return mockDb, nil
	}
	handler, err := DataHandler.New(config, connectDb)
	if err != nil {
		t.Fatalf("failed to create handler, Error:%s", err)
	}
	_, e := handler.Patch("test", "test01/attr4[testCmt01]", "testCmt01")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e := handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr4"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr4"].([]interface{})[0].(string) != "testCmt01" {
		t.Fatalf("patch failed")
	}
	_, e = handler.Patch("test", "test01/attr4[testCmt01]", "testCmt02")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr4"].([]interface{})) != 1 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr4"].([]interface{})[0].(string) != "testCmt02" {
		t.Fatalf("patch failed")
	}
	_, e = handler.Patch("test", "test01/attr4[testCmt01]", "testCmt03")
	if e == nil {
		t.Fatalf("failed to catch CmtRef Error")
	}
	_, e = handler.Patch("test", "test01/attr4[testCmt01]", "testCmt01")
	if e != nil {
		t.Fatalf("failed to patch data. Error:%s", e)
	}
	data, e = handler.Get("test", "test01")
	if e != nil {
		t.Fatalf("failed to get patched data")
	}
	if len(data[Record.Data].(map[string]interface{})["attr4"].([]interface{})) != 2 {
		t.Fatalf("patch failed")
	}
	if data[Record.Data].(map[string]interface{})["attr4"].([]interface{})[1].(string) != "testCmt01" {
		t.Fatalf("patch failed")
	}
	_, e = handler.Patch("test", "test01/attr4[testCmt01]", "testCmt02")
	if e == nil {
		t.Fatalf("failed to catch duplicate error")
	}
}
