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
	"reflect"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/SchemaPath"
)

func PrepareConn(schemaStr string, recordStr string) *SchemaPath.Connection {
	getSchema := func(dataType string) (*SchemaDoc.SchemaDoc, error) {
		schemaMap := map[string]interface{}{}
		err := json.Unmarshal([]byte(schemaStr), &schemaMap)
		if err != nil {
			return nil, fmt.Errorf("failed to ummarshal schema map str, Error:%s", err)
		}
		schemaData, ok := schemaMap[dataType].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema [type]=[%s] does not exists", dataType)
		}
		return SchemaDoc.New(schemaData, dataType, nil)
	}
	getRecord := func(dataType string, dataId string) (*Record.Record, error) {
		recordMap := map[string]interface{}{}
		err := json.Unmarshal([]byte(recordStr), &recordMap)
		if err != nil {
			return nil, fmt.Errorf("failed to ummarshal schema map str, Error:%s", err)
		}
		data, ok := recordMap[dataType].(map[string]interface{})[dataId].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("record [%s/%s] does not exists", dataType, dataId)
		}
		record, err := Record.LoadMap(data)
		if err != nil {
			return nil, fmt.Errorf("failed to load data as Record. Error:%s", err)
		}
		return record, nil
	}
	conn := SchemaPath.Connection{
		GetSchema: getSchema,
		GetRecord: getRecord,
	}
	return &conn

}

func TestConn(t *testing.T) {
	schemaStr := `
		{
			"testSch01": {
				"name": "testSch01",
				"description": "Test Schema 01",
				"properties": {
					"testAttr01": {
						"type": "string"
					}
				}
			}
		}
	`
	recordStr := `
		{
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
	conn := PrepareConn(schemaStr, recordStr)
	schema, err := conn.GetSchema("testSch01")
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

func TestWalkInObjectAndMap(t *testing.T) {
	schemaStr := `
	{
		"schema1": {
			"name": "schema1",
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
					"properties": {
						"value1": {
							"type": "string"
						},
						"value2": {
							"type": "string"
						}
					}
				}
			}

		}
	}`
	recordStr := `{
		"schema1": {
			"data1": {
				"__id": "data1",
				"__type": "schema1",
				"__ver": "0.0.1",
				"data": {
					"name": "data1",
					"value": {
						"value1": "01",
						"value2": "02"
					},
					"mapStr": {
						"keyExists": "exists"
					}
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schema1/data1/value/value1"
	schemaPath, err := SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	value1, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if value1.(string) != "01" {
		t.Errorf("invalid value from [path]=[%s], [%s]!=[01]", queryPath, value1.(string))
	}
	queryPath = "schema1/data1/mapStr/keyExists"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	valueExists, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if valueExists.(string) != "exists" {
		t.Errorf("invalid value from [path]=[%s], [%s]!=[exists]", queryPath, valueExists.(string))
	}
	queryPath = "schema1/data1/mapStr/keyNotExists"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	valueNotExists, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if valueNotExists != nil {
		t.Errorf("invalid value from [path]=[%s], [%s]!=[nil]", queryPath, valueNotExists.(string))
	}
}

func TestWalkInArray(t *testing.T) {
	schemaStr := `
		{
			"schemaWitArray": {
				"name": "schemaWitArray",
				"description": "schema of object with array of object in attribute",
				"properties": {
					"attrArray": {
						"type": "array",
						"items": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						}
					}
				},
				"definitions": {
					"itemObj": {
						"description": "item object of an array",
						"key": "{key1}_{key2}",
						"properties": {
							"key1": {
								"type": "string"
							},
							"key2": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	`
	recordStr := `
	{
		"schemaWitArray": {
			"testArray01": {
				"__id": "testArray01",
				"__type": "schemaWitArray",
				"__ver": "0.0.1",
				"data": {
					"attrArray": [
						{
							"key1": "01",
							"key2": "01"
						},
						{
							"key1": "01",
							"key2": "02"
						}
					]
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWitArray/testArray01/attrArray"
	schemaPath, err := SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	arrayValue, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if reflect.TypeOf(arrayValue).Kind() != reflect.Slice {
		t.Errorf("failed to get array from path=[%s]", queryPath)
	}
	queryPath = "sch emaWitArray/testArray01/attrArray[01_01]"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	itemValue, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if itemValue == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if itemValue.(map[string]interface{})["key2"] != "01" {
		t.Errorf("failed to get the correct value from [path]=[%s]", queryPath)
	}
	queryPath = "schemaWitArray/testArray01/attrArray[01_02]/key2"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	keyValue, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if keyValue == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if keyValue.(string) != "02" {
		t.Errorf("failed to get the correct value=[%s] from [path]=[%s]", keyValue.(string), queryPath)
	}
}

func TestWalkInRef(t *testing.T) {
	schemaStr := `
		{
			"schemaWithRef": {
				"name": "schemaWitArray",
				"description": "schema of object with array of object in attribute",
				"properties": {
					"itemArray": {
						"type": "array",
						"items": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						}
					}
				},
				"definitions": {
					"itemObj": {
						"description": "item object of an array",
						"key": "{key1}_{key2}",
						"properties": {
							"key1": {
								"type": "string"
							},
							"key2": {
								"type": "string"
							},
							"refIdx": {
								"type": "string",
								"contentMediaType": "inventory/schemaRef"
							}
						}
					}
				}
			},
			"schemaRef": {
				"name": "schemaRef",
				"description": "schema of ref object",
				"properties": {
					"data": {
						"type": "object",
						"$ref": "#/definitions/data"
					}
				},
				"definitions": {
					"data": {
						"description": "data wrapper for the keyed map",
						"properties": {
							"name": {
								"type": "string"
							},
							"items": {
								"type": "map",
								"items": {
									"type": "object",
									"$ref": "#/definitions/itemData"
								}
							}
						}
					},
					"itemData": {
						"description": "mapped item data schema",
						"properties": {
							"attr01": {
								"type": "string"
							},
							"attr02": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	`
	recordStr := `
	{
		"schemaWithRef": {
			"refData01": {
				"__id": "refData01",
				"__type": "schemaWithRef",
				"__ver": "0.0.1",
				"data": {
					"itemArray": [
						{
							"key1": "01",
							"key2": "01",
							"refIdx": "ref01/data/items/item01/attr01"
						},
						{
							"key1": "01",
							"key2": "02",
							"refIdx": "ref02/data/items/item02/attr02"
						}
					]
				}
			}
		},
		"schemaRef": {
			"ref01": {
				"__id": "ref01",
				"__type": "schemaRef",
				"__ver": "0.0.1",
				"data" : {
					"data": {
						"name": "ref01",
						"items": {
							"item01": {
								"attr01": "value01-01-01",
								"attr02": "value01-01-02"
							}
						}
					}
				}
			},
			"ref02": {
				"__id": "ref02",
				"__type": "schemaRef",
				"__ver": "0.0.1",
				"data" : {
					"data": {
						"name": "ref02",
						"items": {
							"item02": {
								"attr01": "value02-02-01",
								"attr02": "value02-02-02"
							}
						}
					}
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWithRef/refData01/itemArray[01_01]/refIdx"
	schemaPath, err := SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	refValue, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if refValue == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if refValue.(string) != "value01-01-01" {
		t.Errorf("expect [value01-01-01]!=[%s] from [path]=[%s]", refValue.(string), queryPath)
	}
	queryPath = "schemaWithRef/refData01/itemArray[01_01]/refIdx/$"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	refPath, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if refPath == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if refPath.(string) != "ref01/data/items/item01/attr01" {
		t.Errorf("failed to get the correct value. [%s]!=[ref01/data/items/item01/attr01]", refPath.(string))
	}
	queryPath = "schemaWithRef/refData01/itemArray[01_01]/refIdx?ref"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	refPath, err = schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if refPath == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if refPath.(string) != "ref01/data/items/item01/attr01" {
		t.Errorf("failed to get the correct value. [%s]!=[ref01/data/items/item01/attr01]", refPath.(string))
	}
}

func TestWalkSchema(t *testing.T) {
	schemaStr := `
	{
		"schemaWitArray": {
			"name": "schemaWitArray",
			"description": "schema of object with array of object in attribute",
			"properties": {
				"attrArray": {
					"type": "array",
					"items": {
						"type": "object",
						"$ref": "#/definitions/itemObj"
					}
				}
			},
			"definitions": {
				"itemObj": {
					"name": "itemObj",
					"description": "item object of an array",
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						}
					}
				}
			}
		}
	}
`
	recordStr := `
	{
		"schemaWitArray": {
			"testArray01": {
				"__id": "testArray01",
				"__type": "schemaWitArray",
				"__ver": "0.0.1",
				"data": {
					"attrArray": [
						{
							"key1": "01",
							"key2": "01"
						},
						{
							"key1": "01",
							"key2": "02"
						}
					]
				}
			},
			"testArray02": {
				"__id": "testArray01",
				"__type": "schemaWitArray",
				"__ver": "0.0.1",
				"data": {
					"attrArray": null
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWitArray/testArray01?schema"
	schemaPath, err := SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	schema, err := schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to walk path and return schema on record. Error: %s", err)
	}
	if schema.(map[string]interface{})[JsonKey.Name].(string) != "schemaWitArray" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArray?schema"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	schema, err = schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to walk path and return schema on record. Error: %s", err)
	}
	if schema.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArray[01_02]?schema"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	schema, err = schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to walk path and return schema on record. Error: %s", err)
	}
	if schema.(map[string]interface{})[JsonKey.Name].(string) != "itemObj" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray02/attrArray?schema"
	schemaPath, err = SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	schema, err = schemaPath.WalkValue()
	if err != nil {
		t.Errorf("failed to walk path and return schema on record. Error: %s", err)
	}
	if schema.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
}

func TestWalkFlat(t *testing.T) {
	schemaStr := `
	{
		"schemaWithRef": {
			"name": "schemaWitArray",
			"description": "schema of object with array of object in attribute",
			"properties": {
				"itemArray": {
					"type": "array",
					"items": {
						"type": "object",
						"$ref": "#/definitions/itemObj"
					}
				}
			},
			"definitions": {
				"itemObj": {
					"description": "item object of an array",
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"refIdx": {
							"type": "string",
							"contentMediaType": "inventory/schemaRef"
						}
					}
				}
			}
		},
		"schemaRef": {
			"name": "schemaRef",
			"description": "schema of ref object",
			"properties": {
				"data": {
					"type": "object",
					"$ref": "#/definitions/data"
				}
			},
			"definitions": {
				"data": {
					"description": "data wrapper for the keyed map",
					"properties": {
						"name": {
							"type": "string"
						},
						"items": {
							"type": "map",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemData"
							}
						}
					}
				},
				"itemData": {
					"description": "mapped item data schema",
					"properties": {
						"attr01": {
							"type": "string"
						},
						"attr02": {
							"type": "string"
						}
					}
				}
			}
		}
	}
`
	recordStr := `
	{
		"schemaWithRef": {
			"refData01": {
				"__id": "refData01",
				"__type": "schemaWithRef",
				"__ver": "0.0.1",
				"data": {
					"itemArray": [
						{
							"key1": "01",
							"key2": "01",
							"refIdx": "ref01/data/items/item01/attr01"
						},
						{
							"key1": "01",
							"key2": "02",
							"refIdx": "ref02/data/items/item02/attr02"
						}
					]
				}
			}
		},
		"schemaRef": {
			"ref01": {
				"__id": "ref01",
				"__type": "schemaRef",
				"__ver": "0.0.1",
				"data" : {
					"data": {
						"name": "ref01",
						"items": {
							"item01": {
								"attr01": "value01-01-01",
								"attr02": "value01-01-02"
							}
						}
					}
				}
			},
			"ref02": {
				"__id": "ref02",
				"__type": "schemaRef",
				"__ver": "0.0.1",
				"data" : {
					"data": {
						"name": "ref02",
						"items": {
							"item02": {
								"attr01": "value02-02-01",
								"attr02": "value02-02-02"
							}
						}
					}
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWithRef/refData01?flat"
	flatPath, err := SchemaPath.NewFromPath(conn, queryPath, nil)
	if err != nil {
		t.Errorf("failed to generate SchemaPath. from [path]=[%s], Error: %s", queryPath, err)
	}
	flatVal, err := flatPath.WalkValue()
	if err != nil {
		t.Errorf("failed to parse value from [path]=[%s], Error: %s", queryPath, err)
	}
	if flatVal == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	flatMap, ok := flatVal.(map[string]interface{})
	if !ok {
		t.Errorf("return value is not map. @[path]=[%s]", queryPath)
	}
	flatAry, ok := flatMap["itemArray"].([]interface{})
	if !ok {
		t.Errorf("return value has no key itemArray, or failed to convert it to array of interface{}. @[path]=[%s]", queryPath)
	}
	if flatAry[0].(string) != "01_01" {
		t.Errorf("failed to extract key of item 0 of itemArray, expect [01_01] != [%s], @[path]=[%s]", flatAry[0], queryPath)
	}
}
