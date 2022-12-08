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
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

func LoadIterResult(data interface{}) (*SchemaPath.IteratorResult, error) {
	result := SchemaPath.IteratorResult{}
	err := Json.CopyTo(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func TestIterOneFork(t *testing.T) {
	recordStr := `{
		"schema":{
			"IterEntry": {
				"__id": "IterEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "IterEntry",
					"version": "0.0.1",
					"key": "{iterKey}",
					"properties": {
						"iterKey": {
							"type": "string"
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
								"recursiveKey": {
									"type": "object",
									"$ref": "#"
								}
							}
						}
					}
				}
			},
			"refObj": {
				"__id": "IterEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "refObj",
					"version": "0.0.1",
					"key": "ref{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"recursiveKey": {
							"type": "string",
							"contentMediaType": "inventory/IterEntry"
						}
					}
				}
			}
		},
		"IterEntry": {
			"iter01": {
				"__id": "iter01",
				"__type": "IterEntry",
				"__ver": "0.0.1",
				"data": {
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub01",
								"arrayObj": [],
								"mapObj": {},
								"arrayRef": [],
								"mapRef": {}
							}
						},
						{
							"key1": "01",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub02",
								"arrayObj": [],
								"mapObj": {},
								"arrayRef": [],
								"mapRef": {}
							}
						}
					],
					"mapObj": {
						"01": {
							"key1": "02",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub03",
								"arrayObj": [],
								"mapObj": {},
								"arrayRef": [],
								"mapRef": {}
							}
						},
						"02": {
							"key1": "02",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub04",
								"arrayObj": [],
								"mapObj": {},
								"arrayRef": [],
								"mapRef": {}
							}
						}
					},
					"arrayRef": [
						"ref01_01",
						"ref01_02"
					],
					"mapRef": {
						"01": "ref01_01",
						"02": "ref01_02"
					}
				}
			}
		},
		"refObj": {
			"ref01_01": {
				"__id": "ref01_01",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "01",
					"recursiveKey": "iter01"
				}
			},
			"ref01_02": {
				"__id": "ref01_02",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02",
					"recursiveKey": "iter01"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	path := "IterEntry/iter01/arrayObj[*]?iterator"
	iterResult, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr := LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}
	path = "IterEntry/iter01/mapObj[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}
	path = "IterEntry/iter01/mapObj/*?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/arrayRef[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapRef[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapRef/*?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 2 {
		t.Fatalf("expect result path of [2], got [%d] instead", len(result.QueryResults))
	}
}

func TestIterTwoForks(t *testing.T) {
	recordStr := `{
		"schema":{
			"IterEntry": {
				"__id": "IterEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "IterEntry",
					"version": "0.0.1",
					"key": "{iterKey}",
					"properties": {
						"iterKey": {
							"type": "string"
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
								"recursiveKey": {
									"type": "object",
									"$ref": "#"
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
					"key": "ref{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"recursiveKey": {
							"type": "string",
							"contentMediaType": "inventory/IterEntry"
						}
					}
				}
			}
		},
		"IterEntry": {
			"iter01": {
				"__id": "iter01",
				"__type": "IterEntry",
				"__ver": "0.0.1",
				"data": {
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub01",
								"arrayObj": [
									{
										"key1": "03",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub05",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"01": "ref01_01"
											}
										}
									},
									{
										"key1": "03",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub06",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"02": "ref01_02"
											}
										}
									}
								],
								"mapObj": {
									"01": {
										"key1": "04",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub07",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"03": "ref01_01"
											}
										}
									},
									"02": {
										"key1": "04",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub08",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"04": "ref01_02"
											}
										}
									}
								},
								"arrayRef": [
									"ref01_01",
									"ref01_02"
								],
								"mapRef": {
									"01": "ref01_01",
									"02": "ref01_02"
								}
							}
						},
						{
							"key1": "01",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub02",
								"arrayObj": [
									{
										"key1": "05",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub09",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"05": "ref01_01"
											}
										}
									},
									{
										"key1": "05",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub10",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"06": "ref01_02"
											}
										}
									}
								],
								"mapObj": {},
								"arrayRef": [
									"ref01_01",
									"ref01_02"
								],
								"mapRef": {
									"01": "ref01_01",
									"02": "ref01_02"
								}
							}
						}
					],
					"mapObj": {
						"01": {
							"key1": "02",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub03",
								"arrayObj": [],
								"mapObj": {
									"01": {
										"key1": "06",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub11",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									},
									"02": {
										"key1": "06",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub12",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									}
								},
								"arrayRef": [],
								"mapRef": {}
							}
						},
						"02": {
							"key1": "02",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub04",
								"arrayObj": [],
								"mapObj": {
									"03": {
										"key1": "07",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub11",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									},
									"04": {
										"key1": "07",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub12",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									}
								},
								"arrayRef": [],
								"mapRef": {}
							}
						}
					},
					"arrayRef": [
						"ref01_01",
						"ref01_02"
					],
					"mapRef": {
						"01": "ref01_01",
						"02": "ref01_02"
					}
				}
			}
		},
		"refObj": {
			"ref01_01": {
				"__id": "ref01_01",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "01",
					"recursiveKey": "iter01"
				}
			},
			"ref01_02": {
				"__id": "ref01_02",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02",
					"recursiveKey": "iter01"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	path := "IterEntry/iter01/arrayObj[*]/recursiveKey/arrayObj[*]?iterator"
	iterResult, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr := LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapObj[*]/recursiveKey/mapObj[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapObj/*/recursiveKey/mapObj/*?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/arrayRef[*]/recursiveKey/arrayRef[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapRef[*]/recursiveKey/mapRef[*]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapRef/*/recursiveKey/mapRef/*?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}
}

func TestIterFilter(t *testing.T) {
	recordStr := `{
		"schema":{
			"IterEntry": {
				"__id": "IterEntry",
				"__type": "schema",
				"__ver": "0.0.1",
				"data": {
					"name": "IterEntry",
					"version": "0.0.1",
					"key": "{iterKey}",
					"properties": {
						"iterKey": {
							"type": "string"
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
								"recursiveKey": {
									"type": "object",
									"$ref": "#"
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
					"key": "ref{key1}_{key2}",
					"properties": {
						"key1": {
							"type": "string"
						},
						"key2": {
							"type": "string"
						},
						"recursiveKey": {
							"type": "string",
							"contentMediaType": "inventory/IterEntry"
						}
					}
				}
			}
		},
		"IterEntry": {
			"iter01": {
				"__id": "iter01",
				"__type": "IterEntry",
				"__ver": "0.0.1",
				"data": {
					"arrayObj": [
						{
							"key1": "01",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub01",
								"arrayObj": [
									{
										"key1": "03",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub05",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"01": "ref01_01"
											}
										}
									},
									{
										"key1": "03",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub06",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"02": "ref01_02"
											}
										}
									}
								],
								"mapObj": {
									"01": {
										"key1": "04",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub07",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"03": "ref01_01"
											}
										}
									},
									"02": {
										"key1": "04",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub08",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"04": "ref01_02"
											}
										}
									}
								},
								"arrayRef": [
									"ref01_01",
									"ref01_02"
								],
								"mapRef": {
									"01": "ref01_01",
									"02": "ref01_02"
								}
							}
						},
						{
							"key1": "01",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub02",
								"arrayObj": [
									{
										"key1": "05",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub09",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"01": "ref01_01"
											}
										}
									},
									{
										"key1": "05",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub10",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"02": "ref01_02"
											}
										}
									}
								],
								"mapObj": {
									"01": {
										"key1": "04",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub07",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_01"
											],
											"mapRef": {
												"03": "ref01_01"
											}
										}
									},
									"02": {
										"key1": "04",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub08",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [
												"ref01_02"
											],
											"mapRef": {
												"04": "ref01_02"
											}
										}
									}
								},
								"arrayRef": [
									"ref01_01",
									"ref01_02"
								],
								"mapRef": {
									"01": "ref01_01",
									"02": "ref01_02"
								}
							}
						}
					],
					"mapObj": {
						"01": {
							"key1": "02",
							"key2": "01",
							"recursiveKey": {
								"iterKey": "sub03",
								"arrayObj": [],
								"mapObj": {
									"01": {
										"key1": "06",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub11",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									},
									"02": {
										"key1": "06",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub12",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									}
								},
								"arrayRef": [],
								"mapRef": {}
							}
						},
						"02": {
							"key1": "02",
							"key2": "02",
							"recursiveKey": {
								"iterKey": "sub04",
								"arrayObj": [],
								"mapObj": {
									"03": {
										"key1": "07",
										"key2": "01",
										"recursiveKey": {
											"iterKey": "sub11",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									},
									"04": {
										"key1": "07",
										"key2": "02",
										"recursiveKey": {
											"iterKey": "sub12",
											"arrayObj": [],
											"mapObj": {},
											"arrayRef": [],
											"mapRef": {}
										}
									}
								},
								"arrayRef": [],
								"mapRef": {}
							}
						}
					},
					"arrayRef": [
						"ref01_01",
						"ref01_02"
					],
					"mapRef": {
						"01": "ref01_01",
						"02": "ref01_02"
					}
				}
			}
		},
		"refObj": {
			"ref01_01": {
				"__id": "ref01_01",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "01",
					"recursiveKey": "iter01"
				}
			},
			"ref01_02": {
				"__id": "ref01_02",
				"__type": "Obj",
				"__ver": "0.0.1",
				"data": {
					"key1": "01",
					"key2": "02",
					"recursiveKey": "iter01"
				}
			}
		}
	}`
	conn := PrepareConn(recordStr)
	path := "IterEntry/iter01/arrayObj[*]/recursiveKey/arrayObj[03_01]/?iterator"
	iterResult, err := QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr := LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 1 {
		t.Fatalf("expect result path of [1], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapObj[*]/recursiveKey/mapObj[03]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 1 {
		t.Fatalf("expect result path of [1], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/mapObj/*/recursiveKey/mapObj/03?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}

	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 1 {
		t.Fatalf("expect result path of [1], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/arrayRef[*]/recursiveKey/arrayObj[*]/recursiveKey/mapRef[01]?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}

	path = "IterEntry/iter01/arrayRef[*]/recursiveKey/arrayObj[*]/recursiveKey/mapRef/01?iterator"
	iterResult, err = QueryPath(conn, path)
	if err != nil {
		t.Fatal(err)
	}
	result, cErr = LoadIterResult(iterResult)
	if cErr != nil {
		t.Fatal(cErr)
	}
	if len(result.QueryResults) != 4 {
		t.Fatalf("expect result path of [4], got [%d] instead", len(result.QueryResults))
	}
}
