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
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestWalkInObjectAndMap(t *testing.T) {
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
						"value2": "02"
					},
					"mapStr": {
						"keyExists": "exists"
					}
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	queryPath := "schema1/data1/value/value1"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01" {
		t.Fatalf("invalid value from [path]=[%s], [%s]!=[01]", queryPath, value.(string))
	}
	queryPath = "schema1/data1/mapStr/keyExists"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "exists" {
		t.Fatalf("invalid value from [path]=[%s], [%s]!=[exists]", queryPath, value.(string))
	}
	queryPath = "schema1/data1/mapStr[keyExists]"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "exists" {
		t.Fatalf("invalid value from [path]=[%s], [%s]!=[exists]", queryPath, value.(string))
	}
	queryPath = "schema1/data1/mapStr/keyNotExists"
	_, err = QueryPath(conn, queryPath)
	if err == nil {
		t.Fatalf("key not exists should return error")
	}
	if err.Status != http.StatusNotFound {
		t.Fatalf("query of not exists key return err.Code=[%d], expect err.Code=[%d]", err.Status, http.StatusNotFound)
	}
	queryPath = "schema1/data1/mapStr[keyNotExists]"
	_, err = QueryPath(conn, queryPath)
	if err == nil {
		t.Fatalf("key not exists should return error")
	}
	if err.Status != http.StatusNotFound {
		t.Fatalf("query of not exists key return err.Code=[%d], expect err.Code=[%d]", err.Status, http.StatusNotFound)
	}
}

func TestWalkInArray(t *testing.T) {
	recordStr := `{
		"schema": {
			"schemaWitArray": {
				"__id": "schemaWitArray",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "schemaWitArray",
					"version": "0.0.1",
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
		},
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
	conn := PrepareConn(recordStr)
	queryPath := "schemaWitArray/testArray01/attrArray"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Errorf("failed to get array from path=[%s]", queryPath)
	}
	queryPath = "schemaWitArray/testArray01/attrArray[01_01]"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if value.(map[string]interface{})["key2"] != "01" {
		t.Errorf("failed to get the correct value from [path]=[%s]", queryPath)
	}
	queryPath = "schemaWitArray/testArray01/attrArray[01_02]/key2"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if value == nil {
		t.Errorf("failed to get the value of idx=[01_01], @[path]=[%s]", queryPath)
	}
	if value.(string) != "02" {
		t.Errorf("failed to get the correct value=[%s] from [path]=[%s]", value.(string), queryPath)
	}
}

func TestWalkInAll(t *testing.T) {
	recordStr := `{
		"schema": {
			"SchemaAllPath": {
				"__id": "SchemaAllPath",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "SchemaAllPath",
					"version": "0.0.1",
					"description": "Entry Point for Schema Path All Test",
					"properties": {
						"arrayObj": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemObj"
							}
						},
						"arrayRef": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/refObj"
							}
						},
						"mapObj": {
							"type": "map",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemObj"
							}
						},
						"mapRef": {
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
							"key": "{key1}_{key2}",
							"properties": {
								"key1": {
									"type": "string"
								},
								"key2": {
									"type": "string"
								},
								"key3": {
									"type": "string",
									"required": false
								}
							}
						}
					}
				}
			},
			"refObj": {
				"__id": "refObj",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "refObj",
					"version": "0.0.1",
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"key3": {
							"type": "string",
							"required": false
						}
					}
				}
			}
		},
		"SchemaAllPath": {
			"allPath01": {
				"__id": "allPath01",
				"__type": "SchemaAllPath",
				"__ver": "0.0.1",
				"data": {
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01",
							"key3": "01"
						},
						{
							"key1": "01",
							"key2": "02",
							"key3": "02"
						},
						{
							"key1": "01",
							"key2": "03"
						}
					],
					"arrayRef": [
						"01_01",
						"01_02",
						"01_03"
					],
					"mapObj": {
						"01": {
							"key1": "01",
							"key2": "01",
							"key3": "01"
						},
						"02": {
							"key1": "01",
							"key2": "02",
							"key3": "02"
						},
						"03": {
							"key1": "01",
							"key2": "03"
						}
					},
					"mapRef": {
						"01": "01_01",
						"02": "01_02",
						"03": "01_03"
					}
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
					"key2": "01",
					"key3": "01"
				}
			},
			"01_02": {
				"__id": "01_02",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02",
					"key3": "02"
				}
			},
			"01_03": {
				"__id": "01_03",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "03"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	for _, pathPart := range []string{
		"arrayObj", "arrayRef", "mapObj", "mapRef",
	} {
		path := fmt.Sprintf("SchemaAllPath/allPath01/%s[*]/key2", pathPart)
		value, err := QueryPath(conn, path)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			t.Fatalf("invalid return value type=[%s], expected=[%s], path=[%s]", reflect.TypeOf(value).Kind(), reflect.Slice, path)
		}
		path = fmt.Sprintf("SchemaAllPath/allPath01/%s[*]/key3", pathPart)
		value, err = QueryPath(conn, path)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			t.Fatalf("invalid return value type=[%s], expected=[%s], path=[%s]", reflect.TypeOf(value).Kind(), reflect.Slice, path)
		}
		if len(value.([]interface{})) != 2 {
			t.Fatalf("failed to filter not exists path. get list len=[%d], expected len=[2]", len(value.([]interface{})))
		}

	}
}

func TestWalkRecordCmtRecordDirect(t *testing.T) {
	recordStr := `{
		"schema": {
			"entry": {
				"__id": "entry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "entry",
					"version": "0.0.1",
					"properties": {
						"nextList": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/next"
							}
						}
					}
				}
			},
			"next": {
				"__id": "next",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "next",
					"version": "0.0.1",
					"key": "{itemKey}",
					"properties": {
						"itemKey": {
							"type": "string"
						},
						"value":{
							"type": "string"
						}
					}
				}
			}
		},
		"entry": {
			"01": {
				"__id": "01",
				"__type": "entry",
				"__ver": "0.0.1",
				"data": {
					"nextList": [
						"next-02",
						"aws-dev2-uswest2"
					]
				}
			}
		},
		"next": {
			"aws-dev2-uswest2": {
				"__id": "aws-dev2-uswest2",
				"__type": "next",
				"__ver": "0.0.1",
				"data": {
					"itemKey": "aws-dev2-uswest2",
					"value": "this is next 01"
				}
			},
			"next-02": {
				"__id": "next-02",
				"__type": "next",
				"__ver": "0.0.1",
				"data": {
					"itemKey": "next-02",
					"value": "this is next 02"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	path := "entry/01/nextList"
	value, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid return value type=[%s], expected=[%s], path=[%s]", reflect.TypeOf(value).Kind(), reflect.Slice, path)
	}
	path = "entry/01/nextList[aws-dev2-uswest2]"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid return value type=[%s], expected=[%s], path=[%s]", reflect.TypeOf(value).Kind(), reflect.Map, path)
	}
}
