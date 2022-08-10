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
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
	"github.com/salesforce/UniTAO/lib/Util"
)

func TestSchemaValidate(t *testing.T) {
	log.Print("test Run")
	filePath := "data/infrastructure.json"
	testData, err := Util.LoadJSONMap(filePath)
	if err != nil {
		t.Fatalf("failed loading data from [path]=[%s], Err:%s", filePath, err)
	}
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
		t.Fatalf("failed to load data from file=%s", filePath)
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

func TestSchema(t *testing.T) {
	schemaOfSchema := `
	{
		"name": "schema",
		"description": "schema of schema",
		"additionalProperties": false,
		"properties": {
			"name": {
				"type": "string"
			},
			"description": {
				"type": "string",
				"required": false
			},
			"key": {
				"type": "string",
				"required": false
			},
			"properties": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#/definitions/prop"
				}
			},
			"definitions": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#"
				},
				"required": false
			}
		},
		"definitions": {
			"prop": {
				"additionalProperties": false,
				"properties": {
					"type": {
						"type": "string"
					},
					"items": {
						"type": "object",
						"$ref": "#/definitions/prop",
						"required": false
					},
					"$ref": {
						"type": "string",
						"required": false
					},
					"contentMediaType": {
						"type": "string",
						"required": false
					},
					"required": {
						"type": "boolean",
						"required": false
					}
				}
			}
		}
	}
	`
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
	schema, err := LoadSchema(schemaOfSchema)
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

func validateData(schema *Schema.SchemaOps, dataStr string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return fmt.Errorf("failed to parse dataStr. Error:%s", err)
	}
	err = schema.ValidateData(data)
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
