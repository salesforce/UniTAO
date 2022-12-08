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
	"encoding/json"
	"testing"

	"github.com/salesforce/UniTAO/lib/Schema/Compare"
	"github.com/salesforce/UniTAO/lib/Schema/SchemaDoc"
	"github.com/salesforce/UniTAO/lib/Util/Json"
)

func getSchemaCmp(t *testing.T) *Compare.SchemaCompare {
	schemaStr := `{
		"name": "test",
		"version": "0.0.1",
		"properties": {
			"attrDirect": {
				"type": "string"
			},
			"attrArrayStr": {
				"type": "array",
				"items": {
					"type": "string"
				}
			},
			"attrMapStr": {
				"type": "map",
				"items": {
					"type": "string"
				}
			},
			"attrArrayObj": {
				"type": "array",
				"items": {
					"type": "object",
					"$ref": "#/definitions/itemObj"
				}
			},
			"attrMapObj": {
				"type": "map",
				"items": {
					"type": "object",
					"$ref": "#/definitions/itemObj"
				}
			}
		},
		"definitions": {
			"itemObj" : {
				"name": "itemObj",
				"key": "{attr1}",
				"properties": {
					"attr1": {
						"type": "string"
					},
					"attr2": {
						"type": "string"
					},
					"attrOpt1": {
						"type": "string",
						"required": false
					}
				}
			}
		}
	}`
	schema, err := SchemaDoc.FromString(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	cmp := Compare.SchemaCompare{
		Schema: schema,
	}
	return &cmp
}

func getSource(t *testing.T) map[string]interface{} {
	srcObjStr := `{
		"attrDirect": "testAttrDirect",
		"attrArrayStr": [
			"testArray01",
			"testArray02"
		],
		"attrMapStr": {
			"01": "testMap01",
			"02": "testMap02"
		},
		"attrArrayObj": [
			{
				"attr1": "ary01",
				"attr2": "arytest01",
				"attrOpt1": "aryOpt01"
			},
			{
				"attr1": "ary02",
				"attr2": "arytest02"
			}
		],
		"attrMapObj": {
			"01": {
				"attr1": "map01",
				"attr2": "maptest01",
				"attrOpt1": "mapOpt01"
			},
			"02": {
				"attr1": "map02",
				"attr2": "maptest02"
			}
		}
	}`
	source := map[string]interface{}{}
	err := json.Unmarshal([]byte(srcObjStr), &source)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
func TestRootChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	diffs := cmp.CompareObj(source, nil, "")
	if len(diffs) != 1 {
		t.Fatal("failed to compare obj delete")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	diffs = cmp.CompareObj(nil, source, "")
	if len(diffs) != 1 {
		t.Fatal("failed to compare obj Add")
	}
	if diffs[0].Action != Compare.Add {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Add)
	}
}

func TestDirectAttrChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	tgt, err := Json.Copy(source)
	if err != nil {
		t.Fatal(err)
	}
	tgt.(map[string]interface{})["attrDirect"] = "changed"
	diffs := cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Modify {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Modify)
	}
	if diffs[0].DataPath != "/attrDirect" {
		t.Fatalf("invalid diff path=[%s], [/attrDirect] expected", diffs[0].DataPath)
	}
}

func TestStrArrayChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	tgt, err := Json.Copy(source)
	if err != nil {
		t.Fatal(err)
	}
	tgtAry := tgt.(map[string]interface{})["attrArrayStr"].([]interface{})
	tgt.(map[string]interface{})["attrArrayStr"] = tgtAry[1:]
	diffs := cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrArrayStr[testArray01]" {
		t.Fatalf("invalid diff path=[%s], [/attrArrayStr[testArray01]] expected", diffs[0].DataPath)
	}
	tgt.(map[string]interface{})["attrArrayStr"] = append(tgtAry[1:], "testArray03")
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 2 {
		t.Fatal("failed to catch the change")
	}
}

func TestStrMapChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	tgt, err := Json.Copy(source)
	if err != nil {
		t.Fatal(err)
	}
	tgtMap := tgt.(map[string]interface{})["attrMapStr"].(map[string]interface{})
	tgtMap["03"] = "testMap03"
	diffs := cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Add {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Add)
	}
	if diffs[0].DataPath != "/attrMapStr[03]" {
		t.Fatalf("invalid diff path=[%s], [/attrMapStr[03]] expected", diffs[0].DataPath)
	}
	delete(tgtMap, "03")
	tgtMap["01"] = "testMap03"
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Modify {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Modify)
	}
	if diffs[0].DataPath != "/attrMapStr[01]" {
		t.Fatalf("invalid diff path=[%s], [/attrMapStr[01]] expected", diffs[0].DataPath)
	}
	delete(tgtMap, "01")
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrMapStr[01]" {
		t.Fatalf("invalid diff path=[%s], [/attrMapStr[01]] expected", diffs[0].DataPath)
	}
}

func TestObjArrayChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	tgt, err := Json.Copy(source)
	if err != nil {
		t.Fatal(err)
	}
	tgtArray := tgt.(map[string]interface{})["attrArrayObj"].([]interface{})
	tgtArray[0].(map[string]interface{})["attr2"] = "arytest01-01"
	diffs := cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Modify {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Modify)
	}
	if diffs[0].DataPath != "/attrArrayObj[ary01]/attr2" {
		t.Fatalf("invalid diff path=[%s], [/attrArrayObj[ary01]/attr2] expected", diffs[0].DataPath)
	}
	tgtArray[0].(map[string]interface{})["attr2"] = "arytest01"
	delete(tgtArray[0].(map[string]interface{}), "attrOpt1")
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrArrayObj[ary01]/attrOpt1" {
		t.Fatalf("invalid diff path=[%s], [/attrArrayObj[ary01]/attrOpt1] expected", diffs[0].DataPath)
	}
	tgt.(map[string]interface{})["attrArrayObj"] = tgtArray[1:]
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrArrayObj[ary01]" {
		t.Fatalf("invalid diff path=[%s], [/attrArrayObj[ary01]] expected", diffs[0].DataPath)
	}
	tgtArray[0].(map[string]interface{})["attrOpt1"] = "aryOpt01"
	item3Str := `{
		"attr1": "ary03",
		"attr2": "arytest03",
		"attrOpt1": "aryOpt03"
	}`
	item3 := map[string]interface{}{}
	json.Unmarshal([]byte(item3Str), &item3)
	tgt.(map[string]interface{})["attrArrayObj"] = append(tgtArray, item3)
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Add {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrArrayObj[ary03]" {
		t.Fatalf("invalid diff path=[%s], [/attrArrayObj[ary03]] expected", diffs[0].DataPath)
	}
}

func TestObjMapChange(t *testing.T) {
	cmp := getSchemaCmp(t)
	source := getSource(t)
	tgt, err := Json.Copy(source)
	if err != nil {
		t.Fatal(err)
	}
	tgtMap := tgt.(map[string]interface{})["attrMapObj"].(map[string]interface{})
	tgtMap["01"].(map[string]interface{})["attr2"] = "maptest01-01"
	diffs := cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Modify {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Modify)
	}
	if diffs[0].DataPath != "/attrMapObj[01]/attr2" {
		t.Fatalf("invalid diff path=[%s], [/attrMapObj[01]/attr2] expected", diffs[0].DataPath)
	}
	tgtMap["01"].(map[string]interface{})["attr2"] = "maptest01"
	delete(tgtMap["01"].(map[string]interface{}), "attrOpt1")
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrMapObj[01]/attrOpt1" {
		t.Fatalf("invalid diff path=[%s], [/attrMapObj[01]/attrOpt1] expected", diffs[0].DataPath)
	}
	tgtMap["01"].(map[string]interface{})["attrOpt1"] = "mapOpt01"
	map01 := tgtMap["01"].(map[string]interface{})
	delete(tgtMap, "01")
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Del {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrMapObj[01]" {
		t.Fatalf("invalid diff path=[%s], [/attrMapObj[01]] expected", diffs[0].DataPath)
	}
	tgtMap["01"] = map01
	item3Str := `{
		"attr1": "map03",
		"attr2": "arytest03",
		"attrOpt1": "aryOpt03"
	}`
	item3 := map[string]interface{}{}
	json.Unmarshal([]byte(item3Str), &item3)
	tgtMap["03"] = item3
	diffs = cmp.CompareObj(source, tgt, "")
	if len(diffs) != 1 {
		t.Fatal("failed to catch the change")
	}
	if diffs[0].Action != Compare.Add {
		t.Fatalf("invalid diff action=[%s], [%s] expected", diffs[0].Action, Compare.Del)
	}
	if diffs[0].DataPath != "/attrMapObj[03]" {
		t.Fatalf("invalid diff path=[%s], [/attrMapObj[03]] expected", diffs[0].DataPath)
	}
}
