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

/*
func TestPositiveAllPass(t *testing.T) {
	schemaStr := `
	{
		"TopData": {
			"name": "TopData",
			"description": "top layer of data with array attribute",
			"properties": {
				"name": {
					"type": "string"
				},
				"attrArray": {
					"type": "array",
					"items": {
						"type": "object",
						"$ref": "#/definitions/arrayItem01"
					}
				}
			},
			"definitions": {
				"arrayItem01": {
					"name": "arrayItem01",
					"description": "array item object 01",
					"key": "{name}",
					"properties": {
						"name": {
							"type": "string"
						},
						"value": {
							"type": "string",
							"ContentMediaType": "inventory/FirstRef"
						},
						"optValue": {
							"type": "string",
							"required": false,
							"ContentMediaType": "inventory/SecondRef"
						}
					}
				}
			}
		},
		"FirstRef": {
			"name": "FirstRef",
			"description": "first reference data type",
			"properties": {
				"refKey": {
					"type": "string"
				},
				"nextObj": {
					"type": "object",
					"$ref": "#/definitions/FirstRefObject"
				}
			},
			"definitions": {
				"FirstRefObject": {
					"name": "FirstRefObject",
					"description": "Object in first Reference type",
					"properties": {
						"objName": {
							"type": "string"
						}
					}
				}
			}
		},
		"SecondRef": {
			"name": "SecondRef",
			"description": "second reference data type, optional path",
			"properties": {
				"optName": {
					"type": "string"
				},
				"optObj": {
					"type": "object",
					"$ref": "#/definitions/SecondRefOptObj"
				}
			},
			"definitions": {
				"SecondRefOptObj": {
					"name": "SecondRefOptObj",
					"description": "second reference Optional Object",
					"properties": {
						"objName": {
							"type": "string"
						}
					}
				}
			}
		}
	}
	`
	dataStr := `
	{
		"TopData": {
			"data01": {
				"__id": "data01",
				"__type": "TopData",
				"__ver": "0.0.1",
				"data": {
					"name": "data01",
					"attrArray": [
						{
							"name": "item01",
							"value": "value01"
						},
						{
							"name": "item02",
							"value": "value02"
						},
						{
							"name": "item03",
							"value": "value03",
							"optValue": "optValue01"
						}

					]
				}
			},
			"data02": {
				"__id": "data02",
				"__type": "TopData",
				"__ver": "0.0.1",
				"data": {
					"name": "data01",
					"attrArray": [
						{
							"name": "item01",
							"value": "value01"
						},
						{
							"name": "item02",
							"value": "value02"
						},
						{
							"name": "item03",
							"value": "value03"
						}

					]
				}
			}
		},
		"FirstRef": {
			"value01": {
				"__id": "value01",
				"__type": "FirstRef",
				"__ver": "0.0.1",
				"data": {
					"refKey": "value01Key",
					"nextObj": {
						"objName": "value01Object"
					}
				}
			},
			"value02": {
				"__id": "value02",
				"__type": "FirstRef",
				"__ver": "0.0.1",
				"data": {
					"refKey": "value02Key",
					"nextObj": {
						"objName": "value02Object"
					}
				}
			},
			"value03": {
				"__id": "value03",
				"__type": "FirstRef",
				"__ver": "0.0.1",
				"data": {
					"refKey": "value03Key",
					"nextObj": {
						"objName": "value03Object"
					}
				}
			}
		},
		"SecondRef": {
			"optValue01": {
				"__id": "optValue01",
				"__type": "SecondRef",
				"__ver": "0.0.1",
				"data": {
					"optName": "optValue01Name",
					"optObj": {
						"objName": "optValue01Object"
					}
				}
			}
		}
	}
	`
	conn := PrepareConn(schemaStr, dataStr)
	queryPath := "TopData/data01/attrArray[*]/value"
	value, err := QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Errorf("failed, expect value type [slice]!=[%s] @[path]=[%s]", reflect.TypeOf(value).Kind(), queryPath)
	}
	if len(value.([]interface{})) != 3 {
		t.Errorf("failed, expect array length [3]!=[%d]", len(value.([]interface{})))
	}

	queryPath = "TopData/data01/attrArray[*]/optValue"
	value, err = QueryPath(conn, queryPath)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(value).Kind() != reflect.Slice {
		t.Errorf("failed, expect value type [slice]!=[%s] @[path]=[%s]", reflect.TypeOf(value).Kind(), queryPath)
	}
	if len(value.([]interface{})) != 1 {
		t.Errorf("failed, expect array length [3]!=[%d]", len(value.([]interface{})))
	}
	queryPath = "TopData/data02/attrArray[*]/optValue"
	_, err = QueryPath(conn, queryPath)
	if err == nil {
		t.Errorf("failed to raise error when path does not exists")
	}
}
*/
