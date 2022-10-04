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

	"github.com/salesforce/UniTAO/lib/SchemaPath/Node"
)

func TestPathNodeOneObj(t *testing.T) {
	schemaStr := `{
		"SchemaEntry": {
			"name": "SchemaEntry",
			"properties": {
				"test": {
					"type": "string"
				}
			}
		}
	}`
	recordStr := `{}`
	conn := PrepareConn(schemaStr, recordStr)
	dataType := "SchemaEntry"
	dataId := "anything"
	nextPath := ""
	queryNode, err := Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath [%s/%s], nextPath=[%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next != nil {
		t.Fatalf("invalid path build, [%s/%s] should have only 1 node", dataType, dataId)
	}
	nextPath = "test"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath [%s/%s], nextPath=[%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build path. Error:%s", err)
	}
	if queryNode.Next != nil {
		t.Fatalf("invalid path build, [%s/%s] should have only 1 node", dataType, dataId)
	}
	if queryNode.AttrName != "test" {
		t.Fatalf("invalid path build, [%s/%s] failed to parse AttrName=[test]", dataType, dataId)
	}
	nextPath = "attrNotExists"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath [%s/%s], nextPath=[%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build path. Error:%s", err)
	}
	if queryNode.Next != nil {
		t.Fatalf("invalid path build, [%s/%s] should have only 1 node", dataType, dataId)
	}
	if queryNode.AttrName != "attrNotExists" {
		t.Fatalf("invalid path build, [%s/%s] failed to parse AttrName=[test]", dataType, dataId)
	}
	if queryNode.AttrDef != nil {
		t.Fatalf("invalid path build, [%s/%s] AttrName=[attrNotExists], AttrDef expected to be nil", dataType, dataId)
	}
	nextPath = "attrNotExists/further"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate node. Error:%s", err)
	}
	err = queryNode.BuildPath()
	if err == nil {
		t.Fatalf("failed to catch error for walk further on no schema attribute")
	}
}

func TestPathNodeSubItem(t *testing.T) {
	schemaStr := `{
		"SchemaEntry": {
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
	}`
	recordStr := `{}`
	conn := PrepareConn(schemaStr, recordStr)
	dataType := "SchemaEntry"
	dataId := "anything"
	nextPath := "simpleObj"
	queryNode, err := Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to create query node. Error:%s", err)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], didn't get itemObj", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next id=[%s], expect=[itemObj]", dataType, dataId, nextPath, queryNode.Next.Schema.Id)
	}
	nextPath = "simpleObj/key1"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next.AttrName != "key1" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], didn't catch attrName=[key1]", dataType, dataId, nextPath)
	}
	nextPath = "arrayObj"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], no Next when Idx is empty", dataType, dataId, nextPath)
	}
	nextPath = "arrayObj[anything]"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
	nextPath = "mapObj"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], no Next when Idx is empty", dataType, dataId, nextPath)
	}
	nextPath = "mapObj[anything]"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
	nextPath = "mapObj/anything"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing Next when Idx is not empty", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "itemObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], invalid item schema", dataType, dataId, nextPath)
	}
}

func TestPathNodeRefItem(t *testing.T) {
	schemaStr := `{
		"SchemaEntry": {
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
		},
		"refObj": {
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
	}`
	recordStr := `{}`
	conn := PrepareConn(schemaStr, recordStr)
	dataType := "SchemaEntry"
	dataId := "anything"
	nextPath := "simpleRef"
	queryNode, err := Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], missing ref node", dataType, dataId, nextPath)
	}
	nextPath = "arrayRef"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect no Next when no Idx", dataType, dataId, nextPath)
	}
	nextPath = "arrayRef[anything]"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next.Schema.Id)
	}
	nextPath = "mapRef"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect no Next when no Idx", dataType, dataId, nextPath)
	}
	nextPath = "mapRef[anything]"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next.Schema.Id)
	}
	nextPath = "mapRef/anything"
	queryNode, err = Node.New(conn, dataType, dataId, nextPath, nil, nil)
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	err = queryNode.BuildPath()
	if err != nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s]", dataType, dataId, nextPath)
	}
	if queryNode.Next == nil {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], expect Next with Idx", dataType, dataId, nextPath)
	}
	if queryNode.Next.Schema.Id != "refObj" {
		t.Fatalf("failed to build queryPath=[%s/%s/%s], next Schema unexpected Id [%s]!=[refObj]", dataType, dataId, nextPath, queryNode.Next.Schema.Id)
	}
}
