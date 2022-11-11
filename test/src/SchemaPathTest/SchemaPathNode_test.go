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

	"github.com/salesforce/UniTAO/lib/SchemaPath"
	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
)

func TestPathNodeOneObj(t *testing.T) {
	recordStr := `{
		"schema": {
			"SchemaEntry": {
				"__id": "SchemaEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "SchemaEntry",
					"properties": {
						"test": {
							"type": "string"
						}
					}					
				}
			}
		},
		"SchemaEntry": {
			"anything": {
				"__id": "anything",
				"__type": "SchemaEntry",
				"__ver": "0.0.1",
				"data": {
					"test": "test"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	dataType := "SchemaEntry"
	dataId := "anything"
	nextPath := ""
	queryNode, err := SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath [%s/%s], nextPath=[%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 0 {
		t.Fatalf("there should be next node as we are not building further")
	}
	nextPath = "test"
	err = queryNode.BuildPath(nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath [%s/%s], nextPath=[%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("invalid path build, [%s/%s] should have only 1 node", dataType, dataId)
	}
	if queryNode.Next[0].AttrName != "test" {
		t.Fatalf("invalid path build, [%s/%s] failed to parse AttrName=[test]", dataType, dataId)
	}
	queryNode.Next = []*Node.PathNode{}
	nextPath = "attrNotExists"
	err = queryNode.BuildPath(nextPath)
	if err == nil {
		t.Fatalf("failed to catch invalid path. %s", nextPath)
	}
}

func TestPathNodeSubItem(t *testing.T) {
	recordStr := `{
		"schema": {
			"SchemaEntry": {
				"__id": "SchemaEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "SchemaEntry",
					"properties": {
						"simpleObj": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						},
						"arrayObj": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemObj"
							}
						},
						"mapObj": {
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
		"SchemaEntry": {
			"anything": {
				"__id": "anything",
				"__type": "SchemaEntry",
				"__ver": "0.0.1",
				"data": {
					"simpleObj": {
						"key1": "test",
						"key2": "01"
					},
					"arrayObj": [
						{
							"key1": "test",
							"key2": "02"
						}
					],
					"mapObj": {
						"03": {
							"key1": "test",
							"key2": "04"
						}
					}
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	dataType := "SchemaEntry"
	dataId := "anything"
	nextPath := "simpleObj"
	queryNode, err := SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to create query node. Error:%s", err)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], didn't get itemObj", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next id=[%s], expect=[itemObj]", dataType, dataId, nextPath, queryNode.Next[0].Schema.Id)
	}
	nextPath = "simpleObj/key1"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].AttrName != "key1" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], didn't catch attrName=[key1]", dataType, dataId, nextPath)
	}
	nextPath = "arrayObj"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], no Next when Idx is empty", dataType, dataId, nextPath)
	}
	nextPath = "arrayObj[test_02]"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next[0].Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
	nextPath = "mapObj"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], no Next when Idx is empty", dataType, dataId, nextPath)
	}
	nextPath = "mapObj[03]"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next[0].Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
	nextPath = "mapObj/03"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
}

func TestPathNodeRefItem(t *testing.T) {
	recordStr := `{
		"schema": {
			"SchemaEntry": {
				"__id": "SchemaEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "SchemaEntry",
					"properties": {
						"simpleRef": {
							"type": "string",
							"contentMediaType": "inventory/refObj"
						},
						"arrayRef": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/refObj"
							}
						},
						"mapRef": {
							"type": "map",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/refObj"
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
		"SchemaEntry": {
			"entry1": {
				"__id": "entry1",
				"__type": "SchemaEntry",
				"__ver": "0.0.1",
				"data": {
					"simpleRef": "test01_01",
					"arrayRef": [
						"test01_01"
					],
					"mapRef": {
						"01": "test01_01"
					}
				}
			}			
		},
		"refObj": {
			"test01_01": {
				"__id": "test01_01",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key1": "test01",
					"key2": "01",
					"key3": "test"
				}
			}
		}

	}`
	conn := PrepareConn(recordStr)
	dataType := "SchemaEntry"
	dataId := "entry1"
	nextPath := "simpleRef"
	queryNode, err := SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing ref node", dataType, dataId, nextPath)
	}
	nextPath = "arrayRef"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect no Next when no Idx", dataType, dataId, nextPath)
	}
	nextPath = "arrayRef[test01_01]"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 || len(queryNode.Next[0].Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Next[0].Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next[0].Schema.Id)
	}
	nextPath = "mapRef"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect no Next when no Idx", dataType, dataId, nextPath)
	}
	nextPath = "mapRef[01]"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 || len(queryNode.Next[0].Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Next[0].Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next[0].Schema.Id)
	}
	nextPath = "mapRef/01"
	queryNode, err = SchemaPath.BuildNodePath(conn, dataType, dataId, nextPath)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if len(queryNode.Next) != 1 || len(queryNode.Next[0].Next) != 1 {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next[0].Next[0].Next[0].Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next[0].Schema.Id)
	}
}
