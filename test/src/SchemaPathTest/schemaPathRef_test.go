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
	"reflect"
	"testing"
)

func TestRefHappyPath(t *testing.T) {
	recordStr := `{
		"schema": {
			"CmtRef": {
				"__id": "CmtRef",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "CmtRef",
					"version": "0.0.1",
					"description": "test for CMT ref query path",
					"properties": {
						"directRef": {
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
						"arrayObj": {
							"type": "array",
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
						},
						"mapObj": {
							"type": "map",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemObj"
							}
						},
						"noRef": {
							"type": "string"
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
			},
			"refObj": {
				"__id": "refObj",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "refObj",
					"version": "0.0.1",
					"description": "reference object",
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
		"CmtRef": {
			"cmtRef1": {
				"__id": "cmtRef1",
				"__type": "CmtRef",
				"__ver": "0.0.1",
				"data": {
					"directRef": "01_01",
					"arrayRef": [
						"01_01",
						"01_02"
					],
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01"
						},
						{
							"key1": "01",
							"key2": "02"
						}
					],
					"mapRef": {
						"01": "01_01",
						"02": "01_02"
					},
					"mapObj": {
						"01": {
							"key1": "01",
							"key2": "01"
						},
						"02": {
							"key1": "01",
							"key2": "02"
						}
					},
					"noRef": "123"
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
	conn := PrepareConn(recordStr)
	path := "CmtRef/cmtRef1/directRef?ref"
	value, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01_01" {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
	path = "CmtRef/cmtRef1/arrayRef[01_01]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01_01" {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
	path = "CmtRef/cmtRef1/arrayObj[01_01]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01_01" {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
	path = "CmtRef/cmtRef1/mapRef[01]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01_01" {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
	path = "CmtRef/cmtRef1/mapObj[01]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if value.(string) != "01_01" {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
}

func TestInvalidPath(t *testing.T) {
	recordStr := `{
		"schema": {
			"CmtRef": {
				"__id": "CmtRef",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "CmtRef",
					"version": "0.0.1",
					"description": "test for CMT ref query path",
					"properties": {
						"directRef": {
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
						"arrayObj": {
							"type": "array",
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
						},
						"mapObj": {
							"type": "map",
							"items": {
								"type": "object",
								"$ref": "#/definitions/itemObj"
							}
						},
						"noRef": {
							"type": "string"
						},
						"noRefArray": {
							"type": "array",
							"items": {
								"type": "string"
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
			},
			"refObj": {
				"__id": "refObj",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "refObj",
					"version": "0.0.1",
					"description": "reference object",
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
		"CmtRef": {
			"cmtRef1": {
				"__id": "cmtRef1",
				"__type": "CmtRef",
				"__ver": "0.0.1",
				"data": {
					"directRef": "01_01",
					"arrayRef": [
						"01_01",
						"01_02"
					],
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01"
						},
						{
							"key1": "01",
							"key2": "02"
						}
					],
					"mapRef": {
						"01": "01_01",
						"02": "01_02"
					},
					"mapObj": {
						"01": {
							"key1": "01",
							"key2": "01"
						},
						"02": {
							"key1": "01",
							"key2": "02"
						}
					},
					"noRef": "123",
					"noRefArray": [
						"123"
					]
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
	conn := PrepareConn(recordStr)
	path := "CmtRef/cmtRef1/noRef?ref"
	value, err := QueryPath(conn, path)
	if err != nil {
		t.Fatalf("failed to get ref @path=[%s]", path)
	}
	if value != "123" {
		t.Fatalf("failed, when ref not available, get value directly")
	}
	path = "CmtRef/cmtRef1/noRefArray[123]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatalf("failed to get ref on path=[%s]", path)
	}
	if value != "123" {
		t.Fatalf("failed, when ref not available, get value directly")
	}
	path = "CmtRef/cmtRef1/directRef/key1?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatalf("failed to get error on path=[%s]", path)
	}
	if value != "01" {
		t.Fatalf("failed, when ref not available, get value directly")
	}
	path = "CmtRef/cmtRef1/arrayRef[01_01]/key1?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatalf("failed to get ref on path=[%s]", path)
	}
	if value != "01" {
		t.Fatalf("failed, when ref not available, get value directly")
	}
}

func TestRefAllPath(t *testing.T) {
	recordStr := `{
		"schema": {
			"CmtRef": {
				"__id": "CmtRef",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "CmtRef",
					"version": "0.0.1",
					"description": "test for CMT ref query path",
					"properties": {
						"directRef": {
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
					"version": "0.0.1",
					"description": "reference object",
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
		"CmtRef": {
			"cmtRef1": {
				"__id": "cmtRef1",
				"__type": "CmtRef",
				"__ver": "0.0.1",
				"data": {
					"directRef": "01_01",
					"arrayRef": [
						"01_01",
						"01_02"
					],
					"mapRef": {
						"01": "01_01",
						"02": "01_02"
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
					
				}
			},
			"01_02": {
				"__id": "01_02",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	path := "CmtRef/cmtRef1/arrayRef[*]?ref"
	value, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("got invalid value type=[%s] from Ref on idx=[*], expected type=[%s]", reflect.TypeOf(value).Kind(), reflect.Slice)
	}
	path = "CmtRef/cmtRef1/mapRef[*]?ref"
	value, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("got invalid value type=[%s] from Ref on idx=[*], expected type=[%s]", reflect.TypeOf(value).Kind(), reflect.Slice)
	}
}
