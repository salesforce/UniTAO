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

package JournalProcessTest

import (
	"log"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/CmtIndex"
)

// leaf record with category attributes
// with multi-layer of parent records, and registry attribute on a deeper path in the parent record
// multiple leaf register in same parent record on different path

func TestOneLayerWithMultipleIdxPath(t *testing.T) {
	env := prepEnv(t)
	leafSchema := `{
		"__id": "leaf",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "leaf",
			"version": "0.0.1",
			"properties": {
				"layer1": {
					"type": "string",
					"contentMediaType": "inventory/layer1"
				},
				"cat1": {
					"type": "string"
				},
				"cat1a": {
					"type": "string"
				},					
				"cat2": {
					"type": "string"
				},
				"cat2a": {
					"type": "string"
				},					
				"data": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(env, leafSchema)
	layer1Schema := `{
		"__id": "layer1",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "layer1",
			"version": "0.0.1",
			"properties": {
				"name": {
					"type": "string"
				},
				"leafs": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/leaf",
						"indexTemplate": "layer1/{layer1}/leafs"
					}
				},
				"cat1": {
					"type": "map",
					"items": {
						"type": "object",
						"$ref": "#/definitions/cat1"
					}
				},
				"cat2": {
					"type": "array",
					"items": {
						"type": "object",
						"$ref": "#/definitions/cat2"
					}
				}
			},
			"definitions": {
				"cat1": {
					"name": "cat1",
					"key": "{name}",
					"properties": {
						"name": {
							"type": "string"
						},
						"cat1a": {
							"type": "map",
							"items": {
								"type": "object",
								"$ref": "#/definitions/cat1Sub"
							}
						}
					}
				},
				"cat1Sub": {
					"name": "cat1Sub",
					"key": "{name}",
					"properties": {
						"name": {
							"type": "string"
						},
						"leafs": {
							"type": "map",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/leaf",
								"indexTemplate": "layer1/{layer1}/cat1[{cat1}]/cat1a[{cat1a}]/leafs"
							}
						}
					}
				},
				"cat2": {
					"name": "cat2",
					"key": "{name}",
					"properties": {
						"name": {
							"type": "string"
						},
						"cat2a": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/cat2a"
							}
						}
					}
				},
				"cat2a": {
					"name": "cat2a",
					"key": "{name}",
					"properties": {
						"name": {
							"type": "string"
						},
						"leafs": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/leaf",
								"indexTemplate": "layer1/{layer1}/cat2[{cat2}]/cat2a[{cat2a}]/leafs"
							}
						}
					}
				}
			}
		}
	}`
	jC := addSchema(env, layer1Schema)
	if jC != 1 {
		t.Fatalf("invalid journal processed, [%d]!= 1", jC)
	}
	jC = processJournal(env, CmtIndex.KeyCmtIdx, "leaf")
	if jC != 3 {
		t.Fatalf("invalid journal processed, [%d]!= 3", jC)
	}
	layer1Str := `{
		"__id": "layer1-01",
		"__type": "layer1",
		"__ver": "0.0.1",
		"data": {
			"name": "layer1-01",
			"leafs": [],
			"cat1": {
				"cat1-01": {
					"name": "cat1-01",
					"cat1a": {
						"cat1a-01": {
							"name": "cat1a-01",
							"leafs": {}
						},
						"cat1a-02": {
							"name": "cat1a-02",
							"leafs": {}

						}
					}
				}
			},
			"cat2": [
				{
					"name": "cat2-01",
					"cat2a": [
						{
							"name": "cat2a-01",
							"leafs": []
						},
						{
							"name": "cat2a-02",
							"leafs": []
						}
					]
				}
			]

		}
	}`
	jC = addData(env, layer1Str)
	if jC != 1 {
		t.Fatalf("invalid journal processed, [%d]!= 1", jC)
	}
	leafList := []string{
		`{
			"__id": "leaf01",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"cat1": "cat1-01",
				"cat1a": "cat1a-01",
				"cat2": "cat2-01",
				"cat2a": "cat2a-01",
				"data": "leaf01"
			}
		}`,
		`{
			"__id": "leaf02",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"cat1": "cat1-01",
				"cat1a": "cat1a-01",
				"cat2": "cat2-01",
				"cat2a": "cat2a-01",
				"data": "leaf02"
			}
		}`,
		`{
			"__id": "leaf03",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"cat1": "cat1-01",
				"cat1a": "cat1a-02",
				"cat2": "cat2-01",
				"cat2a": "cat2a-02",
				"data": "leaf02"
			}
		}`,
	}
	for _, leafStr := range leafList {
		jC = addData(env, leafStr)
		if jC != 1 {
			t.Fatalf("invalid journal processed, [%d]!= 1", jC)
		}
		jC = processJournal(env, "layer1", "layer1-01")
		if jC != 3 {
			t.Fatalf("invalid journal processed, [%d]!= 3", jC)
		}
	}
}

func TestForAddLayerToExistsIdxTree(t *testing.T) {
	env := prepEnv(t)
	leafSchemaV1 := `{
		"__id": "leaf",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "leaf",
			"version": "0.0.1",
			"properties": {
				"layer1": {
					"type": "string",
					"contentMediaType": "inventory/layer1"
				},				
				"data": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(env, leafSchemaV1)
	Layer1SchemaV1 := `{
		"__id": "layer1",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": { 
			"name": "layer1",
			"version": "0.0.1",
			"properties": {
				"leafs": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/leaf",
						"indexTemplate": "layer1/{layer1}/leafs"
					}
				}
			}
		}
	}`
	addSchema(env, Layer1SchemaV1)
	cmtIdxData, err := env.Handler.GetData(CmtIndex.KeyCmtIdx, "leaf")
	if err != nil {
		t.Fatalf("failed to get [%s] record for [leaf], Error:%s", CmtIndex.KeyCmtIdx, err)
	}
	cmtIdx, ex := CmtIndex.LoadMap(cmtIdxData)
	if ex != nil {
		t.Fatalf("failed to load [%s/leaf], Error:%s", CmtIndex.KeyCmtIdx, ex)
	}
	if _, ok := cmtIdx.Subscriber["layer1"]; !ok {
		t.Fatalf("failed to add subscriber layer1")
	}
	layer1V1 := `{
		"__id": "layer1-01",
		"__type": "layer1",
		"__ver": "0.0.1",
		"data": {
			"leafs": []
		}
	}`
	addData(env, layer1V1)
	leafV1 := []string{
		`{
			"__id": "leaf-01",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"data": "leaf-01"
			}
		}`,
		`{
			"__id": "leaf-02",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"data": "leaf-02"
			}
		}`,
		`{
			"__id": "leaf-03",
			"__type": "leaf",
			"__ver": "0.0.1",
			"data": {
				"layer1": "layer1-01",
				"data": "leaf-03"
			}
		}`,
	}
	for _, leaf := range leafV1 {
		addData(env, leaf)
		jC := processJournal(env, "layer1", "layer1-01")
		if jC != 1 {
			t.Fatalf("invalid journal processed, [%d]!= 1", jC)
		}
	}
	// now we want to add a layer in middle
	leafSchemaV2 := `{
		"__id": "leaf",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": {
			"name": "leaf",
			"version": "0.0.2",
			"properties": {
				"layer1": {
					"type": "string",
					"contentMediaType": "inventory/layer1"
				},
				"layer2PostFix": {
					"type": "string"
				},
				"data": {
					"type": "string"
				}
			}
		}
	}`
	addSchema(env, leafSchemaV2)
	leafV2 := []string{
		`{
			"__id": "leaf-01",
			"__type": "leaf",
			"__ver": "0.0.2",
			"data": {
				"layer1": "layer1-01",
				"layer2PostFix": "01",
				"data": "leaf-01"
			}
		}`,
		`{
			"__id": "leaf-02",
			"__type": "leaf",
			"__ver": "0.0.2",
			"data": {
				"layer1": "layer1-01",
				"layer2PostFix": "01",
				"data": "leaf-02"
			}
		}`,
		`{
			"__id": "leaf-03",
			"__type": "leaf",
			"__ver": "0.0.2",
			"data": {
				"layer1": "layer1-01",
				"layer2PostFix": "02",
				"data": "leaf-03"
			}
		}`,
	}
	for _, leaf := range leafV2 {
		setData(env, leaf)
		jC := processJournal(env, "layer1", "layer1-01")
		if jC != 0 {
			t.Fatalf("invalid journal processed, [%d]!= 0", jC)
		}
	}
	layer2SchemaV1 := `{
		"__id": "layer2",
		"__type": "schema",
		"__ver": "0.0.1",
		"data": { 
			"name": "layer2",
			"version": "0.0.1",
			"properties": {
				"leafs": {
					"type": "array",
					"items": {
						"type": "string",
						"contentMediaType": "inventory/leaf",
						"indexTemplate": "layer2/{layer1}-{layer2PostFix}/leafs"
					}
				}
			}
		}
	}`
	addSchema(env, layer2SchemaV1)
	cmtIdxData, err = env.Handler.GetData(CmtIndex.KeyCmtIdx, "leaf")
	if err != nil {
		t.Fatalf("failed to get [%s] record for [leaf], Error:%s", CmtIndex.KeyCmtIdx, err)
	}
	cmtIdx, ex = CmtIndex.LoadMap(cmtIdxData)
	if ex != nil {
		t.Fatalf("failed to load [%s/leaf], Error:%s", CmtIndex.KeyCmtIdx, ex)
	}
	if _, ok := cmtIdx.Subscriber["layer2"]; !ok {
		t.Fatalf("failed to add subscriber layer2")
	}
	leafList, err := env.Handler.List("leaf")
	if err != nil {
		t.Fatal(err)
	}
	// process journal for all leaves so there should be no change from the journal process
	for _, id := range leafList {
		jC := processJournal(env, "leaf", id.(string))
		if jC != 0 {
			t.Fatalf("invalid journal processed, [%d]!= 0", jC)
		}
	}
	jC := processJournal(env, "layer1", "layer1-01")
	if jC != 0 {
		t.Fatalf("invalid journal processed, [%d]!= 0", jC)
	}
	layer2V1List := []string{
		`{
			"__id": "layer1-01-01",
			"__type": "layer2",
			"__ver": "0.0.1",
			"data": {
				"leafs": []
			}
		}`,
		`{
			"__id": "layer1-01-02",
			"__type": "layer2",
			"__ver": "0.0.1",
			"data": {
				"leafs": []
			}
		}`,
	}
	for _, layer2V1 := range layer2V1List {
		jC = addData(env, layer2V1)
		if jC < 2 {
			t.Fatalf("invalid journal processed, [%d] < 2", jC)
		}
	}
	layer2List, _ := env.Handler.List("layer2")
	for _, id := range layer2List {
		record, _ := env.Handler.Inventory.Get("layer2", id.(string))
		if len(record.Data["leafs"].([]interface{})) == 0 {
			t.Fatalf("failed to add idx for id=[%s]", id.(string))
		}
		for _, leaf := range record.Data["leafs"].([]interface{}) {
			log.Printf("layer2:[%s] - leaf:[%s]", id, leaf)
		}
	}
}
