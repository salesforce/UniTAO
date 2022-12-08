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
	"fmt"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema"
	"github.com/salesforce/UniTAO/lib/Schema/Record"
)

func TestPatchSimpePath(t *testing.T) {
	schemaStr := `{
		"name": "testRoot",
		"version": "0.0.1",
		"properties": {
			"attr1": {
				"type": "string"
			}
		}
	}`
	recordStr := `{
		"__id": "test01",
		"__type": "testRoot",
		"__ver": "0.0.1",
		"data": {
			"attr1": "test"
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record")
	}
	patchData(schemaStr, record, "attr1", "ok")
	if record.Data["attr1"].(string) != "ok" {
		t.Fatalf("failed to set simple attr1 to ok ")
	}
}

func TestPatchArrayStr(t *testing.T) {
	schemaStr := `{
		"name": "testRoot",
		"version": "0.0.1",
		"properties": {
			"attr1": {
				"type": "array",
				"items": {
					"type": "string"
				}
			},
			"attrCmt": {
				"type": "array",
				"items": {
					"type": "string",
					"contentMediaType": "inventory/refType"
				}
			}
		}
	}`
	recordStr := `{
		"__id": "test01",
		"__type": "testRoot",
		"__ver": "0.0.1",
		"data": {
			"attr1": [
				"test",
				"test01"
			],
			"attrCmt": [
				"cmt01",
				"cmt02",
				"cmt03"
			]
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record, Error:%s", err)
	}
	e := patchData(schemaStr, record, "attr1[0]", "ok")
	if e != nil {
		t.Fatal(err)
	}
	if record.Data["attr1"].([]interface{})[0] != "ok" {
		t.Fatal("failed to set data attr1[0] to ok")
	}
	e = patchData(schemaStr, record, "attr1[test]", "ok")
	if e == nil {
		t.Fatal("failed to catch: 'simple array patch should use index failure'")
	}
	e = patchData(schemaStr, record, "attrCmt[cmt01]", "ok")
	if e != nil {
		t.Fatal("failed to update cmt array")
	}
	if record.Data["attrCmt"].([]interface{})[0].(string) != "ok" {
		t.Fatal("failed to update cmt array [0] to ok")
	}
}

func TestPatchArrayObj(t *testing.T) {
	schemaStr := `{
		"name": "testRoot",
		"version": "0.0.1",
		"properties": {
			"attr1": {
				"type": "array",
				"items": {
					"type": "object",
					"$ref": "#/definitions/item"
				}
			}
		},
		"definitions": {
			"item": {
				"name": "item",
				"key": "{type}_{name}",
				"properties": {
					"type": {
						"type": "string"
					},
					"name": {
						"type": "string"
					},
					"value": {
						"type": "string"
					}
				}
			}
		}
	}`
	recordStr := `{
		"__id": "test01",
		"__type": "testRoot",
		"__ver": "0.0.1",
		"data": {
			"attr1": [
				{
					"type": "test",
					"name": "01",
					"value": "test"
				}
			]
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record, Error:%s", err)
	}
	e := patchData(schemaStr, record, "attr1[test_01]/value", "ok")
	if e != nil {
		t.Fatal(err)
	}
	if record.Data["attr1"].([]interface{})[0].(map[string]interface{})["value"].(string) != "ok" {
		t.Fatal("failed to update attr1[test_01] to ok")
	}
	newItem := map[string]interface{}{
		"type":  "ok",
		"name":  "01",
		"value": "test",
	}
	e = patchData(schemaStr, record, "attr1[test_01]", newItem)
	if e != nil {
		t.Fatal(err)
	}
	if record.Data["attr1"].([]interface{})[0].(map[string]interface{})["type"].(string) != "ok" {
		t.Fatal("failed to update attr1[test_01] to ok_01")
	}
}

func TestPatchMapStr(t *testing.T) {
	schemaStr := `{
		"name": "testRoot",
		"version": "0.0.1",
		"properties": {
			"attr1": {
				"type": "map",
				"items": {
					"type": "string"
				}
			},
			"attrCmt": {
				"type": "map",
				"items": {
					"type": "string",
					"contentMediaType": "inventory/itemRef"
				}
			}
		}
	}`
	recordStr := `{
		"__id": "test01",
		"__type": "testRoot",
		"__ver": "0.0.1",
		"data": {
			"attr1": {
				"test01": "test",
				"test02": "test"
			},
			"attrCmt": {
				"test01": "test01",
				"test02": "test02"
			}
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record, Error:%s", err)
	}
	e := patchData(schemaStr, record, "attr1[test01]", "ok")
	if e != nil {
		t.Fatal(err)
	}
	if record.Data["attr1"].(map[string]interface{})["test01"].(string) != "ok" {
		t.Fatal("failed to patch attr1[test01] to ok")
	}
	e = patchData(schemaStr, record, "attrCmt[test01]", "ok")
	if e != nil {
		t.Fatal("failed to update Cmt to ok")
	}
	if _, ok := record.Data["attrCmt"].(map[string]interface{})["ok"]; !ok {
		t.Fatal("failed to change test_01 to ok")
	}
	e = patchData(schemaStr, record, "attrCmt[ok]", "test02")
	if e != nil {
		t.Fatal("failed to update Cmt to ok")
	}
	if _, ok := record.Data["attrCmt"].(map[string]interface{})["ok"]; ok {
		t.Fatal("failed to change ok to test02")
	}
	if len(record.Data["attrCmt"].(map[string]interface{})) != 1 {
		t.Fatal("failed to remove the renamed ok")
	}
}

func TestPatchMapObj(t *testing.T) {
	schemaStr := `{
		"name": "testRoot",
		"version": "0.0.1",
		"properties": {
			"attr1": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#/definitions/item"
				}
			},
			"attr2": {
				"type": "object",
				"additionalProperties": {
					"type": "object",
					"$ref": "#/definitions/item"
				}
			}
		},
		"definitions": {
			"item": {
				"name": "item",
				"key": "{type}_{name}",
				"properties": {
					"type": {
						"type": "string"
					},
					"name": {
						"type": "string"
					},
					"value": {
						"type": "string"
					}
				}
			}
		}
	}`
	recordStr := `{
		"__id": "test01",
		"__type": "testRoot",
		"__ver": "0.0.1",
		"data": {
			"attr1": {
				"test_01":{
					"type": "test",
					"name": "01",
					"value": "test"
				},
				"test_02": {
					"type": "test",
					"name": "02",
					"value": "test"
				}
			},
			"attr2": {
				"test_01":{
					"type": "test",
					"name": "01",
					"value": "test"
				},
				"test_02": {
					"type": "test",
					"name": "02",
					"value": "test"
				}
			}
		}
	}`
	record, err := Record.LoadStr(recordStr)
	if err != nil {
		t.Fatalf("failed to load bad record str as record, Error:%s", err)
	}
	e := patchData(schemaStr, record, "attr1[test_01]/value", "ok")
	if e != nil {
		t.Fatal("failed to patch value attr1[test_01]/value to ok")
	}
	if record.Data["attr1"].(map[string]interface{})["test_01"].(map[string]interface{})["value"].(string) != "ok" {
		t.Fatal("failed to patch value attr1[test_01]/value to ok")
	}
	e = patchData(schemaStr, record, "attr2[test_02]/value", "ok")
	if e != nil {
		t.Fatal("failed to patch value attr2[test_02]/value to ok")
	}
	if record.Data["attr2"].(map[string]interface{})["test_02"].(map[string]interface{})["value"].(string) != "ok" {
		t.Fatal("failed to patch value attr2[test_02]/value to ok")
	}
	e = patchData(schemaStr, record, "attr1[test_01]/type", "ok")
	if e != nil {
		t.Fatal("failed to patch value attr1[test_01]/type to ok")
	}
	if record.Data["attr1"].(map[string]interface{})["ok_01"].(map[string]interface{})["value"].(string) != "ok" {
		t.Fatal("failed to patch value attr1[test_01]/value to ok")
	}
}

func patchData(schemaStr string, record *Record.Record, patchPath string, newData interface{}) error {
	schema, err := LoadSchema(schemaStr)
	if err != nil {
		return fmt.Errorf("failed to load schemaStr, Error: %s", err)
	}
	e := Schema.SetDataOnPath(schema.Schema, record.Data, patchPath, fmt.Sprintf("%s/%s", record.Type, record.Id), newData)
	if e != nil {
		return e
	}
	return nil
}
