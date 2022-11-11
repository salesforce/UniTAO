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
	"reflect"
	"testing"

	"github.com/salesforce/UniTAO/lib/Util"
)

func validateFlatValue(value interface{}) error {
	valueType := reflect.TypeOf(value).Kind()
	switch valueType {
	case reflect.Slice:
		return validateFlatArray(value.([]interface{}))
	case reflect.Map:
		for key, item := range value.(map[string]interface{}) {
			itemType := reflect.TypeOf(item).Kind()
			switch itemType {
			case reflect.Slice:
				err := validateFlatArray(item.([]interface{}))
				if err != nil {
					return fmt.Errorf("value of [%s] is not a flat array, Error: %s", key, err)
				}
			case reflect.Map:
				return fmt.Errorf("value of [%s] is a map", key)
			}
		}
	}
	return nil
}

func validateFlatArray(ary []interface{}) error {
	for idx, item := range ary {
		err := validateSimple(item)
		if err != nil {
			return fmt.Errorf("item[%d] is not a simple value, Error:%s", idx, err)
		}
	}
	return nil
}

func validateSimple(value interface{}) error {
	valueType := reflect.TypeOf(value).Kind()
	switch valueType {
	case reflect.Slice:
		return fmt.Errorf("value is an array")
	case reflect.Map:
		return fmt.Errorf("value is a map")
	default:
		return nil
	}
}

func TestFlatRecord(t *testing.T) {
	recordStr := `{
		"schema": {
			"FlatTest": {
				"__id": "FlatTest",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "FlatTest",
					"description": "test for retrieve flat",
					"properties": {
						"simpleAttr": {
							"type": "string"
						},
						"directRef": {
							"type": "string",
							"contentMediaType": "inventory/refObj"
						},
						"directObj": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						},
						"arraySimple": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"mapSimple": {
							"type": "map",
							"items": {
								"type": "string"
							}
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
									"type": "striing",
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
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"key3": {
							"type": "striing",
							"required": false
						}
					}
				}
			}
		},
		"FlatTest": {
			"test01": {
				"__id": "test01",
				"__type": "FlatTest",
				"__ver": "0.0.1",
				"data": {
					"simpleAttr": "simple",
					"directRef": "01_01",
					"directObj": {
						"key1": "01",
						"key2": "01",
						"key3": "01"
					},
					"arraySimple": [
						"simple01",
						"simple02"
					],
					"mapSimple": {
						"01": "simple01",
						"02": "simple02"
					},
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
					"arrayRef": [
						"01_01",
						"01_02",
						"01_03"
					],
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
	queryPath := "FlatTest/test01?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid parse record. type=[%s], expect=[%s]", reflect.TypeOf(value).Kind(), reflect.Map)
	}
}

func TestFlatSimple(t *testing.T) {
	recordStr := `{
		"schema": {
			"FlatTest": {
				"__id": "FlatTest",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "FlatTest",
					"description": "test for retrieve flat",
					"properties": {
						"simpleAttr": {
							"type": "string"
						},
						"directRef": {
							"type": "string",
							"contentMediaType": "inventory/refObj"
						},
						"directObj": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						},
						"arraySimple": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"mapSimple": {
							"type": "map",
							"items": {
								"type": "string"
							}
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
									"type": "striing",
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
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"key3": {
							"type": "striing",
							"required": false
						}
					}
				}
			}
		},
		"FlatTest": {
			"test01": {
				"__id": "test01",
				"__type": "FlatTest",
				"__ver": "0.0.1",
				"data": {
					"simpleAttr": "simple",
					"directRef": "01_01",
					"directObj": {
						"key1": "01",
						"key2": "01",
						"key3": "01"
					},
					"arraySimple": [
						"simple01",
						"simple02"
					],
					"mapSimple": {
						"01": "simple01",
						"02": "simple02"
					},
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
					"arrayRef": [
						"01_01",
						"01_02",
						"01_03"
					],
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
	queryPath := "FlatTest/test01/arraySimple?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	ex := validateFlatValue(value)
	if ex != nil {
		t.Fatalf(ex.Error())
	}
	queryPath = "FlatTest/test01/mapSimple?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if val1, ok := value.(map[string]interface{})["01"]; !ok || val1 != "simple01" {
		t.Fatal("failed to retrieve simeple value as flat")
	}
	ex = validateFlatValue(value)
	if ex != nil {
		t.Fatalf(ex.Error())
	}
}

func TestFlatObj(t *testing.T) {
	recordStr := `{
		"schema": {
			"FlatTest": {
				"__id": "FlatTest",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "FlatTest",
					"description": "test for retrieve flat",
					"properties": {
						"simpleAttr": {
							"type": "string"
						},
						"directRef": {
							"type": "string",
							"contentMediaType": "inventory/refObj"
						},
						"directObj": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						},
						"arraySimple": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"mapSimple": {
							"type": "map",
							"items": {
								"type": "string"
							}
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
									"type": "striing",
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
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"key3": {
							"type": "striing",
							"required": false
						}
					}
				}
			}
		},
		"FlatTest": {
			"test01": {
				"__id": "test01",
				"__type": "FlatTest",
				"__ver": "0.0.1",
				"data": {
					"simpleAttr": "simple",
					"directRef": "01_01",
					"directObj": {
						"key1": "01",
						"key2": "01",
						"key3": "01"
					},
					"arraySimple": [
						"simple01",
						"simple02"
					],
					"mapSimple": {
						"01": "simple01",
						"02": "simple02"
					},
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
					"arrayRef": [
						"01_01",
						"01_02",
						"01_03"
					],
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
	queryPath := "FlatTest/test01/directObj?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr := validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	queryPath = "FlatTest/test01/arrayObj?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	queryPath = "FlatTest/test01/mapObj?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
}

func TestFlatRef(t *testing.T) {
	recordStr := `{
		"schema": {
			"FlatTest": {
				"__id": "FlatTest",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "FlatTest",
					"description": "test for retrieve flat",
					"properties": {
						"simpleAttr": {
							"type": "string"
						},
						"directRef": {
							"type": "string",
							"contentMediaType": "inventory/refObj"
						},
						"directObj": {
							"type": "object",
							"$ref": "#/definitions/itemObj"
						},
						"arraySimple": {
							"type": "array",
							"items": {
								"type": "string"
							}
						},
						"mapSimple": {
							"type": "map",
							"items": {
								"type": "string"
							}
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
									"type": "striing",
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
					"key": "{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"key3": {
							"type": "striing",
							"required": false
						}
					}
				}
			}
		},
		"FlatTest": {
			"test01": {
				"__id": "test01",
				"__type": "FlatTest",
				"__ver": "0.0.1",
				"data": {
					"simpleAttr": "simple",
					"directRef": "01_01",
					"directObj": {
						"key1": "01",
						"key2": "01",
						"key3": "01"
					},
					"arraySimple": [
						"simple01",
						"simple02"
					],
					"mapSimple": {
						"01": "simple01",
						"02": "simple02"
					},
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
					"arrayRef": [
						"01_01",
						"01_02",
						"01_03"
					],
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
	queryPath := "FlatTest/test01/directRef?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr := validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	queryPath = "FlatTest/test01/arrayRef?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid flat value from ArrayRef flat")
	}
	queryPath = "FlatTest/test01/arrayRef[01_01]?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid flat value from ArrayRef flat")
	}
	queryPath = "FlatTest/test01/mapRef?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid flat value from ArrayRef flat")
	}
	queryPath = "FlatTest/test01/mapRef[01]?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid flat value from ArrayRef flat")
	}
	queryPath = "FlatTest/test01/mapRef/01?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	if reflect.TypeOf(value).Kind() != reflect.Map {
		t.Fatalf("invalid flat value from ArrayRef flat")
	}
}

func TestFlat2LayerArrayAll(t *testing.T) {
	recordStr := `{
		"schema": {
			"entry": {
				"__id": "entry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "entry",
					"properties": {
						"arrayObj": {
							"type": "array",
							"items": {
								"type": "object",
								"$ref": "#/definitions/firstLayer"
							}
						}
					},
					"definitions": {
						"firstLayer": {
							"name": "firstLayer",
							"key": "layer1-{key}",
							"properties": {
								"key": {
									"type": "string"
								},
								"arrayObj": {
									"type": "array",
									"items": {
										"type": "object",
										"$ref": "#/definitions/secondLayer"
									}							
								}
							}
						},
						"secondLayer": {
							"name": "secondLayer",
							"key": "layer2-{key}",
							"properties": {
								"key": {
									"type": "string"
								},
								"arrayObj": {
									"type": "array",
									"items": {
										"type": "object",
										"$ref": "#/definitions/itemObj"
									}
								}
							}
						},
						"itemObj": {
							"name": "itemObj",
							"key": "item-{key}",
							"properties": {
								"key": {
									"type": "string"
								}
							}
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
					"arrayObj": [
						{
							"key": "01",
							"arrayObj": [
								{
									"key": "01",
									"arrayObj": [
										{
											"key": "01"
										},
										{
											"key": "02"
										},
										{
											"key": "03"
										}
									]
								},
								{
									"key": "02",
									"arrayObj": [
										{
											"key": "03"
										},
										{
											"key": "04"
										},
										{
											"key": "05"
										}
									]
								}
							]
						},
						{
							"key": "02",
							"arrayObj": [
								{
									"key": "03",
									"arrayObj": [
										{
											"key": "06"
										},
										{
											"key": "07"
										},
										{
											"key": "08"
										}
									]
								},
								{
									"key": "04",
									"arrayObj": [
										{
											"key": "09"
										},
										{
											"key": "10"
										},
										{
											"key": "11"
										}
									]
								}
							]
						},
						{
							"key": "03",
							"arrayObj": [
								{
									"key": "05",
									"arrayObj": [
										{
											"key": "12"
										},
										{
											"key": "13"
										},
										{
											"key": "14"
										}
									]
								},
								{
									"key": "06",
									"arrayObj": [
										{
											"key": "15"
										},
										{
											"key": "16"
										},
										{
											"key": "17"
										}
									]
								}
							]
						}
					]
				}
			}			
		}
	}`
	conn := PrepareConn(recordStr)
	queryPath := "entry/01/arrayObj[*]?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	idxMap := Util.IdxList(value.([]interface{}))
	if _, ok := idxMap["layer1-01"]; !ok {
		t.Fatal("failed to get correct key of obj, missing layer1-01")
	}
	pathErr := validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	queryPath = "entry/01/arrayObj[*]/arrayObj[*]?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	pathErr = validateFlatValue(value)
	if pathErr != nil {
		t.Fatalf(pathErr.Error())
	}
	idxMap = Util.IdxList(value.([]interface{}))
	if _, ok := idxMap["layer2-01"]; !ok {
		t.Fatal("failed to get correct key of obj, missing layer2-01")
	}
}

func TestFlatDeDupe(t *testing.T) {
	recordStr := `{
		"schema": {
			"entry": {
				"__id": "entry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "entry",
					"properties": {
						"simpleAry": {
							"type": "array",
							"items": {
								"type": "string"						
							}
						},
						"refAry": {
							"type": "array",
							"items": {
								"type": "string",
								"contentMediaType": "inventory/refObj"
							}
						},
						"objMap": {
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
							"key": "obj-{key}",
							"properties": {
								"key": {
									"type": "string"
								},
								"refVal": {
									"type": "string",
									"contentMediaType": "inventory/refObj"
								},
								"simpleAry": {
									"type": "array",
									"items": {
										"type": "string"
									}
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
					"key": "ref-{key}",
					"properties": {
						"key": {
							"type": "string"
						},
						"value": {
							"type": "string"
						},
						"simpleAry": {
							"type": "array",
							"items": {
								"type": "string"
							}
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
					"simpleAry": [
						"01",
						"02",
						"01",
						"04",
						"03",
						"02"
					],
					"refAry": [
						"ref-01",
						"ref-02",
						"ref-03"
					],
					"objMap": {
						"01": {
							"key": "01",
							"refVal": "ref-01",
							"simpleAry": [
								"01",
								"02"
							]
						},
						"02": {
							"key": "02",
							"refVal": "ref-02",
							"simpleAry": [
								"01",
								"02",
								"03"
							]
						},
						"03": {
							"key": "03",
							"refVal": "ref-02",
							"simpleAry": [
								"01",
								"02",
								"03",
								"04"
							]
						}
					}
				}
			}
		},
		"refObj": {
			"ref-01": {
				"__id": "ref-01",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key": "01",
					"value": "01",
					"simpleAry": [
						"01",
						"02"
					]
				}
			},
			"ref-02": {
				"__id": "ref-02",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key": "02",
					"value": "02",
					"simpleAry": [
						"01",
						"02",
						"03"
					]
				}
			},
			"ref-03": {
				"__id": "ref-03",
				"__type": "refObj",
				"__ver": "0.0.1",
				"data": {
					"key": "03",
					"value": "03",
					"simpleAry": [
						"01",
						"02",
						"03",
						"04"
					]
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	queryPath := "entry/01/simpleAry?flat"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid return type on simpleAry. suppose to be [%s]", reflect.Slice)
	}
	if len(value.([]interface{})) != 4 {
		t.Fatalf("failed to dedupe. return item num [%d]!=[4]", len(value.([]interface{})))
	}
	queryPath = "entry/01/refAry[*]/simpleAry?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid return type on simpleAry. suppose to be [%s]", reflect.Slice)
	}
	if len(value.([]interface{})) != 4 {
		t.Fatalf("failed to dedupe. return item num [%d]!=[4]", len(value.([]interface{})))
	}
	queryPath = "entry/01/objMap[*]/simpleAry?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid return type on simpleAry. suppose to be [%s]", reflect.Slice)
	}
	if len(value.([]interface{})) != 4 {
		t.Fatalf("failed to dedupe. return item num [%d]!=[4]", len(value.([]interface{})))
	}
	queryPath = "entry/01/objMap[*]/refVal?flat"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Fatalf("invalid return type on simpleAry. suppose to be [%s]", reflect.Slice)
	}
	if len(value.([]interface{})) != 2 {
		t.Fatalf("failed to dedupe. return item num [%d]!=[2]", len(value.([]interface{})))
	}
}
