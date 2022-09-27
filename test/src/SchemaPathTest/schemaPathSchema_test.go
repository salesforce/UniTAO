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
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/JsonKey"
)

func TestWalkArraySchema(t *testing.T) {
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
				},
				"attrArraySimple": {
					"type": "array",
					"items": {
						"type": "string"
					}
				},
				"attrArrayRef": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/refObj"
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
		},
		"refObj": {
			"name": "refObj",
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
					],
					"attrArraySimple": [
						"01_01",
						"01_02"
					],
					"attrArrayRef": [
						"01_01",
						"01_02"
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
		},
		"refObj": {
			"01_01": {
				"__id": "01_01",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "01"
				}
			},
			"01_02": {
				"__id": "01_02",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02"
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWitArray/testArray01?schema"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "schemaWitArray" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArray?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArray[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "itemObj" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray02/attrArray?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArraySimple?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArraySimple[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.String {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArrayRef?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Array {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWitArray/testArray01/attrArrayRef[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "refObj" {
		t.Errorf("got invalid shema data")
	}
}

func TestWalkMapSchema(t *testing.T) {
	schemaStr := `
	{
		"schemaWithMap": {
			"name": "schemaWithMap",
			"description": "schema of object with Map of object in attribute",
			"properties": {
				"attrMap": {
					"type": "map",
					"items": {
						"type": "object",
						"$ref": "#/definitions/itemObj"
					}
				},
				"attrMapSimple": {
					"type": "map",
					"items": {
						"type": "string"
					}
				},
				"attrMapRef": {
					"type": "map",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/refObj"
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
		},
		"refObj": {
			"name": "refObj",
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
`
	recordStr := `
	{
		"schemaWithMap": {
			"testMap01": {
				"__id": "testMap01",
				"__type": "schemaWithMap",
				"__ver": "0.0.1",
				"data": {
					"attrMap": {
						"01_01": {
							"key1": "01",
							"key2": "01"
						},
						"01_02": {
							"key1": "01",
							"key2": "02"
						}
					},
					"attrMapSimple": {
						"01": "01_01",
						"02": "02_02"
					},
					"attrMapRef": {
						"01": "01_01",
						"02": "02_02"
					}
				}
			},
			"testMap02": {
				"__id": "testMap02",
				"__type": "schemaWithMap",
				"__ver": "0.0.1",
				"data": {
					"attrMap": null
				}
			}
		},
		"refObj": {
			"01_01": {
				"__id": "01_01",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "01"
				}
			},
			"01_02": {
				"__id": "01_02",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02"
				}
			}
		}
	}`
	conn := PrepareConn(schemaStr, recordStr)
	queryPath := "schemaWithMap/testMap01?schema"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "schemaWithMap" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMap?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Map {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMap[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "itemObj" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMap/anything?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "itemObj" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapSimple?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Map {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapSimple[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.String {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapSimple/anything?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.String {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapRef?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Type].(string) != JsonKey.Map {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapRef[01_02]?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "refObj" {
		t.Errorf("got invalid shema data")
	}
	queryPath = "schemaWithMap/testMap01/attrMapRef/anything?schema"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(map[string]interface{})[JsonKey.Name].(string) != "refObj" {
		t.Errorf("got invalid shema data")
	}
}
