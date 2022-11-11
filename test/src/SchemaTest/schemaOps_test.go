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

package SchemaTest

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

func TestSchemaValidate(t *testing.T) {
	log.Print("test Run")
	infraStr := `{
		"schema": {
			"__id": "infrastructure",
			"__type": "schema",
			"__ver": "1.01.01",
			"data": {
				"name": "infrastructure",
				"description": "Infrastructure (root or all data)",
				"properties": {
					"id": {
						"type": "string",
						"required": true
					},
					"regions": {
						"type": "array",
						"items": {
							"type": "string",
							"contentMediaType": "inventory/region"
						}
					}
				}
			}
		},
		"data": {   
			"__id":         "global",
			"__type":       "infrastructure",
			"__ver":        "1.01.01",
			"data": {
				"id": "global", 
				"description": "Global Root of all infrastructure",
				"regions":["North_America", "South_America", "Europe", "Asia", "Middle_East", "Africa"]
			}
		},
		"negativeData": [
			{
				"__id":         "global",
				"__type":       "infrastructure",
				"__ver":        "1.01.01",
				"data": {
					"Id": "global", 
					"description": "Global Root of all infrastructure",
					"Region":["North_America", "South_America", "Europe", "Asia", "Middle_East", "Africa"]
				}
			}
		]
	}`
	testData := map[string]interface{}{}
	json.Unmarshal([]byte(infraStr), &testData)
	schemaRecordData, ok := testData[JsonKey.Schema].(map[string]interface{})
	if !ok {
		t.Fatalf("missing field [%s] from test data", JsonKey.Schema)
	}
	schemaRecord, err := Record.LoadMap(schemaRecordData)
	if err != nil {
		t.Fatalf("failed to load data as record. Error:%s", err)
	}
	schema, err := Schema.LoadSchemaOpsRecord(schemaRecord)
	if err != nil {
		t.Fatalf("failed to load schema record, Error:\n%s", err)
	}
	record, err := Record.LoadMap(testData["data"].(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to load data from infra")
	}
	err = schema.ValidateRecord(record)
	if err != nil {
		t.Fatalf("schema validation failed. Error:\n%s", err)
	}
	for idx, data := range testData["negativeData"].([]interface{}) {
		record, err = Record.LoadMap(data.(map[string]interface{}))
		if err != nil {
			t.Fatalf("failed to load netative data @[%d]", idx)
		}
		err = schema.ValidateRecord(record)
		if err == nil {
			t.Fatalf("failed to alert schema err on negative data [idx]=[%d]", idx)
		}
		log.Printf("spot err at idx=[%d], Err:\n%s", idx, err)
	}
}

func TestSchemaSimpleMap(t *testing.T) {
	log.Print("test simple map definitions")
	simpleMapSchemaStr := `{
		"name": "mapDefintions",
		"description": "map data schema defintions",
		"properties": {
			"simpleMap": {
				"type": "object",
				"additionalProperties": {
					"type": "string"
				}
			}
		}		
	}`
	simpleData1 := `{
		"simpleMap": {
			"test0": "123"
		}
	}`
	simpleData2 := `{
		"simpleMap": {
			"test0": 123
		}
	}`
	schema, err := LoadSchema(simpleMapSchemaStr)
	if err != nil {
		t.Fatalf("failed load simple map schema. Error:%s", err)
	}
	err = validateData(schema, simpleData1)
	if err != nil {
		t.Fatalf("Failed on pass positive simpleData1, Error:%s", err)
	}
	err = validateData(schema, simpleData2)
	if err == nil {
		t.Fatal("Failed on negative simpleData2, Error")
	}
}

func TestSchemaHashObj(t *testing.T) {
	log.Print("test simple map definitions")
	hashObjSchemaStr := `{
		"name": "hashObjectMap",
		"description": "map data schema defintions",
		"properties": {
			"hashObjMap": {
				"type": "object",
				"additionalProperties": {
					"type": "object",
					"$ref": "#/definitions/dataItem"
				}
			}
		},
		"definitions": {
			"dataItem": {
				"name": "dataItem",
				"description": "object data as item for hash map",
				"key": "{field0}",
				"properties": {
					"field0": {
						"type": "string"
					},
					"field1": {
						"type": "string"
					}
				}
			}
		}
	}`
	hashObjectData01 := `{
		"hashObjMap": {
			"testKey0": {
				"field0": "123",
				"field1": "234"
			}
		}
	}`
	hashObjectData02 := `{
		"hashObjMap": {
			"testKey0": {
				"field0": "123"
			}
		}
	}`
	schema, err := LoadSchema(hashObjSchemaStr)
	if err != nil {
		t.Fatalf("failed load simple map schema. Error:%s", err)
	}
	err = validateData(schema, hashObjectData01)
	if err != nil {
		t.Fatalf("Failed on pass positive simpleData1, Error:%s", err)
	}
	err = validateData(schema, hashObjectData02)
	if err == nil {
		t.Fatal("Failed on negative simpleData2, Error")
	}
}

func TestSchemaCustomTypeMap(t *testing.T) {
	hashObjSchemaStr := `{
		"name": "hashObjectMap",
		"description": "map data schema defintions",
		"properties": {
			"hashObjMap": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#/definitions/dataItem"
				}
			}
		},
		"definitions": {
			"dataItem": {
				"name": "dataItem",
				"description": "object data as item for hash map",
				"key": "{field0}",
				"properties": {
					"field0": {
						"type": "string"
					},
					"field1": {
						"type": "string"
					}
				}
			}
		}
	}`
	hashObjectData01 := `{
		"hashObjMap": {
			"testKey0": {
				"field0": "123",
				"field1": "234"
			}
		}
	}`
	hashObjectData02 := `{
		"hashObjMap": {
			"testKey0": {
				"field0": "123"
			}
		}
	}`
	schema, err := LoadSchema(hashObjSchemaStr)
	if err != nil {
		t.Fatalf("failed load simple map schema. Error:%s", err)
	}
	err = validateData(schema, hashObjectData01)
	if err != nil {
		t.Fatalf("Failed on pass positive simpleData1, Error:%s", err)
	}
	err = validateData(schema, hashObjectData02)
	if err == nil {
		t.Fatal("Failed on negative simpleData2, Error")
	}
}

func getSchemaOfSchema() (*Schema.SchemaOps, error) {
	rootDir, err := Util.RootDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get running dir")
	}
	schemaFile, err := filepath.Abs(filepath.Join(rootDir, "lib/Schema/data/schema.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to get ABS path of schema.json")
	}
	schemaData, err := Util.LoadJsonFile(schemaFile)
	if err != nil {
		return nil, err
	}
	schemaList := schemaData.(map[string]interface{})["data"].([]interface{})
	for idx, recObj := range schemaList {
		record, err := Record.LoadMap(recObj.(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to load schema record @[%d]", idx)
		}
		if record.Id == JsonKey.Schema {
			schema, err := Schema.LoadSchemaOpsData(JsonKey.Schema, "0.00.0001", record.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to load schema of schema")
			}
			return schema, nil
		}
	}
	return nil, fmt.Errorf("failed to find schema of schema")
}

func TestSchema(t *testing.T) {
	correctSchema := `
	{
		"name": "testSchema01",
		"properties": {
			"testAttr": {
				"type": "string"
			},
			"testObj": {
				"type": "object",
				"$ref": "#/definitions/obj"
			}
		},
		"definitions": {
			"obj": {
				"name": "obj",
				"properties": {
					"objAttr01": {
						"type": "string",
						"contentMediaType": "inventory/testType"
					}
				}
			}
		}
	}
	`
	badSchma := `
	{
		"name": "testSchema01",
		"properties": {
			"testAttr": {
				"type": "string"
			},
			"testObj": {
				"type": "object",
				"$ref": "#/definitions/obj"
			}
		},
		"definitions": {
			"obj": {
				"name": "obj",
				"properties": {
					"objAttr01": {
						"type": "string",
						"ContentMediaType": "inventory/testType"
					}
				}
			}
		}
	}
	`
	schema, err := getSchemaOfSchema()
	if err != nil {
		t.Fatalf("failed load schema of schema. Error:%s", err)
	}
	err = validateData(schema, correctSchema)
	if err != nil {
		t.Fatalf("Failed on pass positive schema, Error:%s", err)
	}
	err = validateData(schema, badSchma)
	if err == nil {
		t.Fatal("Failed on negative badSchma, Error")
	}
}

func TestValidateRecordId(t *testing.T) {
	schemaStr := `{
		"name": "test",
		"key": "{testAttr}",
		"properties": {
			"testAttr": {
				"type": "string"
			}
		}
	}`
	schema, err := LoadSchema(schemaStr)
	if err != nil {
		t.Fatalf("failed load schemaStr. Error:%s", err)
	}
	badRecordStr := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"testAttr": "123"
		}
	}`
	record, err := Record.LoadStr(badRecordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record")
	}
	err = schema.ValidateRecord(record)
	if err == nil {
		t.Fatalf("failed to catch invalid record id")
	}
}

func TestValidateArrayCheckDup(t *testing.T) {
	schemaStr := `{
		"name": "test",
		"properties": {
			"arrayOfCmt": {
				"type": "array",
				"items": {
					"type": "string",
					"contentMediaType": "inventory/testRef"
				}
				
			},
			"arrayOfStr": {
				"type": "array",
				"items": {
					"type": "string"
				}
			},
			"arrayOfObj": {
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
				"key": "{keyAttr}",
				"properties": {
					"keyAttr": {
						"type": "string"
					}
				}
			}
		}
	}`
	schema, err := LoadSchema(schemaStr)
	if err != nil {
		t.Fatalf("failed to load schemaStr, Error: %s", err)
	}
	recordStr := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"arrayOfStr": [
				"123",
				"234",
				"345",
				"456",
				"123"
			],
			"arrayOfCmt": [
				"123",
				"234",
				"345",
				"456"
			],
			"arrayOfObj": [
				{
					"keyAttr": "1"
				},
				{
					"keyAttr": "2"
				},
				{
					"keyAttr": "3"
				}
			]
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load record. Error:%s", err)
	}
	err = schema.ValidateRecord(record)
	if err != nil {
		t.Fatalf("failed to validate record. Error:%s", err)
	}
	recordDupStr := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"arrayOfStr": [
				"123",
				"234",
				"345",
				"456",
				"123"
			],
			"arrayOfCmt": [
				"123",
				"234",
				"345",
				"456",
				"123"
			],
			"arrayOfObj": [
				{
					"keyAttr": "1"
				},
				{
					"keyAttr": "2"
				},
				{
					"keyAttr": "3"
				}
			]
		}
	}`
	record, err = Record.LoadStr(recordDupStr)
	if err != nil {
		t.Fatalf("failed to load record. Error:%s", err)
	}
	err = schema.ValidateRecord(record)
	if err == nil {
		t.Fatalf("failed to catch duplicate in string array. Error:%s", err)
	}
	recordDupStr = `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"arrayOfStr": [
				"123",
				"234",
				"345",
				"456",
				"123"
			],
			"arrayOfCmt": [
				"123",
				"234",
				"345",
				"456"
			],
			"arrayOfObj": [
				{
					"keyAttr": "1"
				},
				{
					"keyAttr": "2"
				},
				{
					"keyAttr": "3"
				},
				{
					"keyAttr": "3"
				}
			]
		}
	}`
	record, err = Record.LoadStr(recordDupStr)
	if err != nil {
		t.Fatalf("failed to load record. Error:%s", err)
	}
	err = schema.ValidateRecord(record)
	if err == nil {
		t.Fatalf("failed to catch duplicate in string array. Error:%s", err)
	}
}

func TestValidateHashCheckKey(t *testing.T) {
	schemaStr := `{
		"name": "test",
		"properties": {
			"hashStr": {
				"type": "map",
				"items": {
					"type": "string"
				}
			},
			"hashCmt": {
				"type": "map",
				"items": {
					"type": "string",
					"contentMediaType": "inventory/testCmtRef"
				}
			},
			"hashObject": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#/definitions/itemObj"
				}
			}
		},
		"definitions": {
			"itemObj": {
				"name": "itemObj",
				"key": "{keyAttr}",
				"properties": {
					"keyAttr": {
						"type": "string"
					}
				}
			}
		}
	}`
	schema, err := LoadSchema(schemaStr)
	if err != nil {
		t.Fatalf("failed to load schemaStr, Error: %s", err)
	}
	goodRecordstr := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"hashStr": {
				"1": "2",
				"2": "1",
				"3": "2"
			},
			"hashCmt": {
				"1": "1",
				"2": "2",
				"3": "3"
			},
			"hashObject": {
				"1": {
					"keyAttr": "1"
				},
				"2": {
					"keyAttr": "2"
				},
				"3": {
					"keyAttr": "3"
				}
			}
		}
	}`
	record, err := Record.LoadStr(goodRecordstr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record")
	}
	err = schema.ValidateRecord(record)
	if err != nil {
		t.Fatalf("failed to validate good record")
	}
	badRecordstr := `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"hashStr": {
				"1": "2",
				"2": "1",
				"3": "2"
			},
			"hashCmt": {
				"1": "2",
				"2": "2",
				"3": "3"
			},
			"hashObject": {
				"1": {
					"keyAttr": "1"
				},
				"2": {
					"keyAttr": "2"
				},
				"3": {
					"keyAttr": "3"
				}
			}
		}
	}`
	record, err = Record.LoadStr(badRecordstr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record")
	}
	err = schema.ValidateRecord(record)
	if err == nil {
		t.Fatalf("failed to invalid key value hash of cmt")
	}
	badRecordstr = `{
		"__id": "test01",
		"__type": "test",
		"__ver": "0.0.1",
		"data": {
			"hashStr": {
				"1": "2",
				"2": "1",
				"3": "2"
			},
			"hashCmt": {
				"1": "1",
				"2": "2",
				"3": "3"
			},
			"hashObject": {
				"1": {
					"keyAttr": "1"
				},
				"2": {
					"keyAttr": "3"
				},
				"3": {
					"keyAttr": "3"
				}
			}
		}
	}`
	record, err = Record.LoadStr(badRecordstr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record")
	}
	err = schema.ValidateRecord(record)
	if err == nil {
		t.Fatalf("failed to invalid key value hash of cmt")
	}
}

func validateData(schema *Schema.SchemaOps, dataStr string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return fmt.Errorf("failed to parse dataStr. Error:%s", err)
	}
	err = schema.Meta.Validate(data)
	if err != nil {
		return fmt.Errorf("failed to validate data. Error:%s", err)
	}
	return nil
}

func LoadSchema(schemaData string) (*Schema.SchemaOps, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(schemaData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schemaData, Error:%s", err)
	}
	schema, err := Schema.LoadSchemaOpsData(JsonKey.Schema, "0.00.0001", data)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
