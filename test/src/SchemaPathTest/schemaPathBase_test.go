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

package SchemaPathTest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath"
	SchemaPathData "github.com/salesforce/UniTAO/lib/SchemaPath/Data"
	"github.com/salesforce/UniTAO/lib/Util"
	"github.com/salesforce/UniTAO/lib/Util/Http"
)

func TestParseArrayPath(t *testing.T) {
	arrayPath := "abc[1]"
	attrName, attrIdx, err := Util.ParseArrayPath(arrayPath)
	if err != nil {
		t.Errorf("failed to parse array attr: %s, Error:%s", arrayPath, err)
	}
	if attrName != "abc" {
		t.Errorf("parse path failed, expect [abc]!=[%s]", attrName)
	}
	if attrIdx != "1" {
		t.Errorf("parse path failed, expect [1]!=[%s]", attrIdx)
	}
	arrayPath = "abc[]"
	_, _, err = Util.ParseArrayPath(arrayPath)
	if err == nil {
		t.Errorf("failed to caught array path error=[empty idx]. %s", arrayPath)
	}
}

func PrepareConn(recordStr string) *SchemaPathData.Connection {
	getRecord := func(dataType string, dataId string) (*Record.Record, *Http.HttpError) {
		recordMap := map[string]interface{}{}
		err := json.Unmarshal([]byte(recordStr), &recordMap)
		if err != nil {
			return nil, Http.WrapError(err, "failed to ummarshal schema map str", http.StatusInternalServerError)
		}
		var data map[string]interface{}
		if dataType == JsonKey.Schema {
			schemaId, schemaVer, _ := SchemaDoc.ParseDataType(dataId)
			schemaMap := recordMap[dataType].(map[string]interface{})
			schemaData, ok := schemaMap[schemaId].(map[string]interface{})
			if !ok {
				return nil, Http.NewHttpError(fmt.Sprintf("schema [%s/%s] does not exists", dataType, schemaId), http.StatusNotFound)
			}
			data = schemaData
			if schemaVer != "" && data[Record.Version] != schemaVer {
				archiveId := SchemaDoc.ArchivedSchemaId(schemaId, schemaVer)
				schemaData, ok := schemaMap[archiveId].(map[string]interface{})
				if !ok {
					return nil, Http.NewHttpError(fmt.Sprintf("schema [%s/%s] does not exists", dataType, archiveId), http.StatusNotFound)
				}
				data = schemaData
			}
		} else {
			recordData, ok := recordMap[dataType].(map[string]interface{})[dataId].(map[string]interface{})
			if !ok {
				return nil, Http.NewHttpError(fmt.Sprintf("record [%s/%s] does not exists", dataType, dataId), http.StatusNotFound)
			}
			data = recordData
		}
		record, err := Record.LoadMap(data)
		if err != nil {
			return nil, Http.WrapError(err, "failed to load data as Record.", http.StatusInternalServerError)
		}
		return record, nil
	}
	conn := SchemaPathData.Connection{
		FuncRecord: getRecord,
	}
	return &conn
}

func TestConn(t *testing.T) {
	recordStr := `
		{
			"schema": {
				"testSch01": {
					"__id": "testSch01",
					"__type": "schema",
					"__ver": "0.0.1",
					"data": {
						"name": "testSch01",
						"description": "Test Schema 01",
						"properties": {
							"testAttr01": {
								"type": "string"
							}
						}
					}
				}
			},
			"testSch01": {
				"testId01": {
					"__id": "testId01",
					"__type": "testSch01",
					"__ver": "0.0.1",
					"data": {
						"testAttr01": "testValue01"
					}
				}
			}
		}
	`
	conn := PrepareConn(recordStr)
	schema, err := conn.GetRecord(JsonKey.Schema, "testSch01")
	if err != nil {
		t.Errorf("failed while get schema=[testSch01], Error:%s", err)
	}
	if schema.Id != "testSch01" {
		t.Errorf("failed to get schema=[testSch01], got [" + schema.Id + "] instead")
	}
	record, err := conn.GetRecord("testSch01", "testId01")
	if err != nil {
		t.Errorf("failed to get record [type/id]=[testSch01/testId01], Error: %s", err)
	}
	if record.Id != "testId01" {
		t.Errorf("got wrong record.id [%s]!=[testId01]", record.Id)
	}
}

func TestPathNode(t *testing.T) {
	recordStr := `{
		"schema": {
			"schema1": {
				"__id": "schema1",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "schema1",
					"version": "0.0.1",
					"description": "test schema 01",
					"properties": {
						"name": {
							"type": "string"
						},
						"value": {
							"type": "object",
							"$ref": "#/definitions/testValue"
						},
						"mapStr": {
							"type": "map",
							"items": {
								"type": "string"
							}
						}
					},
					"definitions": {
						"testValue": {
							"name": "testValue",
							"properties": {
								"value1": {
									"type": "string"
								},
								"value2": {
									"type": "string",
									"contentMediaType": "inventory/schema2"
								}
							}
						}
					}
				}
			},
			"schema2": {
				"__id": "schema2",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "schema2",
					"version": "0.0.1",
					"description": "cross recod type schema 02",
					"properties": {
						"test": {
							"type": "string"
						}
					}
				}
			}
		},
		"schema1": {
			"data1": {
				"__id": "data1",
				"__type": "schema1",
				"__ver": "0.0.1",
				"data": {
					"name": "data1",
					"value": {
						"value1": "01",
						"value2": "data02"
					},
					"mapStr": {
						"keyExists": "exists"
					}
				}
			}
		},
		"schema2": {
			"data02": {
				"__id": "data02",
				"__type": "schema2",
				"__ver": "0.0.1",
				"data": {
					"test": "testStr"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	queryPath := "/value/value2"
	_, err := SchemaPath.BuildNodePath(conn, "schema1", "data1", queryPath)
	if err != nil {
		t.Fatalf("PathNode-positive: failed to build node path chain. Error: %s", err)
	}
	queryPath = "/value/value2/extra"
	_, err = SchemaPath.BuildNodePath(conn, "schema1", "data1", queryPath)
	if err == nil {
		t.Fatalf("failed to caught error for walk on undefined path")
	}
}

func QueryPath(conn *SchemaPathData.Connection, path string) (interface{}, *Http.HttpError) {
	dataType, nextPath := Util.ParsePath(path)
	query, err := SchemaPath.CreateQuery(conn, dataType, nextPath)
	if err != nil {
		return nil, err
	}
	value, err := query.WalkValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}
